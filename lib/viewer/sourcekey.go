package viewer

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// SourceKey represents a reference to a metric in the form "domain/metric" or "domain/metric_name"
// Examples: "status.global/queries", "innodb/rows_read"
type SourceKey struct {
	Domain string
	Metric string
	Raw    string // Original string for error messages
}

// ParseSourceKey parses a string like "status.global/queries" into domain and metric
func ParseSourceKey(s string) (SourceKey, error) {
	parts := strings.SplitN(s, "/", 2)
	if len(parts) != 2 {
		return SourceKey{}, fmt.Errorf("invalid source key format: %s (expected domain/metric)", s)
	}

	// Support legacy format: "status" maps to "status.global", "variables" maps to "var.global"
	domain := parts[0]
	if domain == "status" {
		domain = "status.global"
	} else if domain == "variables" {
		domain = "var.global"
	}

	return SourceKey{
		Domain: domain,
		Metric: strings.ToLower(parts[1]), // blip uses lowercase
		Raw:    s,
	}, nil
}

// UnmarshalYAML allows SourceKey to be loaded from YAML
func (sk *SourceKey) UnmarshalYAML(value *yaml.Node) error {
	var s string
	if err := value.Decode(&s); err != nil {
		return err
	}

	parsed, err := ParseSourceKey(s)
	if err != nil {
		return err
	}

	*sk = parsed
	return nil
}

// String returns the string representation
func (sk SourceKey) String() string {
	if sk.Raw != "" {
		return sk.Raw
	}
	return fmt.Sprintf("%s/%s", sk.Domain, sk.Metric)
}
