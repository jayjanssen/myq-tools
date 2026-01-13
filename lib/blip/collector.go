package blip

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/cashapp/blip"
	"github.com/cashapp/blip/metrics"
	"github.com/cashapp/blip/monitor"
)

// minIntervalBuffer is the minimum buffer time needed for cleanup operations
// when calculating context timeout in Collect(). The interval must be greater
// than this value to ensure a positive timeout.
const minIntervalBuffer = 500 * time.Millisecond

// Collector wraps blip's monitor.Engine to collect metrics
type Collector struct {
	cfg             blip.ConfigMonitor
	db              *sql.DB
	engine          *monitor.Engine
	plan            blip.Plan
	interval        time.Duration
	levelName       string
	startTime       time.Time
	collectionCount uint
	metricsByDomain map[string][]string
}

// MissingMetricsWarning describes metrics that were requested but not collected
type MissingMetricsWarning struct {
	Domain  string
	Metrics []string
	Hint    string
}

// domainWarnings maps domains to their fix hints for missing metrics
var domainWarnings = map[string]string{
	"innodb": "SET GLOBAL innodb_monitor_enable = '<metric_name>';",
}

// NewCollector creates a new blip-based collector
func NewCollector(cfg blip.ConfigMonitor, db *sql.DB) *Collector {
	return &Collector{
		cfg:    cfg,
		db:     db,
		engine: monitor.NewEngine(cfg, db),
	}
}

// Prepare initializes the collector with a plan for the specified metrics
func (c *Collector) Prepare(interval time.Duration, metricsByDomain map[string][]string) error {
	// Validate interval is greater than minIntervalBuffer to ensure positive context timeout
	// (Collect subtracts minIntervalBuffer for cleanup buffer, so interval must be > minIntervalBuffer)
	if interval <= minIntervalBuffer {
		return fmt.Errorf("interval must be greater than %v, got %s", minIntervalBuffer, interval)
	}

	c.interval = interval
	c.levelName = "default"
	c.startTime = time.Now()
	c.collectionCount = 0
	c.metricsByDomain = metricsByDomain

	// Build the collect map dynamically based on required metrics
	collectMap := make(map[string]blip.Domain)

	for domain, metrics := range metricsByDomain {
		// Check if this domain contains any patterns (wildcards)
		hasPattern := false
		for _, metric := range metrics {
			if strings.Contains(metric, "*") || strings.HasPrefix(metric, "^") {
				hasPattern = true
				break
			}
		}

		// If there are patterns, we need to collect all metrics for that domain
		// because Blip doesn't support pattern matching in the Metrics array
		if hasPattern {
			collectMap[domain] = blip.Domain{
				Name:    domain,
				Options: map[string]string{"all": "yes"},
			}
		} else {
			// Collect only specific metrics
			collectMap[domain] = blip.Domain{
				Name:    domain,
				Metrics: metrics,
				// Don't set "all" option, or set it to "no" - absence means "no"
			}
		}
	}

	// If no domains specified, fall back to collecting all status and variables
	if len(collectMap) == 0 {
		collectMap["status.global"] = blip.Domain{
			Name:    "status.global",
			Options: map[string]string{"all": "yes"},
		}
		collectMap["var.global"] = blip.Domain{
			Name:    "var.global",
			Options: map[string]string{"all": "yes"},
		}
	}

	c.plan = blip.Plan{
		Name:   "myq-tools-plan",
		Source: "myq-tools",
		Levels: map[string]blip.Level{
			c.levelName: {
				Name:    c.levelName,
				Freq:    interval.String(),
				Collect: collectMap,
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
	// Create a context with timeout (leave minIntervalBuffer for cleanup)
	ctx, cancel := context.WithTimeout(context.Background(), c.interval-minIntervalBuffer)
	defer cancel()

	// Collect metrics for this interval
	startTime := time.Now()

	// Increment collection counter for this cycle
	c.collectionCount++
	interval := c.collectionCount

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

// GetMetrics starts a ticker and returns a channel of metrics.
// The goroutine will stop when the context is cancelled.
func (c *Collector) GetMetrics(ctx context.Context) <-chan *blip.Metrics {
	ch := make(chan *blip.Metrics, 1)

	ticker := time.NewTicker(c.interval)
	go func() {
		defer close(ch)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				metricsSlice, err := c.Collect()
				if err != nil {
					// Log error but continue (partial success possible)
					fmt.Fprintf(os.Stderr, "Collection error: %v\n", err)
				}

				// Send all collected metrics to channel
				for _, m := range metricsSlice {
					select {
					case ch <- m:
					case <-ctx.Done():
						return
					}
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

// CheckMissingMetrics compares requested metrics against collected results
// and returns warnings for domains that have known fix hints
func (c *Collector) CheckMissingMetrics(collected []*blip.Metrics) []MissingMetricsWarning {
	var warnings []MissingMetricsWarning

	// Build set of collected metrics by domain
	collectedByDomain := make(map[string]map[string]bool)
	for _, m := range collected {
		for domain, metricValues := range m.Values {
			if collectedByDomain[domain] == nil {
				collectedByDomain[domain] = make(map[string]bool)
			}
			for _, mv := range metricValues {
				collectedByDomain[domain][mv.Name] = true
			}
		}
	}

	// Check each domain that has a warning hint
	for domain, hint := range domainWarnings {
		requestedMetrics, ok := c.metricsByDomain[domain]
		if !ok {
			continue
		}

		var missing []string
		for _, metric := range requestedMetrics {
			if !collectedByDomain[domain][metric] {
				missing = append(missing, metric)
			}
		}

		if len(missing) > 0 {
			warnings = append(warnings, MissingMetricsWarning{
				Domain:  domain,
				Metrics: missing,
				Hint:    hint,
			})
		}
	}

	return warnings
}
