package blip

import (
	"testing"
	"time"

	"github.com/cashapp/blip"
)

func TestNewMetricCache(t *testing.T) {
	cache := NewMetricCache(true)

	if cache == nil {
		t.Fatal("NewMetricCache returned nil")
	}
	if cache.index == nil {
		t.Error("index map not initialized")
	}
	if cache.prevIndex == nil {
		t.Error("prevIndex map not initialized")
	}
	if cache.HasCurrent() {
		t.Error("new cache should not have current metrics")
	}
	if cache.HasPrevious() {
		t.Error("new cache should not have previous metrics")
	}
}

func TestUpdate(t *testing.T) {
	cache := NewMetricCache(true)

	// Create first metrics
	metrics1 := &blip.Metrics{
		Begin: time.Now(),
		End:   time.Now().Add(time.Second),
		Values: map[string][]blip.MetricValue{
			"status.global": {
				{Name: "questions", Value: 100, Type: blip.CUMULATIVE_COUNTER},
				{Name: "threads_running", Value: 5, Type: blip.GAUGE},
			},
		},
	}

	cache.Update(metrics1)

	if !cache.HasCurrent() {
		t.Error("cache should have current metrics after first update")
	}
	if cache.HasPrevious() {
		t.Error("cache should not have previous metrics after first update")
	}

	// Create second metrics
	metrics2 := &blip.Metrics{
		Begin: time.Now().Add(time.Second),
		End:   time.Now().Add(2 * time.Second),
		Values: map[string][]blip.MetricValue{
			"status.global": {
				{Name: "questions", Value: 150, Type: blip.CUMULATIVE_COUNTER},
				{Name: "threads_running", Value: 7, Type: blip.GAUGE},
			},
		},
	}

	cache.Update(metrics2)

	if !cache.HasCurrent() {
		t.Error("cache should have current metrics after second update")
	}
	if !cache.HasPrevious() {
		t.Error("cache should have previous metrics after second update")
	}

	// Verify current values
	if val, ok := cache.GetMetric("status.global", "questions"); !ok || val.Value != 150 {
		t.Errorf("Expected current questions=150, got %v (ok=%v)", val.Value, ok)
	}

	// Verify previous values
	if val, ok := cache.GetPrevMetric("status.global", "questions"); !ok || val.Value != 100 {
		t.Errorf("Expected previous questions=100, got %v (ok=%v)", val.Value, ok)
	}
}

func TestUpdate_NilMetrics(t *testing.T) {
	cache := NewMetricCache(true)

	// Update with nil should not panic
	cache.Update(nil)

	if cache.HasCurrent() {
		t.Error("cache with nil update should not have current metrics")
	}
}

func TestGetMetric(t *testing.T) {
	cache := NewMetricCache(true)

	metrics := &blip.Metrics{
		Begin: time.Now(),
		End:   time.Now().Add(time.Second),
		Values: map[string][]blip.MetricValue{
			"status.global": {
				{Name: "com_select", Value: 100, Type: blip.CUMULATIVE_COUNTER},
				{Name: "threads_running", Value: 5, Type: blip.GAUGE},
			},
			"var.global": {
				{Name: "max_connections", Value: 151, Type: blip.GAUGE},
			},
		},
	}

	cache.Update(metrics)

	// Test existing metric
	if val, ok := cache.GetMetric("status.global", "com_select"); !ok {
		t.Error("Expected to find com_select")
	} else if val.Value != 100 {
		t.Errorf("Expected com_select=100, got %v", val.Value)
	}

	// Test non-existing metric
	if _, ok := cache.GetMetric("status.global", "nonexistent"); ok {
		t.Error("Expected not to find nonexistent metric")
	}

	// Test non-existing domain
	if _, ok := cache.GetMetric("nonexistent.domain", "metric"); ok {
		t.Error("Expected not to find metric in nonexistent domain")
	}
}

