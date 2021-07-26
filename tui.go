package main

import (
	"log"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type TUI struct {
	app        *tview.Application
	grid       *tview.Grid
	primitives []tview.Primitive
	tview      *tview.TextView
}

func newTextView(text string) tview.Primitive {
	return tview.NewTextView().
		SetTextAlign(tview.AlignLeft).
		SetText(text).
		SetWrap(false)
}

func InitTUI(flags Flags, cfg []CmdConfig) TUI {
	var tui TUI

	tui.grid = tview.NewGrid().
		SetBorders(true).
		SetRows(1).
		SetColumns(16)

	// Set up cmd textviews
	col := 1
	for _, c := range cfg {
		// Add header
		tui.grid.AddItem(newTextView(c.Cmd), 0, col, 1, 1, 0, 0, false)

		// Add cmd output
		p := newTextView("")
		tui.primitives = append(tui.primitives, p)
		tui.grid.AddItem(p, 1, col, 1, 1, 0, 0, false)

		col++
	}

	// Add timestamp + header in first column
	timeview := newTextView("")
	tui.grid.AddItem(newTextView("Time"), 0, 0, 1, 1, 0, 0, false)
	tui.grid.AddItem(timeview, 1, 0, 1, 1, 0, 0, false)

	tView, ok := timeview.(*tview.TextView)
	if !ok {
		log.Panic("Error converting timeview")
	}

	tui.tview = tView

	// Register input handlers to sync scrolling between textviews
	allPrimitives := append(tui.primitives, tui.tview)
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

	tui.app = tview.NewApplication().SetRoot(tui.grid, true).EnableMouse(false)

	return tui
}

func (t TUI) AddData(data LogEntry) {
	newlines := ""

	for i := 0; i < strings.Count(data.Text, "\n"); i++ {
		newlines += "\n"
	}

	for c, p := range t.primitives {
		tv, ok := p.(*tview.TextView)
		if !ok {
			log.Panic("Can't convert TV")
		}
		if (c + 1) == data.Col {
			tv.Write([]byte(data.Text))
		} else {
			tv.Write([]byte(newlines))
		}
	}

	t.tview.Write([]byte(data.Ts.Format(time.Stamp)))
	t.tview.Write([]byte(newlines))
}

func (t TUI) Update() {
	t.app.Draw()
}

func (t TUI) Run() {
	// Fire up the tview event loop
	if err := t.app.Run(); err != nil {
		log.Panic(err)
	}
}
