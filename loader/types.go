package loader

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// Source to collect a Sample
type Source struct {
	Name        SourceName
	Description string
	Query       string
	Type        SourceType
	Key         string
}

// A SourceName identifies some unique portion of data gathered from a Source
type SourceName string

// A SourceType identifies how source query results are parsed and stored
type SourceType int

const (
	STRING SourceType = iota
	MAP
)

func (st *SourceType) UnmarshalYAML(value *yaml.Node) error {
	switch value.Value {
	case `String`:
		*st = STRING
	case `Map`:
		*st = MAP
	default:
		return fmt.Errorf(`invalid SourceType: %s`, value.Value)
	}
	return nil
}
