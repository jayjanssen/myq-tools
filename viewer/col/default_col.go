package col

import (
	"fmt"

	"github.com/jayjanssen/myq-tools2/loader"
)

// A defaultCol contains base attributes and methods shared by all Cols
type defaultCol struct {
	Name        string
	Description string
	Sources     []string
	Type        string
	Length      int
}

func (c defaultCol) GetName() string {
	return c.Name
}

// Single line help for the view
func (c defaultCol) GetShortHelp() string {
	return fmt.Sprintf("%s: %s", c.Name, c.Description)
}

// A list of sources that this view requires
func (c defaultCol) GetSources() ([]*loader.Source, error) {
	var sources []*loader.Source
	for _, source_str := range c.Sources {
		sp, err := loader.GetSource(source_str)
		if err == nil {
			sources = append(sources, sp)
		} else {
			return nil, err
		}
	}
	return sources, nil
}

// Header for this view, unclear if state is needed
func (c defaultCol) GetHeader(sr loader.StateReader) (result []string) {
	result = append(result, fmt.Sprintf("%*s", c.Length, c.Name))
	return
}

// helper function to fit a plain string to our Length
func (c defaultCol) fitString(input string) string {
	if len(input) > int(c.Length) {
		return input[0:c.Length] // First width characters
	} else {
		return fmt.Sprintf(`%*s`, c.Length, input)
	}
}
