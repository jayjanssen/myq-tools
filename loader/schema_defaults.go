package loader

import "gopkg.in/yaml.v3"

var (
	Schemas []MyqSchema
)

func LoadDefaultSchemas() error {
	return ParseSchemas(default_schemas_yaml)
}

func ParseSchemas(yaml_str string) error {
	return yaml.Unmarshal([]byte(yaml_str), &Schemas)
}

const default_schemas_yaml = `---
- name: status
  description: "SHOW GLOBAL STATUS or SELECT * FROM performance_schema.global_status"
  auto: true
  keys:
    Uptime: INT
    Threads_running: INT
    Ssl_cipher_list: STRING
`
