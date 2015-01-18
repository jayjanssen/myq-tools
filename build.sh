#!/bin/sh

# Run tests
go test -bench=. ./...

# Do builds
gox -os="linux darwin freebsd" -arch="386 amd64 arm" -build-toolchain
gox -os="linux darwin freebsd" -arch="386 amd64 arm" -output="bin/myq_status.{{.OS}}-{{.Arch}}"

# Create upload files
tar cvzf myq_tools.tgz bin/*
zip myq_tools.zip bin/*
