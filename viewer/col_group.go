package viewer

import (
	"github.com/jayjanssen/myq-tools2/loader"
)

// A GroupCol is a list of (related) cols
type GroupCol struct {
	defaultCol `yaml:",inline"`
	Cols       StateViewerList `yaml:"cols"`
}

// Header for this Group, the name of the Group is first, then the headers of each individual col
func (gc GroupCol) GetHeader(sr loader.StateReader) (result []string) {
	colOuts := gc.groupColOutput(func(sv StateViewer) []string {
		return sv.GetHeader(sr)
	})

	// Determne the length of this Group by the first line of output from the Cols
	if gc.Length == 0 && len(colOuts) > 0 {
		gc.Length = len(colOuts[0])
	}
	result = append(result, fitStringLeft(gc.Name, gc.Length))
	result = append(result, colOuts...)
	return
}

func (gc GroupCol) GetData(sr loader.StateReader) []string {
	return gc.groupColOutput(func(sv StateViewer) []string {
		return sv.GetData(sr)
	})
}

func (gc GroupCol) groupColOutput(getColOut func(sv StateViewer) []string) (result []string) {
	// Collect the string arrays from each column
	colsOutput := make([][]string, len(gc.Cols))
	maxLines := 0
	for i, c := range gc.Cols {
		colsOutput[i] = getColOut(c)
		if maxLines < len(colsOutput[i]) {
			maxLines = len(colsOutput[i])
		}
	}

	// Each col will output one or more lines, and they may output different amounts of lines. We use blank lines when a col doesn't have a value for a given line

	// Output maxLines # of lines to result
	for line := 0; line < maxLines; line += 1 {
		lineStr := ``
		for colI, colOut := range colsOutput {
			colLines := len(colOut) // How many lines does this col have?

			// Figure out which colOut line we should be printing
			staggeredI := line - (maxLines - colLines)

			// If staggeredI is negative, print a Blank, otherwise use the colOut
			if staggeredI < 0 {
				col := gc.Cols[colI]
				lineStr += col.GetBlankLine()
			} else {
				lineStr += colOut[staggeredI]
			}

			// Add a space for the next line
			lineStr += ` `
		}
		// Append the lineStr less 1 character (trailing space)
		result = append(result, lineStr[:len(lineStr)-1])
	}
	return
}
