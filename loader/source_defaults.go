package loader

import (
	_ "embed"
	"fmt"

	"gopkg.in/yaml.v3"
)

var (
	sources   []*Source
	sourceMap map[SourceName]*Source
)

//go:embed sources_defaults.yaml
var defaultSourcesYaml string

func LoadDefaultSources() error {
	return ParseSources(defaultSourcesYaml)
}

// Lookup a source given its name
func GetSource(source_name SourceName) (*Source, error) {
	sp, ok := sourceMap[source_name]
	if !ok {
		return nil, fmt.Errorf("source not found: %s", source_name)
	}
	return sp, nil
}

func ParseSources(yaml_str string) error {
	err := yaml.Unmarshal([]byte(yaml_str), &sources)
	if err != nil {
		return err
	}

	sourceMap = make(map[SourceName]*Source)

	for _, source := range sources {
		sourceMap[source.Name] = source
	}
	return nil
}
