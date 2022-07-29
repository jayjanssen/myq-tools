package viewer

import (
	"fmt"
	"sort"
	"strings"
	"unicode/utf8"

	"gopkg.in/yaml.v3"
)

// A colNum is an abstract object that contains Units and Precision values.  Implementation of those are left to the "subclasses"
type colNum struct {
	defaultCol `yaml:",inline"`
	Units      UnitsType `yaml:"units"`
	Precision  int       `yaml:"precision"`
}

// The type of numeric value
type UnitsType int

const (
	NUMBER UnitsType = iota
	MEMORY
	SECOND
	MICROSECOND
	NANOSECOND
	PICOSECOND
	PERCENT
)

// Units Definitions allow us to collapse numbers and use a postfix instead
type UnitsDef map[float64]string

// Lookup map given a UnitsType
var unitsLookup = map[UnitsType]UnitsDef{
	NUMBER: {
		1:          ``,
		1000:       `k`,
		1000000:    `m`,
		1000000000: `g`,
	},
	MEMORY: {
		1:             `b`,
		1024:          `K`,
		1048576:       `M`,
		1073741824:    `G`,
		1099511627776: `T`,
	},
	SECOND: {
		1000:        `ks`,
		1:           `s`,
		0.001:       `ms`,
		0.000001:    `µs`,
		0.000000001: `ns`,
	},
	MICROSECOND: {
		1000000000: `ks`,
		1000000:    `s`,
		1000:       `ms`,
		1:          `µs`,
	},
	NANOSECOND: {
		1000000000: `s`,
		1000000:    `ms`,
		1000:       `µs`,
		1:          `ns`,
	},
	PICOSECOND: {
		1000000000000: `s`,
		1000000000:    `ms`,
		1000000:       `µs`,
		1000:          `ns`,
		1:             `ps`,
	},
	PERCENT: {
		1: `%`,
	},
}

// Convert UnitTypes in yaml string form to our internal const representation
func (ut *UnitsType) UnmarshalYAML(value *yaml.Node) error {
	switch value.Value {
	case `Number`:
		*ut = NUMBER
	case `Memory`:
		*ut = MEMORY
	case `Second`:
		*ut = SECOND
	case `Microsecond`:
		*ut = MICROSECOND
	case `Nanosecond`:
		*ut = NANOSECOND
	case `Picosecond`:
		*ut = PICOSECOND
	case `Percent`:
		*ut = PERCENT
	default:
		return fmt.Errorf("invalid UnitType: %s", value.Value)
	}
	return nil
}

// Given the value, fit it into our Precision, Length, and Units
// callers should pass the Col.Precision value as the second argument
func (nc colNum) fitNumber(value float64, precision int) string {
	// Get the units we will be using
	units := unitsLookup[nc.Units]

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
		str := fmt.Sprintf(`%.*f%s`, precision, raw, unit)
		left := nc.Length - utf8.RuneCountInString(str)

		// fmt.Printf("%f, %d, %d, %s, %f, %s, %d\n", value, nc.Length, nc.Precision, unit, raw, str, left)

		if raw >= 0 && (nc.Length+precision)-utf8.RuneCountInString(str) >= 0 {
			// Our number is > 0 and fits into nc.Length + precision
			if left < 0 {
				if precision > 0 {
					// No space left, try to chop the precision
					return nc.fitNumber(value, precision-1)
				} else {
					// Nothing to chop, any bigger factors will be too wide, so return here.
					return str
				}
			} else if left > 1 && factor != 1 {
				// If we have space for some extra precision, use it
				return fmt.Sprintf(`%.*f%s`, left-1, raw, unit)
			} else {
				if factor != 1 && raw < 1 && left > 0 && fmt.Sprintf(`%.1f`, raw) != `1.0` {
					// Raw is < 1, therefore str is rounded up.  Let's print a decimal instead
					return fmt.Sprintf(`%0.*f%s`, precision+left, raw, unit)[1:]
				} else if factor != 1 && str == fmt.Sprintf("0%s", unit) {
					if left > 0 {
						// There's still some space left to print something intelligent
						return fmt.Sprintf(`%.*f%s`, precision+1, raw, unit)[1:]
					}

					// if we are returning 0m, 0k, etc, then we can't fit this number into the size given
					return strings.Repeat(`#`, nc.Length)
				} else {
					// Just return what we've got
					return str
				}
			}
		}
	}

	// We're past the highest factor and nothing fits
	str := fmt.Sprintf(`%.*f`, precision, value)
	if len(str) > nc.Length && precision > 0 {
		// We can try chopping precision here for a fit
		return nc.fitNumber(value, precision-1)
	} else {
		// Just print it (too wide)
		// return str
		return strings.Repeat(`#`, int(nc.Length))
	}
}
