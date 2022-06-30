package col

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

type numCol struct {
	defaultCol
	Units     UnitType
	Precision int
}

// The type of numeric value
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
		return fmt.Errorf("invalid UnitType: %s", value.Value)
	}
	return nil
}
