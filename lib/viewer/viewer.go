package viewer

import (
	"fmt"

	myblip "github.com/jayjanssen/myq-tools/lib/blip"
	"gopkg.in/yaml.v3"
)

// A Viewer represents the output of data from metrics into a (usually) constrained width with a header and one or more lines of output
type Viewer interface {
	// Get name of the view
	GetName() string

	// Single line help for this viewer
	GetShortHelp() string

	// Detailed help for this viewer
	GetDetailedHelp() []string

	// A list of domains that this view requires
	GetDomains() []string

	// A list of source keys (domain/metric pairs) that this viewer requires
	GetRequiredMetrics() []SourceKey

	// A map of domain to list of metric names required by this viewer
	GetMetricsByDomain() map[string][]string

	// Header for this view
	GetHeader(*myblip.MetricCache) []string

	// Data for this view based on the metrics
	GetData(*myblip.MetricCache) []string

	// Blank for this view when we need to pad extra lines
	GetBlank() string
}

// A list of things that implement Viewer
type ViewerList []Viewer

type typesucker struct {
	Type string `yaml:"type"`
}

// Convert ViewerList entries into their individual types
func (svl *ViewerList) UnmarshalYAML(value *yaml.Node) error {
	var newlist ViewerList
	for _, content := range value.Content {
		typeobj := typesucker{}
		err := content.Decode(&typeobj)
		if err != nil {
			return err
		}

		switch typeobj.Type {
		case `String`:
			c := StringCol{}
			err := content.Decode(&c)
			if err != nil {
				return err
			}
			newlist = append(newlist, c)
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
		case `Switch`:
			c := SwitchCol{}
			err := content.Decode(&c)
			if err != nil {
				return err
			}
			newlist = append(newlist, c)
		case `Subtract`:
			c := SubtractCol{}
			err := content.Decode(&c)
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
