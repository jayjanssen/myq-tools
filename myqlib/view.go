package myqlib

import (
	"bytes"
	"fmt"
)

// All Views must implement the following
type View interface {
	// outputs (write to the buffer)
	Help(b *bytes.Buffer) // help
	
	ShortHelp(b *bytes.Buffer) // Brief help
	
	 // Use this timecol in the output
	SetTimeCol( timecol *Col )
	
	// header to print above data
	Header(state *MyqState) chan string
	
	// A full line of output given the state
	Data(state *MyqState) chan string

	// All the cols (including time col)
	Cols() []Col
	
	// put a filler for the column into the buffer (usually because we can't put something useful)
	Filler() (string)
	Blank() (string) // line blank, but only spaces
	
	WriteString(val string) (string) // output the given val to fit the width of the column
	
	Width() int64 // width of the view
	
}

// NormalView
type NormalView struct {
	DefaultCol // Views are columns too
	cols []Col  // slice of columns in this view
	timecol *Col // timecol to use
}

func NewNormalView(help string, cols ...Col) *NormalView {
	return &NormalView{DefaultCol:DefaultCol{help: help}, cols: cols}
}

func (v *NormalView) Help(b *bytes.Buffer) {
	v.ShortHelp(b)
	b.WriteString("\n")
	for _, col := range v.cols {
		col.Help(b)
		b.WriteString("\n")
	}
}

func (v *NormalView) ShortHelp(b *bytes.Buffer) {
	b.WriteString(v.help)
	b.WriteString("\n")
}

func (v *NormalView) SetTimeCol( timecol *Col ) {
	v.timecol = timecol
}

func (v *NormalView) Header(state *MyqState) (chan string) {
	var column_channels []chan string
	for _, col := range v.Cols() {
		column_channels = append( column_channels, col.Header(state))
	}
	
	ch := make( chan string )
	go v.ordered_col_output( ch, column_channels )
	
	return ch
}

func (v *NormalView) Data(state *MyqState) (chan string) {
	var column_channels []chan string
	for _, col := range v.Cols() {
		column_channels = append( column_channels, col.Data(state))
	}
	
	ch := make( chan string )
	go v.ordered_col_output( ch, column_channels )
	
	return ch
}

func (v *NormalView) ordered_col_output( ch chan string, column_channels []chan string ) {
	defer close( ch )
	for {
		var hdrline bytes.Buffer
		got_something := false
		space := false
		for i, col := range v.Cols() {
			if space {
				hdrline.WriteString(" ")
			} else {
				space = true
			}
			if str, more := <- column_channels[i]; more {
				hdrline.WriteString( str )
				got_something = true
			} else {
				hdrline.WriteString( col.Blank())
			}
		}
		if got_something {
			ch <- hdrline.String()
		} else {
			break
		}
	}
}

// All columns preceeded by the time column
func (v *NormalView) Cols() []Col {
	if v.timecol == nil {
		return v.cols
	} else {
		return append( []Col{ *v.timecol }, v.cols... )
	}
}

func (v *NormalView) Width() (w int64) {
	for _, col := range v.Cols() {
		w += col.Width() + 1
	}
	w -= 1
	return
}

// ExtraHeaderView
type ExtraHeaderView struct {
	NormalView
	extra_header func(state *MyqState) (chan string)
}

func NewExtraHeaderView(help string, extra_header func(state *MyqState) (chan string), cols ...Col) *ExtraHeaderView {
	return &ExtraHeaderView{NormalView{DefaultCol:DefaultCol{help: help}, cols: cols}, extra_header}
}

func (v *ExtraHeaderView) Header(state *MyqState) (chan string) {
	ch := make(chan string)
	
	go func() {
		defer close(ch)
		extrach := v.extra_header(state)	
		for extrastr := range extrach {
			ch <- extrastr
		}
		normalch := v.NormalView.Header(state )
		for normalstr := range normalch {
			ch <- normalstr
		}
	}()
	
	return ch
}

// ExtraHeaderView
type GroupCol struct {
	NormalView
	title string
}

func NewGroupCol(title, help string, cols ...Col) *GroupCol {
	return &GroupCol{NormalView{DefaultCol:DefaultCol{help: help}, cols: cols}, title}
}

// All columns preceeded by the time column
func (v *GroupCol) Cols() []Col {
	return v.cols
}

func (v *GroupCol) Header(state *MyqState) (chan string) {
	ch := make( chan string )

	go func() {
		defer close(ch)
		
		str := v.title
		if len(str) > int(v.Width()) {
			str = v.title[0:v.Width()]
		}
		ch <- fmt.Sprintf(fmt.Sprint(`%-`, v.Width(), `s`), str)
		
		viewch := v.NormalView.Header(state)
		for viewstr := range viewch {
			ch <- viewstr
		}
	}()
	
	return ch
}