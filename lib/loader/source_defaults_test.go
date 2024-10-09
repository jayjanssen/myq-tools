package loader

import (
	"testing"
)

func TestSourceParse(t *testing.T) {
	err := LoadDefaultSources()
	if err != nil {
		t.Error(err)
	}
	if len(sources) < 1 {
		t.Error("No Sources parsed!")
	}

	// Check the status Source
	status := sources[0]
	if status.Name != "status" {
		t.Error("First view is not named `status`")
	}
}

func TestGetSource(t *testing.T) {
	source, err := GetSource("status")
	if err != nil {
		t.Error(err)
	}
	if source.Name != "status" {
		t.Errorf("Unexpected status name: %s", source.Name)
	}
	if source.Description != "MySQL server global status counters" {
		t.Errorf("Unexpected status description: %s", source.Description)
	}
}

func TestGetSourceErr(t *testing.T) {
	_, err := GetSource("fooey")
	if err == nil {
		t.Error("Expected error!")
	}
}
