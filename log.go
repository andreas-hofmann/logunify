package main

import (
	"encoding/gob"
	"io"
	"log"
	"os"
	"time"
)

type LogEntry struct {
	Ts   time.Time
	Col  int
	Text string
}

type Log struct {
	handle io.ReadWriteCloser
	writer *gob.Encoder
	reader *gob.Decoder
}

func (l *Log) Write(entry LogEntry) {
	if l.writer != nil {
		l.writer.Encode(entry)
	}
}

func (l *Log) Replay(ch chan LogEntry, realtime bool) {
	if l.reader == nil {
		return
	}

	go func() {
		var lastTime time.Time
		initialized := false

		for {
			var entry LogEntry

			if err := l.reader.Decode(&entry); err != nil {
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

			ch <- entry
		}
		close(ch)
	}()
}

func (l *Log) Close() {
	l.handle.Close()
}

func InitLogfile(filename string) Log {
	var l Log

	// Set up a log writer, if a logfile is given
	if len(filename) > 0 {
		var err error

		l.handle, err = os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0644)
		if err != nil {
			log.Fatal("Could not open logfile: ", err.Error())
		}
		l.reader = gob.NewDecoder(l.handle)
		l.writer = gob.NewEncoder(l.handle)
	}

	return l
}