func TestGetPrevMetric(t *testing.T) {
	cache := NewMetricCache(true)

	metrics1 := &blip.Metrics{
		Begin: time.Now(),
		End:   time.Now().Add(time.Second),
		Values: map[string][]blip.MetricValue{
			"status.global": {
				{Name: "questions", Value: 100, Type: blip.CUMULATIVE_COUNTER},
			},
		},
	}

	metrics2 := &blip.Metrics{
		Begin: time.Now().Add(time.Second),
		End:   time.Now().Add(2 * time.Second),
		Values: map[string][]blip.MetricValue{
			"status.global": {
				{Name: "questions", Value: 200, Type: blip.CUMULATIVE_COUNTER},
			},
		},
	}

	cache.Update(metrics1)
	cache.Update(metrics2)

	// Test previous metric exists
	if val, ok := cache.GetPrevMetric("status.global", "questions"); !ok {
		t.Error("Expected to find previous questions")
	} else if val.Value != 100 {
		t.Errorf("Expected previous questions=100, got %v", val.Value)
	}

	// Test non-existing previous metric
	if _, ok := cache.GetPrevMetric("status.global", "nonexistent"); ok {
		t.Error("Expected not to find nonexistent previous metric")
	}
}

func TestGetMetricValue(t *testing.T) {
	cache := NewMetricCache(true)

	metrics := &blip.Metrics{
		Begin: time.Now(),
		End:   time.Now().Add(time.Second),
		Values: map[string][]blip.MetricValue{
			"status.global": {
				{Name: "com_select", Value: 42, Type: blip.CUMULATIVE_COUNTER},
			},
		},
	}

	cache.Update(metrics)

	// Test existing metric
	val := cache.GetMetricValue("status.global", "com_select")
	if val != 42 {
		t.Errorf("Expected com_select=42, got %v", val)
	}

	// Test non-existing metric returns 0
	val = cache.GetMetricValue("status.global", "nonexistent")
	if val != 0 {
		t.Errorf("Expected 0 for nonexistent metric, got %v", val)
	}
}

func TestSecondsDiff(t *testing.T) {
	cache := NewMetricCache(true)

	start := time.Now()
	metrics1 := &blip.Metrics{
		Begin: start,
		End:   start.Add(time.Second),
	}

	metrics2 := &blip.Metrics{
		Begin: start.Add(time.Second),
		End:   start.Add(3 * time.Second),
	}

	// Before any updates
	if diff := cache.SecondsDiff(); diff != 0 {
		t.Errorf("Expected 0 seconds diff before updates, got %v", diff)
	}

	// After first update
	cache.Update(metrics1)
	if diff := cache.SecondsDiff(); diff != 0 {
		t.Errorf("Expected 0 seconds diff with only current, got %v", diff)
	}

	// After second update
	cache.Update(metrics2)
	diff := cache.SecondsDiff()
	if diff != 2.0 {
		t.Errorf("Expected 2.0 seconds diff, got %v", diff)
	}
}

func TestSecondsDiff_FileMode_UptimeBased(t *testing.T) {
	// Bug 2 fix: In file mode, SecondsDiff should use uptime differences,
	// not timestamp differences, to handle irregular sample intervals correctly
	cache := NewMetricCache(false) // File mode

	startTime := time.Date(2024, 1, 1, 14, 30, 0, 0, time.UTC)

	// First sample: uptime=100s, synthetic timestamp increments by 5s
	metrics1 := &blip.Metrics{
		Begin: startTime,
		End:   startTime.Add(5 * time.Second), // Configured interval is 5s
		Values: map[string][]blip.MetricValue{
			"status.global": {
				{Name: "uptime", Value: 100, Type: blip.GAUGE},
				{Name: "questions", Value: 1000, Type: blip.CUMULATIVE_COUNTER},
			},
		},
	}

	cache.Update(metrics1)

	// Second sample: uptime=110s (actual diff is 10s), but synthetic timestamp increments by 5s
	// This simulates a case where actual sample interval differs from configured interval
	metrics2 := &blip.Metrics{
		Begin: startTime.Add(5 * time.Second),
		End:   startTime.Add(10 * time.Second), // Synthetic timestamp: +5s
		Values: map[string][]blip.MetricValue{
			"status.global": {
				{Name: "uptime", Value: 110, Type: blip.GAUGE}, // Actual uptime diff: +10s
				{Name: "questions", Value: 2000, Type: blip.CUMULATIVE_COUNTER},
			},
		},
	}

	cache.Update(metrics2)

	diff := cache.SecondsDiff()
	// Should use uptime difference (10s), not timestamp difference (5s)
	if diff != 10.0 {
		t.Errorf("Expected 10.0 seconds diff (uptime-based), got %v", diff)
	}

	// Third sample: uptime=115s (actual diff is 5s), synthetic timestamp increments by 5s
	metrics3 := &blip.Metrics{
		Begin: startTime.Add(10 * time.Second),
		End:   startTime.Add(15 * time.Second), // Synthetic timestamp: +5s
		Values: map[string][]blip.MetricValue{
			"status.global": {
				{Name: "uptime", Value: 115, Type: blip.GAUGE}, // Actual uptime diff: +5s
				{Name: "questions", Value: 2500, Type: blip.CUMULATIVE_COUNTER},
			},
		},
	}

	cache.Update(metrics3)

	diff = cache.SecondsDiff()
	// Should use uptime difference (5s), not timestamp difference (5s)
	// In this case they match, but the logic should still use uptime
	if diff != 5.0 {
		t.Errorf("Expected 5.0 seconds diff (uptime-based), got %v", diff)
	}
}

