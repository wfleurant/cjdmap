package main

import (
	"encoding/binary"
	"encoding/hex"
	"github.com/inhies/go-cjdns/admin"
	"strings"
	"time"
)

func log2x64(number uint64) uint {
	var out uint
	for number != 0 {
		number = number >> 1
		out++
	}
	return out
}

func isBehind(destination uint64, midPath uint64) bool {
	if midPath > destination {
		return false
	}
	mask := ^uint64(0) >> (64 - log2x64(midPath))
	return (destination & mask) == (midPath & mask)
}

// WARNING: this depends on implementation quirks of the router and will be broken in the future.
func isOneHop(destination uint64, midPath uint64) bool {
	c := destination >> log2x64(midPath)
	if c&1 != 0 {
		return log2x64(c) == 4
	}
	if c&3 != 0 {
		return log2x64(c) == 7
	}
	return log2x64(c) == 10
}

func getHops(table []*Route, fullPath uint64) (output []*Route) {
	for _, route := range table {
		if isBehind(fullPath, route.RawPath) {
			output = append(output, route)
		}
	}
	return
}

func runTrace(user *admin.Admin, t *target, hops []*Route) *Host {
	startTime := time.Now().Unix()
	trace := &Trace{Proto: "CJDNS"}
	for y, p := range hops {
		if y == 0 {
			continue
		}

		logger.Printf("Pinging %v\t", p.IP)
		pong, err := admin.RouterModule_pingNode(user, p.IP, 1024)
		if err != nil {
			logger.Println(err)
			return nil
		}
		if pong.Error == "timeout" {
			logger.Println("timeout")
			return nil
		}
		rtt := float32(pong.Time)
		if rtt == 0 {
			rtt = 1
		}

		hop := &Hop{
			TTL:    y,
			RTT:    rtt,
			IPAddr: p.IP,
		}
		trace.Hops = append(trace.Hops, hop)
	}

	endTime := time.Now().Unix()
	h := &Host{
		StartTime: startTime,
		EndTime:   endTime,
		Status: &Status{
			State:     HostStateUp,
			Reason:    "pingNode",
			ReasonTTL: 56,
		},
		Address: newAddress(t.addr),
		Trace:   trace,
		//Times: &Times{ // Don't know what to do with this element yet.
		//	SRTT:   1,
		//	RTTVar: 1,
		//	To:     1,
		//},
	}

	if t.name != "" {
		h.Hostnames = []*Hostname{&Hostname{Name: t.name, Type: HostnameTypeUser}}
	}
	return h
}

func (t *target) traceRoute(user *admin.Admin) (*Host, error) {
	// Ping to force a lookup if there isn't a route already.
	_, err := admin.RouterModule_pingNode(user, t.addr, 0)
	if err != nil {
		return nil, err
	}

	response, err := admin.RouterModule_lookup(user, t.addr)
	if err != nil {
		return nil, err
	}
	s := response["result"].(string)
	if len(s) > 19 { // Got an address@route
		s = s[40:]
	}
	s = strings.Replace(s, ".", "", -1)
	b, err := hex.DecodeString(s)
	if err != nil {
		return nil, err
	}
	path := binary.BigEndian.Uint64(b)

	table := getTable(user)
	hops := getHops(table, path)
	return runTrace(user, t, hops), nil
}

func (t *target) traceRoutes(user *admin.Admin) (traces []*Host, err error) {
	table := getTable(user)
	logger.Println("Finding all routes to", t.addr)

	for i := range table {
		if table[i].IP != t.addr {
			continue
		}
		if table[i].Link < 1 {
			continue
		}

		hops := getHops(table, table[i].RawPath)
		if hops == nil {
			continue
		}

		trace := runTrace(user, t, hops)
		if err != nil {
			logger.Println(err)
			continue
		}

		traces = append(traces, trace)
	}
	return
}
