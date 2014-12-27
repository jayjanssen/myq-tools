package myqlib

import (
  "bytes"
  "fmt"
  "errors"
  // "reflect"
)

type MetricType uint8
const (
	Gauge MetricType = iota
	Rate
  Function
)

type Units uint8
const (
  None Units = iota
  Bytes
  Seconds
)

type colDef struct {
  name string           // name of the metric
  header string         // column header
  format_hdr string     // printf format of header
  format_val string     // printf format of value
  valtype MetricType    // How to compute the value of the col 
  valunits Units        // how to colapse and label values
}

func NewcolDef(name, header, format_hdr, format_val string ) *colDef {
    return &colDef{
      name: name,
      header: header,
      format_hdr: format_hdr,
      format_val: format_val,
      valtype: Gauge,
      valunits: None,
    }
}

// Write a column's data
func (c colDef) Output( b *bytes.Buffer, state MyqState) {
	switch c.valtype {
  case Function:
    fallthrough
	case Rate: 
    // !! still not sure I like the uptime here
    diff, err :=  calculate_rate( state.Cur[c.name], state.Prev[c.name], state.TimeDiff )
    if err != nil {
      b.WriteString( fmt.Sprint( "-")) // !! don't like this either
    } else {
      b.WriteString( fmt.Sprintf( c.format_val, diff ))
    }
	case Gauge:
    fallthrough
  default:
    b.WriteString( fmt.Sprintf( c.format_val, state.Cur[c.name]))
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