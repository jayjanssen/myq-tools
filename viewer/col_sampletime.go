package viewer

import "github.com/jayjanssen/myq-tools2/loader"

type SampleTimeCol struct {
	defaultCol
}

func NewSampleTimeCol() SampleTimeCol {
	tc := SampleTimeCol{}
	tc.Name = "time"
	tc.Length = 8

	return tc
}

func (c SampleTimeCol) GetSources() []loader.SourceName {
	return []loader.SourceName{}
}

// Asks the StateReader for what time to print
func (c SampleTimeCol) GetData(sr loader.StateReader) []string {
	return []string{FitString(sr.GetTimeString(), c.Length)}
}
