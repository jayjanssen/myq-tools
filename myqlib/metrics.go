package myqlib

// MyqSamples are K->V maps
type MyqSample map[string]interface{}

// Number of keys in the sample
func (s MyqSample) Length() int {
	return len(s)
}

// MyqState contains the current and previous SHOW STATUS outputs.  Also SHOW VARIABLES.
// Prev and Vars might be nil
type MyqState struct {
	Cur, Prev, Vars MyqSample
	TimeDiff        float64
	Count           uint
}
