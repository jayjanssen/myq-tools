#!/bin/sh

BUILDOS=$(uname -s)
BUILDARCH=$(uname -p)
OUTFILE="build/myq_status.$BUILDOS-$BUILDARCH"

go build -i -o $OUTFILE myq_status.go && echo "Built $OUTFILE"
