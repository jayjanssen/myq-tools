package viewer

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// A view is made up of Groups of Cols
type View struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`

	// Usually a view would have Groups OR Cols, but not both.  If both, print groups first, then individual cols
	Groups []Colgroup      `yaml:"groups"`
	Cols   StateViewerList `yaml:"cols"`
}

// A list of things that implement StateViewer
type StateViewerList []StateViewer

// Convert StateViewerList entries into their individual types
func (svl *StateViewerList) UnmarshalYAML(value *yaml.Node) error {
	var newlist StateViewerList
	for _, content := range value.Content {
		rawmap := make(map[string]string)
		// yaml.Unmarshal(content, &rawmap)
		content.Decode(&rawmap)
		fmt.Printf("content type: %v\n", rawmap["type"])

		switch rawmap["type"] {
		case `Rate`:
			fmt.Println("here")
			c := Col{}
			content.Decode(&c)
			fmt.Printf("content decoded: %v\n", c)
			newlist = append(newlist, c)
		case `Gauge`:
			c := Col{}
			content.Decode(&c)
			newlist = append(newlist, c)
		default:
			return fmt.Errorf("invalid type: %s", rawmap["type"])
		}
	}
	*svl = newlist
	return nil
}

// A colgroup is a list of (related) cols
type Colgroup struct {
	Name        string          `yaml:"name"`
	Description string          `yaml:"description"`
	Cols        StateViewerList `yaml:"cols"`
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
