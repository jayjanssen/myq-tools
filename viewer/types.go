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
		err := content.Decode(&rawmap)
		if err != nil {
			return err
		}

		switch rawmap["type"] {
		case `Rate`:
			c := RateCol{}
			err = content.Decode(&c)
			if err != nil {
				return err
			}
			newlist = append(newlist, c)
		case `Gauge`:
			c := GaugeCol{}
			err = content.Decode(&c)
			if err != nil {
				return err
			}
			newlist = append(newlist, c)
		default:
			return fmt.Errorf("invalid column type: %s", rawmap["type"])
		}
	}
	*svl = newlist
	return nil
}
