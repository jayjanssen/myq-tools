//go:build integration
// +build integration

package blip

import (
	"context"
	"testing"
	"time"

	"github.com/jayjanssen/myq-tools/lib/testutil"
)

func TestPrepare_WithWildcards_Integration(t *testing.T) {
	db := testutil.ConnectMySQL(t)
	defer db.Close()

	cfg := testutil.GetTestConfig()

	collector := NewCollector(cfg, db)

	metricsByDomain := map[string][]string{
		"status.global": {"com_*", "threads_running"},
		"var.global":    {"max_connections"},
	}

	err := collector.Prepare(1*time.Second, metricsByDomain)
	if err != nil {
		t.Fatalf("Prepare failed: %v", err)
	}

	// Verify plan was created
	if collector.plan.Name != "myq-tools-plan" {
		t.Errorf("Expected plan name 'myq-tools-plan', got '%s'", collector.plan.Name)
	}

	// Verify domains are in the plan
	if collector.plan.Levels == nil {
		t.Fatal("Plan levels should not be nil")
	}

	level, ok := collector.plan.Levels["default"]
	if !ok {
		t.Fatal("Plan should have 'default' level")
	}

	if level.Collect == nil {
		t.Fatal("Level collect map should not be nil")
	}

	// Verify status.global domain exists
	statusDomain, ok := level.Collect["status.global"]
	if !ok {
		t.Fatal("Plan should have 'status.global' domain")
	}

	// When wildcards are used, "all" option should be set
	if statusDomain.Options == nil || statusDomain.Options["all"] != "yes" {
		t.Error("status.global domain should have 'all' option set to 'yes' when wildcards are used")
	}

	// Verify var.global domain exists
	varDomain, ok := level.Collect["var.global"]
	if !ok {
		t.Fatal("Plan should have 'var.global' domain")
	}

	// max_connections is a specific metric, not a wildcard
	if varDomain.Metrics == nil || len(varDomain.Metrics) == 0 {
		t.Error("var.global domain should have specific metrics when no wildcards are used")
	}
}

func TestPrepare_SpecificMetrics_Integration(t *testing.T) {
	db := testutil.ConnectMySQL(t)
	defer db.Close()

	cfg := testutil.GetTestConfig()

	collector := NewCollector(cfg, db)

	metricsByDomain := map[string][]string{
		"status.global": {"uptime", "threads_connected"},
		"var.global":    {"max_connections", "version"},
	}

	err := collector.Prepare(1*time.Second, metricsByDomain)
	if err != nil {
		t.Fatalf("Prepare failed: %v", err)
	}

	// Verify plan was created
	if collector.plan.Name != "myq-tools-plan" {
		t.Errorf("Expected plan name 'myq-tools-plan', got '%s'", collector.plan.Name)
	}

	// Verify specific metrics are in the plan (not "all")
	level := collector.plan.Levels["default"]
	statusDomain := level.Collect["status.global"]

	if statusDomain.Options != nil && statusDomain.Options["all"] == "yes" {
		t.Error("status.global domain should not have 'all' option when specific metrics are requested")
	}

	if statusDomain.Metrics == nil || len(statusDomain.Metrics) == 0 {
		t.Error("status.global domain should have specific metrics")
	}
}

func TestPrepare_EmptyDomains_Integration(t *testing.T) {
	db := testutil.ConnectMySQL(t)
	defer db.Close()

	cfg := testutil.GetTestConfig()

	collector := NewCollector(cfg, db)

	// Empty domains should fall back to collecting all status and variables
	err := collector.Prepare(1*time.Second, map[string][]string{})
	if err != nil {
		t.Fatalf("Prepare failed: %v", err)
	}

	// Verify fallback domains are in the plan
	level := collector.plan.Levels["default"]

	if _, ok := level.Collect["status.global"]; !ok {
		t.Error("Plan should have 'status.global' domain as fallback")
	}

	if _, ok := level.Collect["var.global"]; !ok {
		t.Error("Plan should have 'var.global' domain as fallback")
	}
}

