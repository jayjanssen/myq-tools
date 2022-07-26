package viewer

import (
	"testing"

	"github.com/jayjanssen/myq-tools2/loader"
)

func getTestSortedExpandedCountsCol() SortedExpandedCountsCol {
	sk := loader.SourceKey{SourceName: "status", Key: "com_set.*"}
	rc := SortedExpandedCountsCol{}
	rc.Name = "set"
	rc.Description = "SET commands per second"
	rc.Type = "RateSum"
	rc.Keys = []loader.SourceKey{sk}
	rc.Length = 5
	rc.Units = NUMBER
	rc.Precision = 0

	return rc
}

func TestSortedExpandedCountsColGetData(t *testing.T) {
	col := getTestSortedExpandedCountsCol()
	sr := getTestRateSumState(`100`, `1000`)

	output := col.GetData(sr)
	if len(output) != 1 {
		t.Fatalf(`unexpected output lines: %d`, len(output))
	}

	if output[0] != `       900 [com_set_option com_set_password com_set_resource_group]` {
		t.Errorf(`unexpected output: '%s'`, output[0])
	}
}
