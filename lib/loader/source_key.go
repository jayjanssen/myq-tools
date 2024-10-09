package loader

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// A SourceKey identifies a single key from a single source
type SourceKey struct {
	SourceName SourceName
	Key        string
}

// Convert UnitTypes in yaml string form to our internal const representation
func (sk *SourceKey) UnmarshalYAML(value *yaml.Node) error {
	name, key, found := strings.Cut(value.Value, `/`)
	if !found {
		return fmt.Errorf("sourcekey invalid format: %s", value.Value)
	}

	skey := SourceKey{
		SourceName: SourceName(name),
		Key:        key,
	}
	*sk = skey
	return nil
}
