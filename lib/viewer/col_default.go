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

// Header for this view
func (c defaultCol) GetHeader(cache *myblip.MetricCache) []string {
	return []string{FitString(c.Name, c.Length)}
}

// Blank space for this col
func (c defaultCol) GetBlank() string {
	return FitString(` `, c.Length)
}
