package viewer

import (
	"testing"

	"github.com/jayjanssen/myq-tools2/loader"
)

// Create a state reader to test with
func getTestSampleTimeState() loader.StateReader {
	sp := loader.NewState()

	cursamp := loader.NewSample()
	sp.GetCurrentWriter().SetSample(`status`, cursamp)

	cursamp.Data[`threads_connect`] = "10"

	return sp
}

func TestSampleTimeColGetHeader(t *testing.T) {
	tc := NewSampleTimeCol()
	sr := getTestSampleTimeState()

	h := tc.GetHeader(sr)
	if len(h) != 1 {
		t.Errorf(`got wrong number of header lines: %d`, len(h))
	}

	if h[0] != `    time` {
		t.Errorf(`got wrong time header: '%s'`, h[0])
	}
}

func TestTimeColGetData(t *testing.T) {
	tc := NewSampleTimeCol()
	sr := getTestSampleTimeState()

	h := tc.GetData(sr)
	if len(h) != 1 {
		t.Errorf(`got wrong number of data lines: %d`, len(h))
	}

	if h[0] != `      0s` {
		t.Errorf(`got wrong time data: '%s'`, h[0])
	}
}
