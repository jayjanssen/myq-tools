//go:build integration
// +build integration

package viewer

import (
	"strings"
	"testing"
	"time"

	"github.com/jayjanssen/myq-tools/lib/blip"
	"github.com/jayjanssen/myq-tools/lib/testutil"
)

/*
Integration tests for view compatibility with MySQL.

Running:

	# Requires MySQL running locally or via environment variables
	export MYSQL_HOST=localhost
	export MYSQL_PORT=3306
	export MYSQL_USER=root
	export MYSQL_PASSWORD=

	go test -tags=integration ./lib/viewer -v

Views in viewsToSkip are skipped (e.g., Percona-specific views like wsrep).
*/

// viewsToSkip contains views that should be skipped in MySQL compatibility tests
var viewsToSkip = map[string]string{
	"wsrep": "Percona/Galera specific metrics",
}

// isPattern returns true if the metric name contains wildcard patterns
func isPattern(metric string) bool {
	return strings.Contains(metric, "*") || strings.HasPrefix(metric, "^")
}

func TestViewCompatibility_Integration(t *testing.T) {
	// Connect to MySQL once for all views
	db := testutil.ConnectMySQL(t)
	if db == nil {
		return
	}
	defer db.Close()

	// Load all default views
	err := LoadDefaultViews()
	if err != nil {
		t.Fatalf("Failed to load views: %v", err)
	}

	// Test each view
	for _, viewName := range ListViews() {
		viewName := viewName // capture for closure
		t.Run(viewName, func(t *testing.T) {
			// Skip check
			if reason, skip := viewsToSkip[viewName]; skip {
				t.Skipf("Skipping view: %s", reason)
			}

			// Get the view and extract required metrics
			viewer, err := GetViewer(viewName)
			if err != nil {
				t.Fatalf("Failed to get viewer: %v", err)
			}

			view, ok := viewer.(View)
			if !ok {
				t.Fatalf("Viewer is not a View type")
			}

			metricsByDomain := view.GetMetricsByDomain()

			// Skip if view has no metrics
			if len(metricsByDomain) == 0 {
				t.Skip("View has no metrics to validate")
			}

			// Create collector and prepare
			cfg := testutil.GetTestConfig()
			collector := blip.NewCollector(cfg, db)
			interval := 1 * time.Second

			err = collector.Prepare(interval, metricsByDomain)
			if err != nil {
				t.Fatalf("Failed to prepare collector: %v", err)
			}

			// Collect metrics once
			metricsSlice, err := collector.Collect()
			if err != nil {
				t.Fatalf("Failed to collect metrics: %v", err)
			}

			// Build map of collected metrics: domain -> metric -> exists
			collected := make(map[string]map[string]bool)
			for _, metrics := range metricsSlice {
				for domain, metricValues := range metrics.Values {
					if collected[domain] == nil {
						collected[domain] = make(map[string]bool)
					}
					for _, mv := range metricValues {
						collected[domain][mv.Name] = true
					}
				}
			}

			// Validate each required metric exists
			for domain, requiredMetrics := range metricsByDomain {
				for _, metric := range requiredMetrics {
					if isPattern(metric) {
						// For patterns, verify domain returned some metrics
						if len(collected[domain]) == 0 {
							t.Errorf("Domain %s returned no metrics (needed for pattern %s) in view '%s'",
								domain, metric, viewName)
						}
						continue
					}

					// For specific metrics, validate exact match
					if !collected[domain][metric] {
						t.Errorf("Missing metric: %s/%s in view '%s' - metric may not exist in this MySQL version",
							domain, metric, viewName)
					}
				}
			}
		})
	}
}
