package main

import (
	"github.com/inhies/go-cjdns/admin"
	"math"
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

func (h *Host) Traceroute(user *admin.Admin) {
	var hopSets [][]*Route
	table := getTable(user)

	for _, route := range table {
		if route.IP != h.Address.Addr {
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

	traceChan := make(chan *Trace)
	for _, hops := range hopSets {
		go trace(user, hops, traceChan)
	}
	for i := 0; i < len(hopSets); i++ {
		trace := <-traceChan
		if trace == nil {
			continue
		}
		h.mu.Lock()
		h.Traces = append(h.Traces, trace)
		h.mu.Unlock()
	}
	return
}

func trace(user *admin.Admin, hops []*Route, results chan *Trace) {
	t := &Trace{Proto: "cjdns"}
	for y, p := range hops[1:] {
		tRoute := &Ping{}
		tRoute.Target = p.Path
		err := pingNode(user, tRoute)
		if err != nil || tRoute.Error == "timeout" {
			results <- nil
			return
		}

		h := &Hop{
			TTL:    y,
			RTT:    tRoute.TTime,
			IPAddr: p.IP,
		}
		t.Hops = append(t.Hops, h)
	}
	results <- t
}
