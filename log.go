package main

import (
	"encoding/gob"
	"time"
)

type LogEntry struct {
	Ts   time.Time
	Col  int
	Text string
}

func ReplayLog(decoder *gob.Decoder, realtime bool, output chan LogEntry) {
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
