package main

import (
	"context"
	"log"
)

var version string = "undefined"

func main() {
	// First off, parse command line arguments
	flags := parseFlags()

	cfg, initCmds, flags := ReadConfig(flags)

	if !flags.valid() {
		log.Fatal("Invalid config.")
	}

	logfile := NewLogFile(flags.LogFileName, flags.writeLogFile())
	defer logfile.Close()

	logremote := NewLogRemote(flags.remoteAddr(), flags.Connect)
	defer logremote.Close()

	if flags.Replay {
		// Try to read cmd config from all available channels.
		if flags.remoteConnection() {
			log.Println("Reading cmd-config from remote.")
			cfg = logremote.ReadCfg(cfg)
		} else {
			log.Println("Reading cmd-config from logfile.")
			cfg = logfile.ReadCfg(cfg)
		}
	}

	if len(cfg) == 0 {
		log.Fatal("Could not read cmd-config! -replay flag missing?")
	}

	var ui TUI

	if flags.UI {
		ui = InitTUI(cfg, flags.MaxLines)
	}

	// Setup the log channel
	logchan := make(chan LogEntry)

	// Set up a context, to allwo terminating background processes
	ctx := context.Background()
	defer ctx.Done()

	if flags.Replay {
		if flags.remoteConnection() {
			logremote.Replay(logchan, flags.Realtime)
		} else {
			logfile.Replay(logchan, flags.Realtime)
		}
	} else {
		// We're not replaying. Write out the config, before starting anything else.
		logfile.WriteCfg(cfg)
		logremote.WriteCfg(cfg)

		for _, c := range initCmds {
			if err := RunInitCmd(c); err != nil {
				log.Fatal("Failed to run init cmd", c, ":", err)
			}
		}

		col := 1
		for _, c := range cfg {
			go RunCmd(ctx, c, col, logchan)
			col++
		}
	}

	// Receive log data in the background and send it to logfile + views
	go func() {
		for l := range logchan {
			if !flags.Replay && flags.remoteConnection() {
				logremote.Write(l)
			}
			if flags.writeLogFile() {
				logfile.Write(l)
			}

			ui.AddData(l)

			if flags.writeLogFile() || flags.Realtime {
				ui.Update()
			}
		}

		ui.Update()
	}()

	ui.Run()
}
