package main

import (
	"encoding/xml"
	"fmt"
	"log"
	"os"
	"time"
)

const (
	Version = "0.0.1"

	magicalLinkConstant = 5366870.0 //Determined by cjd way back in the dark ages.

	defaultPingTimeout  = 5000 //5 seconds
	defaultPingCount    = 0
	defaultPingInterval = float64(1)

	defaultLogLevel    = "DEBUG"
	defaultLogFile     = ""
	defaultLogFileLine = 0

	defaultPass      = ""
	defaultAdminBind = ""
)

var (
	logger       *log.Logger
	PingTimeout  int
	PingCount    int
	PingInterval float64

	LogLevel    string
	LogFile     string
	LogFileLine int

	File, OutFile string

	AdminPassword string
	AdminBind     string
)

func usage() {
	fmt.Println("usage:", os.Args[0], "HOST [HOST ...]")
	os.Exit(1)
}

func main() {
	if len(os.Args) == 1 {
		usage()
	}
	logger = log.New(os.Stderr, "", 0)

	user, err := adminConnect()
	if err != nil {
		logger.Fatalln("Error:", err)
	}

	var args string
	for _, arg := range os.Args {
		args = args + " " + arg
	}

	startTime := time.Now()
	run := &NmapRun{
		Scanner:          "cjdmap",
		Args:             args,
		Start:            startTime.Unix(),
		Startstr:         startTime.String(),
		Version:          "0.0a",
		XMLOutputVersion: "1.04",
	}

	for _, arg := range os.Args[1:] {
		switch arg {
		case "-h", "--help":
			usage()

		default:
			target, err := newTarget(arg)
			if err != nil {
				logger.Println("Error:", err)
				continue
			}
			trace, err := target.traceRoute(user)
			if err != nil {
				logger.Println("Error:", err)
			}
			run.Hosts = append(run.Hosts, trace)
		}
	}

	stopTime := time.Now()
	run.Finished = &Finished{
		Time:    stopTime.Unix(),
		TimeStr: stopTime.String(),
		//Elapsed: (stopTime.Sub(startTime)*time.Millisecond).String(),
		Exit: "success",
	}

	oX, err := xml.Marshal(run)
	if err != nil {
		logger.Fatalln("Error:", err)
	}

	fmt.Fprint(os.Stdout, xml.Header)
	fmt.Fprintln(os.Stdout, `<?xml-stylesheet href="file:///usr/bin/../share/nmap/nmap.xsl" type="text/xsl"?>`)
	os.Stdout.Write(oX)
}
