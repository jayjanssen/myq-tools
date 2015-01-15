package myqlib

import (
	"bytes"
	"fmt"
)

// All Columns must implement the following
type Col interface {
	// outputs (write to the buffer)
	Help(b *bytes.Buffer)    // short help
	Header1(b *bytes.Buffer) // if empty, must print width spaces
	Header2(b *bytes.Buffer) // header to print above data

	// A full line of output given the state
	Data(b *bytes.Buffer, state *MyqState)
	
	// put a filler for the column into the buffer (usually because we can't put something useful)
	Filler(b *bytes.Buffer)

	Width() uint8 // width of the column
}

// 'Default' column -- "inherited" by others
type DefaultCol struct {
	name  string // name/header of the group
	help  string // short description of the group
	width uint8  // width of the column output (header and data)
}

func (c DefaultCol) Help(b *bytes.Buffer) {
	b.WriteString(fmt.Sprint(c.name, ": ", c.help))
}
func (c DefaultCol) Width() uint8 { return c.width }
func (c DefaultCol) Header1(b *bytes.Buffer) {
	b.WriteString(fmt.Sprintf(fmt.Sprint(`%-`, c.Width(), `s`), ""))
}
func (c DefaultCol) Header2(b *bytes.Buffer) {
	str := c.name
	if len(str) > int(c.Width()) {
		str = c.name[0:c.Width()]
	}
	b.WriteString(fmt.Sprintf(fmt.Sprint(`%`, c.Width(), `s`), str))
}

func (c DefaultCol) Filler(b *bytes.Buffer) {
	b.WriteString(fmt.Sprintf(fmt.Sprint(`%`, c.Width(), `s`), "-"))
}

// Column that is a Group of columns
type GroupCol struct {
	DefaultCol
	cols []Col // slice of columns in this group
}

func (c GroupCol) Help(b *bytes.Buffer) {
	b.WriteString(c.help)
	b.WriteString("\n")
	for _, col := range c.cols {
		b.WriteString("  ")
		col.Help(b)
		b.WriteString("\n")
	}
}
func (c GroupCol) Width() (w uint8) {
	for _, col := range c.cols {
		w += col.Width() + 1
	}
	w -= 1
	return
}
func (c GroupCol) Header1(b *bytes.Buffer) {
	str := c.name
	if len(str) > int(c.Width()) {
		str = c.name[0:c.Width()]
	}
	b.WriteString(fmt.Sprintf(fmt.Sprint(`%-`, c.Width(), `s`), str))
}
func (c GroupCol) Header2(b *bytes.Buffer) {
	space := false
	for _, col := range c.cols {
		if space {
			b.WriteString(" ")
		} else {
			space = true
		}
		col.Header2(b)
	}
}
func (c GroupCol) Data(b *bytes.Buffer, state *MyqState) {
	space := false
	for _, col := range c.cols {
		if space {
			b.WriteString(" ")
		} else {
			space = true
		}
		col.Data(b, state)
	}
}

// More conventional column types

// Gauge Columns simply display a SHOW STATUS variable
type GaugeCol struct {
	DefaultCol
	variable_name string // SHOW STATUS variable of this column
	precision     uint8  // # of decimals to show on floats (optional)
	units         UnitsDef
}
func (c GaugeCol) Data(b *bytes.Buffer, state *MyqState) {
	if val, err := state.Cur.getFloat(c.variable_name); err == nil {
		// got a number
		cv := collapse_number(val, int64(c.width), int64(c.precision), c.units)
		b.WriteString(fmt.Sprintf(fmt.Sprint(`%`, c.width, `s`), cv))
	} else if val, err := state.Cur.getString(c.variable_name); err == nil {
		if len(val) > int(c.Width()) {
			b.WriteString(val[0:c.Width()]) // first 'width' chars
		} else {
			b.WriteString(fmt.Sprintf(fmt.Sprint(`%`, c.Width(), `s`), val))
		}
	} else {
		// must be missing, just filler
		c.Filler(b)
	}
}

// Rate Columns the rate of change of a SHOW STATUS variable
type RateCol struct {
	DefaultCol
	variable_name string // SHOW STATUS variable of this column
	precision     uint8  // # of decimals to show on floats (optional)
	units         UnitsDef
}
func (c RateCol) Data(b *bytes.Buffer, state *MyqState) {
	cnum, cerr := state.Cur.getFloat(c.variable_name)
	pnum, _ := state.Prev.getFloat(c.variable_name)
	
	if cerr != nil { // we only care about cerr, if perr is set, it should be a 0.0
		c.Filler(b)
	}
	
	rate := calculate_rate(cnum, pnum, state.SecondsDiff)
	cv := collapse_number(rate, int64(c.width), int64(c.precision), c.units)
	b.WriteString(fmt.Sprintf(fmt.Sprint(`%`, c.width, `s`), cv))
}

