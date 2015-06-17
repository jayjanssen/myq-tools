package myqlib

import (
	"bufio"
	"bytes"
	"io"
	"log"
	"strconv"
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
func parseSamples(reader io.Reader, ch chan MyqSample, interval time.Duration) {
	outputtype := BATCH // default to BATCH
	typechecked := false
	recordmatch := []byte(`Variable_name`)

	// We can't have intervals smaller than 1s
	// if the interval is larger, we check samples for intervals
	// so we can avoid parsing them fully.
	check_intervals := false
	uptime_str := []byte(`Uptime`)
	var prev_uptime float64
	if interval.Nanoseconds() > 1000000 {
		check_intervals = true
	}
	// Scan back for the Uptime in the given record and return true if it can be skipped
	skip_interval := func(record []byte) (skippable bool) {
		upt_pos := bytes.Index(record, uptime_str) + len(uptime_str) // After the Uptime
		if upt_pos >= 0 {
			upt_nl := bytes.IndexByte(record[upt_pos:], '\n') + upt_pos    // Find the next newline
			uptime_str := string(bytes.Trim(record[upt_pos:upt_nl], `| `)) // Trim extra chars
			current_uptime, _ := strconv.ParseFloat(uptime_str, 64)        // Parse the str to float
			if prev_uptime == 0 {
				prev_uptime = current_uptime
			} else {
				if current_uptime-prev_uptime < interval.Seconds() {
					// This sample's uptime is too early, skip it
					return true
				}
			}
		}
		return false
	}

	// This scanner will look for the start of a new set of SHOW STATUS output
	scanner := bufio.NewScanner(reader)
	scanner.Split(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		// Check if this looks like a TABULAR file, but only once
		if !typechecked {
			if bytes.HasPrefix(data, []byte(`+`)) || bytes.HasPrefix(data, []byte(`|`)) {
				outputtype, recordmatch = TABULAR, []byte(`| Variable_name`)
			}
			typechecked = true
		}

		if end := bytes.Index(data, recordmatch); end > 0 {
			// Found a new record
			nl := bytes.IndexByte(data[end:], '\n') // Find the subsequent newline

			if check_intervals && skip_interval(data[0:end]) {
				return end + nl + 1, nil, nil
			}
			return end + nl + 1, data[0:end], nil
		}

		// if we're at EOF and have data, return it, otherwise let it fall through
		if atEOF && len(data) > 0 {
			return len(data), data, nil
		}

		return 0, nil, nil
	})

	for scanner.Scan() {
		// The scanner sends complete samples
		parseBatch(ch, bytes.NewBuffer(scanner.Bytes()), outputtype)
	}

	// Not sure if we care here or not, remains to be seen
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}

func parseBatch(ch chan MyqSample, buffer *bytes.Buffer, outputtype showoutputtype) {
	var divideridx int

	timesample := make(MyqSample)
	scanner := bufio.NewScanner(buffer)

	for scanner.Scan() {
		line := scanner.Bytes()
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

		timesample[strings.ToLower(string(key))] = string(value)
	}

	if timesample.Length() > 0 {
		ch <- timesample
	}
}
