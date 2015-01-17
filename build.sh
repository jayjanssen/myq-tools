#!/bin/sh

gox -os="linux darwin freebsd" -arch="386 amd64 arm" -output="bin/myq_status.{{.OS}}-{{.Arch}}"
