package myqlib

import (
	"bytes"
	"fmt"
)

// All Views must implement the following
type View interface {
	// Column help
	Help() chan string
	ShortHelp() chan string

	// Header/Data functions return a channel of strings
	Header(state *MyqState) chan string
	Data(state *MyqState) chan string

	// Use this timecol in the output
	SetTimeCol(timecol *Col)

	// All the cols (including time col)
	all_cols() []Col
}

// NormalView
type NormalView struct {
	DefaultCol       // Views are columns too
	cols       []Col // slice of columns in this view
	timecol    *Col  // timecol to use
}

func NewNormalView(help string, cols ...Col) *NormalView {
	return &NormalView{DefaultCol: DefaultCol{help: help}, cols: cols}
}

func (v *NormalView) Help() chan string {
	ch := make(chan string)

	go func() {
		defer close(ch)
		for shortst := range v.ShortHelp() {
			ch <- shortst
		}

		for _, col := range v.cols {
			for colst := range col.Help() {
				ch <- fmt.Sprint("\t", colst)
			}
		}
	}()

	return ch
}

func (v *NormalView) ShortHelp() chan string {
	ch := make(chan string, 1)
	defer close(ch)
	ch <- fmt.Sprint(v.help)
	return ch
}

func (v *NormalView) SetTimeCol(timecol *Col) {
	v.timecol = timecol
}

func (v *NormalView) Header(state *MyqState) chan string {
	return v.ordered_col_output(func(c Col) chan string {
		return c.Header(state)
	})
}

func (v *NormalView) Data(state *MyqState) chan string {
	return v.ordered_col_output(func(c Col) chan string {
		return c.Data(state)
	})
}

func (v *NormalView) ordered_col_output(get_col_chan func(c Col) chan string) chan string {
	var column_channels []chan string
	for _, col := range v.all_cols() {
		column_channels = append(column_channels, get_col_chan(col))
	}

	ch := make(chan string)
	go func() {
		defer close(ch)
		for {
			var hdrline bytes.Buffer
			got_something := false
			space := false
			for i, col := range v.all_cols() {
				if space {
					hdrline.WriteString(" ")
				} else {
					space = true
				}
				if str, more := <-column_channels[i]; more {
					hdrline.WriteString(str)
					got_something = true
				} else {
					hdrline.WriteString(column_blank(col))
				}
			}
			if got_something {
				ch <- hdrline.String()
			} else {
				break
			}
		}
	}()
	return ch
}

// All columns preceeded by the time column
func (v *NormalView) all_cols() []Col {
	if v.timecol == nil {
		return v.cols
	} else {
		return append([]Col{*v.timecol}, v.cols...)
	}
}

func (v *NormalView) Width() (w int64) {
	for _, col := range v.all_cols() {
		w += col.Width() + 1
	}
	w -= 1
	return
}

// ExtraHeaderView
type ExtraHeaderView struct {
	NormalView
	extra_header func(state *MyqState) chan string
}

func NewExtraHeaderView(help string, extra_header func(state *MyqState) chan string, cols ...Col) *ExtraHeaderView {
	return &ExtraHeaderView{NormalView{DefaultCol: DefaultCol{help: help}, cols: cols}, extra_header}
}

func (v *ExtraHeaderView) Header(state *MyqState) chan string {
	ch := make(chan string)

	go func() {
		defer close(ch)
		normalch := v.NormalView.Header(state)
		for normalstr := range normalch {
			ch <- normalstr
		}
		// Extra headers come out above normal headers, which means we send them later
		extrach := v.extra_header(state)
		for extrastr := range extrach {
			ch <- extrastr
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
	return &GroupCol{NormalView{DefaultCol: DefaultCol{help: help}, cols: cols}, title}
}

// All columns preceeded by the time column
func (v *GroupCol) all_cols() []Col {
	return v.cols
}

func (v *GroupCol) Header(state *MyqState) chan string {
	ch := make(chan string)

	go func() {
		defer close(ch)

		// Output the columns first
		viewch := v.NormalView.Header(state)
		for viewstr := range viewch {
			ch <- viewstr
		}

		// Then our title (reverse order)
		str := v.title
		if len(str) > int(v.Width()) {
			str = v.title[0:v.Width()]
		}
		ch <- fmt.Sprintf(fmt.Sprint(`%-`, v.Width(), `s`), str)
	}()

	return ch
}
