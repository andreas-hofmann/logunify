# logunify

A small tool, which allows running multiple shellcommands in parallel, and log their output to a unified logfile. The logged data can be replayed later.

To use it, edit the included config.yaml file, and add your commands.

Available config parameters:
 - `loop: [true|false]` -> Run the command in an endless loop.
 - `intervalMs: <milliseconds>` -> Interval between looped command executions.

You only need the yaml config when running commands. For replaying, a logfile is sufficient. If your device is short on storage space, sending out the log data over a TCP connection is supported, too.

This was created from the need to gather debugging info on an Android system, but should work on Linux as well (or any system with /bin/bash or /bin/sh, for that matter). To cross compile it, simply export `GOOS=android` and `GOARCH=arm` (or arm64) before running `go build`.