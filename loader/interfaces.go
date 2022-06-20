package loader

import "time"

// Loads data from somewhere to produce samples
type Loader interface {

	// Setup the loader to load the given schema(s) and error-check
	Initialize(interval time.Duration, sources []Source) error

	// Produces a state for every interval.  The state.Prev is maintained automatically
	GetStateChannel() <-chan *State
}

// Functions to read a State
type StateReader interface {
}
