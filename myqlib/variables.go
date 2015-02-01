package myqlib

import (
	"regexp"
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
