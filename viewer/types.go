package viewer

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// A view is made up of Groups of Cols
type View struct {
	Name        string     `yaml:"name"`
	Description string     `yaml:"description"`
	Groups      []Colgroup `yaml:"groups"`
}

// A colgroup is a list of (related) cols
type Colgroup struct {
	Name        string
	Description string
	Cols        []Col
}

// The type of Col
type ColType int

const (
	// Given a key, this subtracts the prev key value from cur and divides by the num of seconds (/s)
	RATE ColType = iota
	// Simply output the cur value of the key
	GAUGE
	// Given a key, subtract the prev key value from the cur and emit the result
	DIFF
	// Given a key with a string value, emit the first Length chars
	STRING

	// Like string, but only shows the last Length chars
	RIGHTMOST

	// Given a list of keys, emit the rate of the sum of their values
	RATESUM

	// Given two keys, emit the ratio of the two (numerator / denominator)
	PERCENT

	// Takes a custom function, should probably just be special types
	FUNC
)

// Convert ColTypes in yaml string form to our internal const representation
func (ct *ColType) UnmarshalYAML(value *yaml.Node) error {
	switch value.Value {
	case `Rate`:
		*ct = RATE
	case `Gauge`:
		*ct = GAUGE
	default:
		return fmt.Errorf("Invalid ColType: %s", value.Value)
	}
	return nil
}

// The type of value
type UnitType int

const (
	NUMBER UnitType = iota
)

// Convert UnitTypes in yaml string form to our internal const representation
func (ut *UnitType) UnmarshalYAML(value *yaml.Node) error {
	switch value.Value {
	case `Number`:
		*ut = NUMBER
	default:
		return fmt.Errorf("Invalid UnitType: %s", value.Value)
	}
	return nil
}
