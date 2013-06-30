package main

import (
	"github.com/inhies/go-cjdns/admin"
	"math"
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

/* TODO
Store pings in a per route/host map for caching.
*/

func (t *target) Traceroutes(user *admin.Admin) []*Host {
	var hopSets [][]*Route
	table := getTable(user)

	for _, route := range table {
		if route.IP != t.addr {
			continue
		}
		if route.Link < 1 {
			continue
		}
		hops := getHops(table, route.RawPath)
		if len(hops) < 1 {
			continue
		}

		hopSets = append(hopSets, hops)
	}

	traceChan := make(chan *Host)
	traces := make([]*Host, 0, len(hopSets))
	for _, hops := range hopSets {
		go trace(user, t, hops, traceChan)
	}
	for i := 0; i < len(hopSets); i++ {
		trace := <-traceChan
		if trace == nil {
			continue
		}
		traces = append(traces, trace)
	}
	return traces
}

func trace(user *admin.Admin, t *target, hops []*Route, results chan *Host) {
	startTime := time.Now().Unix()
	trace := &Trace{Proto: "cjdns"}
	var lastHop *Hop
	for y, p := range hops {
		if y == 0 {
			continue
		}
		/*
			tRoute := &Ping{}
			tRoute.Target = p.Path
			err := pingNode(user, tRoute)
			print("\a")
			if err != nil || tRoute.Error == "timeout" {
				results <- nil
				return
			}
		*/

		lastHop = &Hop{
			TTL:    y,
			RTT:    1,
			IPAddr: p.IP,
		}
		trace.Hops = append(trace.Hops, lastHop)
	}

	endTime := time.Now().Unix()
	h := &Host{
		StartTime: startTime,
		EndTime:   endTime,
		Status: &Status{
			State:     HostStateUp,
			Reason:    "CJDNS-Ping",
			ReasonTTL: 56,
		},
		Address: newAddress(t.addr),
		Trace:   trace,
		Times: &Times{ // Don't know what to do with this element yet.
			SRTT:   1,
			RTTVar: 1,
			To:     1,
		},
	}

	if t.name != "" {
		h.Hostnames = []*Hostname{&Hostname{Name: t.name, Type: HostnameTypeUser}}
	}

	results <- h
}
