# logunify

A small tool, which allows running multiple shellcommands in parallel, and log their output to a unified logfile. The logged data can be replayed later.

To use it, take one of the included config files, and add your commands, and tweak it to your needs.

You only need the yaml config when running commands. For replaying, a logfile is sufficient. If your device is short on storage space, sending out the log data over a TCP connection is supported, too.

This was created from the need to gather debugging info on an Android system, but should work on Linux as well (or any system with /bin/bash or /bin/sh, for that matter). To cross compile it, simply export `GOOS=android` and `GOARCH=arm` (or arm64) before running `go build` (or use the included build.sh script).

## Config file

The config file has three sections:

* *flags*
* *init*
* *runtime*

The *flags* section allows setting (or overriding) the command line flags, which can make usage on an embedded system more convenient. Just push the binary + logunify.yaml to the device, and run it.

The *init* section contains shell-commands, which are run once before logging starts. Enabling log-messages, clearing buffers, etc. can be done here.

The *runtime* section contains all commands, which shall be used to collect log info. Available options are:

* _loop_: {True|False} - Run the command in an endless loop
* _intervalMs_: \<milliseconds\> - Sleep time between looped calls

Note that each command must end with a colon, even when no options are supplied.
See the included example configs for a starting point.

## Some usage examples:

Load a connect config, send logdata to the remote and additionally write a local logfile.

    ./logunify -config cfg_examples/cfg_connect.yaml -logfile local.log

Listen for connections + display received data via TUI:

    ./logunify -replay -listen -ui

Listen for connections + write received data to a logfile:

    ./logunify -replay -listen -logfile unified.log

Split a unified logfile to distinct files:

    ./logunify -replay -logfile unified.log -splitlog