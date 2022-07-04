package loader

// import (
// 	"bytes"
// 	"os"
// 	"os/exec"
// 	"reflect"
// 	"strings"
// 	"syscall"
// 	"time"

// 	"github.com/jayjanssen/myq-tools/model"
// )

// const (
// 	MYSQLCLI string = "mysql"

// 	// These next two must match
// 	END_STRING  string = "MYQTOOLSEND"
// 	END_COMMAND string = "SELECT 'MYQTOOLSEND'"

// 	// The commands we send to the mysql cli
// 	STATUS_COMMAND    string = "SHOW GLOBAL STATUS"
// 	VARIABLES_COMMAND string = "SHOW GLOBAL VARIABLES"
// )

// // Build the argument list
// var MYSQLCLIARGS []string = []string{
// 	"-B", // Batch mode (tab-separated output)
// 	"-n", // Unbuffered
// 	"-N", // Skip column names
// }

// func getMySQLCLICmd(extra_args string) (*exec.Cmd, error) {
// 	// Make sure we have MYSQLCLI
// 	path, err := exec.LookPath(MYSQLCLI)
// 	if err != nil {
// 		return nil, err
// 	}

// 	all_args := MYSQLCLIARGS
// 	if extra_args != "" {
// 		all_args = append(all_args, strings.Split(extra_args, ` `)...)
// 	}

// 	// Initialize the command
// 	return exec.Command(path, all_args...), nil
// }

// type LiveLoader struct {
// 	interval time.Duration
// 	args     string
// }

// func NewLiveLoader(args string, i time.Duration) (*LiveLoader, error) {

// 	// Use the given args to connect to mysql and get a status, if this fails we can't connect to mysql
// 	cmd, err := getMySQLCLICmd(args + ` -e status`)
// 	if err != nil {
// 		return nil, err
// 	}

// 	_, err = cmd.CombinedOutput()
// 	if err != nil {
// 		return nil, err
// 	}
// 	return &LiveLoader{i, args}, nil
// }

// // Creates a MyqData struct given a query that returns records with two fields (a key and value)
// func (l LiveLoader) LoadFromCli(query string) (<-chan *model.MyqData, error) {

// 	cmd, err := getMySQLCLICmd(l.args)
// 	if err != nil {
// 		return nil, err
// 	}

// 	// Send the subprocess a SIGTERM when we exit
// 	attr := new(syscall.SysProcAttr)
// 	r := reflect.ValueOf(attr)
// 	f := reflect.Indirect(r).FieldByName(`Pdeathsig`)
// 	if f.IsValid() {
// 		f.Set(reflect.ValueOf(syscall.SIGTERM))
// 		cmd.SysProcAttr = attr
// 	}

// 	// Collect Stderr in a buffer
// 	var stderr bytes.Buffer
// 	cmd.Stderr = &stderr

// 	// Create a pipe for Stdout
// 	stdout, err := cmd.StdoutPipe()
// 	if err != nil {
// 		return nil, err
// 	}

// 	// Create a pipe for Stdin -- we input our command here every interval
// 	stdin, err := cmd.StdinPipe()
// 	if err != nil {
// 		return nil, err
// 	}

// 	// Start the command
// 	if err := cmd.Start(); err != nil {
// 		return nil, err
// 	}

// 	// Handle if the subcommand exits
// 	go func() {
// 		err := cmd.Wait()
// 		if err != nil {
// 			os.Stderr.WriteString(stderr.String())
// 			os.Exit(1)
// 		}
// 	}()

// 	// Construct the command to send to the subprocess
// 	full_command := strings.Join([]string{query, END_COMMAND, "\n"}, "; ")
// 	send_command := func() {
// 		// We don't check if the write failed, it's assumed the cmd.Wait() above will catch the sub proc dying

// 		stdin.Write([]byte(full_command)) // command we're harvesting
// 	}
// 	// send the first command immediately
// 	send_command()

// 	// produce more output every interval
// 	ticker := time.NewTicker(l.interval)
// 	go func() {
// 		defer stdin.Close()
// 		for range ticker.C {
// 			send_command()
// 		}
// 	}()

// 	// parse samples in the background
// 	ch := make(chan *model.MyqData)
// 	go func() {
// 		defer close(ch)
// 		parseSamples(stdout, ch, l.interval)
// 	}()

// 	// Got this far, the channel should start getting samples
// 	return ch, nil
// }

// // Returns a channel where new MyqSamples are collected and sent every l.interval from the l.db connection.
// func (l LiveLoader) GetSamples() <-chan *model.MyqSample {

// 	// Setup channels for MyqData for Status and variable commands
// 	status_ch, err := l.LoadFromCli(STATUS_COMMAND)
// 	if err != nil {
// 		return nil
// 	}
// 	var_ch, err := l.LoadFromCli(VARIABLES_COMMAND)
// 	if err != nil {
// 		return nil
// 	}

// 	ch := make(chan *model.MyqSample)
// 	go func() {
// 		for {
// 			status_data, status_ok := <-status_ch
// 			var_data, var_ok := <-var_ch

// 			if !status_ok || !var_ok {
// 				break
// 			}

// 			ch <- model.NewMyqSample(*status_data, *var_data)
// 		}
// 	}()
// 	return ch
// }
