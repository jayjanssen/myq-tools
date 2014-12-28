package main

import (
	"bytes"
	"fmt"
	"os"
	"./myqlib"
)

func main() {
	// Parse arguments
	var file = "./testdata/mysqladmin.lots"
	var view = "cttf"

	var hdrEvery = int64(20)

	fmt.Println(file, view)

	// Load default and custom Views/MetricDefs
	views := myqlib.DefaultViews()
	v, ok := views[view]
	if !ok { panic("Unknown view") }

	time := myqlib.UPTIME
	v.SetTime(time)

	// Load data
	samples, err := myqlib.GetSamplesFile(file)
	if err != nil {
		panic(err)
	}

	// Apply selected view to output each sample
	i := int64(1)
	state := myqlib.MyqState{}
	for cur := range samples {
		var buf bytes.Buffer

		// Set the state for this sample
		cur[`iteration`] = i
		state.Cur = cur
		if state.Prev != nil {
			state.TimeDiff = float64(state.Cur["uptime"].(int64) - state.Prev["uptime"].(int64))
		}

		// Output a header if necessary
		if i % hdrEvery == 0 {
			v.Header(&buf)
		}
		// Output data
		v.Data(&buf, state)
		buf.WriteTo(os.Stdout)

		// Set the state for the next round
		state.Prev = cur
		i++
	}
}
