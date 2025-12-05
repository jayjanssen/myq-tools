package main

import (
	"database/sql"
	"flag"
	"fmt"
	"math"
	"os"
	"os/signal"
	"runtime/pprof"
	"syscall"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jayjanssen/myq-tools/lib/blip"
	"github.com/jayjanssen/myq-tools/lib/clientconf"
	"github.com/jayjanssen/myq-tools/lib/viewer"
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

		for _, name := range viewer.ListViews() {
			view, _ := viewer.GetViewer(name)
			fmt.Fprintf(os.Stderr, "   %s\n", view.GetShortHelp())
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
		for _, helpst := range view.GetDetailedHelp() {
			fmt.Fprintln(os.Stderr, helpst)
		}
		os.Exit(OK)
	}

	// Extract required metrics from the view
	metricsByDomain := view.GetMetricsByDomain()

	// Create metrics channel based on mode (live or file)
	var metricsChan <-chan *blip.Metrics

	if *statusfile == "" {
		// Live mode: connect to MySQL using blip
		mysqlConfig, err := clientconf.GenerateConfig()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error generating config: %v\n", err)
			os.Exit(LOADER_ERROR)
		}

		// Convert to blip config
		blipCfg, err := blip.ConfigFromMySQL(mysqlConfig)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error converting config: %v\n", err)
			os.Exit(LOADER_ERROR)
		}

		// Open database connection
		dsn, err := blip.MakeDSN(blipCfg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating DSN: %v\n", err)
			os.Exit(LOADER_ERROR)
		}

		db, err := sql.Open("mysql", dsn)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error connecting to MySQL: %v\n", err)
			os.Exit(LOADER_ERROR)
		}
		defer db.Close()

		// Create and initialize collector
		collector := blip.NewCollector(blipCfg, db)
		err = collector.Prepare(*interval, metricsByDomain)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error preparing collector: %v\n", err)
			os.Exit(LOADER_ERROR)
		}
		defer collector.Stop()

		metricsChan = collector.GetMetrics()
	} else {
		// File mode: parse mysqladmin output
		parser := blip.NewFileParser(*statusfile, *varfile)
		err = parser.Initialize(*interval)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error initializing file parser: %v\n", err)
			os.Exit(LOADER_ERROR)
		}

		metricsChan = parser.GetMetrics()
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

	// Create metric cache
	cache := blip.NewMetricCache()

	// Main loop through metrics
	for metrics := range metricsChan {
		// Update cache with new metrics
		cache.Update(metrics)

		// Reprint a header whenever lines == 0
		if linesSinceHeader == 0 {
			for _, headerLn := range view.GetHeader(cache) {
				printOutput(headerLn)
				linesSinceHeader += 1
			}
		}

		// Output data
		for _, dataLn := range view.GetData(cache) {
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
