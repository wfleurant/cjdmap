package main

import (
	"encoding/binary"
	"encoding/hex"
	"github.com/3M3RY/go-cjdns/admin"
	"sort"
	"strings"
	"time"
)

type Routes []*Route

func (s Routes) Len() int      { return len(s) }
func (s Routes) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

type ByPath struct{ Routes }

func (s ByPath) Less(i, j int) bool { return s.Routes[i].RawPath < s.Routes[j].RawPath }

//Sorts with highest quality link at the top
type ByQuality struct{ Routes }

func (s ByQuality) Less(i, j int) bool { return s.Routes[i].RawLink > s.Routes[j].RawLink }

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

	sort.Sort(ByPath{output})
	return
}

func (t *target) runTrace(user *admin.Admin, hops []*Route) (*Host, error) {
	startTime := time.Now().Unix()
	trace := &Trace{Proto: "CJDNS"}
	for y, p := range hops {
		if y == 0 {
			continue
		}

		logger.Printf("\tpinging %v", p.IP)
		// Ping by path so we don't get RTT for a different route.
		rtt, err := admin.SwitchPinger_ping(user, p.Path, 0)
		if err != nil {
			logger.Println(err)
			return nil, err
		}
		if rtt == 0 {
			rtt = 1
		}

		hop := &Hop{
			TTL:    y,
			RTT:    rtt,
			IPAddr: p.IP,
			//Host:   p.Path,
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
	return h, nil
}

func (t *target) traceRoute(user *admin.Admin) (*Host, error) {
	// Ping to try and force CJDNS to determine a fresh route?
	//_, _, err := admin.RouterModule_pingNode(user, t.addr, 0)
	// should put the version string somewhere in the XML output
	//if err != nil {
	// return nil, err
	//}

	pathS, err := admin.RouterModule_lookup(user, t.addr)
	if err != nil {
		return nil, err
	}
	if len(pathS) > 19 { // Got an address@route
		logger.Print(t.name, " is not in routing table, but is accessible through ", pathS)
		pathS = pathS[40:]
	} else {
		logger.Print(t.name, " path: ", pathS)
	}

	pathS = strings.Replace(pathS, ".", "", -1)
	pathB, err := hex.DecodeString(pathS)
	if err != nil {
		return nil, err
	}
	pathI := binary.BigEndian.Uint64(pathB)

	table := getTable(user)
	hops := getHops(table, pathI)
	return t.runTrace(user, hops)
}

func traceAll(user *admin.Admin) (traces []*Host) {
	table := getTable(user)
	sort.Sort(ByQuality{table})

	traced := make(map[string]bool)

	for _, route := range table {
		if traced[route.IP] || route.Link < 1 {
			continue
		}

		hops := getHops(table, route.RawPath)
		if hops == nil {
			continue
		}

		t, err := newTarget(route.IP)
		if err != nil {
			continue
		}
		logger.Println("Tracing route to", route.IP, "through path", route.Path)

		if trace, err := t.runTrace(user, hops); err == nil {
			traces = append(traces, trace)
			traced[route.IP] = true
		} else {
			logger.Println("Error tracing ", route.Path, ": ", err)
		}
	}
	return
}
