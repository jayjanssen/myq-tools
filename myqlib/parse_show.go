package myqlib

import (
	"bufio"
	"log"
	// "regexp"
	"github.com/glenn-brown/golang-pkg-pcre/src/pkg/pcre"
	"strconv"
	"strings"
)

var mysqlShowRE pcre.Regexp = pcre.MustCompile(`^\|[[:space:]]([[:word:]]+)[[:space:]]+\|[[:space:]]([[:graph:]]+)[[:space:]]+\|$`, 0)

// Parse lines from mysql SHOW output.
func scanMySQLShowLines(scanner *bufio.Scanner, ch chan MyqSample) {
	timesample := make(MyqSample)

	for scanner.Scan() {
		match := mysqlShowRE.MatcherString(scanner.Text(), 0)
		if match.Matches() {
			if match.GroupString(1) == "Variable_name" {
				// Send the old sample (if any) and start a new one
				if timesample.Length() > 0 {
					ch <- timesample
					timesample = make(MyqSample)
				}
			} else {
				// normalize keys to lowercase
				lowerkey := strings.ToLower(match.GroupString(1))
				timesample[lowerkey] = convert(match.GroupString(2))
			}
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
	if ans, err := strconv.ParseInt(s, 0, 64); err == nil {
		return ans
	} else if ans, err := strconv.ParseFloat(s, 64); err == nil {
		return ans
	} else if s == `ON` {
		return true
	} else if s == `OFF` {
		return false
	} else {
		return s
	} // Just leave it as a string
}
