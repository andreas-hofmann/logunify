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

type DataLog struct {
	handle io.ReadWriteCloser
	writer *gob.Encoder
	reader *gob.Decoder
}

func (l DataLog) WriteCfg(cfg []CmdConfig) {
	if l.writer == nil {
		return
	}

	for _, c := range cfg {
		l.writer.Encode(c)
	}
}

func (l DataLog) ReadCfg() (cfg []CmdConfig) {
	if l.reader == nil {
		return cfg
	}

	for {
		var c CmdConfig
		if err := l.reader.Decode(&c); err != nil {
			break
		}
		cfg = append(cfg, c)
	}

	return cfg
}

func (l DataLog) Write(entry LogEntry) {
	if l.writer == nil {
		return
	}

	l.writer.Encode(entry)
}

func (l DataLog) Replay(ch chan LogEntry, realtime bool) {
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

func (l DataLog) Close() {
	if l.handle != nil {
		l.handle.Close()
	}
}

func NewLogFile(filename string, write bool) (l DataLog) {
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

func NewLogRemote(peer string, write bool) (l DataLog) {
	if len(peer) <= 0 {
		return
	}

	var err error

	if write {
		log.Println("Connecting to", peer)
		for {
			l.handle, err = net.Dial("tcp", peer)
			if err == nil {
				l.writer = gob.NewEncoder(l.handle)
				break
			}
			time.Sleep(time.Second * 2)
			log.Println("Retrying connection:", err)
		}
	} else {
		conn, err := net.Listen("tcp", peer)
		if err != nil {
			log.Fatal("Could not listen for connections:", err.Error())
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
