package viewer

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// this needs some error handling and testing love
func GetTermSize() (height, width int64) {
	cmd := exec.Command("stty", "size")
	cmd.Stdin = os.Stdin
	out, _ := cmd.Output()
	vals := strings.Split(strings.TrimSpace(string(out)), " ")

	height, _ = strconv.ParseInt(vals[0], 10, 64)
	width, _ = strconv.ParseInt(vals[1], 10, 64)
	return
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

// A buffer that only allows lines maxwidth long
type FixedWidthBuffer struct {
	bytes.Buffer
	maxwidth int64
}

// Set the new maximum width for the buffer
func (b *FixedWidthBuffer) SetWidth(w int64) {
	b.maxwidth = w
}

// Write a string to the buffer, truncate anything longer than maxwidth
func (b *FixedWidthBuffer) WriteString(s string) (n int, err error) {
	runes := bytes.Runes([]byte(s))
	if b.maxwidth != 0 && len(runes) > int(b.maxwidth) {
		return b.Buffer.WriteString(string(runes[:b.maxwidth]))
	} else {
		return b.Buffer.WriteString(s)
	}
}

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
func fitString(input string, length int) string {
	if len(input) > int(length) {
		return input[0:length] // First width characters
	} else {
		return fmt.Sprintf(`%*s`, length, input)
	}
}

// helper function to fit a plain string to our Length
func fitStringLeft(input string, length int) string {
	if len(input) > int(length) {
		return input[0:length] // First width characters
	} else {
		return fmt.Sprintf(`%-*s`, length, input)
	}
}
