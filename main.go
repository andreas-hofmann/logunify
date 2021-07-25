package main

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/rivo/tview"
)

func main() {
	// First off, parse command line arguments
	flags := ParseFlags()

	cfg := ReadConfig(flags)
	logging := InitLog(flags)

	// Set up a context, to allwo terminating background processes
	ctx := context.Background()
	defer ctx.Done()

	// Set up grid
	tui := InitTUI(flags, cfg)

	col := 1
	for _, c := range cfg {
		if !flags.Replay {
			go RunCmd(ctx, c, col, logging.channel)
		}
		col++
	}

	logging.Replay()

	app := tview.NewApplication().SetRoot(tui.grid, true).EnableMouse(false)

	// Receive log data in the background and send it to logfile + views
	go func() {
		for l := range logging.channel {
			logging.Write(l)

			lines := strings.Split(strings.TrimRight(l.Text, "\n"), "\n")
			for linenr, line := range lines {
				for c, p := range tui.primitives {
					tv, ok := p.(*tview.TextView)
					if !ok {
						log.Panic("Can't convert TV")
					}
					if (c + 1) == l.Col {
						tv.Write([]byte(line + "\n"))
					} else {
						tv.Write([]byte("\n"))
					}
				}

				if linenr == 0 {
					tui.tview.Write([]byte(l.Ts.Format(time.Stamp) + "\n"))
				} else {
					tui.tview.Write([]byte("\n"))
				}
			}

			if !flags.Replay || flags.Realtime {
				app.Draw()
			}
		}

		app.Draw()
	}()

	// Fire up the tview event loop
	if err := app.Run(); err != nil {
		log.Panic(err)
	}
}
