#!/bin/sh
VERSION=`git describe --all --long`
go build -ldflags "-X main.version=$VERSION"
