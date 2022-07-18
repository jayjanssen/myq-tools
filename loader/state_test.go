package loader

import "testing"

func TestStateSetPrev(t *testing.T) {
	state := NewState()

	var ssp *SampleSet
	state.SetPrevious(ssp)

	prev := state.GetPrevious()
	if prev != nil {
		t.Errorf(`expected prev to be nil: %v`, prev)
	}
}

func TestSecondsDiff(t *testing.T) {
	state := NewState()
	diff := state.SecondsDiff(`foo`)
	if diff != 0 {
		t.Errorf(`bad diff: %f`, diff)
	}
}
