package myqlib

import (
  "fmt"
  "bytes"
)

type MyqView struct {  
  cols []colDef // slice of columns in this view
  help_description string // short description of the view
}

// Write a header line for a view
func (v MyqView) WriteHeader( b *bytes.Buffer) {
  // First line
  // Second line
  b.WriteString( fmt.Sprintf( "%8s", "Time" ))
  
  // Print all col headers for each in order
  for _, col := range v.cols {
    b.WriteString( " " ) // one space before each column
    b.WriteString( fmt.Sprintf( col.format_hdr, col.header ))
  }  
}


// Write a data line for a view given the state
func (v MyqView) WriteData( b *bytes.Buffer, state MyqState, useUptime bool ) {  
  // Every view outputs a time field 8 chars long
  if useUptime {
    b.WriteString( fmt.Sprintf( "%8d", state.Cur["uptime"] ))    
  } else {
    // date
    b.WriteString( fmt.Sprintf( "%8d", 1 ))
  }
  
  // Output all the col values in order based on their format
  for _, col := range v.cols {
    b.WriteString( " " ) // one space before each column
    col.Output( b, state )
  }
}