package main

import (
	"github.com/inhies/go-cjdns/admin"
	"fmt"
	"net"
	"time"
)

var UnknownHostStatus *Status

func init() {
	UnknownHostStatus = &Status{State: HostStateUnknown}
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
	return
}

func newHost(host string) (h *Host, err error) {
	name, addr, err := getNameAddr(host)
	if err != nil {
		return nil, err
	}
	now := time.Now().Unix()
	h = &Host{
		StartTime: now,
		EndTime:   now,
		Status:    UnknownHostStatus,
		Address:   newAddress(addr),
		Times:     new(Times),
	}
	if name != "" {
		h.Hostnames = []*Hostname{&Hostname{Name: name, Type: HostnameTypeUser}}
	}
	return
}

func (h *Host) CheckStatus(user *admin.Admin) {
	//start := time.Now()
	ping := new(Ping)
	err := pingNode(user, ping)
	//end := time.Now()
	if err != nil {
		h.Status.State = HostStateDown
		h.Status.Reason = err.Error()
		return
	}

	h.Status.State = HostStateUp
	h.Status.Reason = "CJDNS ping"
	h.Status.ReasonTTL = 64
}
