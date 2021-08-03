package main

import (
	"fmt"
	"io"
	"os"
	"time"
)

type SplitLog struct {
	prefix  string
	handles []io.WriteCloser
}

func NewSplitLog(filename string, cfg []CmdConfig) *SplitLog {
	if len(filename) <= 0 {
		return nil
	}

	l := SplitLog{}

	l.prefix = filename

	for i, c := range cfg {
		handle, err := os.Create(l.prefix + "." + fmt.Sprintf("%d", i))
		if err == nil {
			l.handles = append(l.handles, handle)
			handle.Write([]byte(fmt.Sprintf("CMD: %s\n  Interval: %d\n  Loop: %v\n", c.Cmd, c.Params.IntervalMs, c.Params.Loop)))
		}
	}

	return &l
}

func (log SplitLog) Write(l LogEntry) {
	for c, h := range log.handles {
		if c+1 == l.Col {
			h.Write([]byte("\n" + l.Ts.Format(time.Stamp) + ": " + l.Text + "\n"))
		}
	}

}

func (log SplitLog) Close() {
	for _, h := range log.handles {
		h.Close()
	}
}
