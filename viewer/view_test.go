package viewer

import (
	"fmt"
	"testing"

	"github.com/jayjanssen/myq-tools2/loader"
)

func getTestView() View {
	view := View{}
	view.Name = "Test View"
	view.Description = "My Test View"
	view.Groups = make([]GroupCol, 1)
	view.Groups[0] = getTestGroupCol()

	return view
}

func TestViewImplementsStateViewer(t *testing.T) {
	view := getTestView()
	var _ StateViewer = view
}

// Create a state reader to test with
func getTestViewState() loader.StateReader {
	sp := loader.NewState()
	prevss := loader.NewSampleSet()

	cursamp := loader.NewSample()
	sp.GetCurrentWriter().SetSample(`status`, cursamp)

	prevsamp := loader.NewSample()
	prevss.SetSample(`status`, prevsamp)
	sp.SetPrevious(prevss)

	cursamp.Data[`connections`] = `15`
	prevsamp.Data[`connections`] = `10`

	cursamp.Data[`threads_connect`] = `4`
	prevsamp.Data[`threads_connect`] = `3`

	return sp
}

func TestViewGetDetailedHelp(t *testing.T) {
	view := getTestView()
	helpStrs := view.GetDetailedHelp()

	if len(helpStrs) != 3 {
		fmt.Printf(`unexpected GetDetailedHelp length: %d`, len(helpStrs))
	}
}

func TestViewGetSources(t *testing.T) {
	view := getTestView()
	sources := view.GetSources()

	if len(sources) != 1 {
		fmt.Printf(`unexpected GetSources length: %d`, len(sources))
	}
}

func TestViewGetHeader(t *testing.T) {
	view := getTestView()
	sr := getTestGroupState()

	lines := view.GetHeader(sr)

	expectedLines := []string{
		`         Connects `,
		`    time cons conn`,
	}

	if len(lines) != len(expectedLines) {
		t.Errorf(`unexpected # of lines: %d`, len(lines))
	}
	for i, expected := range expectedLines {
		if lines[i] != expected {
			t.Errorf(`unexpected line %d output: '%s'`, i, lines[i])
		}
	}
}

func TestViewGetData(t *testing.T) {
	view := getTestView()
	sr := getTestViewState()

	lines := view.GetData(sr)

	expectedLines := []string{
		`      0s    5    4`,
	}

	if len(lines) != len(expectedLines) {
		t.Errorf(`unexpected # of lines: %d`, len(lines))
	}
	for i, expected := range expectedLines {
		if lines[i] != expected {
			t.Errorf(`unexpected line %d output: '%s'`, i, lines[i])
		}
	}
}
