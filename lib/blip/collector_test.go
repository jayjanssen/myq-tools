package blip

import (
	"testing"

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

// Note: TestPrepare requires a real database connection and is tested
// in integration tests rather than unit tests

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
