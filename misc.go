/*
 * You may redistribute this program and/or modify it under the terms of
 * the GNU General Public License as published by the Free Software Foundation,
 * either version 3 of the License, or (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */
package main

import (
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/3M3RY/go-cjdns/admin"
	"github.com/3M3RY/go-cjdns/config"
	"io/ioutil"
	"math/rand"
	"net"
	"os/user"
	"regexp"
	"strconv"
	"strings"
)

type CjdnsAdmin struct {
	Address  string `json:"addr"`
	Port     int    `json:"port"`
	Password string `json:"password"`
	Config   string `json:"config,omitempty"`
}

// gotYes will read from stdin and if it is any variation of 'y' or 'yes' then it returns true
// If defaultYes is set to true and the user presses enter without entering anything else it returns true
func gotYes(defaultYes bool) bool {
	var choice string
	n, _ := fmt.Scanln(&choice)
	if n == 0 {
		if defaultYes {
			return true
		} else {
			return false
		}
	}
	if strings.ToLower(choice) == "y" || strings.ToLower(choice) == "yes" {
		return true
	}
	return false
}

// Reads the .cjdnsadmin file and returns the structured contents
func readCjdnsadmin(file string) (admin *CjdnsAdmin, err error) {
	rawFile, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	raw, err := stripComments(rawFile)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(raw, &admin)
	if err != nil {
		return nil, err
	}
	if err != nil {
		// BUG(inhies): Find a better way of dealing with these errors.
		if e, ok := err.(*json.SyntaxError); ok {
			// BUG(inhies): Instead of printing x amount of characters, print the previous and following 2 lines
			fmt.Println("Invalid JSON") //" at byte", e.Offset, "(after stripping comments...)")
			fmt.Println("----------------------------------------")
			fmt.Println(string(raw[e.Offset-60 : e.Offset+60]))
			fmt.Println("----------------------------------------")
		} else if _, ok := err.(*json.InvalidUTF8Error); ok {
			fmt.Println("Invalid UTF-8")
		} else if e, ok := err.(*json.InvalidUnmarshalError); ok {
			fmt.Println("Invalid unmarshall type", e.Type)
			fmt.Println(err)
		} else if e, ok := err.(*json.UnmarshalFieldError); ok {
			fmt.Println("Invalid unmarshall field", e.Field, e.Key, e.Type)
		} else if e, ok := err.(*json.UnmarshalTypeError); ok {
			fmt.Println("Invalid JSON")
			fmt.Println("Expected", e.Type, "but received a", e.Value)
			fmt.Println("I apologize for not being more helpful")
		} else if e, ok := err.(*json.UnsupportedTypeError); ok {
			fmt.Println("Invalid JSON")
			fmt.Println("I am unable to utilize type", e.Type)
		} else if e, ok := err.(*json.UnsupportedValueError); ok {
			fmt.Println("Invalid JSON")
			fmt.Println("I am unable to utilize value", e.Value, e.Str)
		}
		return nil, err
	}
	return

}

// Reads the configuration file specified in global variable File
// and sets the admin credentials
func readConfig() (conf *config.Config, err error) {
	conf, err = config.LoadMinConfig(File)
	if err != nil || len(conf.Admin.Password) == 0 {
		return
	}

	AdminPassword = conf.Admin.Password
	AdminBind = conf.Admin.Bind

	return
}

