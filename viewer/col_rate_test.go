package viewer

import (
	"testing"

	"github.com/jayjanssen/myq-tools2/loader"
)

func getTestRateCol() RateCol {
	sk := loader.SourceKey{"status", "connections"}
	rc := RateCol{}
	rc.Name = "test"
	rc.key = sk
	rc.Length = 4
	rc.Units = NUMBER
	rc.Precision = 0

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
