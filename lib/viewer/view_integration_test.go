//go:build integration
// +build integration

package viewer

import (
	"strings"
	"testing"

	"github.com/jayjanssen/myq-tools/lib/testutil"
)

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

			// Will add validation logic in Task 3
		})
	}
}
