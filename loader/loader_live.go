package loader

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

const (
	// The commands we send to the mysql cli
	STATUS_QUERY    string = "SELECT VARIABLE_NAME, VARIABLE_VALUE FROM performance_schema.global_status"
	VARIABLES_QUERY string = "SELECT VARIABLE_NAME, VARIABLE_VALUE FROM performance_schema.global_variables"
)

// SHOW output via mysqladmin on a live server
type LiveLoader struct {
	interval time.Duration
	dsn      string
	db       *sql.DB
}

// Create a new SqlLoader
// - dsn https://pkg.go.dev/github.com/go-sql-driver/mysql#Config
// - i:  interval for GetSamples
func NewLiveLoader(dsn string) *LiveLoader {
	ll := &LiveLoader{}
	ll.dsn = dsn

	return ll
}

// Connect to the DB and report any errors
func (l *LiveLoader) Initialize(interval time.Duration, sources []SourceName) error {
	l.interval = interval

	// Open the db connection and confirm it works
	db, err := sql.Open("mysql", l.dsn)
	if err != nil {
		return err
	}
	db.SetMaxOpenConns(1)
	l.db = db

	return nil
}

// Returns a channel where new MyqSamples are collected and sent every l.interval from the l.db connection.
func (l *LiveLoader) GetStateChannel() <-chan StateReader {
	ch := make(chan StateReader)

	// Closure to build the next state and send to down the channel
	var prev_ssp *SampleSet
	generateState := func() {
		ssp := l.getSampleSet()

		state := NewState()
		state.Live = true
		state.SetCurrent(ssp)
		state.SetPrevious(prev_ssp)

		ch <- state
		prev_ssp = ssp
	}

	// Start a ticker in a goroutine to collect samples every l.interval
	ticker := time.NewTicker(l.interval)
	go func() {
		// Generate the first state right away
		generateState()

		// Send another State every tick
		for range ticker.C {
			generateState()
		}
	}()
	return ch
}

// Collects a Sampleset
func (l *LiveLoader) getSampleSet() *SampleSet {
	ssp := NewSampleSet()

	status := l.getSample(STATUS_QUERY)
	variables := l.getSample(VARIABLES_QUERY)

	ssp.SetSample(`status`, status)
	ssp.SetSample(`variables`, variables)

	return ssp
}

// Create a Sample given a query
func (l *LiveLoader) getSample(query string) *Sample {
	sample := NewSample()

	rows, err := l.db.Query(query)
	if err != nil {
		sample.err = fmt.Errorf("cannot run query (%s): %s", query, err)
		return sample
	}
	defer rows.Close()

	for rows.Next() {
		var name, value string
		if err := rows.Scan(&name, &value); err != nil {
			sample.err = fmt.Errorf("Error parsing query results (%s): %s", query, err)
			return sample
		}
		// All data keys are lower case
		sample.Data[strings.ToLower(name)] = value
	}
	return sample
}
