package main

import (
	"context"
	"log"
	"os"
	"os/exec"
	"time"
)

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

func RunCmd(ctx context.Context, cmd CmdConfig, col int, output chan LogEntry) {
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