func TestSecondsDiff_FileMode_MissingUptime(t *testing.T) {
	// Test fallback to timestamp difference when uptime is missing
	cache := NewMetricCache(false) // File mode

	startTime := time.Date(2024, 1, 1, 14, 30, 0, 0, time.UTC)

	// First sample without uptime
	metrics1 := &blip.Metrics{
		Begin: startTime,
		End:   startTime.Add(5 * time.Second),
		Values: map[string][]blip.MetricValue{
			"status.global": {
				{Name: "questions", Value: 1000, Type: blip.CUMULATIVE_COUNTER},
			},
		},
	}

	cache.Update(metrics1)

	// Second sample without uptime
	metrics2 := &blip.Metrics{
		Begin: startTime.Add(5 * time.Second),
		End:   startTime.Add(10 * time.Second),
		Values: map[string][]blip.MetricValue{
			"status.global": {
				{Name: "questions", Value: 2000, Type: blip.CUMULATIVE_COUNTER},
			},
		},
	}

	cache.Update(metrics2)

	diff := cache.SecondsDiff()
	// Should fallback to timestamp difference (5s) when uptime is missing
	if diff != 5.0 {
		t.Errorf("Expected 5.0 seconds diff (timestamp fallback), got %v", diff)
	}
}

func TestHasCurrent_HasPrevious(t *testing.T) {
	cache := NewMetricCache(true)

	// Initially empty
	if cache.HasCurrent() {
		t.Error("new cache should not have current")
	}
	if cache.HasPrevious() {
		t.Error("new cache should not have previous")
	}

	// After first update
	metrics1 := &blip.Metrics{Begin: time.Now(), End: time.Now()}
	cache.Update(metrics1)

	if !cache.HasCurrent() {
		t.Error("cache should have current after first update")
	}
	if cache.HasPrevious() {
		t.Error("cache should not have previous after first update")
	}

	// After second update
	metrics2 := &blip.Metrics{Begin: time.Now(), End: time.Now()}
	cache.Update(metrics2)

	if !cache.HasCurrent() {
		t.Error("cache should have current after second update")
	}
	if !cache.HasPrevious() {
		t.Error("cache should have previous after second update")
	}
}

