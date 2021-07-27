package main

import (
	"context"
)

func main() {
	// First off, parse command line arguments
	flags := ParseFlags()
	cfg := ReadConfig(flags)
	logging := InitLog(flags)
	defer logging.Close()

	// Set up a context, to allwo terminating background processes
	ctx := context.Background()
	defer ctx.Done()

	var ui UI

	ui = InitTUI(flags, cfg)

	if !flags.Replay {
		col := 1
		for _, c := range cfg {
			go RunCmd(ctx, c, col, logging.channel)
			col++
		}
	} else {
		logging.Replay(flags.Realtime)
	}

	// Receive log data in the background and send it to logfile + views
	go func() {
		for l := range logging.channel {
			logging.Write(l)

			ui.AddData(l)

			if !flags.Replay || flags.Realtime {
				ui.Update()
			}
		}

		ui.Update()
	}()

	ui.Run()
}
