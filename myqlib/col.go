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
	WriteString(b *bytes.Buffer, val string) // output the given val to fit the width of the column

	Width() int64 // width of the column
}

// 'Default' column -- "inherited" by others
type DefaultCol struct {
	name  string // name/header of the group
	help  string // short description of the group
	width int64  // width of the column output (header and data)
}

func (c DefaultCol) Help(b *bytes.Buffer) {
	b.WriteString(fmt.Sprint(c.name, ": ", c.help))
}
func (c DefaultCol) Width() int64 { return c.width }
func (c DefaultCol) Header1(b *bytes.Buffer) {
	b.WriteString(fmt.Sprintf(fmt.Sprint(`%-`, c.Width(), `s`), ""))
}
func (c DefaultCol) Header2(b *bytes.Buffer) {
	str := c.name
	if len(str) > int(c.Width()) {
		str = c.name[0:c.Width()]
	}
	c.WriteString(b, str)
}
func (c DefaultCol) Filler(b *bytes.Buffer) {
	c.WriteString(b, "-")
}

func (c DefaultCol) WriteString(b *bytes.Buffer, val string) {
	b.WriteString(fmt.Sprintf(fmt.Sprint(`%`, c.Width(), `s`), val))
}

// Column that is a Group of columns
type GroupCol struct {
	DefaultCol
	cols []Col // slice of columns in this group
}

