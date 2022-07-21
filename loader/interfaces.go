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
	SecondsDiff() float64

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

	// Get Time data for the Set
	GetTimeGenerated() time.Time
	GetUptime() int64

	// Fetch the given SourceKey and parse it into the given type
	GetString(SourceKey) (string, error)
	GetInt(SourceKey) (int64, error)
	GetFloat(SourceKey) (float64, error)

	// Same as above, just ignore the error
	GetI(SourceKey) int64
	GetF(SourceKey) float64
	GetStr(SourceKey) string

	// Get a Sum of a series of SourceKeys
	GetFloatSum([]SourceKey) float64

	// Given a SourceKey list with a patterns, expand that to the full list of SourceKeys without patterns.  The result of this should be cached!
	ExpandSourceKeys([]SourceKey) []SourceKey

	// Gets either a float or an int (check type of result), or an error
	GetNumeric(SourceKey) (interface{}, error)
}

type SampleSetWriter interface {
	SetSample(key SourceName, s SampleReader)
	SetUptime(int64)
}

// Functions to read a Sample
type SampleReader interface {
	// Timestamp when the Sample was parsed
	GetTimeGenerated() time.Time

	// Number of keys in the Sample
	Length() int

	// Get a list of all key strings in this stample
	GetKeys() []string

	// Get the String value of a given key
	GetString(key string) (string, error)

	// Get the error from this Sample collection, if any
	Error() error
}
