package viewer

import "gopkg.in/yaml.v3"

var (
	Views []View
)

func LoadDefaultViews() error {
	return ParseViews(default_view_yaml)
}

func ParseViews(yaml_str string) error {
	return yaml.Unmarshal([]byte(yaml_str), &Views)
}

const default_view_yaml = `---
- name: cttf
  description: "Connections, Threads, Tables, and Files"
  groups:
    - name: Connects
      description: "Connection related metrics"
      cols:
        - name: cons
          description: "Connections per second"
          length: 4
          var_key: 'connections'
          precision: 0
          units: 'Number'
`
