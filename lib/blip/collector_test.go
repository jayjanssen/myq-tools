package blip

import (
	"strings"
	"testing"
	"time"

	"github.com/cashapp/blip"
)

func TestNewCollector(t *testing.T) {
	cfg := blip.ConfigMonitor{
		MonitorId: "test",
		Hostname:  "localhost:3306",
		Username:  "testuser",
		Password:  "testpass",
	}

	// Note: We can't actually connect without a real database,
	// but we can test the constructor
	collector := NewCollector(cfg, nil)

	if collector == nil {
		t.Fatal("NewCollector returned nil")
	}

	if collector.cfg.MonitorId != "test" {
		t.Errorf("Expected MonitorId 'test', got '%s'", collector.cfg.MonitorId)
	}

	if collector.engine == nil {
		t.Error("engine not initialized")
	}
}

func TestPrepare_IntervalValidation(t *testing.T) {
	cfg := blip.ConfigMonitor{
		MonitorId: "test",
		Hostname:  "localhost:3306",
	}

	collector := NewCollector(cfg, nil)
	metricsByDomain := map[string][]string{
		"status.global": {"uptime"},
	}

	// Test that exactly 500ms is rejected (would result in 0 timeout)
	err := collector.Prepare(500*time.Millisecond, metricsByDomain)
	if err == nil {
		t.Error("Expected error for interval of exactly 500ms")
	}
	if err != nil && !strings.Contains(err.Error(), "greater than 500ms") {
		t.Errorf("Expected error message about 'greater than 500ms', got: %v", err)
	}

	// Test that less than 500ms is rejected
	err = collector.Prepare(400*time.Millisecond, metricsByDomain)
	if err == nil {
		t.Error("Expected error for interval less than 500ms")
	}

	// Test that greater than 500ms is accepted (we can't actually prepare without DB, but validation should pass)
	// Actually, Prepare will fail when trying to connect, but the interval validation should pass first
	// So we can't fully test this without a DB, but we verified the rejection cases above
}

func TestStop(t *testing.T) {
	cfg := blip.ConfigMonitor{
		MonitorId: "test",
		Hostname:  "localhost:3306",
	}

	collector := NewCollector(cfg, nil)

	// Stop should not panic even if engine is not fully initialized
	collector.Stop()

	// Calling Stop multiple times should be safe
	collector.Stop()
}

func TestListDomains(t *testing.T) {
	domains := ListDomains()

	if len(domains) == 0 {
		t.Error("Expected ListDomains to return at least some domains")
	}

	// Check for some expected common domains
	expectedDomains := []string{"status.global", "var.global"}
	domainMap := make(map[string]bool)
	for _, domain := range domains {
		domainMap[domain] = true
	}

	for _, expected := range expectedDomains {
		if !domainMap[expected] {
			t.Errorf("Expected domain '%s' to be in list", expected)
		}
	}
}
