package loader

import (
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/hashicorp/go-multierror"
)

// A collection of Samples at a given time
type SampleSet struct {
	// The samples collected
	Samples map[SourceName]SampleReader

	// Timestamp when the Sample was generated, in case of a File loader, this could be the same (or very close) in every set.
	Timestamp time.Time

	// Uptime of the server relative to start of loading
	Uptime int64
}

// Create new SampleSet
func NewSampleSet() *SampleSet {
	ss := SampleSet{}
	ss.Timestamp = time.Now()
	ss.Samples = make(map[SourceName]SampleReader)
	return &ss
}

// Store a sample for the given key into this set
func (ssp *SampleSet) SetSample(key SourceName, sp SampleReader) {
	ssp.Samples[key] = sp
}

// Store the uptime for the sample
func (ssp *SampleSet) SetUptime(u int64) {
	ssp.Uptime = u
}

// Check if the given Source is in this set
func (ssp *SampleSet) HasSource(sn SourceName) bool {
	_, ok := ssp.Samples[sn]
	return ok
}

// Collect errors from all the Samples
func (ssp *SampleSet) GetErrors() error {
	var errs *multierror.Error
	for _, sample := range ssp.Samples {
		if sample == nil {
			continue
		}
		if sample.Error() != nil {
			errs = multierror.Append(errs, sample.Error())
		}
	}
	return errs.ErrorOrNil()
}

// Get time data this Set was generated
func (ssp *SampleSet) GetTimeGenerated() time.Time {
	return ssp.Timestamp
}
func (ssp *SampleSet) GetUptime() int64 {
	return ssp.Uptime
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
		return 0.0, err
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

// Get a Sum of a series of SourceKeys
func (ssp *SampleSet) GetFloatSum(sks []SourceKey) float64 {
	var total float64
	for _, sk := range sks {
		total += ssp.GetF(sk)
	}
	return total
}

// Takes a list of SourceKeys where the .Key might contain a regex
func (ssp *SampleSet) ExpandSourceKeys(sks []SourceKey) (results []SourceKey) {
	// Go through every input in sks
	for _, sk := range sks {
		re, err := regexp.Compile(sk.Key)
		// Not a regex?
		if err != nil {
			results = append(results, sk)
			continue
		}

		// Get the sample
		sp, ok := ssp.Samples[sk.SourceName]
		// no such source?
		if !ok {
			continue
		}

		// Check our regex against every key in the sample
		for _, key := range sp.GetKeys() {
			if re.MatchString(key) {
				results = append(results, SourceKey{sk.SourceName, key})
			}
		}
	}

	return
}
