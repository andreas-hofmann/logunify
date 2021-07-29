package main

import (
	"flag"
)

type Flags struct {
	ConfigFileName string
	LogFileName    string
	Replay         bool
	Realtime       bool
	Remote         string
	Port           int
	Listen         bool
	Connect        bool
	NoUI           bool
	MaxLines       int
}

func ParseFlags() Flags {
	var f Flags

	flag.StringVar(&f.ConfigFileName, "config", "./logunify.yaml", "Config file to use")
	flag.StringVar(&f.LogFileName, "logfile", "logunify.log", "Log file to write to")
	flag.BoolVar(&f.Replay, "replay", false, "Replay a stored log file")
	flag.BoolVar(&f.Realtime, "realtime", false, "Replay a stored log file in real time")
	flag.IntVar(&f.Port, "port", 20000, "Port to use when logging over TCP")
	flag.BoolVar(&f.Listen, "listen", false, "Listen to incoming connections")
	flag.StringVar(&f.Remote, "remote", "", "Remote to connect to")
	flag.IntVar(&f.MaxLines, "maxlines", 500, "Maximum lines in UI buffer. 0 for unlimited scrollback.")
	flag.BoolVar(&f.NoUI, "noui", false, "Disable UI, just log data.")

	flag.Parse()

	f.Connect = false

	if !f.Listen && len(f.Remote) <= 0 {
		f.Port = 0
	}

	if f.MaxLines < 0 {
		f.MaxLines = 0
	}

	if f.Listen {
		f.LogFileName = ""
		f.Replay = true
		f.Realtime = true
		f.Connect = true
	}

	if len(f.Remote) > 0 {
		f.Replay = false
		f.Connect = true
	}

	return f
}
