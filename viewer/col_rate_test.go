package viewer

import (
	"testing"

	"github.com/jayjanssen/myq-tools2/loader"
)

func getTestRateCol() RateCol {
	rc := RateCol{}
	rc.Name = "test"
	rc.Sources = []loader.SourceName{"status"}
	rc.Length = 4
	rc.Units = NUMBER
	rc.Precision = 0
	rc.Key = "connections"

	return rc
}

func TestRateCol(t *testing.T) {
	rc := getTestRateCol()
	if rc.Name != "test" {
		t.Errorf("Unexpected col name (test): %s", rc.Name)
	}
}

func TestRateColImplementsStateViewer(t *testing.T) {
	var _ StateViewer = getTestRateCol()
}
