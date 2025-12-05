package blip

import (
	"testing"
	"time"

	"github.com/cashapp/blip"
)

func TestNewMetricCache(t *testing.T) {
	cache := NewMetricCache()

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
	cache := NewMetricCache()

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
	cache := NewMetricCache()

	// Update with nil should not panic
	cache.Update(nil)

	if cache.HasCurrent() {
		t.Error("cache with nil update should not have current metrics")
	}
}

func TestGetMetric(t *testing.T) {
	cache := NewMetricCache()

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
	cache := NewMetricCache()

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
	cache := NewMetricCache()

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
	cache := NewMetricCache()

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

func TestHasCurrent_HasPrevious(t *testing.T) {
	cache := NewMetricCache()

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
	cache := NewMetricCache()

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
	cache := NewMetricCache()

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
	cache := NewMetricCache()

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
