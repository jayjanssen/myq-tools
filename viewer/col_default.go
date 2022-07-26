package viewer

import (
	"fmt"

	"github.com/jayjanssen/myq-tools2/loader"
)

// A defaultCol contains base attributes and methods shared by all Cols
type defaultCol struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Type        string `yaml:"type"`
	Length      int    `yaml:"length"`
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

// Header for this view, unclear if state is needed
func (c defaultCol) GetHeader(sr loader.StateReader) []string {
	return []string{FitString(c.Name, c.Length)}
}

// Blank line for this view
func (c defaultCol) GetBlankLine() string {
	return FitString(` `, c.Length)
}
