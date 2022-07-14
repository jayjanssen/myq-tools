package viewer

import (
	"reflect"
	"testing"

	"github.com/jayjanssen/myq-tools2/loader"
)

func TestDefsParse(t *testing.T) {
	err := LoadDefaultViews()
	if err != nil {
		t.Error(err)
	}
	if len(Views) < 1 {
		t.Fatal("No views parsed!")
	}
	name := ViewNames[0]
	if name != "cttf" {
		t.Fatal("First view is not named `cttf`!")
	}

	cttf, ok := Views[name]
	if !ok {
		t.Fatal("Could not get `cttf` view")
	}

	if len(cttf.Groups) == 0 {
		t.Fatal("No groups parsed for `cttf`")
	}

	group := cttf.Groups[0]
	if group.Name != "Connects" {
		t.Fatal("First cttf group not  Connects")
	}

	if len(group.Cols) == 0 {
		t.Fatal("No cols parsed for `Connects` group")
	}
	cons := group.Cols[0]
	if cons.GetName() != "cons" {
		t.Fatalf("First Connects column is not cons: %s", cons.GetName())
	}

	mycons := RateCol{}
	mycons.Name = "cons"
	mycons.Description = "Connections per second"
	mycons.Key = loader.SourceKey{`status`, `connections`}
	mycons.Type = "Rate"
	mycons.Units = NUMBER
	mycons.Length = 4
	mycons.Precision = 0

	if !reflect.DeepEqual(cons, mycons) {
		t.Error("cons not matching!")
		t.Logf("got: %+v", cons)
		t.Logf("expected: %+v", mycons)
	}

}
