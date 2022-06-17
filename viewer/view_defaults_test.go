package viewer

import (
	"testing"
)

func TestDefsParse(t *testing.T) {
	err := LoadDefaultViews()
	if err != nil {
		t.Error(err)
	}
	if len(Views) < 1 {
		t.Error("No views parsed!")
	}
	cttf := Views[0]
	if cttf.Name != "cttf" {
		t.Error("First view is not named `cttf`!")
	}
}
