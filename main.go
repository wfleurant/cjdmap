package main

import (
	"fmt"
	"os"
	//"sync"
	"encoding/xml"
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

func main() {
	user, err := adminConnect()
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
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

	targets := make([]*target, 0, len(os.Args[1:]))

	for _, arg := range os.Args[1:] {
		target, err := newTarget(arg)
		if err != nil {
			fmt.Println("Error:", err)
			continue
		}
		targets = append(targets, target)
	}

	for _, target := range targets {
		traces := target.Traceroutes(user)
		run.Hosts = append(run.Hosts, traces...)
	}

	hostsTotal := cap(run.Hosts)
	hostsUp := len(run.Hosts)
	hostsDown := hostsTotal - hostsUp

	stopTime := time.Now()
	run.Finished = &Finished{
		Time:    stopTime.Unix(),
		TimeStr: stopTime.String(),
		//Elapsed: (stopTime.Sub(startTime)*time.Millisecond).String(),
		Exit:    "success",
	}
	run.HostStats = &Hosts{
		Up:    hostsUp,
		Down:  hostsDown,
		Total: hostsTotal,
	}

	oX, err := xml.Marshal(run)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	fmt.Fprint(os.Stdout, xml.Header)
	fmt.Fprintln(os.Stdout, `<?xml-stylesheet href="file:///usr/bin/../share/nmap/nmap.xsl" type="text/xsl"?>`)
	os.Stdout.Write(oX)

}

/*
	startTime := time.Now()
	run := &Nmaprun{
		Scanner:          "cjdnsmap",
		Args:             fmt.Sprint(os.Args),
		Start:            startTime.Unix(),
		Startstr:         startTime.String(),
		Version:          "0.1",
		Xmloutputversion: "1.04",
	}



	stopTime := time.Now()

	run.Runstats = &Runstats{
		&Finished{
			Time:    stopTime.Unix(),
			Timestr: stopTime.String(),
			Elapsed: int64(stopTime.Sub(startTime) * time.Second),
			Exit:    "success",
		},
		&Hosts{
			Up:    1,
			Down:  0,
			Total: 1,
		},
	}

	output, err := xml.Marshal(run)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

*/
