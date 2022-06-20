package viewer

import (
	"github.com/jayjanssen/myq-tools2/loader"
)

// A Viewer
type Viewer interface {
	// Single line help for the view
	GetShortHelp() string

	// Detailed multi-line help for the view
	GetHelp() []string

	// A list of sources that this view requires
	GetSources() []loader.Source

	// Header for this view, unclear if state is needed
	GetHeader(state *loader.State) []string

	// Data for this view based on the state
	GetData(state *loader.State) []string

	// Live views output the current timestamp whereas Runtime views output the delta in Uptime since the first state
	SetLive()
	SetRuntime()
}
