package loader

// The current State of the monitored server
type State struct {
	// The current and most recent SampleSets
	Current, Previous SampleSetReader

	// Uptime of the server from the SampleSet
	Uptime int64
}

func newState() *State {
	return &State{}
}

// Seconds between Cur and Prev samples for the given SourceName, return 0 if Source not found, or other error
func (sp *State) SecondsDiff(sn SourceName) float64 {
	var curr, prev float64
	if sp.Current == nil {
		curr = sp.Current.GetSecondsComparable(sn)
	}
	if sp.Previous == nil {
		prev = sp.Previous.GetSecondsComparable(sn)
	}
	return curr - prev
}

// Get the Current and Previous Samplesets, could be nil!
func (sp *State) GetCurrent() SampleSetReader {
	return sp.Current
}
func (sp *State) GetPrevious() SampleSetReader {
	return sp.Previous
}

// Set the Current and Previous Samplesets
func (sp *State) SetCurrent(ssr SampleSetReader) {
	sp.Current = ssr
}
func (sp *State) SetPrevious(ssr SampleSetReader) {
	sp.Previous = ssr
}
