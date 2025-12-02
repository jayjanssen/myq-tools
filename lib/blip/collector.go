package blip

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/cashapp/blip"
	"github.com/cashapp/blip/metrics"
	"github.com/cashapp/blip/monitor"
)

// Collector wraps blip's monitor.Engine to collect metrics
type Collector struct {
	cfg       blip.ConfigMonitor
	db        *sql.DB
	engine    *monitor.Engine
	plan      blip.Plan
	interval  time.Duration
	levelName string
}

// NewCollector creates a new blip-based collector
func NewCollector(cfg blip.ConfigMonitor, db *sql.DB) *Collector {
	return &Collector{
		cfg:    cfg,
		db:     db,
		engine: monitor.NewEngine(cfg, db),
	}
}

// Prepare initializes the collector with a plan containing all blip domains
func (c *Collector) Prepare(interval time.Duration) error {
	c.interval = interval
	c.levelName = "default"

	// Create a simple plan with all available domains
	c.plan = blip.Plan{
		Name:   "myq-tools-plan",
		Source: "myq-tools",
		Levels: map[string]blip.Level{
			c.levelName: {
				Name: c.levelName,
				Freq: interval.String(),
				Collect: map[string]blip.Domain{
					"status.global": {
						Name:    "status.global",
						Metrics: []string{}, // Empty means collect all
						Options: map[string]string{"all": "yes"},
					},
					"var.global": {
						Name:    "var.global",
						Metrics: []string{}, // Empty means collect all
						Options: map[string]string{"all": "yes"},
					},
					"innodb": {
						Name: "innodb",
					},
					"innodb.buffer-pool": {
						Name: "innodb.buffer-pool",
					},
					"repl": {
						Name: "repl",
					},
					"repl.lag": {
						Name: "repl.lag",
					},
					"size.database": {
						Name: "size.database",
					},
					"size.table": {
						Name: "size.table",
					},
					"size.binlog": {
						Name: "size.binlog",
					},
					"trx": {
						Name: "trx",
					},
					"autoinc": {
						Name: "autoinc",
					},
					"tls": {
						Name: "tls",
					},
					"wait.io.table": {
						Name: "wait.io.table",
					},
					"stmt.current": {
						Name: "stmt.current",
					},
				},
			},
		},
	}

	// Prepare the engine with the plan
	ctx := context.Background()
	before := func() {}
	after := func() {}

	err := c.engine.Prepare(ctx, c.plan, before, after)
	if err != nil {
		return fmt.Errorf("error preparing blip engine: %w", err)
	}

	return nil
}

// Collect collects metrics from all domains and returns them
func (c *Collector) Collect() ([]*blip.Metrics, error) {
	// Create a context with timeout (engine max runtime)
	ctx, cancel := context.WithTimeout(context.Background(), c.interval-500*time.Millisecond)
	defer cancel()

	// Collect metrics for this interval
	startTime := time.Now()
	interval := uint(time.Since(time.Time{}) / c.interval) // Simple interval counter

	metrics, err := c.engine.Collect(ctx, interval, c.levelName, startTime)
	if err != nil && len(metrics) == 0 {
		return nil, fmt.Errorf("error collecting metrics: %w", err)
	}

	// Return metrics (could be partial success with error)
	return metrics, err
}

// Stop stops the collector and cleans up
func (c *Collector) Stop() {
	if c.engine != nil {
		c.engine.Stop()
	}
}

// GetMetrics starts a ticker and returns a channel of metrics
func (c *Collector) GetMetrics() <-chan *blip.Metrics {
	ch := make(chan *blip.Metrics, 1)

	ticker := time.NewTicker(c.interval)
	go func() {
		defer close(ch)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				metricsSlice, err := c.Collect()
				if err != nil {
					// Log error but continue (partial success possible)
					fmt.Fprintf(os.Stderr, "Collection error: %v\n", err)
				}

				// Send all collected metrics to channel
				for _, m := range metricsSlice {
					ch <- m
				}
			}
		}
	}()

	return ch
}

// ListDomains returns all available blip domains
func ListDomains() []string {
	return metrics.List()
}