// Attempt to connect to cjdns
func adminConnect() (user *admin.Admin, err error) {
	// If nothing else has already set this
	if AdminBind == "" || AdminPassword == "" {
		// If we still have no idea which configuration file to use
		if File == "" {
			// Try getting it from the .cjdnsadmin file
			cjdAdmin, err := loadCjdnsadmin()
			if err != nil {
				err = fmt.Errorf("Unable to load configuration file: %s", err.Error())
				return nil, err
			}

			// Set the admin credentials from the .cjdnsadmin file
			AdminPassword = cjdAdmin.Password
			AdminBind = cjdAdmin.Address + ":" + strconv.Itoa(cjdAdmin.Port)

			// If File and OutFile aren't already set, set them
			// Note that they could still be empty if the "config"
			// entry isnt in .cjdnsadmin
			if File == "" {
				File = cjdAdmin.Config
			}

			if OutFile == "" {
				OutFile = cjdAdmin.Config
			}
		} else {
			// File is set so use it
			_, err = readConfig()
			if err != nil {
				err = fmt.Errorf("Unable to load configuration file:", err.Error())
				return nil, err
			}
		}
	}
	user, err = admin.Connect(AdminBind, AdminPassword)
	if err != nil {
		if e, ok := err.(net.Error); ok {
			if e.Timeout() {
				fmt.Println("\nConnection timed out")
			} else if e.Temporary() {
				fmt.Println("\nTemporary error (not sure what that means!)")
			} else {
				fmt.Println("\nUnable to connect to cjdns:", e)
			}
		} else {
			fmt.Println("\nError:", err)
		}
		return
	}
	return
}

// Attempt to read the .cjdnsadmin file from the users home directory
func loadCjdnsadmin() (cjdnsAdmin *CjdnsAdmin, err error) {
	tUser, err := user.Current()
	if err != nil {
		return
	}
	cjdnsAdmin, err = readCjdnsadmin(tUser.HomeDir + "/.cjdnsadmin")
	if err != nil {
		return
	}
	return
}

// Fills out an IPv6 address to the full 32 bytes
// This shouldn't be needed in newer versions of cjdns

func padIPv6(ip net.IP) string {
	raw := hex.EncodeToString(ip)
	parts := make([]string, len(raw)/4)
	for i := range parts {
		parts[i] = raw[i*4 : (i+1)*4]
	}
	return strings.Join(parts, ":")
}

type Route struct {
	IP      string
	Path    string
	RawPath uint64
	Link    float64
	RawLink int64
	Version int64
}

// Dumps the entire routing table and structures it
func getTable(user *admin.Admin) (table []*Route) {
	page := 0
	var more int64
	table = make([]*Route, 0)
	for more = 1; more != 0; page++ {
		response, err := admin.NodeStore_dumpTable(user, page)
		if err != nil {
			fmt.Printf("%v\n", err)
			return
		}
		// If an error field exists, and we have an error, return it
		if _, ok := response["error"]; ok {
			if response["error"] != "none" {
				err = fmt.Errorf(response["error"].(string))
				fmt.Printf("Error: %v\n", err)
				return
			}
		}
		//Thanks again to SashaCrofter for the table parsing
		rawTable := response["routingTable"].([]interface{})
		for i := range rawTable {
			item := rawTable[i].(map[string]interface{})
			rPath := item["path"].(string)
			sPath := strings.Replace(rPath, ".", "", -1)
			bPath, err := hex.DecodeString(sPath)
			if err != nil || len(bPath) != 8 {
				//If we get an error, or the
				//path is not 64 bits, discard.
				//This should also prevent
				//runtime errors.
				continue
			}
			path := binary.BigEndian.Uint64(bPath)
			table = append(table, &Route{
				IP:      item["ip"].(string),
				RawPath: path,
				Path:    rPath,
				RawLink: item["link"].(int64),
				Link:    float64(item["link"].(int64)) / magicalLinkConstant,
				Version: item["version"].(int64),
			})

		}

		if response["more"] != nil {
			more = response["more"].(int64)
		} else {
			break
		}
	}

	return
}

// Returns a random alphanumeric string where length is <= max >= min
func randString(min, max int) string {
	r := myRand(min, max, "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789")
	return r
}

// Returns a random character from the specified string where length is <= max >= min
func myRand(min, max int, char string) string {

	var length int

	if min < max {
		length = min + rand.Intn(max-min)
	} else {
		length = min
	}

	buf := make([]byte, length)
	for i := 0; i < length; i++ {
		buf[i] = char[rand.Intn(len(char)-1)]
	}
	return string(buf)
}

func stripComments(b []byte) ([]byte, error) {
	regComment, err := regexp.Compile("(?s)//.*?\n|/\\*.*?\\*/")
	if err != nil {
		return nil, err
	}
	out := regComment.ReplaceAllLiteral(b, nil)
	return out, nil
}
