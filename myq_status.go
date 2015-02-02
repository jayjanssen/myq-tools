package main

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/jayjanssen/myq-tools/myqlib"
	"os"
	"os/signal"
	"runtime/pprof"
	"sort"
	"syscall"
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
	profile := flag.String("profile", "", "enable profiling and store the result in this file")
	header := flag.Int64("header", 20, "repeat the header after this many data points")
	mysql_args := flag.String("mysqlargs", "", "Arguments to pass to mysqladmin (used for connection options)")
	flag.StringVar(mysql_args, "a", "", "Short for -mysqlargs")
	interval := flag.Duration("interval", time.Second, "Time between samples (example: 1s or 1h30m)")
	flag.DurationVar(interval, "i", time.Second, "short for -interval")

	statusfile := flag.String("file", "", "parse mysqladmin ext output file instead of connecting to mysql")
	flag.StringVar(statusfile, "f", "", "short for -file")
	varfile := flag.String("varfile", "", "parse mysqladmin variables file instead of connecting to mysql, for optional use with -file")
	flag.StringVar(varfile, "vf", "", "short for -varfile")

	flag.Parse()

	// Enable profiling if set
	if *profile != "" {
		fmt.Println("Starting profiling to:", *profile)
		f, _ := os.Create(*profile)
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()

		// Need to trap interrupts in order for the profile to flush
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
		go func() {
			<-sigs
			pprof.StopCPUProfile()
			os.Exit(OK)
		}()

	}

	// Load default Views
	views := myqlib.DefaultViews()

	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage:\n  myq_status [flags] <view>\n")
		fmt.Fprintln(os.Stderr, "Description:\n  iostat-like views for MySQL servers\n")

		fmt.Fprintln(os.Stderr, "Options:")
		flag.PrintDefaults()
		fmt.Fprintln(os.Stderr, "\nViews:")

		var view_usage bytes.Buffer

		var sorted_views []string
		for name, _ := range views {
			sorted_views = append(sorted_views, name)
		}
		sort.Strings(sorted_views)
		for _, name := range sorted_views {
			view := views[name]
			view_usage.WriteString(fmt.Sprint("  ", name, ": "))
			for shortst := range view.ShortHelp() {
				view_usage.WriteString(fmt.Sprint(shortst, "\n"))
			}
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
		view_usage.WriteString(fmt.Sprint(`'`, view, `': `))
		for helpst := range v.Help() {
			view_usage.WriteString( fmt.Sprint( helpst, "\n"))
		}
		view_usage.WriteTo(os.Stderr)
		os.Exit(OK)
	}

	// The Loader and Timecol we will use
	var loader myqlib.Loader

	if *statusfile != "" {
		// File given, load it (and the optional varfile)
		loader = myqlib.NewFileLoader(*interval, *statusfile, *varfile)
		v.SetTimeCol( &myqlib.Runtime_col )
	} else {
		// No file given, this is a live collection and we use timestamps
		loader = myqlib.NewLiveLoader(*interval, *mysql_args)
		v.SetTimeCol( &myqlib.Timestamp_col )
	}

	states, err := myqlib.GetState(loader)
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
			lines = 0
			for headerln := range v.Header(state) {
				buf.WriteString( fmt.Sprint( headerln, "\n"))
				lines += 1
			}

			// Recalculate the height of the next header
			*header = myqlib.GetTermHeight() - 3
		}
		// Output data
		for dataln := range v.Data(state) {
			buf.WriteString( fmt.Sprint( dataln, "\n"))
			lines += 1
		}
		buf.WriteTo(os.Stdout)
	}

	os.Exit(OK)
}
