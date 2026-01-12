//go:build integration
// +build integration

package viewer

import (
	"strings"
)

// viewsToSkip contains views that should be skipped in MySQL compatibility tests
var viewsToSkip = map[string]string{
	"wsrep": "Percona/Galera specific metrics",
}

// isPattern returns true if the metric name contains wildcard patterns
func isPattern(metric string) bool {
	return strings.Contains(metric, "*") || strings.HasPrefix(metric, "^")
}
