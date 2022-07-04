package loader

import (
	"bytes"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Different types of files to parse
type showoutputtype uint8

const (
	F_END_STRING string         = "MYQTOOLSEND"
	BATCH        showoutputtype = iota
	TABULAR
)

type FileParser struct {
	scanner    *Scanner
	outputtype showoutputtype
	fileName   string
}

func NewFileParser(fileName string) *FileParser {
	f := FileParser{fileName: fileName}
	return &f
}

func (f *FileParser) Initialize(interval time.Duration) error {
	// Open the given file
	r, err := os.OpenFile(f.fileName, os.O_RDONLY, 0)
	if err != nil {
		return err
	}

	// Check the interval
	if interval.Seconds() < 1 {
		return fmt.Errorf("Interval cannot be less than 1s (%s)", interval.String())
	}

	uptime_str := []byte(`Uptime`)
	var prev_uptime float64

	// Scan back for the Uptime in the given record and return true if it can be skipped
	skip_interval := func(record []byte) (skippable bool) {
		upt_pos := bytes.Index(record, uptime_str) + len(uptime_str) // After the Uptime
		if upt_pos >= 0 {
			// Find the next newline
			upt_nl := bytes.IndexByte(record[upt_pos:], '\n') + upt_pos
			// Trim extra chars
			uptime_str := string(bytes.TrimSpace(bytes.Trim(record[upt_pos:upt_nl], `| `)))
			// Parse the str to float
			current_uptime, _ := strconv.ParseFloat(uptime_str, 64)

			// if current and previous uptimes differ less than the interval, skip
			if prev_uptime > 0 && current_uptime-prev_uptime < interval.Seconds() {
				return true
			}

			prev_uptime = current_uptime
		}
		return false
	}

	// This scanner will look for the start of a new set of SHOW STATUS output
	f.scanner = NewScanner(r)
	typechecked := false                // if we've checked for TABULAR yet or not
	recordmatch := []byte(F_END_STRING) // How to match records (type dependant)
	f.outputtype = BATCH                // default to BATCH

	f.scanner.Split(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		// Check if this looks like a TABULAR file, but only once
		if !typechecked {
			if bytes.HasPrefix(data, []byte(`+`)) || bytes.HasPrefix(data, []byte(`|`)) {
				f.outputtype, recordmatch = TABULAR, []byte(`| Variable_name`)
			}
			typechecked = true
		}

		// Find a new record
		if end := bytes.Index(data, recordmatch); end >= 0 {
			nl := bytes.IndexByte(data[end:], '\n') // Find the subsequent newline

			// if our record match is at position 0, we skip this line and start from the next
			if end == 0 {
				return end + nl + 1, nil, nil
			}

			// If we are checking interval, see if we should skip this record
			if interval.Seconds() > 1 && skip_interval(data[0:end]) {
				return end + nl + 1, nil, nil
			}
			// fmt.Println( "Found record: ", string(data[0:end]))
			return end + nl + 1, data[0:end], nil
		}

		// if we're at EOF and have data, return it, otherwise let it fall through
		if atEOF && len(data) > 0 {
			return len(data), data, nil
		}

		// Didn't see a record end or a EOF, ask for more data
		return 0, nil, nil
	})

	return nil
}

// Scan for the next record set in the file and return it
// If the return is (nil, nil), it indicates end of file
func (f *FileParser) GetNextSample() *Sample {
	if !f.scanner.Scan() {
		if err := f.scanner.Err(); err != nil {
			return NewSampleErr(err)
		} else {
			return nil // EOF
		}
	}

	buffer := bytes.NewBuffer(f.scanner.Bytes())
	var divideridx int

	sample := NewSample()
	chunkScanner := NewScanner(buffer)

	for chunkScanner.Scan() {
		line := chunkScanner.Bytes()
		var key, value []byte

		switch f.outputtype {
		case TABULAR:
			// Line here looks like this: (value can contain spaces)
			// | varname   | value    |
			if !bytes.HasPrefix(line, []byte(`|`)) {
				continue
			}

			// Get the position of the divider if we don't have it already
			if divideridx == 0 {
				divideridx = bytes.Index(line, []byte(` | `))
			} else if len(line) < divideridx {
				// line truncated, probably EOF
				continue
			}

			// Grab the key and value and trim the whitespace
			key = bytes.Trim(line[:divideridx], `| `)
			value = bytes.Trim(line[divideridx:], `| `)
		case BATCH:
			// Batch is much easier, just split on the tab
			raw := bytes.Split(line, []byte("\t"))
			// If we don't get 2 fields, skip it.
			if len(raw) != 2 {
				continue
			}
			key, value = raw[0], raw[1]
		}

		sample.Data[strings.ToLower(string(key))] = string(value)
	}

	if len(sample.Data) > 0 {
		return sample
	} else {
		return f.GetNextSample()
	}
}
