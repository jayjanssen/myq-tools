package loader

import (
	"fmt"
)

// The current State of the monitored server
type State struct {
	// The current and most recent SampleSets
	Current, Previous *SampleSet

	// Is this a Live state?
	Live bool
}

func NewState() *State {
	sp := &State{}
	sp.Current = NewSampleSet()
	return sp
}

// Seconds between Cur and Prev samples for the given SourceName, return 0 if Source not found, there is no Prev sample, or other error
func (sp *State) SecondsDiff(sn SourceName) float64 {
	// No prev sample, this is the first.
	if sp.Previous == nil {
		return 0
	}

	// Live state
	if sp.Live {
		curTime := sp.GetCurrent().GetTimeGenerated()
		prevTime := sp.GetPrevious().GetTimeGenerated()
		diff := curTime.Sub(prevTime)
		return diff.Seconds()
	}

	// File loader state
	curUptime := sp.GetCurrent().GetUptime()
	prevUptime := sp.GetCurrent().GetUptime()
	return float64(curUptime - prevUptime)
}

// Get what to print in the timestamp col
func (sp *State) GetTimeString() string {
	if sp.Live {
		return sp.GetCurrent().GetTimeGenerated().Format(`15:04:05`)
	} else {
		return fmt.Sprintf(`%ds`, sp.GetCurrent().GetUptime())
	}
}

// Get the Current and Previous Samplesets, could be nil!
func (sp *State) GetCurrent() SampleSetReader {
	return sp.Current
}
func (sp *State) GetPrevious() SampleSetReader {
	if sp.Previous == nil {
		return nil
	}
	return sp.Previous
}

// Set Previous Samplesets
func (sp *State) SetPrevious(ssr *SampleSet) {
	sp.Previous = ssr
}

// Get the interface to write to the Current SS
func (sp *State) GetCurrentWriter() SampleSetWriter {
	return sp.Current
}