func TestFindMetrics(t *testing.T) {
	cache := NewMetricCache(true)

	metrics := &blip.Metrics{
		Begin: time.Now(),
		End:   time.Now().Add(time.Second),
		Values: map[string][]blip.MetricValue{
			"status.global": {
				{Name: "com_select", Value: 100, Type: blip.CUMULATIVE_COUNTER},
				{Name: "com_insert", Value: 50, Type: blip.CUMULATIVE_COUNTER},
				{Name: "com_update", Value: 25, Type: blip.CUMULATIVE_COUNTER},
				{Name: "threads_running", Value: 5, Type: blip.GAUGE},
			},
		},
	}

	cache.Update(metrics)

	// Test wildcard pattern
	results := cache.FindMetrics("status.global", "com_*")
	if len(results) != 3 {
		t.Errorf("Expected 3 com_* metrics, got %d", len(results))
	}

	// Test exact match
	results = cache.FindMetrics("status.global", "threads_running")
	if len(results) != 1 {
		t.Errorf("Expected 1 threads_running metric, got %d", len(results))
	}

	// Test no match
	results = cache.FindMetrics("status.global", "nonexistent_*")
	if len(results) != 0 {
		t.Errorf("Expected 0 matches for nonexistent_*, got %d", len(results))
	}

	// Test non-existent domain
	results = cache.FindMetrics("nonexistent.domain", "com_*")
	if len(results) != 0 {
		t.Errorf("Expected 0 matches for nonexistent domain, got %d", len(results))
	}
}

func TestMatchPattern(t *testing.T) {
	tests := []struct {
		name     string
		pattern  string
		testName string
		expected bool
	}{
		// Exact match
		{"exact match", "com_select", "com_select", true},
		{"exact no match", "com_select", "com_insert", false},

		// Wildcard
		{"wildcard prefix match", "com_*", "com_select", true},
		{"wildcard prefix match 2", "com_*", "com_insert", true},
		{"wildcard no match", "com_*", "threads_running", false},
		{"wildcard empty", "*", "anything", true},

		// Caret (treated as prefix)
		{"caret prefix", "^com_", "com_select", true},
		{"caret no match", "^threads", "com_select", false},

		// Edge cases
		{"empty pattern", "", "", true},
		{"pattern longer than name", "very_long_pattern*", "short", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matchPattern(tt.testName, tt.pattern)
			if result != tt.expected {
				t.Errorf("matchPattern(%q, %q) = %v, expected %v",
					tt.testName, tt.pattern, result, tt.expected)
			}
		})
	}
}

func TestDomainExists(t *testing.T) {
	cache := NewMetricCache(true)

	metrics := &blip.Metrics{
		Begin: time.Now(),
		End:   time.Now().Add(time.Second),
		Values: map[string][]blip.MetricValue{
			"status.global": {
				{Name: "questions", Value: 100, Type: blip.CUMULATIVE_COUNTER},
			},
			"var.global": {
				{Name: "max_connections", Value: 151, Type: blip.GAUGE},
			},
		},
	}

	cache.Update(metrics)

	// Test existing domains
	if !cache.DomainExists("status.global") {
		t.Error("Expected status.global to exist")
	}
	if !cache.DomainExists("var.global") {
		t.Error("Expected var.global to exist")
	}

	// Test non-existing domain
	if cache.DomainExists("nonexistent.domain") {
		t.Error("Expected nonexistent.domain not to exist")
	}
}

func TestGetAllDomains(t *testing.T) {
	cache := NewMetricCache(true)

	// Empty cache
	domains := cache.GetAllDomains()
	if len(domains) != 0 {
		t.Errorf("Expected 0 domains in empty cache, got %d", len(domains))
	}

	// With metrics
	metrics := &blip.Metrics{
		Begin: time.Now(),
		End:   time.Now().Add(time.Second),
		Values: map[string][]blip.MetricValue{
			"status.global": {
				{Name: "questions", Value: 100, Type: blip.CUMULATIVE_COUNTER},
			},
			"var.global": {
				{Name: "max_connections", Value: 151, Type: blip.GAUGE},
			},
			"innodb": {
				{Name: "buffer_pool_size", Value: 134217728, Type: blip.GAUGE},
			},
		},
	}

	cache.Update(metrics)

	domains = cache.GetAllDomains()
	if len(domains) != 3 {
		t.Errorf("Expected 3 domains, got %d", len(domains))
	}

	// Verify all domains are present
	domainMap := make(map[string]bool)
	for _, domain := range domains {
		domainMap[domain] = true
	}

	if !domainMap["status.global"] {
		t.Error("Expected status.global in domains list")
	}
	if !domainMap["var.global"] {
		t.Error("Expected var.global in domains list")
	}
	if !domainMap["innodb"] {
		t.Error("Expected innodb in domains list")
	}
}

