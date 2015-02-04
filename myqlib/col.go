package myqlib

import (
	"fmt"
)

// All Columns must implement the following
type Col interface {
	// Column help
	Help() chan string

	// Write a line (or lines) of header to the returned channel
	Header(state *MyqState) chan string

	// A full line of output given the state
	Data(state *MyqState) chan string

	// width of the column
	Width() int64
}

// 'Default' column -- "inherited" by others
type DefaultCol struct {
	name  string // name/header of the col
	help  string // short description of the group
	width int64  // width of the column output (header and data)
}

func (c DefaultCol) Help() chan string {
	ch := make(chan string, 1)
	defer close(ch)
	ch <- fmt.Sprint(c.name, ": ", c.help)
	return ch
}
func (c DefaultCol) Width() int64 { return c.width }

func (c DefaultCol) Header(state *MyqState) chan string {
	ch := make(chan string, 1)
	defer close(ch)

	str := c.name
	if len(str) > int(c.Width()) {
		str = c.name[0:c.Width()]
	}
	ch <- fmt.Sprintf(fmt.Sprint(`%`, c.Width(), `s`), str)

	return ch
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

func (c GaugeCol) Data(state *MyqState) chan string {
	ch := make(chan string, 1)
	defer close(ch)

	if val, err := state.Cur.getFloat(c.variable_name); err == nil {
		ch <- fit_string(collapse_number(val, c.Width(), c.precision, c.units), c.Width())
	} else if val, err := state.Cur.getString(c.variable_name); err == nil {
		ch <- fit_string(val, c.Width())
	} else {
		// must be missing, just filler
		ch <- column_filler(c)
	}

	return ch
}

// Rate Columns the rate of change of a SHOW STATUS variable
type RateCol struct {
	GaugeCol
}

func NewRateCol(name, help string, width int64, variable_name string, precision int64, units UnitsDef) RateCol {
	return RateCol{GaugeCol{DefaultCol{name, help, width}, NumCol{precision, units}, variable_name}}
}

func (c RateCol) Data(state *MyqState) chan string {
	ch := make(chan string, 1)
	defer close(ch)

	cnum, cerr := state.Cur.getFloat(c.variable_name)
	pnum, _ := state.Prev.getFloat(c.variable_name)

	if cerr != nil { // we only care about cerr, if perr is set, it should be a 0.0
		ch <- column_filler(c)
	} else {
		cv := collapse_number(calculate_rate(cnum, pnum, state.SecondsDiff),
			c.Width(), c.precision, c.units)
		ch <- fit_string(cv, c.Width())
	}
	return ch
}

// Diff Columns the difference of a SHOW STATUS variable between samples
type DiffCol struct {
	GaugeCol
}

func NewDiffCol(name, help string, width int64, variable_name string, precision int64, units UnitsDef) DiffCol {
	return DiffCol{GaugeCol{DefaultCol{name, help, width}, NumCol{precision, units}, variable_name}}
}

func (c DiffCol) Data(state *MyqState) chan string {
	ch := make(chan string, 1)
	defer close(ch)

	cnum, cerr := state.Cur.getFloat(c.variable_name)
	pnum, _ := state.Prev.getFloat(c.variable_name)

	if cerr != nil { // we only care about cerr, if perr is set, it should be a 0.0
		ch <- column_filler(c)
	} else {
		cv := collapse_number(
			calculate_diff(cnum, pnum), c.Width(), c.precision, c.units)
		ch <- fit_string(cv, c.Width())
	}
	return ch
}

// Func Columns run a custom function to produce their output
type FuncCol struct {
	DefaultCol
	fn func(state *MyqState, c Col) chan string // takes the state and returns the (unformatted) value
}

func (c FuncCol) Data(state *MyqState) chan string {
	return c.fn(state, c)
}
func NewFuncCol(name, help string, width int64, fn func(*MyqState, Col) chan string) FuncCol {
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

func (c PercentCol) Data(state *MyqState) chan string {
	ch := make(chan string, 1)
	defer close(ch)

	numerator, nerr := state.Cur.getFloat(c.numerator)
	denomenator, derr := state.Cur.getFloat(c.denomenator)

	// Must have both
	if nerr != nil || derr != nil || denomenator == 0 {
		ch <- column_filler(c)
	} else {
		cv := collapse_number((numerator/denomenator)*100, c.Width(), c.precision, c.units)
		ch <- fit_string(cv, c.Width())
	}
	return ch
}

// String Columns show a string (or substring up to width)
type StringCol struct {
	DefaultCol
	variable_name string // SHOW STATUS variable of this column
}

func NewStringCol(name, help string, w int64, variable_name string) StringCol {
	return StringCol{DefaultCol{name, help, w}, variable_name}
}

func (c StringCol) Data(state *MyqState) chan string {
	ch := make(chan string, 1)
	defer close(ch)
	val := state.Cur.getStr(c.variable_name)
	ch <- fit_string(val, c.Width())
	return ch
}

// RightmostCol shows width rightmost chars of the variable_name
type RightmostCol struct {
	StringCol
}

func NewRightmostCol(name, help string, w int64, variable_name string) RightmostCol {
	return RightmostCol{StringCol{DefaultCol{name, help, w}, variable_name}}
}

func (c RightmostCol) Data(state *MyqState) chan string {
	ch := make(chan string, 1)
	defer close(ch)
	ch <- right_fit_string(state.Cur.getStr(c.variable_name), c.Width())
	return ch
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

func (c CurDiffCol) Data(state *MyqState) chan string {
	ch := make(chan string, 1)
	defer close(ch)

	bnum, _ := state.Cur.getFloat(c.bigger)
	snum, _ := state.Cur.getFloat(c.smaller)

	cv := collapse_number(calculate_diff(bnum, snum), c.Width(), c.precision, c.units)
	ch <- fit_string(cv, c.Width())
	return ch
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

func (c RateSumCol) Data(state *MyqState) chan string {
	ch := make(chan string, 1)
	defer close(ch)

	c.expand_variables(state.Cur)

	cursum := calculate_sum(state.Cur, c.expanded_variable_names)
	prevsum := calculate_sum(state.Prev, c.expanded_variable_names)

	rate := calculate_rate(cursum, prevsum, state.SecondsDiff)
	cv := collapse_number(rate, c.Width(), c.precision, c.units)
	ch <- fit_string(cv, c.Width())

	return ch
}

func (c *RateSumCol) expand_variables(sample MyqSample) {
	if len(c.expanded_variable_names) == 0 {
		c.expanded_variable_names = expand_variables(c.variable_names, sample)
	}
}
