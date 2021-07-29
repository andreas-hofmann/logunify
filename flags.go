package main

import (
	"flag"
	"fmt"
)

type Flags struct {
	ConfigFileName string
	LogFileName    string
	Replay         bool
	Realtime       bool
	Remote         string
	Listen         bool
	NoUI           bool
	MaxLines       int
}

func ParseFlags() Flags {
	var f Flags

	var remote string
	var port int

	flag.StringVar(&f.ConfigFileName, "config", "./logunify.yaml", "Config file to use")
	flag.StringVar(&f.LogFileName, "logfile", "logunify.log", "Log file to write to")
	flag.BoolVar(&f.Replay, "replay", false, "Replay a stored log file")
	flag.BoolVar(&f.Realtime, "realtime", false, "Replay a stored log file in real time")
	flag.BoolVar(&f.Listen, "listen", false, "Listen to incoming connections")
	flag.IntVar(&f.MaxLines, "maxlines", 500, "Maximum lines in UI buffer. 0 for unlimited scrollback.")
	flag.BoolVar(&f.NoUI, "noui", false, "Disable UI, just log data.")

	flag.StringVar(&remote, "remote", "", "Remote to connect to")
	flag.IntVar(&port, "port", 20000, "Port to use when logging over TCP")

	flag.Parse()

	connect := false

	if f.MaxLines < 0 {
		f.MaxLines = 0
	}

	if f.Listen {
		f.LogFileName = ""
		f.Replay = true
		f.Realtime = true
		connect = true
	}

	if len(f.Remote) > 0 {
		f.Replay = false
		connect = true
	}

	if connect {
		f.Remote = f.Remote + ":" + fmt.Sprint(port)
	}

	return f
}