func TestGetTimeString_LiveMode(t *testing.T) {
	cache := NewMetricCache(true) // Live mode

	testTime := time.Date(2024, 1, 1, 14, 30, 45, 0, time.UTC)
	metrics := &blip.Metrics{
		Begin: testTime,
		End:   testTime.Add(time.Second),
		Values: map[string][]blip.MetricValue{
			"status.global": {
				{Name: "questions", Value: 100, Type: blip.CUMULATIVE_COUNTER},
			},
		},
	}

	cache.Update(metrics)

	timeStr := cache.GetTimeString()
	expected := "14:30:45"
	if timeStr != expected {
		t.Errorf("Expected time string %q, got %q", expected, timeStr)
	}
}

func TestGetTimeString_FileMode(t *testing.T) {
	cache := NewMetricCache(false) // File mode

	startTime := time.Date(2024, 1, 1, 14, 30, 0, 0, time.UTC)

	// First sample with uptime=100s
	metrics1 := &blip.Metrics{
		Begin: startTime,
		End:   startTime.Add(time.Second),
		Values: map[string][]blip.MetricValue{
			"status.global": {
				{Name: "questions", Value: 100, Type: blip.CUMULATIVE_COUNTER},
				{Name: "uptime", Value: 100, Type: blip.GAUGE},
			},
		},
	}

	cache.Update(metrics1)
	timeStr := cache.GetTimeString()
	if timeStr != "0s" {
		t.Errorf("Expected time string %q for first sample, got %q", "0s", timeStr)
	}

	// Second sample with uptime=101s (1 second later)
	metrics2 := &blip.Metrics{
		Begin: startTime.Add(1 * time.Second),
		End:   startTime.Add(2 * time.Second),
		Values: map[string][]blip.MetricValue{
			"status.global": {
				{Name: "questions", Value: 200, Type: blip.CUMULATIVE_COUNTER},
				{Name: "uptime", Value: 101, Type: blip.GAUGE},
			},
		},
	}

	cache.Update(metrics2)
	timeStr = cache.GetTimeString()
	if timeStr != "1s" {
		t.Errorf("Expected time string %q for second sample, got %q", "1s", timeStr)
	}

	// Third sample with uptime=165s (65 seconds after first = 1m5s)
	metrics3 := &blip.Metrics{
		Begin: startTime.Add(65 * time.Second),
		End:   startTime.Add(66 * time.Second),
		Values: map[string][]blip.MetricValue{
			"status.global": {
				{Name: "questions", Value: 300, Type: blip.CUMULATIVE_COUNTER},
				{Name: "uptime", Value: 165, Type: blip.GAUGE},
			},
		},
	}

	cache.Update(metrics3)
	timeStr = cache.GetTimeString()
	if timeStr != "1m5s" {
		t.Errorf("Expected time string %q for third sample, got %q", "1m5s", timeStr)
	}

	// Fourth sample with uptime=200s (100 seconds after first = 1m40s)
	metrics4 := &blip.Metrics{
		Begin: startTime.Add(100 * time.Second),
		End:   startTime.Add(101 * time.Second),
		Values: map[string][]blip.MetricValue{
			"status.global": {
				{Name: "questions", Value: 400, Type: blip.CUMULATIVE_COUNTER},
				{Name: "uptime", Value: 200, Type: blip.GAUGE},
			},
		},
	}

	cache.Update(metrics4)
	timeStr = cache.GetTimeString()
	if timeStr != "1m40s" {
		t.Errorf("Expected time string %q for fourth sample, got %q", "1m40s", timeStr)
	}
}

