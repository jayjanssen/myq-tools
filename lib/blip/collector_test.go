package blip

import (
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

func TestPrepare_WithWildcards(t *testing.T) {
	t.Skip("Requires real database connection - this tests plan building logic")

	cfg := blip.ConfigMonitor{
		MonitorId: "test",
		Hostname:  "localhost:3306",
	}

	collector := NewCollector(cfg, nil)

	metricsByDomain := map[string][]string{
		"status.global": {"com_*", "threads_running"},
		"var.global":    {"max_connections"},
	}

	// Note: This requires a real database connection
	// The test validates plan building logic
	collector.Prepare(1*time.Second, metricsByDomain)

	// Verify plan was created
	if collector.plan.Name != "myq-tools-plan" {
		t.Errorf("Expected plan name 'myq-tools-plan', got '%s'", collector.plan.Name)
	}
}

func TestPrepare_SpecificMetrics(t *testing.T) {
	t.Skip("Requires real database connection - integration test")
}

func TestPrepare_EmptyDomains(t *testing.T) {
	t.Skip("Requires real database connection - integration test")
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

func TestGetMetrics_Cancellation(t *testing.T) {
	t.Skip("Requires real database connection - integration test")
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

func TestPrepare_WildcardDetection(t *testing.T) {
	t.Skip("Requires real database connection - integration test")
}

func TestCollector_IntervalStorage(t *testing.T) {
	t.Skip("Requires real database connection - integration test")
}
