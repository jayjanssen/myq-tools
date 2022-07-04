package loader

import (
	"reflect"
	"testing"
)

// Sample

func newSample() *Sample {
	sample := NewSample()
	sample.Data["string"] = "String"
	sample.Data["int"] = "10"
	sample.Data["float"] = "1.4256"
	return sample
}

// Length works
func TestSampleLength(t *testing.T) {
	sample := newSample()
	if sample.Length() != 3 {
		t.Error("Expecting 3 KV, got", sample.Length())
	}
}

// Get functions return err on missing key
func TestSampleGetMissingKey(t *testing.T) {
	sample := newSample()
	_, err := sample.GetString("what key?")
	if err == nil {
		t.Error("No error on missing key")
	}
}

// Type conversions
func TestSampleGetConversions(t *testing.T) {
	sample := newSample()
	var v interface{}
	var err error
	var ok bool

	// int64
	v, err = sample.GetInt(`int`)
	if err != nil {
		t.Error(err)
	}
	_, ok = v.(int64)
	if !ok {
		t.Errorf("Expected int64, got %s", reflect.TypeOf(v))
	}

	// float64
	v, err = sample.GetFloat(`float`)
	if err != nil {
		t.Error(err)
	}
	_, ok = v.(float64)
	if !ok {
		t.Errorf("Expected float64, got %s", reflect.TypeOf(v))
	}

	// string
	v, err = sample.GetString(`string`)
	if err != nil {
		t.Error(err)
	}
	_, ok = v.(string)
	if !ok {
		t.Errorf("Expected string, got %s", reflect.TypeOf(v))
	}
}

// Type error handling
func TestSampleGetErrors(t *testing.T) {
	sample := newSample()
	var err error

	// String should throw errors to GetFloat and GetInt
	_, err = sample.GetFloat(`string`)
	if err == nil {
		t.Error("Missing error")
	}
	_, err = sample.GetInt(`string`)
	if err == nil {
		t.Error("Missing error")
	}

	// String should return default values for GetF and GetI
	f := sample.GetF(`string`)
	if f != 0.0 {
		t.Error("Mishandled error GetF")
	}
	i := sample.GetI(`string`)
	if i != 0 {
		t.Error("Mishandled error GetI")
	}

	// GetInt should not parse a float
	_, err = sample.GetInt(`float`)
	if err == nil {
		t.Error("Missing error")
	}

	// GetFloat should parse an int
	f, err = sample.GetFloat(`int`)
	if err != nil {
		t.Errorf("Can't parse int to float: %v", f)
	}
}

// GetNumeric
func TestSampleGetNumeric(t *testing.T) {
	sample := newSample()

	i, err := sample.GetNumeric(`int`)
	if err != nil {
		t.Errorf("Could not parse int as Numeric: %v", err)
	} else if i != int64(10) {
		t.Errorf("int does not equal 10: %v", i)
	}

	f, err := sample.GetNumeric(`float`)
	if err != nil {
		t.Errorf("Could not parse float as Numeric: %v", err)
	}
	if f != float64(1.4256) {
		t.Errorf("float does not equal 1.4256: %f", f)
	}

}
