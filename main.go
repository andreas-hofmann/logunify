package main

import (
	"context"
	"fmt"
	"log"
)

var version string = "undefined"

func main() {
	// First off, parse command line arguments
	flags := parseFlags()

	cfg, initCmds, flags := ReadConfig(flags)

	if err := flags.checkError(); err != nil {
		log.Fatal("Invalid parameters: ", err.Error())
	}

	logfile := NewLogFile(flags.LogFileName, flags.writeLogFile())
	defer logfile.Close()

	logremote := NewLogRemote(flags.remoteAddr(), flags.Connect)
	defer logremote.Close()

	if flags.Replay {
		var logversion string = "undefined"

		// Try to read cmd config from all available channels.
		if flags.remoteConnection() {
			logversion = logremote.ReadVersion()
			log.Println("Reading cmd-config from remote.")
			cfg = logremote.ReadCfg(cfg)
		} else {
			logversion = logfile.ReadVersion()
			log.Println("Reading cmd-config from logfile.")
			cfg = logfile.ReadCfg(cfg)
		}

		if version != logversion {
			log.Printf("Warning: Log was written with a different version (%s)!\n", logversion)
		}
	}

	if len(cfg) == 0 {
		log.Fatal("Could not read cmd-config! -replay flag missing?")
	}

	var logsplitter *SplitLog

	if flags.SplitLogFiles {
		logsplitter = NewSplitLog(flags.LogFileName, cfg)
		defer logsplitter.Close()
		if !flags.UI {
			log.Println("Splitting logfile...")
		}
	}

	var ui *TUI

	if flags.UI {
		ui = InitTUI(cfg, flags.MaxLines)
	}

	// Setup the log channel
	logchan := make(chan LogEntry)

	// Set up a context, to allwo terminating background processes
	ctx := context.Background()
	defer ctx.Done()

	if flags.Replay {
		if flags.writeLogFile() {
			logfile.WriteVersion(version)
			logfile.WriteCfg(cfg)
		}

		if flags.remoteConnection() {
			go logremote.Replay(logchan, flags.Realtime)
		} else {
			go logfile.Replay(logchan, flags.Realtime)
		}
	} else {
		// We're not replaying. Write out the version + config, before starting anything else.
		logfile.WriteVersion(version)
		logfile.WriteCfg(cfg)
		logremote.WriteVersion(version)
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

	spincnt := 0
	spinfunc := func() {
		spincnt++
		switch spincnt {
		case 0:
			fmt.Print("\r-")
		case 3:
			fmt.Print("\r\\")
		case 6:
			fmt.Print("\r|")
		case 9:
			fmt.Print("\r/")
			spincnt = 0
		}
	}

	// Receive log data in the background and send it to logfile + views
	recvfunc := func() {
		for l := range logchan {
			if flags.writeLogRemote() {
				logremote.Write(l)
			}
			if flags.writeLogFile() {
				logfile.Write(l)
			}

			if logsplitter != nil {
				logsplitter.Write(l)
			}

			if ui != nil {
				ui.AddData(l)
				if flags.writeLogFile() || flags.Realtime {
					ui.Update()
				}
			} else {
				spinfunc()
			}
		}

		if ui != nil {
			ui.Update()
		} else {
			fmt.Print("\n")
			log.Println("Done.")
		}
	}

	if ui != nil {
		go recvfunc()
		ui.Run()
	} else {
		recvfunc()
	}

}
