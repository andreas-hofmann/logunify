package main

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func newTextView(text string) tview.Primitive {
	return tview.NewTextView().
		SetTextAlign(tview.AlignLeft).
		SetText(text).
		SetWrap(false)
}

func main() {
	// First off, parse command line arguments
	flags := ParseFlags()

	cfg := ReadConfig(flags)
	logging := InitLog(flags)

	// Set up a context, to allwo terminating background processes
	ctx := context.Background()
	defer ctx.Done()

	// Set up grid
	grid := tview.NewGrid().
		SetBorders(true).
		SetRows(1).
		SetColumns(16)

	// Set up cmd textviews
	var primitives []tview.Primitive

	col := 1
	for _, c := range cfg {
		// Add header
		grid.AddItem(newTextView(c.Cmd), 0, col, 1, 1, 0, 0, false)

		// Add cmd output
		p := newTextView("")
		primitives = append(primitives, p)
		grid.AddItem(p, 1, col, 1, 1, 0, 0, false)

		if !flags.Replay {
			go RunCmd(ctx, c, col, logging.channel)
		}

		col++
	}

	logging.Replay()

	// Add timestamp + header in first column
	timeview := newTextView("")
	grid.AddItem(newTextView("Time"), 0, 0, 1, 1, 0, 0, false)
	grid.AddItem(timeview, 1, 0, 1, 1, 0, 0, false)

	tView, ok := timeview.(*tview.TextView)
	if !ok {
		log.Panic("Error converting timeview")
	}

	// Register input handlers to sync scrolling between textviews
	allPrimitives := append(primitives, tView)
	for _, tv := range allPrimitives {
		tv := tv.(*tview.TextView)

		inputwrapper := func(event *tcell.EventKey) *tcell.EventKey {
			row, _ := tv.GetScrollOffset()

			for _, tv2 := range allPrimitives {
				tv2 := tv2.(*tview.TextView)
				if event.Key() == tcell.KeyEnter {
					tv2.ScrollToEnd()
				} else if tv != tv2 {
					_, col := tv2.GetScrollOffset()
					tv2.ScrollTo(row, col)
				}
			}

			return event
		}

		tv.SetInputCapture(inputwrapper)
	}

	app := tview.NewApplication().SetRoot(grid, true).EnableMouse(false)

	// Receive log data in the background and send it to logfile + views
	go func() {
		for l := range logging.channel {
			logging.Write(l)

			lines := strings.Split(strings.TrimRight(l.Text, "\n"), "\n")
			for linenr, line := range lines {
				for c, p := range primitives {
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
					tView.Write([]byte(l.Ts.Format(time.Stamp) + "\n"))
				} else {
					tView.Write([]byte("\n"))
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
