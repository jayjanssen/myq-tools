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

// A col represents a single output unit in a view
type Col struct {
	Name        string
	Description string
	Source      string
	Key         string
	Type        ColType
	Units       UnitType
	Length      int
	Precision   int
}

// The type of Col
type ColType int

const (
	RATE ColType = iota
	GAUGE
)

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

func (ut *UnitType) UnmarshalYAML(value *yaml.Node) error {
	switch value.Value {
	case `Number`:
		*ut = NUMBER
	default:
		return fmt.Errorf("Invalid UnitType: %s", value.Value)
	}
	return nil
}
