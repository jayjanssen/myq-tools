package viewer

import (
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
	curss := loader.NewSampleSet()
	prevss := loader.NewSampleSet()

	cursamp := loader.NewSample()
	curss.SetSample(`status`, cursamp)

	prevsamp := loader.NewSample()
	prevss.SetSample(`status`, prevsamp)

	sp.SetCurrent(curss)
	sp.SetPrevious(prevss)

	cursamp.Data[`connections`] = `15`
	prevsamp.Data[`connections`] = `10`

	cursamp.Data[`threads_connect`] = `4`
	prevsamp.Data[`threads_connect`] = `3`

	return sp
}

func TestViewGetHeader(t *testing.T) {
	view := getTestView()
	sr := getTestGroupState()

	lines := view.GetHeader(sr)

	expectedLines := []string{
		`Test View`,
		`Connects `,
		`cons conn`,
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
	sr := getTestGroupState()

	lines := view.GetData(sr)

	expectedLines := []string{
		`   5    4`,
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
