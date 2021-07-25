package main

import (
	"context"
)

func main() {
	// First off, parse command line arguments
	flags := ParseFlags()

	cfg := ReadConfig(flags)
	logging := InitLog(flags)

	// Set up a context, to allwo terminating background processes
	ctx := context.Background()
	defer ctx.Done()

	var tui UI
	tui = InitTUI(flags, cfg)

	col := 1
	for _, c := range cfg {
		if !flags.Replay {
			go RunCmd(ctx, c, col, logging.channel)
		}
		col++
	}

	logging.Replay()

	// Receive log data in the background and send it to logfile + views
	go func() {
		for l := range logging.channel {
			logging.Write(l)

			tui.AddData(l)

			if !flags.Replay || flags.Realtime {
				tui.Update()
			}
		}

		tui.Update()
	}()

	tui.Run()
}
