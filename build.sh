#!/bin/sh

# Full build on the local OS
make

# Linux builds
GOOS=linux GOARCH=386 make build
GOOS=linux GOARCH=amd64 make build
GOOS=linux GOARCH=arm make build

# Other OSes
GOOS=darwin GOARCH=amd64 make build
GOOS=freebsd GOARCH=amd64 make build
GOOS=openbsd GOARCH=amd64 make build
GOOS=windows GOARCH=amd64 make build