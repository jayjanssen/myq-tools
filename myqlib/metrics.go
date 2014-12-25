package myqlib

// MyqSamples are K->V maps
type MyqSample map[string]interface{}

// Number of keys in the sample
func (s MyqSample) Length() int {
  return len(s)
}
