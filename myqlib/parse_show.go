package myqlib

import (
	"bufio"
	"log"
	"strconv"
	"strings"
)

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

	for scanner.Scan() {
		// The scanner sends complete lines
		line := scanner.Text()
		
		// Check if this looks like a TABULAR file, but only once
		if !typechecked {
			if strings.HasPrefix( line, `+`) || strings.HasPrefix( line, `|` ){
				outputtype = TABULAR
			}
			typechecked = true
		}
		
		var key, value string
		
		switch outputtype {
		case TABULAR:
			// Line here looks like this: (value can contain spaces)
			// | varname   | value    |
			raw := strings.Split( line, ` | `)
			if len(raw) != 2 { continue } 
			
			key = strings.Trim( raw[0], `| `  )
			value = strings.Trim( raw[1], `| ` )
		case BATCH:
			raw := strings.Split( line, "\t")
			if len(raw) != 2 { continue } 

			key = raw[0]
			value = raw[1]
		}

		if key == "Variable_name" {
			// Send the old sample (if any) and start a new one
			if timesample.Length() > 0 {
				ch <- timesample
				timesample = make(MyqSample)
			}
		} else {
			// normalize keys to lowercase
			timesample[strings.ToLower(key)] = convert(value)
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

// Detect the type of the input string based on regexes
func convert(s string) interface{} {
	// Check for numeric types first (Int should be most common anyway)
	if ans, err := strconv.ParseInt(s, 0, 64); err == nil {
		return ans
	}
	if ans, err := strconv.ParseFloat(s, 64); err == nil {
		return ans
	}
	
	switch s {
	case `ON`:
		return true
	case `OFF`:
		return false
	default:
		// Just leave it as a string
		return s
	} 
}
