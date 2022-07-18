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

	Sources []loader.SourceName
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

// A list of sources that this view requires
func (c defaultCol) GetSources() ([]loader.SourceName, error) {
	return c.Sources, nil
}

// Header for this view, unclear if state is needed
func (c defaultCol) GetHeader(sr loader.StateReader) (result []string) {
	result = append(result, fmt.Sprintf("%*s", c.Length, c.Name))
	return
}

// Blank line for this view
func (c defaultCol) GetBlankLine() string {
	return FitString(` `, c.Length)
}
