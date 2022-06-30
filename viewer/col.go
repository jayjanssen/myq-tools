package viewer

import (
	"fmt"

	"github.com/jayjanssen/myq-tools2/loader"
)

// A col represents a single output unit in a view and implements the StateViewer interface
type Col struct {
	Name        string
	Description string
	Sources     []string
	Key         string
	Type        ColType
	Units       UnitType
	Length      int
	Precision   int
}

// Single line help for the view
func (c Col) GetShortHelp() string {
	return fmt.Sprintf("%s: %s", c.Name, c.Description)
}

// A list of sources that this view requires
func (c Col) GetSources() ([]*loader.Source, error) {
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
func (c Col) GetHeader(sr loader.StateReader) (result []string) {
	result = append(result, fmt.Sprintf("%*s", c.Length, c.Name))
	return
}

// Data for this view based on the state
func (c Col) GetData(sr loader.StateReader) (result []string) {
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
func (c Col) getRate(sr loader.StateReader) (result string) {
	// This sucks because currently Col has sources and a key, but that should:
	// a) allow a mechanism to specify a source/key in one attribute
	// b) allow different types of Cols to take one key, two keys, or a list of depending on what the calculation is.    This implies this function should be in a subclass
	// cur, prev := sr.GetKeyCurPrev

	// Apply precision
	// return fmt.Sprintf("%.2s")
	return "   5"
}
