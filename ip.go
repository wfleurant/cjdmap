package main

import (
	"regexp"
)

const (
	hostRegex = "^([a-zA-Z0-9]([a-zA-Z0-9\\-\\.]{0,}[a-zA-Z0-9]))$"
	ipRegex   = "^fc[a-f0-9]{1,2}:([a-f0-9]{0,4}:){2,6}[a-f0-9]{1,4}$"
)

func validHost(input string) (result bool) {
	result, _ = regexp.MatchString(hostRegex, input)
	return
}

func validIP(input string) (result bool) {
	result, _ = regexp.MatchString(ipRegex, input)
	return
}

func newAddress(arg string) *Address { return &Address{Addr: arg, AddrType: "ipv6"} }
