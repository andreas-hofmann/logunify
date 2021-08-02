#!/bin/sh
VERSION=`git describe --all --long --dirty`
go build -ldflags "-X main.version=$VERSION"
