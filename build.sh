#!/bin/sh
VERSION=`git describe --all --long --dirty`

export GOOS=linux GOARCH=amd64
go build -ldflags "-X main.version=$VERSION"

export GOOS=windows GOARCH=amd64
go build -ldflags "-X main.version=$VERSION"

export GOOS=android GOARCH=arm64
go build -ldflags "-X main.version=$VERSION" -o logunify_android
