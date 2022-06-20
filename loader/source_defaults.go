package loader

import "gopkg.in/yaml.v3"

var (
	Sources []Source
)

func LoadDefaultSources() error {
	return ParseSources(defaultSourcesYaml)
}

func ParseSources(yaml_str string) error {
	return yaml.Unmarshal([]byte(yaml_str), &Sources)
}

const defaultSourcesYaml = `---
- name: status
  description: "Mysql server global status counters"
`
