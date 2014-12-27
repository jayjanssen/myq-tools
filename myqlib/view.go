package myqlib

import(
  "bytes"
  "fmt"
)

// All Views must implement the following
type View interface {
  // outputs (write to the buffer)
  Help(b *bytes.Buffer) // short help
  Header(b *bytes.Buffer) // header to print above data
  
  // A full line of output given the state
  Data(b *bytes.Buffer, state MyqState, useUptime bool)
}

// NormalView
type NormalView struct {
  cols []Col // slice of columns in this view
  help string // short description of the view
}

func (v NormalView) Help(b *bytes.Buffer) {
  b.WriteString( v.help )
}

func (v NormalView) Header(b *bytes.Buffer) {
  // First line
  // Second line
  b.WriteString( fmt.Sprintf( "%8s", "Time" ))
  
  // Print all col headers for each in order
  for _, col := range v.cols {
    b.WriteString( " " ) // one space before each column
    col.Header( b )
  }  
  b.WriteString( "\n")
}

func (v NormalView) Data( b *bytes.Buffer, state MyqState, useUptime bool ) {  
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
    col.Data( b, state )
  }
  b.WriteString( "\n")
}