package main

import (
	"encoding/binary"
	"encoding/hex"
	"github.com/inhies/go-cjdns/admin"
	"math"
	"strings"
	"time"
)

func getHops(table []*Route, fullPath uint64) (output []*Route) {
	for i := range table {
		candPath := table[i].RawPath

		g := 64 - uint64(math.Log2(float64(candPath)))
		h := uint64(uint64(0xffffffffffffffff) >> g)

		if h&fullPath == h&candPath {
			output = append(output, table[i])
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

		pong, err := admin.RouterModule_pingNode(user, p.IP, 1024)
		if err != nil {
			logger.Println(err)
			return nil
		}
		if pong.Error == "timeout" {
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
