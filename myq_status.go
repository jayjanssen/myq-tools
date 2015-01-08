package main

import (
	"./myqlib"
	"bytes"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"
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
	flag.StringVar(mysql_args, "a", "", "Short for -mysqlargs")
	interval := flag.Duration("interval", time.Second, "Time between samples (example: 1s or 1h30m)")
	flag.DurationVar(interval, "i", time.Second, "short for -interval")

	statusfile := flag.String("file", "", "parse mysqladmin ext output file instead of connecting to mysql")
	flag.StringVar(statusfile, "f", "", "short for -file")
	varfile := flag.String("varfile", "", "parse mysqladmin variables file instead of connecting to mysql, for optional use with -file")
	flag.StringVar(varfile, "vf", "", "short for -varfile")

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
			view_usage.WriteString(fmt.Sprint("  ", name, ": "))
			view.Help(&view_usage, true)
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
		view_usage.WriteString(fmt.Sprint("'", view, "' Help: "))
		v.Help(&view_usage, false)
		view_usage.WriteTo(os.Stderr)
		os.Exit(OK)
	}
	
	// The Loader and Timecol we will use
	var loader myqlib.Loader
	var timecol myqlib.Col
	
	if *statusfile != "" {
		// File given, load it (and the optional varfile)
		loader = myqlib.NewFileLoader(*interval, *statusfile, *varfile)
		timecol = myqlib.Timestamp_col
	} else {
		// No file given, this is a live collection and we use timestamps
		loader = myqlib.NewLiveLoader(*interval, *mysql_args)
		timecol = myqlib.Runtime_col
	}

	states, err := myqlib.GetState( loader )
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(LOADER_ERROR)
	}

	// Apply selected view to output each sample
	lines := int64(0)
	for state := range states {
		var buf bytes.Buffer

		// Output a header if necessary
		if lines % *header == 0 {
			var hd1 bytes.Buffer
			timecol.Header1(&hd1)
			hd1.WriteString(" ")
			v.Header1(&hd1)
			hdr1 := strings.TrimSpace(hd1.String())
			if hdr1 != "" {
				buf.WriteString(hd1.String())
			}

			timecol.Header2(&buf)
			buf.WriteString(" ")
			v.Header2(&buf)
			
			*header = myqlib.GetTermHeight() - 3
			// fmt.Println( "New height = ", *header )
		}
		// Output data
		timecol.Data(&buf, state)
		buf.WriteString(" ")
		v.Data(&buf, state)
		buf.WriteTo(os.Stdout)

		lines++
	}
	
	os.Exit(OK)
}
