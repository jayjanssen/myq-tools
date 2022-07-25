package loader

import (
	"reflect"
	"testing"
)

func newTestSampleSet() *SampleSet {
	ssp := NewSampleSet()
	s := newTestSample()
	ssp.SetSample(`testing`, s)

	return ssp
}

// Get functions return err on missing key
func TestSampleGetMissingKey(t *testing.T) {
	ssp := newTestSampleSet()
	_, err := ssp.GetString(SourceKey{`testing`, `what key?`})
	if err == nil {
		t.Error("No error on missing key")
	}
}

// Type conversions
func TestSampleGetConversions(t *testing.T) {
	ssp := newTestSampleSet()
	var v interface{}
	var err error
	var ok bool

	// int64
	v, err = ssp.GetInt(SourceKey{`testing`, `int`})
	if err != nil {
		t.Error(err)
	}
	_, ok = v.(int64)
	if !ok {
		t.Errorf("Expected int64, got %s", reflect.TypeOf(v))
	}

	// int error
	_, err = ssp.GetInt(SourceKey{`testing`, `intbad`})
	if err == nil {
		t.Error(`expected error`)
	}

	// float64
	v, err = ssp.GetFloat(SourceKey{`testing`, `float`})
	if err != nil {
		t.Error(err)
	}
	_, ok = v.(float64)
	if !ok {
		t.Errorf("Expected float64, got %s", reflect.TypeOf(v))
	}

	// float error
	_, err = ssp.GetFloat(SourceKey{`testing`, `floatbad`})
	if err == nil {
		t.Error(`expected error`)
	}

	// string
	v, err = ssp.GetString(SourceKey{`testing`, `string`})
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
	ssp := newTestSampleSet()
	var err error

	// String should throw errors to GetFloat and GetInt
	_, err = ssp.GetFloat(SourceKey{`testing`, `string`})
	if err == nil {
		t.Error("Missing error")
	}
	_, err = ssp.GetInt(SourceKey{`testing`, `string`})
	if err == nil {
		t.Error("Missing error")
	}

	// String should return default values for GetF and GetI
	f := ssp.GetF(SourceKey{`testing`, `string`})
	if f != 0.0 {
		t.Error("Mishandled error GetF")
	}
	i := ssp.GetI(SourceKey{`testing`, `string`})
	if i != 0 {
		t.Error("Mishandled error GetI")
	}

	// GetInt should not parse a float
	_, err = ssp.GetInt(SourceKey{`testing`, `float`})
	if err == nil {
		t.Error("Missing error")
	}

	// GetFloat should parse an int
	f, err = ssp.GetFloat(SourceKey{`testing`, `int`})
	if err != nil {
		t.Errorf("Can't parse int to float: %v", f)
	}
}

// GetStr
func TestGetStr(t *testing.T) {
	ssp := newTestSampleSet()

	str := ssp.GetStr(SourceKey{`testing`, `string`})
	if str != "String" {
		t.Errorf("unexpected string: %s", str)
	}

	str = ssp.GetStr(SourceKey{`testing`, `int`})
	if str != "10" {
		t.Errorf("unexpected int to str: %s", str)
	}

	str = ssp.GetStr(SourceKey{`testing`, `missing`})
	if str != "" {
		t.Errorf("missing string not empty: '%s'", str)
	}
}

// GetNumeric
func TestSampleGetNumeric(t *testing.T) {
	ssp := newTestSampleSet()

	i, err := ssp.GetNumeric(SourceKey{`testing`, `int`})
	if err != nil {
		t.Errorf("Could not parse int as Numeric: %v", err)
	} else if i != int64(10) {
		t.Errorf("int does not equal 10: %v", i)
	}

	f, err := ssp.GetNumeric(SourceKey{`testing`, `float`})
	if err != nil {
		t.Errorf("Could not parse float as Numeric: %v", err)
	}
	if f != float64(1.4256) {
		t.Errorf("float does not equal 1.4256: %f", f)
	}

	_, err = ssp.GetNumeric(SourceKey{`testing`, `string`})
	if err == nil {
		t.Error("expected error GetNumeric on string")
	}
}

func TestSampleSetGetFloatSum(t *testing.T) {
	ssp := newTestSampleSet()

	tot := ssp.GetFloatSum([]SourceKey{
		{`testing`, `int`},
		{`testing`, `float`},
	})

	if tot != 11.4256 {
		t.Errorf("unxpected GetFloatSum: %f", tot)
	}
}

func TestSampleSetExpandSourceKeys(t *testing.T) {
	ssp := newTestSampleSet()

	sample := NewSample()
	sample.Data["prefix"] = "String"
	sample.Data["prefab"] = "10"
	sample.Data["something else"] = "1.4256"

	ssp.SetSample(`testing`, sample)

	expanded := ssp.ExpandSourceKeys([]SourceKey{{`testing`, `pre*`}})

	if len(expanded) != 2 {
		t.Fatalf(`unexpected amount of expanded keys: %d`, len(expanded))
	}
}

func TestSampleSetExpandSourceKeysRegex(t *testing.T) {
	ssp := newTestSampleSet()

	sample := NewSample()
	sample.Data["prefix"] = "String"
	sample.Data["prefab"] = "10"
	sample.Data["fabpre"] = "1.4256"

	ssp.SetSample(`testing`, sample)

	expanded := ssp.ExpandSourceKeys([]SourceKey{{`testing`, `^pre*`}})

	if len(expanded) != 2 {
		t.Fatalf(`unexpected amount of expanded keys: %d`, len(expanded))
	}
}
