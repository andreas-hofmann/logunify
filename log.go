package main

import (
	"encoding/gob"
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
	file     *os.File
	writer   *gob.Encoder
	reader   *gob.Decoder
	channel  chan LogEntry
	realtime bool
}

func (l *Log) Write(entry LogEntry) {
	if l.writer != nil {
		l.writer.Encode(entry)
	}
}

func (l *Log) Replay() {
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

			if l.realtime {
				if !initialized {
					initialized = true
					lastTime = entry.Ts
				} else {
					time.Sleep(entry.Ts.Sub(lastTime))
					lastTime = entry.Ts
				}
			}

			l.channel <- entry
		}
		close(l.channel)
	}()
}

func (l *Log) Close() {
	l.file.Close()
}

func InitLog(flags Flags) Log {
	var l Log

	l.channel = make(chan LogEntry)
	l.realtime = flags.Realtime

	// Set up a log writer, if a logfile is given
	if len(flags.LogFileName) > 0 {
		var err error

		if flags.Replay {
			l.file, err = os.Open(flags.LogFileName)
			if err != nil {
				log.Fatal("Could not open logfile: ", err.Error())
			}
			l.reader = gob.NewDecoder(l.file)
		} else {
			l.file, err = os.Create(flags.LogFileName)
			if err != nil {
				log.Fatal("Could not create logfile: ", err.Error())
			}
			l.writer = gob.NewEncoder(l.file)
		}
	}

	return l
}
