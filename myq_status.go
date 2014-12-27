package main

import (
    "fmt"
    "./myqlib"
    "bytes"
)

func main() {
  // Parse arguments
  var file = "./testdata/mysqladmin.lots"
  var view = "cttf"
  
  var hdrEvery = 10
  
  fmt.Println( file, view )
  
  // Load default and custom Views/MetricDefs
  views := myqlib.DefaultMyqViews()
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
  var buf bytes.Buffer
  for cur := range samples {   
    state.Cur = cur
    
    if i % hdrEvery == 0 {
      v.WriteHeader( &buf )
      buf.WriteString( fmt.Sprintln())
    } 
    v.WriteData( &buf, state, true )
    buf.WriteString( fmt.Sprintln())
    
    fmt.Print( buf.String())
    buf.Reset()
    
    state.Prev = cur
    i++
  }  
}
