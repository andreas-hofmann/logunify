package main

import (
	"context"
	"log"
)

func main() {
	// First off, parse command line arguments
	flags := ParseFlags()

	logging := NewLogFile(flags.LogFileName, !flags.Replay)
	defer logging.Close()

	remote := NewLogRemote(flags.Remote, !flags.Replay)
	defer remote.Close()

	// Set up a context, to allwo terminating background processes
	ctx := context.Background()
	defer ctx.Done()

	// Try to read cmd config from all available channels.
	cfg := logging.ReadCfg()
	if len(cfg) == 0 {
		cfg = remote.ReadCfg()
	}
	if len(cfg) == 0 {
		cfg = ReadConfig(flags)
	}
	if len(cfg) == 0 {
		log.Fatal("Could not read cmd-config!")
	}

	ui := InitTUI(flags, cfg)

	logchan := make(chan LogEntry)

	if !flags.Replay {
		// We're not replaying. Write out the config, before starting anything else.
		logging.WriteCfg(cfg)
		remote.WriteCfg(cfg)

		col := 1
		for _, c := range cfg {
			go RunCmd(ctx, c, col, logchan)
			col++
		}
	} else {
		logging.Replay(logchan, flags.Realtime)
		remote.Replay(logchan, flags.Realtime)
	}

	// Receive log data in the background and send it to logfile + views
	go func() {
		for l := range logchan {
			if !flags.Replay {
				logging.Write(l)
				remote.Write(l)
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
