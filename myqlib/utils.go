package myqlib

import (
	"bytes"
	"os"
	"os/exec"
	"reflect"
	"strconv"
	"strings"
	"syscall"
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

// Set OS-specific SysProcAttrs if they exist
func cleanupSubcmd(c *exec.Cmd) {
	// Send the subprocess a SIGTERM when we exit
	attr := new(syscall.SysProcAttr)

	r := reflect.ValueOf(attr)
	f := reflect.Indirect(r).FieldByName(`Pdeathsig`)

	if f.IsValid() {
		f.Set(reflect.ValueOf(syscall.SIGTERM))
		c.SysProcAttr = attr
	}
}

//
type FixedWidthBuffer struct {
	bytes.Buffer
	maxwidth int64
}

func (b *FixedWidthBuffer) SetWidth(w int64) {
	b.maxwidth = w
}
func (b *FixedWidthBuffer) WriteString(s string) (n int, err error) {
	runes := bytes.Runes([]byte(s))
	if b.maxwidth != 0 && len(runes) > int(b.maxwidth) {
		return b.Buffer.WriteString(string(runes[:b.maxwidth]))
	} else {
		return b.Buffer.WriteString(s)
	}
}
