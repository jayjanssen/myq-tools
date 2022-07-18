package loader

import "time"

// Loads data from somewhere to produce samples
type Loader interface {

	// Setup the loader to load the given schema(s) and error-check
	Initialize(interval time.Duration, sources []SourceName) error

	// Produces a state for every interval.  The state.Prev is maintained automatically
	GetStateChannel() <-chan StateReader
}

// Functions to read a State
type StateReader interface {
	// Seconds between Cur and Prev samples for the given SourceName
	SecondsDiff(SourceName) float64

	// Get what to print in the timestamp col
	GetTimeString() string

	// Get the Current and Previous Samplesets, could be nil!
	GetCurrent() SampleSetReader
	GetPrevious() SampleSetReader
}

type StateWriter interface {
	SetCurrent(SampleSetReader)
	SetPrevious(SampleSetReader)
}

// Functions to read a SampleSet
type SampleSetReader interface {
	// Check if the given Source is in this set
	HasSource(SourceName) bool

	// Collect errors from all the Samples
	GetErrors() error

	// Get time data from a Sample, or "nil-value"
	GetSecondsComparable(SourceName) float64
	GetTimeGenerated(SourceName) time.Time

	// Fetch the given SourceKey and parse it into the given type
	GetString(SourceKey) (string, error)
	GetInt(SourceKey) (int64, error)
	GetFloat(SourceKey) (float64, error)

	// Same as above, just ignore the error
	GetI(SourceKey) int64
	GetF(SourceKey) float64
	GetStr(SourceKey) string

	// Gets either a float or an int (check type of result), or an error
	GetNumeric(SourceKey) (interface{}, error)
}

type SampleSetWriter interface {
	SetSample(key SourceName, s SampleReader)
}

// Functions to read a Sample
type SampleReader interface {
	// A number representing seconds that can be compared (subtracted) from other Samples from the same target/source.
	// This could be Unix seconds (since 1970), seconds since the mysql server started, or some other basis entirely
	GetSecondsComparable() float64

	// Timestamp when the Sample was parsed
	GetTimeGenerated() time.Time

	// Number of keys in the Sample
	Length() int

	// Get the String value of a given key
	GetString(key string) (string, error)

	// Get the error from this Sample collection, if any
	Error() error
}
