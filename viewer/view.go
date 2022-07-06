package viewer

import "github.com/jayjanssen/myq-tools2/loader"

// A view is made up of Groups of Cols
type View struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`

	// Usually a view would have Groups OR Cols, but not both.  If both, print groups first, then individual cols
	Groups []Colgroup      `yaml:"groups"`
	Cols   StateViewerList `yaml:"cols"`
}

func (v View) GetName() string {
	return v.Name
}

// Single line help for the view
func (v View) GetShortHelp() string {
	return ""
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
func (v View) GetHeader(sr loader.StateReader) []string {
	return []string{""}
}

// Data for this view based on the state
func (v View) GetData(sr loader.StateReader) []string {
	return []string{""}
}

// Live views output the current timestamp whereas Runtime views output the delta in Uptime since the first state
func (v View) SetLive() {

}

func (v View) SetRuntime() {

}
