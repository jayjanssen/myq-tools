package viewer

import (
	"testing"

	"github.com/jayjanssen/myq-tools/lib/loader"
)

func getTestGroupCol() GroupCol {
	gc := GroupCol{}
	gc.Name = "Connects"
	gc.Description = "Connection related metrics"
	gc.Type = "Group"

	gc.Cols = make(StateViewerList, 2)
	gc.Cols[0] = getTestRateCol()
	gc.Cols[1] = getTestGaugeCol()

	return gc
}

func TestGroupColImplementsStateViewer(t *testing.T) {
	gc := getTestGroupCol()
	var _ StateViewer = gc
}

// Create a state reader to test with
func getTestGroupState() loader.StateReader {
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

func TestGroupColGetHeader(t *testing.T) {
	gc := getTestGroupCol()
	sr := getTestGroupState()

	lines := gc.GetHeader(sr)
	if len(lines) != 2 {
		t.Errorf(`unexpected # of lines: %d`, len(lines))
	}

	if lines[0] != `Connects ` {
		t.Errorf(`unexpected header first line output: '%s'`, lines[0])
	}
	if lines[1] != `cons conn` {
		t.Errorf(`unexpected header second line output: '%s'`, lines[1])
	}

}

func TestGroupColGetData(t *testing.T) {
	gc := getTestGroupCol()
	sr := getTestGroupState()

	lines := gc.GetData(sr)
	if len(lines) != 1 {
		t.Errorf(`unexpected # of lines: %d`, len(lines))
	}

	if lines[0] != `   5    4` {
		t.Errorf(`unexpected GetData output: '%s'`, lines[0])
	}
}
