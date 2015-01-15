package myqlib

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
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
	val := state.Cur[c.variable_name]

	switch v := val.(type) {
	case int64:
		// format number here
		cv := collapse_number(float64(v), int64(c.width), int64(c.precision), c.units)
		b.WriteString(
			fmt.Sprintf(fmt.Sprint(`%`, c.width, `s`), cv))
	case float64:
		// format number here
		// precision subtracts from total width (+ the decimal point)
		cv := collapse_number(v, int64(c.width), int64(c.precision), c.units)
		b.WriteString(fmt.Sprintf(fmt.Sprint(`%`, c.width, `s`), cv))
	case string:
		b.WriteString(v[0:c.width]) // first 'width' chars
	default:
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
	rate, err := calculate_rate(state.Cur[c.variable_name], state.Prev[c.variable_name], state.SecondsDiff)
	if err != nil {
		c.Filler(b)
	} else {
		cv := collapse_number(rate, int64(c.width), int64(c.precision), c.units)
		b.WriteString(fmt.Sprintf(fmt.Sprint(`%`, c.width, `s`), cv))
	}
}

func calculate_rate(cur, prev interface{}, seconds float64) (float64, error) {
	diff, err := calculate_diff(cur, prev)
	if err != nil {
		return 0.00, err
	}

	if seconds <= 0 {
		return diff, nil
	} else {
		return diff / seconds, nil
	}
}

// Diff Columns the difference of a SHOW STATUS variable between samples
type DiffCol struct {
	DefaultCol
	variable_name string // SHOW STATUS variable of this column
	precision     uint8  // # of decimals to show on floats (optional)
	units         UnitsDef
}

func (c DiffCol) Data(b *bytes.Buffer, state *MyqState) {
	diff, err := calculate_diff(state.Cur[c.variable_name], state.Prev[c.variable_name])
	if err != nil {
		// Can't output, just put a filler
		// fmt.Println( err )
		c.Filler(b)
	} else {
		cv := collapse_number(diff, int64(c.width), int64(c.precision), c.units)
		b.WriteString(fmt.Sprintf(fmt.Sprint(`%`, c.width, `s`), cv))
	}
}

func calculate_diff(cur, prev interface{}) (float64, error) {
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

	if prev == nil {
		return c, nil
	} else if c < p {
		// special case -- if c is < p, the number rolled over or was reset, so best effort answer here.
		return c, nil
	} else {
		return c - p, nil
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
	var numerator, denomenator float64

	nval := state.Cur[c.numerator_name]
	switch nv := nval.(type) {
	case int64:
		numerator = float64(nv)
	case float64:
		numerator = nv
	default:
		c.Filler(b)
		return
	}

	dval := state.Cur[c.denomenator_name]
	switch dv := dval.(type) {
	case int64:
		denomenator = float64(dv)
	case float64:
		denomenator = dv
	default:
		c.Filler(b)
		return
	}

	cv := collapse_number((numerator/denomenator)*100, int64(c.width), int64(c.precision), PercentUnits)

	b.WriteString(
		fmt.Sprintf(fmt.Sprint(`%`, c.width, `s`), cv))

}

// String Columns show a string (or substring up to width)
type StringCol struct {
	DefaultCol
	variable_name string // SHOW STATUS variable of this column
}

func (c StringCol) Data(b *bytes.Buffer, state *MyqState) {
	val := state.Cur[c.variable_name]

	switch v := val.(type) {
	case string:
		if len(v) > int(c.Width()) {
			b.WriteString(v[0:c.Width()]) // first 'width' chars
		} else {
			b.WriteString(fmt.Sprintf(fmt.Sprint(`%`, c.Width(), `s`), v))
		}
	default:
		c.Filler(b)
	}
}

// RightmostCol shows width rightmost chars of the variable_name
type RightmostCol struct {
	DefaultCol
	variable_name string
}

func (c RightmostCol) Data(b *bytes.Buffer, state *MyqState) {
	// We show the least-significant width digits of the value
	id := strconv.Itoa(int(state.Cur[c.variable_name].(int64)))
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
	diff, err := calculate_diff(state.Cur[c.bigger], state.Cur[c.smaller])
	if err != nil {
		c.Filler(b)
	} else {
		cv := collapse_number(diff, int64(c.width), int64(c.precision), c.units)
		b.WriteString(fmt.Sprintf(fmt.Sprint(`%`, c.width, `s`), cv))
	}
}

// RateSum Columns the rate of change of a sum of variables
type RateSumCol struct {
	DefaultCol
	variable_names []string
	precision     uint8  // # of decimals to show on floats (optional)
	units         UnitsDef
}

func (c RateSumCol) Data(b *bytes.Buffer, state *MyqState) {
	sumfunc := func( sample MyqSample ) (sum float64){
		for _, v := range c.variable_names {
			v := sample[v]
			switch v := v.(type) {
			case int64:
				sum += float64(v)
			case float64:
				sum += v
			}
		}
		return sum
	}
	cursum := sumfunc(state.Cur)
	prevsum := sumfunc(state.Prev)
	
	rate, err := calculate_rate(cursum, prevsum, state.SecondsDiff)
	if err != nil {
		c.Filler(b)
	} else {
		cv := collapse_number(rate, int64(c.width), int64(c.precision), c.units)
		b.WriteString(fmt.Sprintf(fmt.Sprint(`%`, c.width, `s`), cv))
	}
}