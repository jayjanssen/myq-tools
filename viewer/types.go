package viewer

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// A list of things that implement StateViewer
type StateViewerList []StateViewer

type typesucker struct {
	Type string `yaml:"type"`
}

// Convert StateViewerList entries into their individual types
func (svl *StateViewerList) UnmarshalYAML(value *yaml.Node) error {
	var newlist StateViewerList
	for _, content := range value.Content {
		typeobj := typesucker{}
		err := content.Decode(&typeobj)
		if err != nil {
			return err
		}

		switch typeobj.Type {
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
		case `RateSum`:
			c := RateSumCol{}
			err = content.Decode(&c)
			if err != nil {
				return err
			}
			newlist = append(newlist, c)
		case `Diff`:
			c := DiffCol{}
			err = content.Decode(&c)
			if err != nil {
				return err
			}
			newlist = append(newlist, c)
		case `Percent`:
			c := PercentCol{}
			err = content.Decode(&c)
			if err != nil {
				return err
			}
			newlist = append(newlist, c)
		case `SortedExpandedCounts`:
			c := SortedExpandedCountsCol{}
			err = content.Decode(&c)
			if err != nil {
				return err
			}
			newlist = append(newlist, c)
		default:
			return fmt.Errorf("invalid column type: %s", typeobj.Type)
		}
	}
	*svl = newlist
	return nil
}
