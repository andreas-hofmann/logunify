package main

import (
	"flag"
)

type Flags struct {
	ConfigFileName string
	LogFileName    string
	Replay         bool
	Realtime       bool
}

func ParseFlags() Flags {
	var f Flags

	flag.StringVar(&f.ConfigFileName, "config", "./logunify.yaml", "Config file to use")
	flag.StringVar(&f.LogFileName, "logfile", "logunify.log", "Log file to write to")
	flag.BoolVar(&f.Replay, "replay", false, "Replay a stored log file")
	flag.BoolVar(&f.Realtime, "realtime", false, "Replay a stored log file in real time")

	flag.Parse()

	return f
}
