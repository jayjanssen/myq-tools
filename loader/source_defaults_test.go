package loader

import (
	"testing"
)

func TestSourceParse(t *testing.T) {
	err := LoadDefaultSources()
	if err != nil {
		t.Error(err)
	}
	if len(Sources) < 1 {
		t.Error("No Sources parsed!")
	}

	// Check the status Source
	status := Sources[0]
	if status.Name != "status" {
		t.Error("First view is not named `status`")
	}
}
