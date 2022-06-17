package viewer

import (
	"github.com/jayjanssen/myq-tools2/loader"
)

type Viewer interface {
	GetShortHelp() string
	GetHelp() []string

	GetHeader(state *loader.MyqState) []string
	GetData(state *loader.MyqState) []string

	SetTimeCol(timecol *Col)
}
