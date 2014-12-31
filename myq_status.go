package main

import (
	"bytes"
	"fmt"
	"os"
  "flag"
  "strings"
  "time"
	"./myqlib"  
)

// Exit codes
const (
  OK int = iota
  BAD_ARGS
  LOADER_ERROR
  
)

func main() {
	// Parse arguments  
  help := flag.Bool("help", false, "this help text")
  header := flag.Int64("header", 20, "repeat the header after this many data points")
  // host := flag.String("host", "", "MySQL hostname")
  // port := flag.Int64("port", 0, "MySQL port")
  //
  // user := flag.String("user", "", "MySQL username")
  // pass := flag.String("pass", "", "MySQL password (askpass recommended instead)")
  // ask_pass := flag.Bool("askpass", false, "Prompt for MySQL password")
  mysql_args := flag.String("mysqlargs", "", "Arguments to pass to mysqladmin (used for connection options)")
  flag.StringVar(mysql_args,"a", "", "Short for -mysqlargs")
  interval := flag.Duration("interval", time.Second, "Time between samples (example: 1s or 1h30m)")
  flag.DurationVar(interval, "i", time.Second, "short for -interval")
  
  
  file := flag.String("file", "", "parse mysqladmin ext output file instead of connecting to mysql")
  flag.StringVar(file, "f", "", "short for -file")
  
  flag.Parse()
  
	// Load default Views
	views := myqlib.DefaultViews()
  
  flag.Usage = func() {
    fmt.Fprintln(os.Stderr, "Usage:\n  myq_status [flags] <view>\n")
    fmt.Fprintln(os.Stderr, "Description:\n  iostat-like views for MySQL servers\n")
    
    fmt.Fprintln(os.Stderr, "Options:")
    flag.PrintDefaults()
    fmt.Fprintln(os.Stderr, "\nViews:")
    
		var view_usage bytes.Buffer
    for name, view := range views {
  		view_usage.WriteString( fmt.Sprint("  ", name, ": "))
      view.Help(&view_usage, true)
  		view_usage.WriteString("\n")
    }
    view_usage.WriteTo(os.Stderr)
    os.Exit(BAD_ARGS)    
  }
  
  if flag.NArg() != 1 {
    flag.Usage()
  }
  
  view := flag.Arg(0)
	v, ok := views[view]
	if !ok { 
    fmt.Fprintln(os.Stderr, "Error: view", view, "not found")
    flag.Usage()
  }
  
  if *help {
      var view_usage bytes.Buffer
    view_usage.WriteString( fmt.Sprint("'", view, "' Help: "))
    v.Help(&view_usage, false)
    view_usage.WriteTo(os.Stderr)
    os.Exit(OK)
  }

	// Load data
  loader, timecol := func() (myqlib.Loader, myqlib.Col) {
    if( *file == "" ) {
      // collect samples from myqladmin
      load := new( myqlib.MySQLAdminStatusLoader)
      load.Interval = *interval
      load.Args = *mysql_args
      return load, myqlib.Timestamp_col
    } else {
      // collect samples from the named file
      load := new( myqlib.FileLoader )
      load.Filename = *file
      return load, myqlib.Runtime_col
    }
  }()
  
	samples, err := loader.GetSamples()
	if err != nil {
    fmt.Fprintln(os.Stderr, err)
    os.Exit(LOADER_ERROR)
  }

	// Apply selected view to output each sample
	lines := int64(0); state := myqlib.MyqState{}
	for sample := range samples {    
		var buf bytes.Buffer
    
		// Set the state for this sample
    state.Cur = sample
    if( state.FirstUptime == 0) {
      state.FirstUptime = state.Cur["uptime"].(int64)
    }
		if state.Prev != nil {
			state.TimeDiff = float64(state.Cur["uptime"].(int64) - state.Prev["uptime"].(int64))
      
      // Skip to the next sample if TimeDiff is < the interval
      if state.TimeDiff < interval.Seconds() { continue }
		}

		// Output a header if necessary
		if lines % *header == 0 {
      var hd1 bytes.Buffer 
      timecol.Header1(&hd1); hd1.WriteString(" "); v.Header1(&hd1)
    	hdr1 := strings.TrimSpace(hd1.String())
    	if hdr1 != "" {
        buf.WriteString(hd1.String());
    	}
      
      timecol.Header2(&buf); buf.WriteString(" "); v.Header2(&buf)
    }
		// Output data
		timecol.Data(&buf, state); buf.WriteString(" "); v.Data(&buf, state)
		buf.WriteTo(os.Stdout)

		// Set the state for the next round
		state.Prev = sample; lines++
	}
  os.Exit(OK)  
}