package myqlib

import (
	"regexp"
)

// Given a variable list (potentially with regexes) and a sample, expand the variables to all possible matches
func expand_variables(variables []string, sample MyqSample) (expanded []string) {
	
	for _, variable := range variables {		
		re, err := regexp.Compile( variable )
		if err != nil {
			// Just push the variable into the resultset
			expanded = append( expanded, variable )
		} else {
			// Got a regex, loop through all keys to try to find matches
			for key, _ := range sample {
				if re.MatchString( key ) {
					expanded = append( expanded, key )
				}
			}
		}
	}
	return
}