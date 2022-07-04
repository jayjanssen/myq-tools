package loader

import (
	"errors"
	"fmt"
	"strconv"
	"time"
)

// The values for a Source for a specifc time
type Sample struct {
	// Timestamp when the SampleSet was generated, in case of a File loader, this could be the same (or very close) in every set.
	Timestamp time.Time

	// The sample map --
	Data map[string]string

	// Any errors from trying to collect this sample
	Error error
}

func NewSample() *Sample {
	s := new(Sample)
	s.Data = make(map[string]string)
	s.Timestamp = time.Now()
	s.Error = nil
	return s
}

func NewSampleErr(err error) *Sample {
	s := new(Sample)
	s.Error = err
	s.Timestamp = time.Now()
	return s
}

// Number of keys in the Sample
func (s Sample) Length() int {
	return len(s.Data)
}

// Get methods for the given key. Returns a value of the appropriate type (error is nil) or default value and an error if it can't parse
func (s Sample) GetString(key string) (string, error) {
	val, ok := s.Data[key]
	if !ok {
		return "", errors.New("Key not found")
	}
	return val, nil // no errors possible here
}
func (s Sample) GetInt(key string) (int64, error) {
	val, ok := s.Data[key]
	if !ok {
		return 0, errors.New("Key not found")
	}

	conv, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return 0, err
	}
	return conv, nil
}
func (s Sample) GetFloat(key string) (float64, error) {
	val, ok := s.Data[key]
	if !ok {
		return 0.0, errors.New("Key not found")
	}

	conv, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return 0.0, err
	} else {
		return conv, nil
	}
}

// Same as above, just ignore the error
func (s Sample) GetI(key string) int64 {
	i, _ := s.GetInt(key)
	return i
}
func (s Sample) GetF(key string) float64 {
	f, _ := s.GetFloat(key)
	return f
}
func (s Sample) GetStr(key string) string {
	str, _ := s.GetString(key)
	return str
}

// Gets either a float or an int (check type of result), or an error
func (s Sample) GetNumeric(key string) (interface{}, error) {
	// Ints can be parsed as a Float, but not the converse, try Int parsing first
	if val, err := s.GetInt(key); err == nil {
		return val, nil
	} else if val, err := s.GetFloat(key); err == nil {
		return val, nil
	} else {
		return nil, fmt.Errorf("Value is not numeric: `%v`", s.GetStr(key))
	}
}
