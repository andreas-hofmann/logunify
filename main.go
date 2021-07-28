package main

import (
	"context"
)

func main() {
	// First off, parse command line arguments
	flags := ParseFlags()
	cfg := ReadConfig(flags)
	logging := InitLogfile(flags.LogFileName)
	defer logging.Close()

	// Set up a context, to allwo terminating background processes
	ctx := context.Background()
	defer ctx.Done()

	var ui UI

	ui = InitTUI(flags, cfg)

	logchan := make(chan LogEntry)

	if !flags.Replay {
		col := 1
		for _, c := range cfg {
			go RunCmd(ctx, c, col, logchan)
			col++
		}
	} else {
		logging.Replay(logchan, flags.Realtime)
	}

	// Receive log data in the background and send it to logfile + views
	go func() {
		for l := range logchan {
			if !flags.Replay {
				logging.Write(l)
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
