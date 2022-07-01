package viewer

// Calculate diff between two numbers, if negative, just return bigger
func CalculateDiff(bigger, smaller float64) float64 {
	if bigger < smaller {
		// special case -- if c is < p, the number rolled over or was reset, so best effort answer here.
		return bigger
	} else {
		return bigger - smaller
	}
}

// Calculate the rate of change between two values, given the time difference between them
func CalculateRate(bigger, smaller, seconds float64) float64 {
	diff := CalculateDiff(bigger, smaller)

	if seconds <= 0 { // negative seconds is weird
		return diff
	} else {
		return diff / seconds
	}
}

// Return the sum of all variables in the given data
// func CalculateSum(data model.MyqData, variable_names []string) (sum float64) {
// 	for _, v := range variable_names {
// 		v, _ := data.GetFloat(v)
// 		sum += v
// 	}
// 	return sum
// }
