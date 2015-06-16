package myqlib

import (
	"bufio"
	"bytes"
	"log"
	"strings"
)

// Different types of files to parse
type showoutputtype uint8

const (
	BATCH showoutputtype = iota
	TABULAR
)

// Parse lines from mysql SHOW output.
func scanMySQLShowLines(scanner *bufio.Scanner, ch chan MyqSample) {
	timesample := make(MyqSample)
	outputtype := BATCH // default to BATCH
	typechecked := false
	var divideridx int

	for scanner.Scan() {
		// The scanner sends complete lines
		line := scanner.Bytes()

		// Check if this looks like a TABULAR file, but only once
		if !typechecked {
			if bytes.HasPrefix(line, []byte(`+`)) || bytes.HasPrefix(line, []byte(`|`)) {
				outputtype = TABULAR
			}
			typechecked = true
		}

		var key, value []byte

		switch outputtype {
		case TABULAR:
			// Line here looks like this: (value can contain spaces)
			// | varname   | value    |
			if !bytes.HasPrefix(line, []byte(`|`)) {
				continue
			}

			if divideridx == 0 {
				divideridx = bytes.Index(line, []byte(` | `))
			} else if len(line) < divideridx {
				continue // line truncated, probably EOF
			}

			key = bytes.Trim(line[:divideridx], `| `)
			value = bytes.Trim(line[divideridx:], `| `)
		case BATCH:
			raw := bytes.Split(line, []byte("\t"))
			if len(raw) != 2 {
				continue
			}
			key, value = raw[0], raw[1]
		}

		if bytes.Equal(key, []byte("Variable_name")) {
			// Send the old sample (if any) and start a new one
			if timesample.Length() > 0 {
				ch <- timesample
				timesample = make(MyqSample)
			}
		} else {
			// normalize keys to lowercase
			timesample[strings.ToLower(string(key))] = string(value)
		}
	}
	// Send the last one
	if timesample.Length() > 0 {
		ch <- timesample
	}

	// Not sure if we care here or not, remains to be seen
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}
