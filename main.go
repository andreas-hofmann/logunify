package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
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

var version string = "undefined"

func parseFlags() Flags {
	var f Flags

	var remote string
	var port int
	var printversion bool

	flag.StringVar(&f.ConfigFileName, "config", "./logunify.yaml", `Config file to use. Only required when recording,
not for replaying.`)
	flag.StringVar(&f.LogFileName, "logfile", "", "Log file to write to.")
	flag.BoolVar(&f.Replay, "replay", false, "Replay a stored log file.")
	flag.BoolVar(&f.Realtime, "realtime", false, "Replay in real time (including pauses).")
	flag.BoolVar(&f.Listen, "listen", false, "Listen for incoming connections for replay data.")
	flag.IntVar(&f.MaxLines, "maxlines", 500, `Maximum lines in UI buffer. 0 for unlimited scrollback."
When replaying, scrollback is always set to unlimited.`)
	flag.BoolVar(&f.NoUI, "noui", false, "Disable UI, just log data.")

	flag.StringVar(&remote, "remote", "", "Remote to connect to.")
	flag.IntVar(&port, "port", 20000, "Port to use when logging over TCP.")
	flag.BoolVar(&printversion, "version", false, "Print the program version.")

	flag.Parse()

	if printversion {
		fmt.Println("logunify version: ", version)
		os.Exit(0)
	}

	connect := false

	if f.MaxLines < 0 {
		f.MaxLines = 0
	}

	if f.Listen {
		f.Replay = true
		f.Realtime = true
		connect = true
	}

	if len(remote) > 0 {
		f.Replay = false
		connect = true
	}

	if connect {
		f.Remote = remote + ":" + fmt.Sprint(port)
	}

	if f.Replay {
		f.MaxLines = 0
	}

	return f
}

func main() {
	// First off, parse command line arguments
	flags := parseFlags()

	logfile := NewLogFile(flags.LogFileName, !flags.Replay)
	defer logfile.Close()

	logremote := NewLogRemote(flags.Remote, !flags.Replay)
	defer logremote.Close()

	// Try to read cmd config from all available channels.
	cfg := logfile.ReadCfg()
	if len(cfg) == 0 {
		cfg = logremote.ReadCfg()
	}
	if len(cfg) == 0 {
		cfg = ReadConfig(flags)
	}
	if len(cfg) == 0 {
		log.Fatal("Could not read cmd-config!")
	}

	// Init the UI
	ui := InitTUI(flags, cfg)

	// Setup the log channel
	logchan := make(chan LogEntry)

	// Set up a context, to allwo terminating background processes
	ctx := context.Background()
	defer ctx.Done()

	if !flags.Replay {
		// We're not replaying. Write out the config, before starting anything else.
		logfile.WriteCfg(cfg)
		logremote.WriteCfg(cfg)

		col := 1
		for _, c := range cfg {
			go RunCmd(ctx, c, col, logchan)
			col++
		}
	} else {
		logfile.Replay(logchan, flags.Realtime)
		logremote.Replay(logchan, flags.Realtime)
	}

	// Receive log data in the background and send it to logfile + views
	go func() {
		for l := range logchan {
			if !flags.Replay {
				logfile.Write(l)
				logremote.Write(l)
			}

			ui.AddData(l)

			if !flags.Replay || flags.Realtime {
				ui.Update()
			}
		}

		ui.Update()
	}()

	ui.Run()
}
