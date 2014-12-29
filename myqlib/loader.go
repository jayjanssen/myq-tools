package myqlib

import (
	"bufio"
	"os"
  "os/exec"
  "strconv"
  "fmt"
  "bytes"
)

const (
  MYSQLADMIN string = "mysqladmin"
  STATUS_COMMAND string = "extended-status"
)

type Loader interface {
  GetSamples() (chan MyqSample, error)
}

// Load mysql status output from a mysqladmin output file
type FileLoader struct {
  Filename string
}
func (l FileLoader) GetSamples() (chan MyqSample, error) {
	file, err := os.OpenFile(l.Filename, os.O_RDONLY, 0)
	if err != nil { return nil, err }

	scanner := bufio.NewScanner(file)
	var ch = make(chan MyqSample)

	// The file scanning goes into the background
	go func() {
		defer file.Close(); defer close(ch)
		scanMySQLShowLines(scanner, ch)
	}()

	return ch, nil
}

// SHOW STATUS output via mysqladmin
type MySQLAdminStatusLoader struct {
  Interval int64 // -i option to mysqladmin
  Args string // other args for mysqladmin (like -u, -p, -h, etc.)
}
func (l MySQLAdminStatusLoader) GetSamples() (chan MyqSample, error) {
  // Make sure we have MYSQLADMIN
	path, err := exec.LookPath(MYSQLADMIN)
  if( err != nil ) { return nil, err }
  
  // Build the argument list
  args := []string{
    STATUS_COMMAND, "-i", strconv.FormatInt(l.Interval,10)} 
  if l.Args != "" { args = append( args, l.Args )}
  
  // Initialize the command
  cmd := exec.Command( path, args...)
	
  var stderr bytes.Buffer
	cmd.Stderr = &stderr
  
  stdout, err := cmd.StdoutPipe()
  if err != nil { return nil, err }
  
  if err := cmd.Start(); err != nil { return nil, err }
  
  scanner := bufio.NewScanner(stdout)
	var ch = make(chan MyqSample)

	// The file scanning goes into the background
	go func() {
    defer close( ch )
		scanMySQLShowLines(scanner, ch)
	}()
  
  go func() {
    err := cmd.Wait()
    fmt.Println( MYSQLADMIN, "exited: ", err, stderr.String())
  }()

	return ch, nil
}