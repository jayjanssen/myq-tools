package loader

// The current State of the monitored server
type State struct {
	// The current and most recent SampleSets
	Current, Previous *SampleSet

	// Uptime of the server from the SampleSet
	Uptime int64
}

func newState() *State {
	return &State{}
}