func TestGetTimeString_FileMode_NonSequentialUptime(t *testing.T) {
	cache := NewMetricCache(false) // File mode

	startTime := time.Date(2024, 1, 1, 14, 30, 0, 0, time.UTC)

	// First sample with uptime=1000s
	metrics1 := &blip.Metrics{
		Begin: startTime,
		End:   startTime.Add(time.Second),
		Values: map[string][]blip.MetricValue{
			"status.global": {
				{Name: "uptime", Value: 1000, Type: blip.GAUGE},
			},
		},
	}

	cache.Update(metrics1)
	timeStr := cache.GetTimeString()
	if timeStr != "0s" {
		t.Errorf("Expected time string %q for first sample, got %q", "0s", timeStr)
	}

	// Second sample with uptime=1005s (5 seconds later, not 1 second)
	metrics2 := &blip.Metrics{
		Begin: startTime.Add(5 * time.Second),
		End:   startTime.Add(6 * time.Second),
		Values: map[string][]blip.MetricValue{
			"status.global": {
				{Name: "uptime", Value: 1005, Type: blip.GAUGE},
			},
		},
	}

	cache.Update(metrics2)
	timeStr = cache.GetTimeString()
	if timeStr != "5s" {
		t.Errorf("Expected time string %q for second sample (uptime-based), got %q", "5s", timeStr)
	}

	// Third sample with uptime=1070s (70 seconds after first = 1m10s)
	metrics3 := &blip.Metrics{
		Begin: startTime.Add(70 * time.Second),
		End:   startTime.Add(71 * time.Second),
		Values: map[string][]blip.MetricValue{
			"status.global": {
				{Name: "uptime", Value: 1070, Type: blip.GAUGE},
			},
		},
	}

	cache.Update(metrics3)
	timeStr = cache.GetTimeString()
	if timeStr != "1m10s" {
		t.Errorf("Expected time string %q for third sample (uptime-based), got %q", "1m10s", timeStr)
	}
}

func TestGetTimeString_FileMode_FirstUptimeZero(t *testing.T) {
	// Test case: First sample has uptime = 0 (server just started)
	// Subsequent samples should not overwrite firstUptime
	cache := NewMetricCache(false)

	// First sample with uptime = 0
	cache.Update(&blip.Metrics{
		Begin: time.Now(),
		End:   time.Now(),
		Values: map[string][]blip.MetricValue{
			"status.global": {
				{Name: "uptime", Value: 0, Type: blip.GAUGE},
			},
		},
	})

	firstTime := cache.GetTimeString()
	if firstTime != "0s" {
		t.Errorf("Expected '0s' for first sample with uptime=0, got '%s'", firstTime)
	}

	// Second sample with uptime = 5
	cache.Update(&blip.Metrics{
		Begin: time.Now().Add(1 * time.Second),
		End:   time.Now().Add(1 * time.Second),
		Values: map[string][]blip.MetricValue{
			"status.global": {
				{Name: "uptime", Value: 5, Type: blip.GAUGE},
			},
		},
	})

	secondTime := cache.GetTimeString()
	// Should show 5s elapsed (5 - 0), not reset to 0s
	if secondTime != "5s" {
		t.Errorf("Expected '5s' for second sample (5-0), got '%s'", secondTime)
	}

	// Third sample with uptime = 10
	cache.Update(&blip.Metrics{
		Begin: time.Now().Add(2 * time.Second),
		End:   time.Now().Add(2 * time.Second),
		Values: map[string][]blip.MetricValue{
			"status.global": {
				{Name: "uptime", Value: 10, Type: blip.GAUGE},
			},
		},
	})

	thirdTime := cache.GetTimeString()
	// Should show 10s elapsed (10 - 0), not reset
	if thirdTime != "10s" {
		t.Errorf("Expected '10s' for third sample (10-0), got '%s'", thirdTime)
	}
}

func TestGetTimeString_FileMode_NoUptime(t *testing.T) {
	cache := NewMetricCache(false) // File mode

	startTime := time.Date(2024, 1, 1, 14, 30, 0, 0, time.UTC)

	// Sample without uptime metric
	metrics1 := &blip.Metrics{
		Begin: startTime,
		End:   startTime.Add(time.Second),
		Values: map[string][]blip.MetricValue{
			"status.global": {
				{Name: "questions", Value: 100, Type: blip.CUMULATIVE_COUNTER},
			},
		},
	}

	cache.Update(metrics1)
	timeStr := cache.GetTimeString()
	// When uptime is missing, both current and first uptime are 0, so difference is 0
	if timeStr != "0s" {
		t.Errorf("Expected time string %q when uptime is missing, got %q", "0s", timeStr)
	}
}
