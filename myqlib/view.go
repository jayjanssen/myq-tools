package myqlib

import (
	"bytes"
	"strings"
)

// All Views must implement the following
type View interface {
	// outputs (write to the buffer)
	Help(b *bytes.Buffer, short bool) // help
	
	SetTimeCol( timecol *Col ) // Use this timecol in the output

	ExtraHeader(b *bytes.Buffer, state *MyqState) // header to print above data

	Header1(b *bytes.Buffer) // header to print above data
	Header2(b *bytes.Buffer) // header to print above data

	// A full line of output given the state
	Data(b *bytes.Buffer, state *MyqState) (lines int64)

	// All the cols (including time col)
	Cols() []Col
}

// NormalView
type NormalView struct {
	help string // short description of the view
	cols []Col  // slice of columns in this view
	timecol *Col // timecol to use
}

func NewNormalView(help string, cols ...Col) *NormalView {
	return &NormalView{help: help, cols: cols}
}

func (v *NormalView) Help(b *bytes.Buffer, short bool) {
	b.WriteString(v.help)
	b.WriteString("\n")
	if !short {
		b.WriteString("\n")
		for _, col := range v.Cols() {
			col.Help(b)
			b.WriteString("\n")
		}
	}
}

func (v *NormalView) SetTimeCol( timecol *Col ) {
	v.timecol = timecol
}

func (v *NormalView) ExtraHeader(b *bytes.Buffer, state *MyqState) {
}

func (v *NormalView) Header1(b *bytes.Buffer) {
	// Print all col header1s for each in order
	var header1 bytes.Buffer
	for _, col := range v.Cols() {
		col.Header1(&header1)
		header1.WriteString(" ")
	}
	// If the header1 buffer is all spaces, skip printing it
	hdr := strings.TrimSpace(header1.String())
	if hdr != "" {
		b.WriteString(header1.String())
		b.WriteString("\n")
	}
}

func (v *NormalView) Header2(b *bytes.Buffer) {
	// Print all col header2s for each in order
	for _, col := range v.Cols() {
		col.Header2(b)
		b.WriteString(" ") // one space before each column
	}
	b.WriteString("\n")
}

func (v *NormalView) Data(b *bytes.Buffer, state *MyqState) (int64) {
	// Output all the col values in order based on their format
	for _, col := range v.Cols() {
		col.Data(b, state)
		b.WriteString(" ") // one space before each column
	}
	b.WriteString("\n")
	
	return 1
}

// All columns preceeded by the time column
func (v *NormalView) Cols() []Col {
	return append( []Col{ *v.timecol }, v.cols... )
}

// ExtraHeaderView
type ExtraHeaderView struct {
	NormalView
	extra_header func(b *bytes.Buffer, state *MyqState)
}

func NewExtraHeaderView(help string, extra_header func(b *bytes.Buffer, state *MyqState), cols ...Col) *ExtraHeaderView {
	return &ExtraHeaderView{NormalView{help: help, cols: cols}, extra_header}
}

func (v *ExtraHeaderView) ExtraHeader(b *bytes.Buffer, state *MyqState) {
	v.extra_header(b, state)
	b.WriteString("\n")
}
