package viewer

import (
	"reflect"
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
	name := ViewNames[0]
	if name != "cttf" {
		t.Error("First view is not named `cttf`!")
	}

	cttf, ok := Views[name]
	if !ok {
		t.Error("Could not get `cttf` view")
	}

	if len(cttf.Groups) == 0 {
		t.Fatal("No groups parsed for `cttf`")
	}

	group := cttf.Groups[0]
	if group.Name != "Connects" {
		t.Error("First cttf group not  Connects")
	}

	if len(group.Cols) == 0 {
		t.Fatal("No cols parsed for `Connects` group")
	}
	cons := group.Cols[0]
	if cons.GetName() != "cons" {
		t.Error("First Connects column is not cons")
	}

	sources := []string{"status"}
	mycons := Col{
		Name:        "cons",
		Description: "Connections per second",
		Sources:     sources,
		Key:         "connections",
		Type:        RATE,
		Units:       NUMBER,
		Length:      4,
		Precision:   0,
	}

	if reflect.DeepEqual(cons, mycons) {
		t.Error("cons not matching!")
	}

}
