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

/* TODO
Store pings in a per route/host map for caching and averaging.
*/
func runTrace(user *admin.Admin, t *target, hops []*Route) *Host {
	startTime := time.Now().Unix()
	trace := &Trace{Proto: "lookup"}
	for y, p := range hops {
		if y == 0 {
			continue
		}

		response, err := admin.RouterModule_pingNode(user, p.IP, 1024)
		if err != nil {
			logger.Println(err)
			return nil
		}
		if response.Error == "timeout" {
			return nil
		}
		rtt := float32(response.Time)
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
		Times: &Times{ // Don't know what to do with this element yet.
			SRTT:   1,
			RTTVar: 1,
			To:     1,
		},
	}

	if t.name != "" {
		h.Hostnames = []*Hostname{&Hostname{Name: t.name, Type: HostnameTypeUser}}
	}
	return h
}

func traceAll(user *admin.Admin) []*Host {
	var hopSets [][]*Route
	table := getTable(user)

	for _, route := range table {
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
		t, _ := newTarget(hops[len(hops)-1].IP)
		go func() {
			traceChan <- runTrace(user, t, hops)
		}()
	}
	for i := 0; i < len(hopSets); i++ {
		trace := <-traceChan
		if trace != nil {
			traces = append(traces, trace)
		}
	}
	return traces
}

func (t *target) traceRoute(user *admin.Admin) (*Host, error) {
	_, err := admin.RouterModule_pingNode(user, t.addr, 0)
	if err != nil {
		return nil, err
	}

	response, err := admin.RouterModule_lookup(user, t.addr)
	if err != nil {
		logger.Println("the error was ", err)
		return nil, err
	}
	b, err := hex.DecodeString(strings.Replace(response["result"].(string), ".", "", -1))
	if err != nil {
		return nil, err
	}
	path := binary.BigEndian.Uint64(b)

	table := getTable(user)
	hops := getHops(table, path)
	return runTrace(user, t, hops), nil
}
