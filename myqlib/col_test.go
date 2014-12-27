package myqlib

import (
  "testing"
  "bytes"
)

func TestIntCol(t *testing.T) {
  var b bytes.Buffer
  col := NewcolDef("threads_running", "trun", "%5s", "%5d")

  state := MyqState{}
  state.Cur = make( MyqSample )
  state.Cur["threads_running"] = 10
  
  col.Output( &b, state )
  if b.String() != "   10" {
    t.Fatal( "Bad output", b.String())
  }
  b.Reset()
}

func TestFloatCol(t *testing.T) {
  var b bytes.Buffer
  col := NewcolDef("wsrep_apply_oooe", "oooe", "%5s", "%1.3f")

  state := MyqState{}
  state.Cur = make( MyqSample )
  state.Cur["wsrep_apply_oooe"] = 0.015600 
  
  col.Output( &b, state )
  if b.String() != "0.016" {
    t.Fatal( "Bad output", b.String())
  }
  b.Reset()
}

func TestDiff(t *testing.T) {
  // nils
  _, e1 := calculate_rate(nil, 10, 1)
  if e1 == nil {
    t.Error( "no error on nil")
  }
  
  // non-numeric types (basically not int or float64)
  _, e3 := calculate_rate("foo", 10, 1)
  _, e4 := calculate_rate(10, "bar", 1)
  if e3 == nil || e4 == nil {
    t.Error( "cur is non-numeric")    
  }
  
  // int
  d1, _ := calculate_rate( 20, 10, 1 )
  if d1 != 10.0 {
    t.Error( "d1 is", d1)
  }
  
  // float
  d2, _ := calculate_rate( 1.750, 1.500, 5)
  if d2 != 0.05 {
    t.Error( "d2 is ", d2)
  }
  
  // cur < prev
  d3, _ := calculate_rate( 10, 200, 2 )
  if d3 != 5.0 {
    t.Error( "d3 is", d3)
  }
  
  // time <= 0
  d4, _ := calculate_rate( 20, 10, -10 )
  if d4 != 20.0 {
    t.Error( "d4 is", d4)
  }
}

func TestRateCol(t *testing.T) {
  var b bytes.Buffer
  col := colDef{"connections", "cons", "%5s", "%5.0f", Rate, None}

  state := MyqState{}
  state.Cur = make( MyqSample )
  state.Cur["connections"] = 10
  state.Cur["uptime"] = 1
  
  state.Prev = make( MyqSample )
  state.Prev["connections"] = 20
  state.Prev["uptime"] = 5
  
  state.TimeDiff = 5.0
  
  col.Output( &b, state )
  if b.String() != "    2" {
    t.Fatal( "Bad output", b.String(),`.`)
  }
  b.Reset()
}

  // implement large number collapsing first
  // state.Cur["threads_running"] = "100000"
  // col.Output( &b, state )
  // if b.String() != "10.0k" {
  //   t.Fatal( "Bad output", b.String())
  // }