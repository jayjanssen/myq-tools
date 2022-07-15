package viewer

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// A list of things that implement StateViewer
type StateViewerList []StateViewer

// Convert StateViewerList entries into their individual types
func (svl *StateViewerList) UnmarshalYAML(value *yaml.Node) error {
	var newlist StateViewerList
	for _, content := range value.Content {
		rawmap := make(map[string]string)
		content.Decode(&rawmap)

		switch rawmap["type"] {
		case `Rate`:
			c := RateCol{}
			content.Decode(&c)
			newlist = append(newlist, c)
		case `Gauge`:
			c := GaugeCol{}
			content.Decode(&c)
			newlist = append(newlist, c)
		default:
			return fmt.Errorf("invalid column type: %s", rawmap["type"])
		}
	}
	*svl = newlist
	return nil
}
