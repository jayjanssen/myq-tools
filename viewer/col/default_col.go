package col

import (
	"fmt"

	"github.com/jayjanssen/myq-tools2/loader"
	"gopkg.in/yaml.v3"
)

// A defaultCol contains base attributes and methods shared by all Cols
type defaultCol struct {
	Name        string
	Description string
	Sources     []string
	Type        ColType
	Length      int
}

func (c defaultCol) GetName() string {
	return c.Name
}

// Single line help for the view
func (c defaultCol) GetShortHelp() string {
	return fmt.Sprintf("%s: %s", c.Name, c.Description)
}

// A list of sources that this view requires
func (c defaultCol) GetSources() ([]*loader.Source, error) {
	var sources []*loader.Source
	for _, source_str := range c.Sources {
		sp, err := loader.GetSource(source_str)
		if err == nil {
			sources = append(sources, sp)
		} else {
			return nil, err
		}
	}
	return sources, nil
}

// Header for this view, unclear if state is needed
func (c defaultCol) GetHeader(sr loader.StateReader) (result []string) {
	result = append(result, fmt.Sprintf("%*s", c.Length, c.Name))
	return
}

// Data for this view based on the state
func (c defaultCol) GetData(sr loader.StateReader) (result []string) {
	var raw string
	switch c.Type {
	case RATE:
		raw = c.getRate(sr)
	}

	var res []string
	res = append(res, raw)
	return res
}

// Get data when it is a number, apply precision
func (c defaultCol) getRate(sr loader.StateReader) (result string) {
	// This sucks because currently Col has sources and a key, but that should:
	// a) allow a mechanism to specify a source/key in one attribute
	// b) allow different types of Cols to take one key, two keys, or a list of depending on what the calculation is.    This implies this function should be in a subclass
	// cur, prev := sr.GetKeyCurPrev

	// Apply precision
	// return fmt.Sprintf("%.2s")
	return "   5"
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
		return fmt.Errorf("invalid ColType: %s", value.Value)
	}
	return nil
}
