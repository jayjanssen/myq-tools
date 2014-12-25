package myqlib

import (
    "testing"
    // "reflect"
)

func TestMetric(t *testing.T) {
  sample := make( MyqSample )
  
  sample["key"] = "value"
  
  if( sample.Length() != 1 ) {
    t.Fatal("Expecting 1 KV, got", sample.Length())
  }
}
//
// func TestConversion(t *testing.T) {
//   first, second := make( MyqSample ), make( MyqSample)
//   first["connections"] = "500"
//   second["connections"] = "550"
//
//   t.Log( reflect.TypeOf( first["connections"] ))
//   v:= reflect.ValueOf( first["connections"])
//   t.Log( "Kind:", v.Kind() )
//   t.Log( "Kind:", v.Kind() )
//
//
//   t.Fail()
//
// }