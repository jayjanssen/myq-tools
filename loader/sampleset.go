package loader

import (
	"fmt"
	"strconv"

	"github.com/hashicorp/go-multierror"
)

// A collection of Samples at a given time
type SampleSet struct {
	// The samples collected
	Samples map[SourceName]*Sample
}

// Create new SampleSet
func NewSampleSet() *SampleSet {
	ss := SampleSet{}
	ss.Samples = make(map[SourceName]*Sample)
	return &ss
}

// Store a sample for the given key into this set
func (ssp *SampleSet) SetSample(key SourceName, sp *Sample) {
	ssp.Samples[key] = sp
}

// Collect errors from all the Samples
func (ssp *SampleSet) GetErrors() error {
	var errs *multierror.Error
	for _, sample := range ssp.Samples {
		if sample == nil {
			continue
		}
		if sample.Error != nil {
			errs = multierror.Append(errs, sample.Error)
		}
	}
	return errs.ErrorOrNil()
}

// Fetch the string value of the the given SourceKey
func (ssp *SampleSet) GetString(sk SourceKey) (string, error) {
	sp, ok := ssp.Samples[sk.SourceName]
	if !ok {
		return "", fmt.Errorf("source (%s) not found", sk.SourceName)
	}

	return sp.GetString(sk.Key)
}

func (ssp *SampleSet) GetInt(sk SourceKey) (int64, error) {
	val, err := ssp.GetString(sk)
	if err != nil {
		return 0, err
	}

	conv, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return 0, err
	}
	return conv, nil
}
func (ssp *SampleSet) GetFloat(sk SourceKey) (float64, error) {
	val, err := ssp.GetString(sk)
	if err != nil {
		return 0, err
	}

	conv, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return 0.0, err
	} else {
		return conv, nil
	}
}

// Same as above, just ignore the error
func (ssp *SampleSet) GetI(sk SourceKey) int64 {
	i, _ := ssp.GetInt(sk)
	return i
}
func (ssp *SampleSet) GetF(sk SourceKey) float64 {
	f, _ := ssp.GetFloat(sk)
	return f
}
func (ssp *SampleSet) GetStr(sk SourceKey) string {
	str, _ := ssp.GetString(sk)
	return str
}

// Gets either a float or an int (check type of result), or an error
func (ssp *SampleSet) GetNumeric(sk SourceKey) (interface{}, error) {
	// Ints can be parsed as a Float, but not the converse, try Int parsing first
	if val, err := ssp.GetInt(sk); err == nil {
		return val, nil
	} else if val, err := ssp.GetFloat(sk); err == nil {
		return val, nil
	} else {
		return nil, fmt.Errorf("value is not numeric: `%v`", ssp.GetStr(sk))
	}
}
