package main

import (
	"encoding/gob"
	"io"
	"log"
	"net"
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
	if l.handle != nil {
		l.handle.Close()
	}
}

func NewLogFile(filename string, write bool) (l Log) {
	if len(filename) <= 0 {
		return
	}

	var err error

	l.handle, err = os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		log.Fatal("Could not open logfile: ", err.Error())
	}

	if write {
		l.writer = gob.NewEncoder(l.handle)
	} else {
		l.reader = gob.NewDecoder(l.handle)
	}

	return
}

func NewLogRemote(peer string, write bool) (l Log) {
	if len(peer) <= 0 {
		return
	}

	var err error

	if write {
		l.handle, err = net.Dial("tcp", peer)
		if err == nil {
			l.writer = gob.NewEncoder(l.handle)
		}
	} else {
		conn, err := net.Listen("tcp", peer)
		if err != nil {
			log.Fatal("Could not listen for connections: ", err.Error())
		}

		log.Println("Listening for connections...")

		l.handle, err = conn.Accept()
		if err != nil {
			log.Fatal("Error accepting connection: ", err.Error())
		}
		l.reader = gob.NewDecoder(l.handle)
	}

	return
}
