package viewer

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// A view is made up of Groups of Cols
type View struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`

	// Usually a view would have Groups OR Cols, but not both.  If both, print groups first, then individual cols
	Groups []Colgroup      `yaml:"groups"`
	Cols   StateViewerList `yaml:"cols"`
}

// A list of things that implement StateViewer
type StateViewerList []StateViewer

// Convert StateViewerList entries into their individual types
func (svl *StateViewerList) UnmarshalYAML(value *yaml.Node) error {
	var newlist StateViewerList
	for _, content := range value.Content {
		rawmap := make(map[string]string)
		// yaml.Unmarshal(content, &rawmap)
		content.Decode(&rawmap)
		fmt.Printf("content type: %v\n", rawmap["type"])

		switch rawmap["type"] {
		case `Rate`:
			fmt.Println("here")
			c := Col{}
			content.Decode(&c)
			fmt.Printf("content decoded: %v\n", c)
			newlist = append(newlist, c)
		case `Gauge`:
			c := Col{}
			content.Decode(&c)
			newlist = append(newlist, c)
		default:
			return fmt.Errorf("invalid column type: %s", rawmap["type"])
		}
	}
	*svl = newlist
	return nil
}

// A colgroup is a list of (related) cols
type Colgroup struct {
	Name        string          `yaml:"name"`
	Description string          `yaml:"description"`
	Cols        StateViewerList `yaml:"cols"`
}
