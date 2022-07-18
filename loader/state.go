package loader

import "fmt"

// The current State of the monitored server
type State struct {
	// The current and most recent SampleSets
	Current, Previous *SampleSet

	// Uptime of the server from the SampleSet
	Uptime int64
}

func NewState() *State {
	return &State{}
}

// Seconds between Cur and Prev samples for the given SourceName, return 0 if Source not found, there is no Prev sample, or other error
func (sp *State) SecondsDiff(sn SourceName) float64 {
	var curr, prev float64
	if sp.Current != nil {
		curr = sp.Current.GetSecondsComparable(sn)
	}

	// No prev sample, this is the first.
	if sp.Previous == nil {
		return 0
	}
	return curr - prev
}

// Get what to print in the timestamp col
func (sp *State) GetTimeString() string {
	return fmt.Sprintf(`%ds`, sp.Uptime)
}

// Get the Current and Previous Samplesets, could be nil!
func (sp *State) GetCurrent() SampleSetReader {
	if sp.Current == nil {
		return nil
	}
	return sp.Current
}
func (sp *State) GetPrevious() SampleSetReader {
	if sp.Previous == nil {
		return nil
	}
	return sp.Previous
}

// Set the Current and Previous Samplesets
func (sp *State) SetCurrent(ssr *SampleSet) {
	sp.Current = ssr
}
func (sp *State) SetPrevious(ssr *SampleSet) {
	sp.Previous = ssr
}
