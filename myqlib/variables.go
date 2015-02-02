package myqlib

import (
	"regexp"
	"fmt"
)

// Given a variable list (potentially with regexes) and a sample, expand the variables to all possible matches
func expand_variables(variables []string, sample MyqSample) (expanded []string) {
	hash := map[string]int{}
	for _, variable := range variables {
		re, err := regexp.Compile(variable)
		if err != nil {
			// Just pass it through as-is
			hash[variable] = 1
		} else {
			// Got a regex, loop through all keys to try to find matches
			for key, _ := range sample {
				if re.MatchString(key) {
					hash[key] = 1
				}
			}
		}
	}
	for key, _ := range hash {
		expanded = append(expanded, key)
	}
	return
}

// Fit a given string into a width
func fit_string(val string, width int64) (string) {
	if len(val) > int(width) {
		return val[0:width] // First width characters
	} else {
		return fmt.Sprintf(fmt.Sprint(`%`, width, `s`), val)
	}
}

// Fit a given string into a width
func right_fit_string(val string, width int64) (string) {
	if len(val) > int(width) {
		return val[len(val)-int(width):]
	} else {
		return fmt.Sprintf(fmt.Sprint(`%`, width, `s`), val)
	}
}

func column_filler(c Col) (string) {
	return fit_string("-", c.Width())
}
func column_blank(c Col) (string) {
	return fit_string(" ", c.Width())
}
