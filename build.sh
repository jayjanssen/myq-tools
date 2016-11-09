#!/bin/sh

# Run tests
go test -bench=. ./...

# Do builds
gox -os="linux darwin freebsd" -arch="386 amd64 arm" -ldflags "-X main.build_version=manual -X main.build_timestamp=`date -u +%Y%m%d.%H%M%S`" -output="bin/myq_status.{{.OS}}-{{.Arch}}"

# Create upload files
tar cvzf myq_tools.tgz bin/*
zip myq_tools.zip bin/*
