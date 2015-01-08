package myqlib

import (
	"bufio"
	"log"
	"regexp"
	"strconv"
	"strings"
)

// Why can't I put these in a const?  no idea.  I'm using globals here just so I'm not recompiling these regexes very often
var mysqlShowRE *regexp.Regexp = regexp.MustCompile(`^\|\s(\w+)\s+\|\s(\S+)\s+\|$`)
var mysqlUIntType *regexp.Regexp = regexp.MustCompile(`^\d+$`)
var mysqlFloatType *regexp.Regexp = regexp.MustCompile(`^\d+\.\d+$`)
var mysqlBoolType *regexp.Regexp = regexp.MustCompile(`^(ON|OFF)$`)

// Parse lines from mysql SHOW output.
func scanMySQLShowLines(scanner *bufio.Scanner, ch chan MyqSample) {
	timesample := make(MyqSample)

	for scanner.Scan() {
		matches := mysqlShowRE.FindStringSubmatch(scanner.Text())
		if matches != nil {
			if matches[1] == "Variable_name" {
				// Send the old sample (if any) and start a new one
				if timesample.Length() > 0 {
					ch <- timesample
					timesample = make(MyqSample)
				}
			} else {
				// normalize keys to lowercase
				lowerkey := strings.ToLower(matches[1])
				timesample[lowerkey] = convert(matches[2])
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
	if mysqlUIntType.MatchString(s) {
		// To int or uint, that is the question
		ans, _ := strconv.ParseInt(s, 0, 64)
		return ans
	} else if mysqlFloatType.MatchString(s) {
		ans, _ := strconv.ParseFloat(s, 64)
		return ans
	} else if mysqlBoolType.MatchString(s) {
		if s == "ON" {
			return true
		} else {
			return false
		}
	} else {
		// Just leave it as a string
		return s
	}
}
