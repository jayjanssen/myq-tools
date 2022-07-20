package loader

import (
	"errors"
	"testing"
)

// Sample

func newTestSample() *Sample {
	sample := NewSample()
	sample.Data["string"] = "String"
	sample.Data["int"] = "10"
	sample.Data["float"] = "1.4256"
	return sample
}

// File Loader implements the Loader interface
func TestSampleImplementsSampleReader(t *testing.T) {
	var _ SampleReader = newTestSample()
}

func TestSampleErr(t *testing.T) {
	serr := NewSampleErr(errors.New("test error"))
	if serr.Error() == nil {
		t.Error("missing error")
	}
}

// Length works
func TestSampleLength(t *testing.T) {
	sample := newTestSample()
	if sample.Length() != 3 {
		t.Error("Expecting 3 KV, got", sample.Length())
	}
}
