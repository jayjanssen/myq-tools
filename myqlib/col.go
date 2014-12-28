package myqlib

import (
	"bytes"
	"errors"
	"fmt"
)

// All Columns must implement the following
type Col interface {
	// outputs (write to the buffer)
	Help(b *bytes.Buffer)    // short help
	Header1(b *bytes.Buffer) // if empty, must print width spaces
	Header2(b *bytes.Buffer) // header to print above data

	// A full line of output given the state
	Data(b *bytes.Buffer, state MyqState)

	Width() uint8 // width of the column
}

// Groups of columns
type GroupCol struct {
	name string // name/header of the group
	help string // short description of the group

	cols []Col // slice of columns in this group
}

func (c GroupCol) Help(b *bytes.Buffer) {
	b.WriteString(c.help)
}
func (c GroupCol) Width() uint8 {
	var w uint8
	for _, col := range c.cols {
		w += col.Width() + 1
	}
	w -= 1
	return w
}

func (c GroupCol) Header1(b *bytes.Buffer) {
	b.WriteString(fmt.Sprintf(fmt.Sprint(`%-`, c.Width(), `s`),
    c.name))
}
func (c GroupCol) Header2(b *bytes.Buffer) {
	space := false
	for _, col := range c.cols {
		if space {
			b.WriteString(" ") // one space before each column
		}
		col.Header2(b)
		space = true
	}
}
func (c GroupCol) Data(b *bytes.Buffer, state MyqState) {

	space := false
	for _, col := range c.cols {
		if space {
			b.WriteString(" ") // one space before each column
		}
		col.Data(b, state)
		space = true
	}
}

// Gauge Columns simply display SHOW STATUS variable
type GaugeCol struct {
	name          string // name/header of the column
	variable_name string // SHOW STATUS variable of this column
	help          string // short description of the view

	width     uint8 // width of the column output (header and data)
	precision uint8 // # of decimals to show on floats (optional)
}

func (c GaugeCol) Help(b *bytes.Buffer) { b.WriteString(c.help) }
func (c GaugeCol) Width() uint8         { return c.width }
func (c GaugeCol) Header1(b *bytes.Buffer) {
	b.WriteString(fmt.Sprintf(fmt.Sprint(`%`, c.width, `s`), ""))
}
func (c GaugeCol) Header2(b *bytes.Buffer) {
	b.WriteString(fmt.Sprintf(fmt.Sprint(`%`, c.width, `s`), c.name))
}

func (c GaugeCol) Data(b *bytes.Buffer, state MyqState) {
	val := state.Cur[c.variable_name]

	switch v := val.(type) {
	case int64:
		// format number here
		b.WriteString(
			fmt.Sprintf(fmt.Sprint(`%`, c.width, `d`), v))
	case float64:
		// format number here
		// precision subtracts from total width (+ the decimal point)
		width := c.width
		if c.precision > 0 {
			width = width - (c.precision + 1)
		}
		b.WriteString(fmt.Sprintf(
			fmt.Sprint(`%`, width, `.`, c.precision, `f`), v))
	case string:
		b.WriteString(v)
	default:
		filler(b, c)
	}
}

// Rate Columns the rate of change of a SHOW STATUS variable
type RateCol struct {
	name          string // name/header of the column
	variable_name string // SHOW STATUS variable of this column
	help          string // short description of the view

	width     uint8 // width of the column output (header and data)
	precision uint8 // # of decimals to show on floats (optional)
}

func (c RateCol) Help(b *bytes.Buffer) { b.WriteString(c.help) }
func (c RateCol) Width() uint8 { return c.width }
func (c RateCol) Header1(b *bytes.Buffer) {
	b.WriteString(fmt.Sprintf(fmt.Sprint(`%`, c.width, `s`), ""))
}
func (c RateCol) Header2(b *bytes.Buffer) {
	b.WriteString(fmt.Sprintf(fmt.Sprint(`%`, c.width, `s`), c.name))
}

func (c RateCol) Data(b *bytes.Buffer, state MyqState) {
	// !! still not sure I like the uptime here
	diff, err := calculate_rate(state.Cur[c.variable_name], state.Prev[c.variable_name], state.TimeDiff)
	if err != nil {
		// Can't output, just put a filler
		// fmt.Println( err )
		filler(b, c)
	} else {
		b.WriteString(fmt.Sprintf(
			fmt.Sprint(`%`, c.width, `.`, c.precision, `f`), diff))
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
	// fmt.Println( reflect.TypeOf( cur ))
	var c, p float64
	switch cu := cur.(type) {
	case int64:
		c = float64(cu)
	case float64:
		c = cu
	default:
		return 0.00, errors.New("cur is not numeric!")
	}

	if prev != nil {
		switch pr := prev.(type) {
		case int64:
			p = float64(pr)
		case float64:
			p = pr
		default:
			return 0.00, errors.New("prev is not numeric!")
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

func filler(b *bytes.Buffer, c Col) {
	b.WriteString(fmt.Sprintf( fmt.Sprint(`%`, c.Width(), `s`), "-"))
}

// Func Columns run a custom function to produce their output
type FuncCol struct {
	name string // name/header of the column
	help string // short description of the view

	width     uint8 // width of the column output (header and data)
	precision uint8 // # of decimals to show on floats (optional)

	fn func(b *bytes.Buffer, state MyqState, c Col) // takes the state and returns the (unformatted) value
}

func (c FuncCol) Help(b *bytes.Buffer) { b.WriteString(c.help) }
func (c FuncCol) Width() uint8         { return c.width }
func (c FuncCol) Header1(b *bytes.Buffer) {
	b.WriteString(fmt.Sprintf(fmt.Sprint(`%`, c.width, `s`), ""))
}
func (c FuncCol) Header2(b *bytes.Buffer) {
	b.WriteString(fmt.Sprintf(fmt.Sprint(`%`, c.width, `s`), c.name))
}

func (c FuncCol) Data(b *bytes.Buffer, state MyqState) {
	c.fn(b, state, c)
}
