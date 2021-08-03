package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
)

type Flags struct {
	ConfigFileName string
	LogFileName    string
	SplitLogFiles  bool
	UI             bool
	MaxLines       int
	Realtime       bool
	Replay         bool
	Listen         bool
	Connect        bool
	Addr           string
	Port           int
}

type FlagError struct {
	err string
}

func (e FlagError) Error() string {
	return e.err
}

func parseFlags() Flags {
	var f Flags

	var printversion bool

	flag.StringVar(&f.ConfigFileName, "config", "./logunify.yaml", `Config file to use. Only required when recording,
not for replaying.`)
	flag.StringVar(&f.LogFileName, "logfile", "", "Log file to write to.")
	flag.BoolVar(&f.SplitLogFiles, "splitlog", false, "Split log output to separate logfiles (one per command).")
	flag.BoolVar(&f.Replay, "replay", false, "Replay log data from a file or remote.")
	flag.BoolVar(&f.Realtime, "realtime", false, "Replay in real time (including pauses).")
	flag.BoolVar(&f.Listen, "listen", false, "Listen for incoming connections.")
	flag.IntVar(&f.MaxLines, "maxlines", 500, "Maximum lines in UI buffer. 0 for unlimited scrollback.")
	flag.BoolVar(&f.UI, "ui", false, "Enable UI. Implied if neither a remote, nor a logfile is given.")
	flag.BoolVar(&f.Connect, "connect", false, "Connect to a remote host.")
	flag.StringVar(&f.Addr, "address", "", "Address to connect/bind to.")
	flag.IntVar(&f.Port, "port", 20000, "Port to use when logging over TCP.")

	flag.BoolVar(&printversion, "version", false, "Print the program version.")

	flag.Parse()

	if printversion {
		fmt.Println("logunify version: ", version)
		os.Exit(0)
	}

	if f.MaxLines < 0 {
		f.MaxLines = 0
	}

	if !(f.logfile() || f.remoteConnection()) {
		f.UI = true
	}

	return f
}

func (flags *Flags) update(flagmap map[string]string) {
	for k, v := range flagmap {
		switch k {
		case "address":
			flags.Addr = v
		case "logfile":
			flags.LogFileName = v
		case "ui":
			flags.UI = v == "true"
		case "replay":
			flags.Replay = v == "true"
		case "realtime":
			flags.Realtime = v == "true"
		case "listen":
			flags.Listen = v == "true"
		case "connect":
			flags.Connect = v == "true"
		case "splitlog":
			flags.SplitLogFiles = v == "true"
		case "addr":
			flags.Addr = v
		case "port":
			p, err := strconv.Atoi(v)
			if err == nil {
				flags.Port = p
			}
		case "maxLines":
			m, err := strconv.Atoi(v)
			if err == nil {
				flags.MaxLines = m
			}
		}
	}
}

func (flags Flags) checkError() error {
	if flags.Listen && flags.Connect {
		return FlagError{"Can only -listen or -connect"}
	}

	if flags.Port <= 0 {
		return FlagError{"Invalid port"}
	}

	if flags.MaxLines < 0 {
		return FlagError{"Invalid maxlines"}
	}

	if flags.SplitLogFiles && !flags.logfile() {
		return FlagError{"Logfile required when splitting log"}
	}

	return nil
}

func (f Flags) writeLogFile() bool {
	return !f.Replay || f.remoteConnection()
}

func (f Flags) writeLogRemote() bool {
	return !f.Replay && f.remoteConnection()
}

func (f Flags) remoteAddr() string {
	if f.remoteConnection() {
		return f.Addr + ":" + fmt.Sprint(f.Port)
	}
	return ""
}

func (f Flags) remoteConnection() bool {
	return f.Connect || f.Listen
}

func (f Flags) logfile() bool {
	return len(f.LogFileName) > 0
}
