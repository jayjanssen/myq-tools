package viewer

import (
	"fmt"

	"github.com/jayjanssen/myq-tools2/loader"
)

// A view is made up of Groups of Cols
type View struct {
	// Views *are* a GroupCol in that they can have a Cols
	GroupCol `yaml:",inline"`

	// Usually a view would have Groups OR Cols, but not both.  If both, print groups first, then individual cols
	Groups []GroupCol `yaml:"groups"`
}

// How to print out the time with our output
var timeCol SampleTimeCol = NewSampleTimeCol()

// Get help for this view
func (v View) GetDetailedHelp() (output []string) {
	// Gather the svs
	var svs StateViewerList
	for _, group := range v.Groups {
		svs = append(svs, group)
	}
	svs = append(svs, v.Cols...)

	// Gather and indent the lines
	output = append(output, v.GetShortHelp())
	for _, sv := range svs {
		for _, line := range sv.GetDetailedHelp() {
			output = append(output, fmt.Sprintf("   %s", line))
		}
	}
	return
}

// A list of sources that this view requires
func (v View) GetSources() ([]loader.SourceName, error) {
	return []loader.SourceName{}, nil
	// return []*loader.Source{
	// 	&loader.Source{},
	// }, nil
}

// Header for this view, unclear if state is needed
func (v View) GetHeader(sr loader.StateReader) []string {
	// Collect all the StateViewers for this view
	var svs StateViewerList
	svs = append(svs, timeCol)
	for _, group := range v.Groups {
		svs = append(svs, group)
	}
	svs = append(svs, v.Cols...)

	// Get the header output of all those svs
	colOuts := pushColOutputDown(svs, func(sv StateViewer) []string {
		return sv.GetHeader(sr)
	})

	// Get the length of this view based on the length of the first colOut
	if v.Length == 0 && len(colOuts) > 0 {
		v.Length = len(colOuts[0])
	}

	return colOuts
}

// Data for this view based on the state
func (v View) GetData(sr loader.StateReader) (result []string) {
	// Collect all the StateViewers for this view
	var svs StateViewerList
	svs = append(svs, timeCol)
	for _, group := range v.Groups {
		svs = append(svs, group)
	}
	svs = append(svs, v.Cols...)

	// Get the data output of all those svs
	return pushColOutputUp(svs, func(sv StateViewer) []string {
		return sv.GetData(sr)
	})
}
