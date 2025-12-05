package blip

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/cashapp/blip"
)

// Different types of files to parse
type showoutputtype uint8

const (
	F_END_STRING string         = "MYQTOOLSEND"
	BATCH        showoutputtype = iota
	TABULAR
)

// FileParser reads mysqladmin ext output and converts to blip.Metrics
type FileParser struct {
	scanner    *bufio.Scanner
	file       *os.File
	outputtype showoutputtype
	fileName   string
	interval   time.Duration
	statusFile string
	varFile    string
}

// NewFileParser creates a new file parser for mysqladmin output
func NewFileParser(statusFile, varFile string) *FileParser {
	return &FileParser{
		statusFile: statusFile,
		varFile:    varFile,
	}
}

// Initialize prepares the file parser
func (f *FileParser) Initialize(interval time.Duration) error {
	f.interval = interval

	// Open the status file
	r, err := os.OpenFile(f.statusFile, os.O_RDONLY, 0)
	if err != nil {
		return fmt.Errorf("cannot open status file %s: %w", f.statusFile, err)
	}
	f.file = r // Store the file handle for later closing
	// Note: File will be kept open for GetMetrics to read from
	// It will be closed when the scanner finishes or when GetMetrics is done

	// Check the interval
	if interval.Seconds() < 1 {
		r.Close() // Close file before returning error
		return fmt.Errorf("interval cannot be less than 1s (%s)", interval.String())
	}

	uptime_str := []byte(`Uptime`)
	var prev_uptime float64

	// Scan back for the Uptime in the given record and return true if it can be skipped
	skip_interval := func(record []byte) (skippable bool) {
		upt_idx := bytes.Index(record, uptime_str)
		if upt_idx < 0 {
			// Uptime not found in record
			return false
		}

		upt_pos := upt_idx + len(uptime_str) // After the Uptime

		// Find the next newline
		nl_idx := bytes.IndexByte(record[upt_pos:], '\n')
		if nl_idx < 0 {
			// No newline found, can't parse uptime
			return false
		}
		upt_nl := nl_idx + upt_pos

		// Trim extra chars
		uptime_str := string(bytes.TrimSpace(bytes.Trim(record[upt_pos:upt_nl], `| `)))
		// Parse the str to float
		current_uptime, _ := strconv.ParseFloat(uptime_str, 64)

		// if current and previous uptimes differ less than the interval, skip
		if prev_uptime > 0 && current_uptime-prev_uptime < interval.Seconds() {
			return true
		}

		prev_uptime = current_uptime
		return false
	}

	// This scanner will look for the start of a new set of SHOW STATUS output
	f.scanner = bufio.NewScanner(r)
	f.scanner.Buffer(make([]byte, 100), bufio.MaxScanTokenSize*16)

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

			// If no newline found, we need more data
			if nl < 0 {
				return 0, nil, nil
			}

			// if our record match is at position 0, we skip this line and start from the next
			if end == 0 {
				return end + nl + 1, nil, nil
			}

			// If we are checking interval, see if we should skip this record
			if interval.Seconds() > 1 && skip_interval(data[0:end]) {
				return end + nl + 1, nil, nil
			}
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

// parseSample parses a single record from mysqladmin ext output
func (f *FileParser) parseSample(data []byte) (map[string]string, error) {
	buffer := bytes.NewBuffer(data)
	var divideridx int
	result := make(map[string]string)

	chunkScanner := bufio.NewScanner(buffer)

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

		result[strings.ToLower(string(key))] = string(value)
	}

	return result, nil
}

// GetMetrics returns a channel that produces blip.Metrics from the file
func (f *FileParser) GetMetrics() <-chan *blip.Metrics {
	ch := make(chan *blip.Metrics, 1)

	go func() {
		defer close(ch)
		defer func() {
			if f.file != nil {
				f.file.Close()
			}
		}()

		var intervalNum uint = 0
		startTime := time.Now()

		for f.scanner.Scan() {
			data := f.scanner.Bytes()
			if len(data) == 0 {
				continue
			}

			// Parse the sample
			statusData, err := f.parseSample(data)
			if err != nil || len(statusData) == 0 {
				continue
			}

			// Convert to blip.Metrics
			metrics := f.convertToBlipMetrics(statusData, intervalNum, startTime)
			ch <- metrics

			intervalNum++
			startTime = startTime.Add(f.interval)
		}

		if err := f.scanner.Err(); err != nil {
			fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
		}
	}()

	return ch
}

// convertToBlipMetrics converts parsed mysqladmin data to blip.Metrics
func (f *FileParser) convertToBlipMetrics(data map[string]string, interval uint, startTime time.Time) *blip.Metrics {
	// Convert all string values to MetricValues
	statusMetrics := []blip.MetricValue{}
	varMetrics := []blip.MetricValue{}

	// Known gauge metrics (rest are assumed to be cumulative counters)
	gauges := map[string]bool{
		"threads_running":                true,
		"threads_connected":              true,
		"prepared_stmt_count":            true,
		"innodb_buffer_pool_pages_dirty": true,
		"innodb_buffer_pool_pages_free":  true,
		"innodb_buffer_pool_pages_total": true,
		"innodb_row_lock_current_waits":  true,
		"innodb_os_log_pending_writes":   true,
		"max_used_connections":           true,
	}

	for key, valStr := range data {
		// Try to convert to float64
		val, err := strconv.ParseFloat(valStr, 64)
		if err != nil {
			// Not a numeric value, skip
			continue
		}

		metricType := blip.CUMULATIVE_COUNTER
		if gauges[key] {
			metricType = blip.GAUGE
		}

		mv := blip.MetricValue{
			Name:  key,
			Value: val,
			Type:  metricType,
		}

		// For now, put everything in status.global
		// In a real implementation, we'd want to categorize better
		statusMetrics = append(statusMetrics, mv)
	}

	return &blip.Metrics{
		Begin:     startTime,
		End:       startTime.Add(f.interval),
		MonitorId: f.statusFile,
		Plan:      "file-replay",
		Level:     "default",
		Interval:  interval,
		State:     blip.STATE_ACTIVE,
		Values: map[string][]blip.MetricValue{
			"status.global": statusMetrics,
			"var.global":    varMetrics, // Empty for now unless varFile is parsed separately
		},
	}
}
