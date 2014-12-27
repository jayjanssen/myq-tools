package main

import (
    "fmt"
    "bytes"
    "os"
    "./myqlib"
)

func main() {
  // Parse arguments
  var file = "./testdata/mysqladmin.lots"
  var view = "cttf"
  
  var hdrEvery = 10
  
  fmt.Println( file, view )
  
  // Load default and custom Views/MetricDefs
  views := myqlib.DefaultViews()
  v, ok := views[view]
  if !ok {
    panic( "Unknown view" )
  }
  
  // Load data
  samples, err := myqlib.GetSamplesFile( file )
  if err != nil {
    panic( err )
  }
  
  // Apply selected view to output each sample
  i:= 0
  state := myqlib.MyqState{}
  for cur := range samples {   
    var buf bytes.Buffer
    state.Cur = cur
    if( state.Prev != nil ) {
      state.TimeDiff = float64( state.Cur["uptime"].(int64) - state.Prev["uptime"].(int64))
    }
    
    if i % hdrEvery == 0 {
      v.Header( &buf )
    } 
    v.Data( &buf, state, true )
    
    buf.WriteTo( os.Stdout )
    
    state.Prev = cur
    i++
  }  
}
