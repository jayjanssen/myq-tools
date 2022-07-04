package loader

import (
	"github.com/hashicorp/go-multierror"
)

// A collection of Samples at a given time
type SampleSet struct {
	// The samples collected
	Samples map[SourceKey]*Sample
}

func newSampleSet() *SampleSet {
	ss := SampleSet{}
	ss.Samples = make(map[SourceKey]*Sample)
	return &ss
}

func (ssp *SampleSet) SetSample(key SourceKey, sp *Sample) {
	ssp.Samples[key] = sp
}

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
