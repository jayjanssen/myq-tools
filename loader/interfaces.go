package loader

import "time"

// Loads data from somewhere to produce samples
type Loader interface {

	// Setup the loader to load the given schema(s) and error-check
	Initialize(interval time.Duration, sources []SourceName) error

	// Produces a state for every interval.  The state.Prev is maintained automatically
	GetStateChannel() <-chan *State
}

// Functions to read a State
type StateReader interface {
	// Number of seconds between Cur and Prev state
	SecondsDiff() int64

	// Cur and Prev value of a given SourceKey
	GetKeyCurPrev(SourceKey) (cur, prev float64)
}
