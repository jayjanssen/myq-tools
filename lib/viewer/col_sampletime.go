package viewer

import myblip "github.com/jayjanssen/myq-tools/lib/blip"

type SampleTimeCol struct {
	defaultCol
}

func NewSampleTimeCol() SampleTimeCol {
	tc := SampleTimeCol{}
	tc.Name = "time"
	tc.Length = 8

	return tc
}

// Asks the MetricCache for what time to print
func (c SampleTimeCol) GetData(cache *myblip.MetricCache) []string {
	return []string{FitString(cache.GetTimeString(), c.Length)}
}
