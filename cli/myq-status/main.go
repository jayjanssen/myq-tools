package main

import (
	"flag"
	"fmt"
	"math"
	"os"
	"os/signal"
	"runtime/pprof"
	"syscall"
	"time"

	"github.com/jayjanssen/myq-tools2/clientconf"
	"github.com/jayjanssen/myq-tools2/loader"
	"github.com/jayjanssen/myq-tools2/viewer"
)

// Exit codes
const (
	OK int = iota
	BAD_ARGS
	LOADER_ERROR
	SOURCES_ERROR
)

// Current Version (passed in on build)
var build_version string
var build_timestamp string

func main() {
	// Parse arguments
	help := flag.Bool("help", false, "this help text")
	version := flag.Bool("version", false, "print the version")

	profile := flag.String("profile", "", "enable profiling and store the result in this file")
	header := flag.Int("header", 0, "repeat the header after this many data points (default: 0, autocalculates)")
	width := flag.Bool("width", false, "Truncate the output based on the width of the terminal")

	interval := flag.Duration("interval", time.Second, "Time between samples (example: 1s or 1h30m)")
	flag.DurationVar(interval, "i", time.Second, "short for -interval")

	statusfile := flag.String("file", "", "parse mysqladmin ext output file instead of connecting to mysql")
	flag.StringVar(statusfile, "f", "", "short for -file")
	varfile := flag.String("varfile", "", "parse mysqladmin variables file instead of connecting to mysql, for optional use with -file")
	flag.StringVar(varfile, "vf", "", "short for -varfile")
	clientconf.SetMySQLFlags()

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

	if *version {
		fmt.Printf("myq-tools %s (%s)\n", build_version, build_timestamp)
		os.Exit(OK)
	}

	// Load default Views
	err := viewer.LoadDefaultViews()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading default views: %s\n", err)
		os.Exit(LOADER_ERROR)
	}

	// Define standard usage output
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "myq-tools %s (%s)\n\n", build_version, build_timestamp)

		fmt.Fprintln(os.Stderr, "Usage:\n  myq_status [flags] <view>")
		fmt.Fprintln(os.Stderr, "Description:\n  iostat-like views for MySQL servers")

		fmt.Fprintln(os.Stderr, "Options:")
		flag.PrintDefaults()
		fmt.Fprintln(os.Stderr, "\nViews:")

		for _, name := range viewer.ViewNames {
			view, _ := viewer.GetViewer(name)
			fmt.Fprintf(os.Stderr, "  %s: %s\n", name, view.GetShortHelp())
		}
		os.Exit(BAD_ARGS)
	}

	// Print usage if we don't have exactly one non-flag cli arg
	if flag.NArg() != 1 {
		flag.Usage()
	}

	// Sanity check interval
	if interval.Seconds() < 1 {
		fmt.Fprintln(os.Stderr, "Error: interval must be >= 1s")
		flag.Usage()
	} else if math.Mod(float64(interval.Nanoseconds()), 1000000000) != 0.0 {
		fmt.Fprintln(os.Stderr, "Warning: interval will be rounded to",
			fmt.Sprintf("%.0f", interval.Seconds()), "seconds")
	}

	// Look for the requested view
	viewName := flag.Arg(0)
	view, err := viewer.GetViewer(viewName)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		flag.Usage()
	}

	// Print help for the requested view
	if *help {
		fmt.Fprintf(os.Stderr, "'%s':\n", viewName)
		for helpst := range view.GetDetailedHelp() {
			fmt.Fprintln(os.Stderr, helpst)
		}
		os.Exit(OK)
	}

	// The Loader and Timecol we will use
	var load loader.Loader

	if *statusfile == "" {
		// No file given, this is a live collection and we use timestamps
		config, err := clientconf.GenerateConfig()
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v", err)
		}
		load = loader.NewLiveLoader(config.FormatDSN())
	} else {
		// File given, load it (and the optional varfile)
		load = loader.NewFileLoader(*statusfile, *varfile)
	}

	sources, err := view.GetSources()
	if err != nil {
		fmt.Fprint(os.Stderr, err)
		os.Exit(SOURCES_ERROR)
	}

	// Initialize the loader
	err = load.Initialize(*interval, sources)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(LOADER_ERROR)
	}

	// How big is our terminal?
	termheight, termwidth := viewer.GetTermSize()

	// How many lines before printing a new header
	headerRepeat := termheight
	if *header != 0 {
		// Use the specified --header count
		headerRepeat = *header
	}

	// Apply selected view to output each sample
	linesSinceHeader := 0

	printOutput := func(s string) {
		if *width {
			s = viewer.FitString(s, termwidth)
		}
		fmt.Println(s)
	}

	// Main loop through loader States
	for state := range load.GetStateChannel() {
		// Reprint a header whenever lines == 0
		if linesSinceHeader == 0 {
			for _, headerLn := range view.GetHeader(state) {
				printOutput(headerLn)
				linesSinceHeader += 1
			}
		}

		// Output data
		for _, dataLn := range view.GetData(state) {
			printOutput(dataLn)
			linesSinceHeader += 1
		}

		// Determine if we need to reset lines to 0 (and trigger a header)
		if linesSinceHeader/headerRepeat >= 1 {
			linesSinceHeader = 0

			// Recalculate terminal size if this affects our width or headerRepeat
			if *width || *header == 0 {
				// Recalculate the size of the terminal now too
				termheight, termwidth = viewer.GetTermSize()
				if *header == 0 {
					headerRepeat = termheight
				}
			}
		}
	}

	os.Exit(OK)
}
