package viewer

import "github.com/jayjanssen/myq-tools2/loader"

// A view is made up of Groups of Cols
type View struct {
	GroupCol `yaml:",inline"`

	// Usually a view would have Groups OR Cols, but not both.  If both, print groups first, then individual cols
	Groups []GroupCol `yaml:"groups"`
}

// Detailed multi-line help for the view
func (v View) GetDetailedHelp() []string {
	return []string{""}
}

// A list of sources that this view requires
func (v View) GetSources() ([]loader.SourceName, error) {
	return []loader.SourceName{}, nil
	// return []*loader.Source{
	// 	&loader.Source{},
	// }, nil
}

// Header for this view, unclear if state is needed
func (v View) GetHeader(sr loader.StateReader) (result []string) {
	// Collect all the StateViewers for this view
	var svs StateViewerList
	for _, group := range v.Groups {
		svs = append(svs, group)
	}
	svs = append(svs, v.Cols...)

	// Get the header output of all those svs
	colOuts := groupColOutput(svs, func(sv StateViewer) []string {
		return sv.GetHeader(sr)
	})

	// Get the length of this view based on the length of the first colOut
	if v.Length == 0 && len(colOuts) > 0 {
		v.Length = len(colOuts[0])
	}

	// Send our name, then the output of our StateViewers
	result = append(result, FitStringLeft(v.Name, v.Length))
	result = append(result, colOuts...)
	return
}

// Data for this view based on the state
func (v View) GetData(sr loader.StateReader) (result []string) {
	// Collect all the StateViewers for this view
	var svs StateViewerList
	for _, group := range v.Groups {
		svs = append(svs, group)
	}
	svs = append(svs, v.Cols...)

	// Get the data output of all those svs
	return groupColOutput(svs, func(sv StateViewer) []string {
		return sv.GetData(sr)
	})
}

// Live views output the current timestamp whereas Runtime views output the delta in Uptime since the first state
func (v View) SetLive() {

}

func (v View) SetRuntime() {

}
