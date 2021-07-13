package main

import (
	"context"
	"encoding/gob"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"sort"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"gopkg.in/yaml.v3"
)

type ConfigParameters struct {
	Loop       bool  `yaml:"loop"`
	IntervalMs int64 `yaml:"intervalMs"`
}

type ConfigMap map[string]ConfigParameters

type CmdConfig struct {
	Cmd    string
	Params ConfigParameters
}

type LogEntry struct {
	Ts   time.Time
	Col  int
	Text string
}

func getShell() []string {
	shells := [][]string{
		{"/bin/bash", "-c"},
		{"/bin/sh", "-c"},
		{"/system/bin/sh", "-c"},
		{"/vendor/bin/sh", "-c"},
	}

	for _, sh := range shells {
		if _, err := os.Stat(sh[0]); os.IsNotExist(err) {
			continue
		}
		return sh
	}

	log.Panic("No suitable shell found!")
	return []string{}
}

func replayLog(decoder *gob.Decoder, realtime bool, output chan LogEntry) {
	var lastTime time.Time
	initialized := false

	for {
		var entry LogEntry

		if err := decoder.Decode(&entry); err != nil {
			break
		}

		if realtime {
			if !initialized {
				initialized = true
				lastTime = entry.Ts
			} else {
				time.Sleep(entry.Ts.Sub(lastTime))
				lastTime = entry.Ts
			}
		}

		output <- entry
	}
}

func runCmd(ctx context.Context, cmd CmdConfig, col int, output chan LogEntry) {
	var executable []string
	executable = append(executable, getShell()...)
	executable = append(executable, cmd.Cmd)

	for {
		c := exec.CommandContext(ctx, executable[0], executable[1:]...)

		o, e := c.StdoutPipe()
		if e != nil {
			output <- LogEntry{time.Now(), col, "[ Error starting " + cmd.Cmd + ": " + e.Error() + " ]"}
		}

		if err := c.Start(); err != nil {
			output <- LogEntry{time.Now(), col, "[ Error starting " + cmd.Cmd + ": " + err.Error() + " ]"}
			continue
		}

		data := make([]byte, 4096, 4096)

		for {
			n, err := o.Read(data)
			if n <= 0 {
				break
			}

			if err == nil {
				output <- LogEntry{time.Now(), col, string(data[0:n])}
			}
		}

		c.Wait()

		if !cmd.Params.Loop {
			break
		}

		if cmd.Params.IntervalMs > 0 {
			time.Sleep(time.Duration(cmd.Params.IntervalMs * 1_000_000))
		}
	}
}

func newTextView(text string) tview.Primitive {
	return tview.NewTextView().
		SetTextAlign(tview.AlignLeft).
		SetText(text).
		SetWrap(false)
}

func readConfig(path string) (config []CmdConfig) {
	cfgfile, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatal("Error reading config: ", err.Error())
	}

	cfg := make(ConfigMap)

	err = yaml.Unmarshal(cfgfile, &cfg)
	if err != nil {
		log.Fatal("Error parsing config: ", err.Error())
	}

	sortedkeys := make([]string, 0, len(cfg))

	for k := range cfg {
		sortedkeys = append(sortedkeys, k)
	}

	sort.Strings(sortedkeys)

	for _, k := range sortedkeys {
		config = append(config, CmdConfig{k, cfg[k]})
	}

	return config
}

func main() {
	// First off, parse command line arguments
	var configFileName string
	var logFileName string
	var replay bool
	var realtime bool

	flag.StringVar(&configFileName, "config", "./config.yaml", "Config file to use")
	flag.StringVar(&logFileName, "logfile", "", "Log file to write to")
	flag.BoolVar(&replay, "replay", false, "Replay a stored log file")
	flag.BoolVar(&realtime, "realtime", false, "Replay a stored log file in real time")

	flag.Parse()

	cfg := readConfig(configFileName)
	logChan := make(chan LogEntry)

	var logfile *os.File = nil
	var logwriter *gob.Encoder = nil
	var logreader *gob.Decoder = nil

	// Set up a log writer, if a logfile is given
	if len(logFileName) > 0 {
		var err error

		if replay {
			logfile, err = os.Open(logFileName)
			if err != nil {
				log.Fatal("Could not open logfile: ", err.Error())
			}
			logreader = gob.NewDecoder(logfile)
		} else {
			logfile, err = os.Create(logFileName)
			if err != nil {
				log.Fatal("Could not create logfile: ", err.Error())
			}
			logwriter = gob.NewEncoder(logfile)
		}

		defer logfile.Close()
	}

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

		if !replay {
			go runCmd(ctx, c, col, logChan)
		}

		col++
	}

	if replay {
		go replayLog(logreader, realtime, logChan)
	}

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
		for l := range logChan {
			if logwriter != nil {
				logwriter.Encode(l)
			}

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

			app.Draw()
		}
	}()

	// Fire up the tview event loop
	if err := app.Run(); err != nil {
		log.Panic(err)
	}
}
