package loader

import (
	"errors"
	"time"
)

// The values for a Source for a specifc time
type Sample struct {
	// Timestamp when the Sample was generated, in case of a File loader, this could be the same (or very close) in every set.
	Timestamp time.Time

	// The sample map --
	Data map[string]string

	// Any errors from trying to collect this sample
	err error
}

func NewSample() *Sample {
	s := new(Sample)
	s.Data = make(map[string]string)
	s.Timestamp = time.Now()
	s.err = nil
	return s
}

func NewSampleErr(err error) *Sample {
	s := new(Sample)
	s.err = err
	s.Timestamp = time.Now()
	return s
}

func (s Sample) GetTimeGenerated() time.Time {
	return s.Timestamp
}

// Number of keys in the Sample
func (s Sample) Length() int {
	return len(s.Data)
}

// Get a list of all key strings in this stample
func (s Sample) GetKeys() (result []string) {
	for k, _ := range s.Data {
		result = append(result, k)
	}
	return
}

// Get methods for the given key. Returns a value of the appropriate type (error is nil) or default value and an error if it can't parse
func (s Sample) GetString(key string) (string, error) {
	val, ok := s.Data[key]
	if !ok {
		return "", errors.New("key not found")
	}
	return val, nil // no errors possible here
}

// Get the error from this Sample collection, if any
func (s Sample) Error() error {
	return s.err
}