func TestGetMetrics_Cancellation_Integration(t *testing.T) {
	db := testutil.ConnectMySQL(t)
	defer db.Close()

	cfg := testutil.GetTestConfig()

	collector := NewCollector(cfg, db)

	metricsByDomain := map[string][]string{
		"status.global": {"uptime"},
	}

	err := collector.Prepare(1*time.Second, metricsByDomain)
	if err != nil {
		t.Fatalf("Prepare failed: %v", err)
	}

	// Create a context that will be cancelled after a short time
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Start collecting metrics
	metricsChan := collector.GetMetrics(ctx)

	// Collect at least one metric
	var collectedCount int
	timeout := time.After(3 * time.Second)
	for {
		select {
		case <-ctx.Done():
			// Context was cancelled, should stop collecting
			if collectedCount == 0 {
				t.Error("Expected to collect at least one metric before cancellation")
			}
			return
		case metric, ok := <-metricsChan:
			if !ok {
				// Channel closed
				return
			}
			if metric != nil {
				collectedCount++
			}
		case <-timeout:
			t.Fatal("Test timed out waiting for metrics or cancellation")
		}
	}
}

func TestPrepare_WildcardDetection_Integration(t *testing.T) {
	db := testutil.ConnectMySQL(t)
	defer db.Close()

	cfg := testutil.GetTestConfig()

	collector := NewCollector(cfg, db)

	tests := []struct {
		name            string
		metricsByDomain map[string][]string
		expectAllOption bool
	}{
		{
			name: "wildcard with *",
			metricsByDomain: map[string][]string{
				"status.global": {"com_*"},
			},
			expectAllOption: true,
		},
		{
			name: "wildcard with ^",
			metricsByDomain: map[string][]string{
				"status.global": {"^com_"},
			},
			expectAllOption: true,
		},
		{
			name: "no wildcard",
			metricsByDomain: map[string][]string{
				"status.global": {"uptime", "threads_connected"},
			},
			expectAllOption: false,
		},
		{
			name: "mixed wildcard and specific",
			metricsByDomain: map[string][]string{
				"status.global": {"com_*", "uptime"},
			},
			expectAllOption: true, // Should use "all" if any wildcard is present
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := collector.Prepare(1*time.Second, tt.metricsByDomain)
			if err != nil {
				t.Fatalf("Prepare failed: %v", err)
			}

			level := collector.plan.Levels["default"]
			domain := level.Collect["status.global"]

			hasAllOption := domain.Options != nil && domain.Options["all"] == "yes"
			if hasAllOption != tt.expectAllOption {
				t.Errorf("Expected 'all' option to be %v, got %v", tt.expectAllOption, hasAllOption)
			}
		})
	}
}

func TestCollector_IntervalStorage_Integration(t *testing.T) {
	db := testutil.ConnectMySQL(t)
	defer db.Close()

	cfg := testutil.GetTestConfig()

	collector := NewCollector(cfg, db)

	metricsByDomain := map[string][]string{
		"status.global": {"uptime"},
	}

	interval := 2 * time.Second
	err := collector.Prepare(interval, metricsByDomain)
	if err != nil {
		t.Fatalf("Prepare failed: %v", err)
	}

	// Verify interval is stored correctly
	if collector.interval != interval {
		t.Errorf("Expected interval %v, got %v", interval, collector.interval)
	}

	// Verify start time is set
	if collector.startTime.IsZero() {
		t.Error("Expected start time to be set")
	}

	// Verify collection count starts at 0
	if collector.collectionCount != 0 {
		t.Errorf("Expected collection count to start at 0, got %d", collector.collectionCount)
	}

	// Collect once and verify count increments
	metrics, err := collector.Collect()
	if err != nil {
		t.Fatalf("Collect failed: %v", err)
	}

	if len(metrics) == 0 {
		t.Error("Expected to collect at least one metric")
	}

	if collector.collectionCount != 1 {
		t.Errorf("Expected collection count to be 1, got %d", collector.collectionCount)
	}
}
