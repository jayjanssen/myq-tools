package viewer

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// this needs some error handling and testing love
func GetTermSize() (int, int) {
	cmd := exec.Command("stty", "size")
	cmd.Stdin = os.Stdin
	out, _ := cmd.Output()
	vals := strings.Split(strings.TrimSpace(string(out)), " ")

	height, _ := strconv.ParseInt(vals[0], 10, 64)
	width, _ := strconv.ParseInt(vals[1], 10, 64)

	return int(height), int(width)
}

// // Set OS-specific SysProcAttrs if they exist
// func cleanupSubcmd(c *exec.Cmd) {
// 	// Send the subprocess a SIGTERM when we exit
// 	attr := new(syscall.SysProcAttr)

// 	r := reflect.ValueOf(attr)
// 	f := reflect.Indirect(r).FieldByName(`Pdeathsig`)

// 	if f.IsValid() {
// 		f.Set(reflect.ValueOf(syscall.SIGTERM))
// 		c.SysProcAttr = attr
// 	}
// }

// Calculate diff between two numbers, if negative, just return bigger
func CalculateDiff(bigger, smaller float64) float64 {
	if bigger < smaller {
		// special case -- if c is < p, the number rolled over or was reset, so best effort answer here.
		return bigger
	} else {
		return bigger - smaller
	}
}

// Calculate the rate of change between two values, given the time difference between them
func CalculateRate(bigger, smaller, seconds float64) float64 {
	diff := CalculateDiff(bigger, smaller)
	if seconds <= 0 { // negative seconds is weird
		return diff
	} else {
		return diff / seconds
	}
}

// Return the sum of all variables in the given data
// func CalculateSum(data model.MyqData, variable_names []string) (sum float64) {
// 	for _, v := range variable_names {
// 		v, _ := data.GetFloat(v)
// 		sum += v
// 	}
// 	return sum
// }

// String functions

// helper function to fit a plain string to our Length
func FitString(input string, length int) string {
	if len(input) > int(length) {
		return input[0:length] // First width characters
	} else {
		return fmt.Sprintf(`%*s`, length, input)
	}
}

// helper function to fit a plain string to our Length
func FitStringLeft(input string, length int) string {
	if len(input) > int(length) {
		return input[0:length] // First width characters
	} else {
		return fmt.Sprintf(`%-*s`, length, input)
	}
}

// Generate a combined set of lines for all given StateViewers
func groupColOutput(svs StateViewerList, getColOut func(sv StateViewer) []string) (result []string) {
	// Collect the string arrays from each column
	colsOutput := make([][]string, len(svs))
	maxLines := 0
	for i, c := range svs {
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
				col := svs[colI]
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
