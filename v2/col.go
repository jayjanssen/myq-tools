package myqlib

import(
  "bytes"
  "fmt"
  "strings"
)

type Col interface {
  // outputs (write to the buffer)
  Help(b *bytes.Buffer) // short help
  Header(b *bytes.Buffer) // header to print above data
  
  // A full line of output given the state
  Data(b *bytes.Buffer, state MyqState, useUptime bool)
  
  Width() // width of the column
}

// Gauge Columns simply display SHOW STATUS variable
type GaugeCol {
  name string // name/header of the column
  width string // width of the column output (header and data)
  help string // short description of the view
  variable_name string // SHOW STATUS variable of this column
}

func (c GaugeCol) Help(b *bytes.Buffer) {
  return c.help
}

func (c GaugeCol) Header(b *bytes.Buffer) {
  format := format.Sprint( `%`, c.width, `s`)
  b.WriteString( fmt.Sprintf( format, col.header ))
}

func (c GaugeCol) Data(b *bytes.Buffer, state MyqState, useUptime bool) {
  b.WriteString( fmt.Sprintf( c.format_val, state.Cur[c.name]))
}

// Rate Columns the rate of change of a SHOW STATUS variable
type GaugeCol {
  name string // name/header of the column
  width string // width of the column output (header and data)
  help string // short description of the view
  variable_name string // SHOW STATUS variable of this column
}

func (c GaugeCol) Help(b *bytes.Buffer) {
  return c.help
}

func (c GaugeCol) Header(b *bytes.Buffer) {
  format := format.Sprint( `%`, c.width, `s`)
  b.WriteString( fmt.Sprintf( format, col.header ))
}

func (c GaugeCol) Data(b *bytes.Buffer, state MyqState, useUptime bool) {
  // !! still not sure I like the uptime here
  diff, err := calculate_rate( state.Cur[c.variable_name], state.Prev[c.variable_name], state.TimeDiff )
  if err != nil {
    // Can't output, just put a filler
    format := fmt.Sprint( `%`, c.width, `s`)
    b.WriteString( fmt.Sprintf( format, "-")) 
  } else {
    b.WriteString( fmt.Sprintf( c.format_val, diff ))
  }
}

// calculate the difference over the time to get the rate.  This is complex, and we need to verify several things:
// 1. input intefaces are non-nil
// 2. cur & prev are int or float64
// 3. if prev is nil  and/or time is <0, we just return cur
// 4. output type always a float, deal with output format later
// 5. handle cur < prev (usually time would be <0 here, but in case), by just returing cur / time
func calculate_rate(cur, prev interface{}, time float64) (float64, error) {
  // cur and prev must not be nil
  if cur == nil {
    return 0.00, errors.New("nil cur")
  }

  // Rates only work on numeric types.  Error on non-numeric and convert numerics to float64 as needed
  var c, p float64
  switch cu := cur.(type) {
  case int:
    c = float64( cu )
  case float64:
    c = cu
  default:
    return 0.00, errors.New( "cur is not numeric!")
  }
  
  if prev != nil {
    switch pr := prev.(type) {
    case int:
      p = float64( pr )
    case float64:
      p = pr
    default:
      return 0.00, errors.New( "prev is not numeric!")
    }
  }
  
  if prev == nil || time <= 0 {
    return c, nil
  } else if c < p {
    return c / time, nil
  } else {
    return (c - p) / time, nil
  }
}
