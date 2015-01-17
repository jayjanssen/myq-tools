#!/bin/sh

go test -bench=. ./...

gox -build-toolchain
gox -os="linux darwin freebsd" -arch="386 amd64 arm" -output="bin/myq_status.{{.OS}}-{{.Arch}}"
tar cvzf myq_tools.tgz bin/*