func NewGroupCol(name, help string, cols ...Col) GroupCol {
	return GroupCol{DefaultCol{name, help, 0}, cols}
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
func (c GroupCol) Width() (w int64) {
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

// Meta-type, displays some kind of number
type NumCol struct {
	precision int64 // # of decimals to show on floats (optional)
	units     UnitsDef
}

// Gauge Columns simply display a SHOW STATUS variable
type GaugeCol struct {
	DefaultCol
	NumCol
	variable_name string // SHOW STATUS variable of this column
}

func NewGaugeCol(name, help string, width int64, variable_name string, precision int64, units UnitsDef) GaugeCol {
	return GaugeCol{DefaultCol{name, help, width}, NumCol{precision, units}, variable_name}
}

func (c GaugeCol) Data(b *bytes.Buffer, state *MyqState) {
	if val, err := state.Cur.getFloat(c.variable_name); err == nil {
		c.WriteString(b, collapse_number(val, c.Width(), c.precision, c.units))
	} else if val, err := state.Cur.getString(c.variable_name); err == nil {
		if len(val) > int(c.Width()) {
			c.WriteString(b, val[0:c.Width()]) // first 'width' chars
		} else {
			c.WriteString(b, val)
		}
	} else {
		// must be missing, just filler
		c.Filler(b)
	}
}

// Rate Columns the rate of change of a SHOW STATUS variable
type RateCol struct {
	GaugeCol
}

func NewRateCol(name, help string, width int64, variable_name string, precision int64, units UnitsDef) RateCol {
	return RateCol{GaugeCol{DefaultCol{name, help, width}, NumCol{precision, units}, variable_name}}
}

func (c RateCol) Data(b *bytes.Buffer, state *MyqState) {
	cnum, cerr := state.Cur.getFloat(c.variable_name)
	pnum, _ := state.Prev.getFloat(c.variable_name)

	if cerr != nil { // we only care about cerr, if perr is set, it should be a 0.0
		c.Filler(b)
	} else {
		rate := calculate_rate(cnum, pnum, state.SecondsDiff)
		cv := collapse_number(rate, c.Width(), c.precision, c.units)
		c.WriteString(b, cv)
	}
}

// Diff Columns the difference of a SHOW STATUS variable between samples
type DiffCol struct {
	GaugeCol
}

func NewDiffCol(name, help string, width int64, variable_name string, precision int64, units UnitsDef) DiffCol {
	return DiffCol{GaugeCol{DefaultCol{name, help, width}, NumCol{precision, units}, variable_name}}
}

func (c DiffCol) Data(b *bytes.Buffer, state *MyqState) {
	cnum, cerr := state.Cur.getFloat(c.variable_name)
	pnum, _ := state.Prev.getFloat(c.variable_name)

	if cerr != nil { // we only care about cerr, if perr is set, it should be a 0.0
		c.Filler(b)
	} else {
		diff := calculate_diff(cnum, pnum)
		cv := collapse_number(diff, c.Width(), c.precision, c.units)
		c.WriteString(b, cv)
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
func NewFuncCol(name, help string, width int64, fn func(*bytes.Buffer, *MyqState, Col)) FuncCol {
	return FuncCol{DefaultCol{name, help, width}, fn}
}

// Percent Columns calculate a ratio between two metrics
type PercentCol struct {
	DefaultCol
	NumCol
	numerator, denomenator string // SHOW STATUS variable of this column
}

func NewPercentCol(name, help string, w int64, numerator, denomenator string, p int64) PercentCol {
	return PercentCol{DefaultCol{name, help, w}, NumCol{p, PercentUnits}, numerator, denomenator}
}

func (c PercentCol) Data(b *bytes.Buffer, state *MyqState) {
	numerator, nerr := state.Cur.getFloat(c.numerator)
	denomenator, derr := state.Cur.getFloat(c.denomenator)

	// Must have both
	if nerr != nil || derr != nil || denomenator == 0 {
		c.Filler(b)
	} else {
		cv := collapse_number((numerator/denomenator)*100, c.Width(), c.precision, c.units)
		c.WriteString(b, cv)
	}

}

// String Columns show a string (or substring up to width)
type StringCol struct {
	DefaultCol
	variable_name string // SHOW STATUS variable of this column
}

func NewStringCol(name, help string, w int64, variable_name string) StringCol {
	return StringCol{DefaultCol{name, help, w}, variable_name}
}

func (c StringCol) Data(b *bytes.Buffer, state *MyqState) {
	val := state.Cur.getStr(c.variable_name)

	if len(val) > int(c.Width()) {
		c.WriteString(b, val[0:c.Width()]) // first 'width' chars
	} else {
		c.WriteString(b, val)
	}
}

// RightmostCol shows width rightmost chars of the variable_name
type RightmostCol struct {
	StringCol
}

func NewRightmostCol(name, help string, w int64, variable_name string) RightmostCol {
	return RightmostCol{StringCol{DefaultCol{name, help, w}, variable_name}}
}

func (c RightmostCol) Data(b *bytes.Buffer, state *MyqState) {
	// We show the least-significant width digits of the value
	id := state.Cur.getStr(c.variable_name)
	if len(id) > int(c.Width()) {
		c.WriteString(b, id[len(id)-int(c.Width()):])
	} else {
		c.WriteString(b, id)
	}
}

// CurDiff Columns the difference between two variables in the same sample (different from DiffCol)
type CurDiffCol struct {
	DefaultCol
	NumCol
	bigger, smaller string // The two variables to subtract
}

func NewCurDiffCol(name, help string, width int64, bigger, smaller string, precision int64, units UnitsDef) CurDiffCol {
	return CurDiffCol{DefaultCol{name, help, width}, NumCol{precision, units}, bigger, smaller}
}

func (c CurDiffCol) Data(b *bytes.Buffer, state *MyqState) {
	bnum, _ := state.Cur.getFloat(c.bigger)
	snum, _ := state.Cur.getFloat(c.smaller)

	cv := collapse_number(calculate_diff(bnum, snum), c.Width(), c.precision, c.units)
	c.WriteString(b, cv)
}

// RateSum Columns the rate of change of a sum of variables
type RateSumCol struct {
	DefaultCol
	NumCol
	variable_names          []string
	expanded_variable_names []string
}

func NewRateSumCol(name, help string, width int64, precision int64, units UnitsDef, variables ...string) RateSumCol {
	return RateSumCol{DefaultCol{name, help, width}, NumCol{precision, units}, variables, []string{}}
}

func (c RateSumCol) Data(b *bytes.Buffer, state *MyqState) {
	c.expand_variables(state.Cur)

	cursum := calculate_sum(state.Cur, c.expanded_variable_names)
	prevsum := calculate_sum(state.Prev, c.expanded_variable_names)

	rate := calculate_rate(cursum, prevsum, state.SecondsDiff)
	cv := collapse_number(rate, c.Width(), c.precision, c.units)
	c.WriteString(b, cv)
}

func (c *RateSumCol) expand_variables(sample MyqSample) {
	if len(c.expanded_variable_names) == 0 {
		c.expanded_variable_names = expand_variables(c.variable_names, sample)
	}
}
