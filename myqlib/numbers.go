package myqlib

import (
	"fmt"
	"sort"
	"strings"
	"unicode/utf8"
)

type UnitsDef map[float64]string

var (
	NumberUnits UnitsDef = UnitsDef{
		1:          ``,
		1000:       `k`,
		1000000:    `m`,
		1000000000: `g`,
	}
	MemoryUnits UnitsDef = UnitsDef{
		1:             `b`,
		1024:          `K`,
		1048576:       `M`,
		1073741824:    `G`,
		1099511627776: `T`,
	}
	SecondUnits UnitsDef = UnitsDef{
		1000:        `ks`,
		1:           `s`,
		0.001:       `ms`,
		0.000001:    `µs`,
		0.000000001: `ns`,
	}
	MicroSecondUnits UnitsDef = UnitsDef{
		1000000000: `ks`,
		1000000:    `s`,
		1000:       `ms`,
		1:          `µs`,
	}
	NanoSecondUnits UnitsDef = UnitsDef{
		1000000000: `s`,
		1000000:    `ms`,
		1000:       `µs`,
		1:          `ns`,
	}
	PercentUnits UnitsDef = UnitsDef{
		1: `%`,
	}
)

// Fit the given value of units into width + at most precision decimals
func collapse_number(value float64, width int64, precision int64, units UnitsDef) string {
	// Load the factors from the given unit and sort them
	var factors []float64
	for k := range units {
		factors = append(factors, k)
	}
	sort.Float64s(factors)

	// Starting from the smallest to the biggest factors
	for _, factor := range factors {
		unit := units[factor]
		raw := value / factor
		str := fmt.Sprintf(fmt.Sprint(`%.`, precision, `f%s`), raw, unit)
		left := width - int64(utf8.RuneCountInString(str))

		// fmt.Printf( "%f, %d, %d, %s, %f, %s, %d\n", value, width, precision, unit, raw, str, left )

		if raw >= 0 && (width+precision)-int64(utf8.RuneCountInString(str)) >= 0 {
			// Our number is > 0 and fits into width + precision
			if left < 0 {
				if precision > 0 {
					// No space left, try to chop the precision
					return collapse_number(value, width, precision-1, units)
				} else {
					// Nothing to chop, any bigger factors will be too wide, so return here.
					return str
				}
			} else if left > 1 && factor != 1 {
				// If we have space for some extra precision, use it
				return fmt.Sprintf(fmt.Sprint(`%.`, left-1, `f%s`), raw, unit)
			} else {
				if factor != 1 && raw < 1 && left > 0 && fmt.Sprintf(`%.1f`, raw ) != `1.0` {
					// Raw is < 1, therefore str is rounded up.  Let's print a decimal instead
					return fmt.Sprintf(fmt.Sprint(`%0.`, precision+left, `f%s`), raw, unit)[1:]
				} else if factor != 1 && str == fmt.Sprintf("0%s", unit) {
					if left > 0 {
						// There's still some space left to print something intelligent
						return fmt.Sprintf(fmt.Sprint(`%.`, precision+1, `f%s`), raw, unit)[1:]
					}

					// if we are returning 0m, 0k, etc, then we can't fit this number into the size given
					return strings.Repeat(`#`, int(width))
				} else {
					// Just return what we've got
					return str
				}
			}
		}
	}

	// We're past the highest factor and nothing fits
	str := fmt.Sprintf(fmt.Sprint(`%.`, precision, `f`), value)
	if int64(len(str)) > width && precision > 0 {
		// We can try chopping precision here for a fit
		return collapse_number(value, width, precision-1, units)
	} else {
		// Just print it (too wide)
		// return str
		return strings.Repeat(`#`, int(width))
	}
}

// Calculate diff between two numbers, if negative, just return bigger
func calculate_diff(bigger, smaller float64) float64 {
	if bigger < smaller {
		// special case -- if c is < p, the number rolled over or was reset, so best effort answer here.
		return bigger
	} else {
		return bigger - smaller
	}
}

// Calculate the rate of change between two values, given the time difference between them
func calculate_rate(bigger, smaller, seconds float64) float64 {
	diff := calculate_diff(bigger, smaller)

	if seconds <= 0 { // negative seconds is weird
		return diff
	} else {
		return diff / seconds
	}
}

// Return the sum of all variables in the given sample
func calculate_sum(sample MyqSample, variable_names []string) (sum float64) {
	for _, v := range variable_names {
		v, _ := sample.getFloat(v)
		sum += v
	}
	return sum
}