// Diff Columns the difference of a SHOW STATUS variable between samples
type DiffCol struct {
	DefaultCol
	variable_name string // SHOW STATUS variable of this column
	precision     uint8  // # of decimals to show on floats (optional)
	units         UnitsDef
}
func (c DiffCol) Data(b *bytes.Buffer, state *MyqState) {
	cnum, cerr := state.Cur.getFloat(c.variable_name)
	pnum, _ := state.Prev.getFloat(c.variable_name)
		
	if cerr != nil { // we only care about cerr, if perr is set, it should be a 0.0
		c.Filler(b)
	} else {
		diff := calculate_diff(cnum, pnum)
		cv := collapse_number(diff, int64(c.width), int64(c.precision), c.units)
		b.WriteString(fmt.Sprintf(fmt.Sprint(`%`, c.width, `s`), cv))
	}
}

// Func Columns run a custom function to produce their output
type FuncCol struct {
	DefaultCol
	fn func(b *bytes.Buffer, state *MyqState, c Col) // takes the state and returns the (unformatted) value
}
func (c FuncCol) Data(b *bytes.Buffer, state *MyqState) {
	c.fn(b, state, c)
}

// Percent Columns calculate a ratio between two metrics
type PercentCol struct {
	DefaultCol
	numerator_name   string // SHOW STATUS variable of this column
	denomenator_name string // SHOW STATUS variable of this column
	precision        uint8  // # of decimals to show on floats (optional)
}
func (c PercentCol) Data(b *bytes.Buffer, state *MyqState) {	
	numerator, nerr := state.Cur.getFloat(c.numerator_name)
	denomenator, derr := state.Cur.getFloat(c.numerator_name)
	
	// Must have both
	if nerr != nil || derr != nil || denomenator == 0 { 
		c.Filler(b)
	} else {
		cv := collapse_number((numerator/denomenator) * 100, int64(c.width), int64(c.precision), PercentUnits)
		b.WriteString( fmt.Sprintf(fmt.Sprint(`%`, c.width, `s`), cv))		
	}

}

// String Columns show a string (or substring up to width)
type StringCol struct {
	DefaultCol
	variable_name string // SHOW STATUS variable of this column
}
func (c StringCol) Data(b *bytes.Buffer, state *MyqState) {
	val := state.Cur.getStr(c.variable_name)

	if len(val) > int(c.Width()) {
		b.WriteString(val[0:c.Width()]) // first 'width' chars
	} else {
		b.WriteString(fmt.Sprintf(fmt.Sprint(`%`, c.Width(), `s`), val))
	}
}

// RightmostCol shows width rightmost chars of the variable_name
type RightmostCol struct {
	DefaultCol
	variable_name string
}
func (c RightmostCol) Data(b *bytes.Buffer, state *MyqState) {
	// We show the least-significant width digits of the value
	id, _ := state.Cur.getString(c.variable_name)
	if len(id) > int(c.Width()) {
		b.WriteString(fmt.Sprintf(fmt.Sprint(`%`, c.Width(), `s`), id[len(id)-int(c.Width()):]))
	} else {
		b.WriteString(fmt.Sprintf(fmt.Sprint(`%`, c.Width(), `s`), id))
	}
}

// CurDiff Columns the difference between two variables in the same sample (different from DiffCol)
type CurDiffCol struct {
	DefaultCol
	bigger,smaller string  // The two variables to subtract
	precision     uint8  // # of decimals to show on floats (optional)
	units         UnitsDef
}
func (c CurDiffCol) Data(b *bytes.Buffer, state *MyqState) {
	bnum, _ := state.Cur.getFloat(c.bigger)
	snum, _ := state.Cur.getFloat(c.smaller)
	
	diff := calculate_diff(bnum, snum)
	cv := collapse_number(diff, int64(c.width), int64(c.precision), c.units)
	b.WriteString(fmt.Sprintf(fmt.Sprint(`%`, c.width, `s`), cv))
}

// RateSum Columns the rate of change of a sum of variables
type RateSumCol struct {
	DefaultCol
	variable_names []string
	precision     uint8  // # of decimals to show on floats (optional)
	units         UnitsDef
}
func (c RateSumCol) Data(b *bytes.Buffer, state *MyqState) {
	cursum := calculate_sum(state.Cur, c.variable_names)
	prevsum := calculate_sum(state.Prev, c.variable_names)
	
	rate := calculate_rate(cursum, prevsum, state.SecondsDiff)
	cv := collapse_number(rate, int64(c.width), int64(c.precision), c.units)
	b.WriteString(fmt.Sprintf(fmt.Sprint(`%`, c.width, `s`), cv))
}


// Helper funcs

// Calculate diff between two numbers, if negative, just return bigger
func calculate_diff(bigger, smaller float64) (float64) {
	if bigger < smaller {
		// special case -- if c is < p, the number rolled over or was reset, so best effort answer here.
		return bigger
	} else {
		return bigger - smaller
	}
}

// Calculate the rate of change between two values, given the time difference between them
func calculate_rate(bigger, smaller, seconds float64) (float64) {
	diff := calculate_diff(bigger, smaller)

	if seconds <= 0 { // negative seconds is weird 
		return diff
	} else {
		return diff / seconds
	}
}

// Return the sum of all variables in the given sample
func calculate_sum( sample MyqSample, variable_names []string ) (sum float64){
	for _, v := range variable_names {
		v, _ := sample.getFloat(v)
		sum += v
	}
	return sum
}
