package main

import (
	"fmt"
	"net"
	"strings"
	"time"
)

var UnknownHostStatus *Status

func init() {
	UnknownHostStatus = &Status{State: HostStateUnknown}
}

type target struct {
	addr string
	name string
	rtt  time.Duration
	xml  *Host
}

func getNameAddr(host string) (hostname, address string, err error) {
	if validIP(host) {
		address = host
	} else if validHost(host) {
		hostname = host
		ips, err := net.LookupHost(host)
		if err != nil {
			return "", "", err
		}

		for _, ip := range ips {
			if validIP(ip) {
				address = ip
			}
		}
		if address == "" {
			return "", "", fmt.Errorf("Could not resolve %s to CJDNS address", host)
		}

	} else {
		return "", "", fmt.Errorf("could not recognize host \"%s\"", host)
	}

	if len(address) != 39 {
		couplets := strings.Split(address, ":")
		address = fmt.Sprintf("%04s", couplets[0])
		for _, couplet := range couplets[1:] {
			address = address + fmt.Sprintf(":%04s", couplet)
		}
	}
	if hostname == "" {
		hostname = address[35:]
	}
	return
}

func newTarget(host string) (t *target, err error) {
	t = new(target)
	t.name, t.addr, err = getNameAddr(host)
	if err != nil {
		return nil, err
	}
	return
}
