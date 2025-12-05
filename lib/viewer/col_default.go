package viewer

import (
	"fmt"

	myblip "github.com/jayjanssen/myq-tools/lib/blip"
)

// A defaultCol contains base attributes and methods shared by all Cols
type defaultCol struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description"`
	Type        string   `yaml:"type"`
	Length      int      `yaml:"length"`
	Domains     []string // List of domains this column needs
}

func (c defaultCol) GetName() string {
	return c.Name
}

// Single line help for the view
func (c defaultCol) GetShortHelp() string {
	return fmt.Sprintf("%s: %s", c.Name, c.Description)
}

// Detailed help -- by default same as short help
func (c defaultCol) GetDetailedHelp() []string {
	result := make([]string, 1)
	result[0] = c.GetShortHelp()
	return result
}

// A list of domains that this view requires
func (c defaultCol) GetDomains() []string {
	return c.Domains
}

// A list of source keys that this column requires
func (c defaultCol) GetRequiredMetrics() []SourceKey {
	return []SourceKey{} // default implementation returns empty
}

// A map of domain to list of metric names (default implementation)
func (c defaultCol) GetMetricsByDomain() map[string][]string {
	result := make(map[string][]string)
	for _, key := range c.GetRequiredMetrics() {
		if key.Domain == "" || key.Metric == "" {
			continue
		}
		if result[key.Domain] == nil {
			result[key.Domain] = []string{}
		}
		result[key.Domain] = append(result[key.Domain], key.Metric)
	}
	return result
}

// Header for this view
func (c defaultCol) GetHeader(cache *myblip.MetricCache) []string {
	return []string{FitString(c.Name, c.Length)}
}

// Blank space for this col
func (c defaultCol) GetBlank() string {
	return FitString(` `, c.Length)
}
