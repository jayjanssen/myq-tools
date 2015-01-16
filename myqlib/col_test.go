package myqlib

import (
	"bytes"
	"testing"
)

func TestIntCol(t *testing.T) {
	var b bytes.Buffer
	col := GaugeCol{DefaultCol{"run", "Threads running", 5}, "threads_running", 0, NumberUnits}

	state := MyqState{}
	state.Cur = make(MyqSample)
	state.Cur["threads_running"] = "10"

	col.Data(&b, &state)
	if b.String() != "   10" {
		t.Fatal("Bad output", b.String())
	}
	b.Reset()
}

func TestFloatCol(t *testing.T) {
	var b bytes.Buffer
	col := GaugeCol{DefaultCol{"oooe", "Galera OOO E", 5}, "wsrep_apply_oooe", 3, NumberUnits}

	state := MyqState{}
	state.Cur = make(MyqSample)
	state.Cur["wsrep_apply_oooe"] = "0.015600"

	col.Data(&b, &state)
	if b.String() != "0.016" {
		t.Fatal("Bad output", b.String())
	}
	b.Reset()
}

func TestDiff(t *testing.T) {
	d2 := calculate_rate(1.750, 1.500, 5)
	if d2 != 0.05 {
		t.Error("d2 is ", d2)
	}

	// cur < prev
	d3 := calculate_rate(10, 200, 2)
	if d3 != 5.0 {
		t.Error("d3 is", d3)
	}

	// time <= 0
	d4 := calculate_rate(20, 10, -10)
	if d4 != 10.0 {
		t.Error("d4 is", d4)
	}
}

func TestRateCol(t *testing.T) {
	var b bytes.Buffer
	col := RateCol{DefaultCol{"cons", "Connections per second", 5}, "connections", 0, NumberUnits}

	state := MyqState{}
	state.Cur = make(MyqSample)
	state.Cur["connections"] = "10"
	state.Cur["uptime"] = "1"

	state.Prev = make(MyqSample)
	state.Prev["connections"] = "20"
	state.Prev["uptime"] = "5"

	state.SecondsDiff = 5.0

	col.Data(&b, &state)
	if b.String() != "    2" {
		t.Fatal("Bad output", b.String(), `.`)
	}
	b.Reset()
}

// implement large number collapsing first
// state.Cur["threads_running"] = "100000"
// col.Data( &b, state )
// if b.String() != "10.0k" {
//   t.Fatal( "Bad output", b.String())
// }
