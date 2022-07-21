package loader

import (
	"math"
	"testing"
	"time"
)

func TestStateSetPrev(t *testing.T) {
	state := NewState()

	var ssp *SampleSet
	state.SetPrevious(ssp)

	prev := state.GetPrevious()
	if prev != nil {
		t.Errorf(`expected prev to be nil: %v`, prev)
	}
}

func TestStateSecondsDiff(t *testing.T) {
	state := NewState()
	diff := state.SecondsDiff()
	if diff != 0 {
		t.Errorf(`bad diff: %f`, diff)
	}

	prevssp := NewSampleSet()
	prevssp.SetUptime(10)
	state.SetPrevious(prevssp)
	state.GetCurrentWriter().SetUptime(15)
	diff = state.SecondsDiff()
	if diff == 5 {
		t.Errorf(`bad diff: %f`, diff)
	}

	state.Live = true
	// -0.0 from time.Sub
	diff = state.SecondsDiff()
	if diff != 0 {
		t.Errorf(`bad diff: %f`, diff)
	}

	prevssp.Timestamp = time.Now().Add(time.Second * -10)
	diff = state.SecondsDiff()
	if math.Round(diff) != 10 {
		t.Errorf(`bad diff: %f`, diff)
	}
}

func TestStateGetTimeString(t *testing.T) {
	state := NewState()
	state.GetCurrentWriter().SetUptime(15)

	// non-live
	ts := state.GetTimeString()
	if ts != "15s" {
		t.Errorf("bad timestring: %s", ts)
	}

	// live
	state.Live = true
	ts = state.GetTimeString()
	if len(ts) != 8 {
		t.Errorf("bad timestring: %s", ts)
	}
}
