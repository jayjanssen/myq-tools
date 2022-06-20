package loader

import (
	"time"
)

// The current and most recent SampleSets
type State struct {
	Current, Previous SampleSet
}

// A collection of Samples at a given time
type SampleSet struct {
	// Timestamp when the SampleSet was generated, in case of a File loader, this could be the same (or very close) in every set.
	Timestamp time.Time

	// Uptime of the server from the SampleSet
	Uptime int64

	// The samples collected, key is the Sample.Source.Name
	Samples map[string]Sample
}

// The values for a Source for a specifc time
type Sample struct {
	// Timestamp when the SampleSet was generated, in case of a File loader, this could be the same (or very close) in every set.
	Timestamp time.Time

	// Uptime of the server from the SampleSet
	Uptime int64

	// The source the Sample was generated from
	SampleSource *Source

	// The sample map --
	Data map[string]interface{}
}

// Source to collect a Sample
type Source struct {
	Name        string
	Description string
	// Needs some attributes that describe how to load this source, live or file
}
