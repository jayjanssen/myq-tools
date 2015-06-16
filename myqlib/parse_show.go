package myqlib

import (
	"bufio"
	"bytes"
	"log"
	"strings"
	"time"
	// "fmt"
)

// Different types of files to parse
type showoutputtype uint8

const (
	BATCH showoutputtype = iota
	TABULAR
)

// Parse lines from mysql SHOW output.
func scanMySQLShowLines(scanner *bufio.Scanner, ch chan MyqSample, interval time.Duration) {
	outputtype := BATCH // default to BATCH
	typechecked := false

	batch := make( []string, 0 )

	for scanner.Scan() {
		// The scanner sends complete lines
		line := scanner.Bytes()

		if len( line ) == 0 {
			continue
		}

		// Check if this looks like a TABULAR file, but only once
		if !typechecked {
			if bytes.HasPrefix( line, []byte(`+`)) || bytes.HasPrefix( line, []byte(`|`)) {
				outputtype = TABULAR
			}
			typechecked = true
		}

		if bytes.HasPrefix( line, []byte(`+`)) {
			continue
		} else if bytes.HasPrefix(line, []byte(`| Variable_name`)) { // Needs to catch TABULAR too
			// Send the old sample (if any) and start a new one
			if len(batch) > 0 {
				parseBatch(ch, batch, outputtype )
				batch = make( []string, 0 )
			}
		// } else if bytes.HasPrefix( line, []byte("| Uptime")) {
			// check the interval based on the current and last uptime
		} else {
			batch = append(batch, string( line ))
		}

	}
	// Send the last one
	if len(batch) > 0 {
		parseBatch(ch, batch, outputtype)
	}

	// Not sure if we care here or not, remains to be seen
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}

func parseBatch( ch chan MyqSample, batch []string, outputtype showoutputtype) {
	var divideridx int

	timesample := make(MyqSample)

	for _, line := range batch {
		var key, value string

		switch outputtype {
			case TABULAR:
				// Line here looks like this: (value can contain spaces)
				// | varname   | value    |
				if !strings.HasPrefix(line, `|`) {
					continue
				}

				if divideridx == 0 {
					divideridx = strings.Index(line, ` | `)
				} else if len(line) < divideridx {
					continue // line truncated, probably EOF
				}

				key = strings.Trim(line[:divideridx], `| `)
				value = strings.Trim(line[divideridx:], `| `)
			case BATCH:
				raw := strings.Split(line, "\t")
				if len(raw) != 2 {
					continue
				}
				key, value = raw[0], raw[1]
		}

		// normalize keys to lowercase
		timesample[strings.ToLower(key)] = value
	}

	ch <- timesample
}
