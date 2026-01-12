# View Compatibility Testing Design

## Goal

Validate that myq-status views remain compatible with MySQL versions by ensuring all metrics referenced by each view actually exist in the target MySQL server. This catches deprecated or removed status variables before they break in production.

## Overview

Create integration tests that validate each view's metric requirements against a live MySQL server. Tests will fail immediately if a view requests a metric that doesn't exist, making CI catch compatibility issues early.

## Test Structure

### Location and Build Tags

- **File**: `lib/viewer/view_integration_test.go`
- **Build tag**: `//go:build integration` (matches existing pattern)
- **Execution**: `go test -tags=integration ./lib/viewer/...`

### Main Test Function

`TestViewCompatibility_Integration(t *testing.T)`:

1. Connect to MySQL once (reuse connection for all subtests)
2. Load all default views via `LoadDefaultViews()`
3. Get view list via `ListViews()`
4. Create subtests for each view using `t.Run(viewName, func(t *testing.T) {...})`

### Per-View Subtest Logic

Each subtest:

1. Check skip list - if view is in skip list, call `t.Skipf()` with reason
2. Get the view using `GetViewer(viewName)` and type assert to `View`
3. Call `view.GetMetricsByDomain()` to get required metrics
4. Create new `Collector` with test config
5. Call `collector.Prepare()` with metrics by domain
6. Call `collector.Collect()` once to get actual metrics
7. Validate all required metrics are present in results

## Handling View Variants

### Skip List Approach

Simple map-based skip list for non-standard MySQL views:

```go
// viewsToSkip contains views that should be skipped in MySQL compatibility tests
var viewsToSkip = map[string]string{
    "wsrep": "Percona/Galera specific metrics",
    // Add more as needed
}
```

In each subtest:

```go
if reason, skip := viewsToSkip[viewName]; skip {
    t.Skipf("Skipping view: %s", reason)
}
```

This keeps it explicit and easy to modify. Future Percona testing can use a separate test with different skip list.

## Metric Validation

### Collecting Required Metrics

```go
view := views[viewName].(View)
metricsByDomain := view.GetMetricsByDomain()

cfg := getTestConfig()
collector := NewCollector(cfg, db)

interval := 1 * time.Second
err := collector.Prepare(interval, metricsByDomain)
// error handling

metrics, err := collector.Collect()
// error handling
```

### Building Collected Metrics Map

```go
// Build map: domain -> metric -> exists
collected := make(map[string]map[string]bool)
for _, m := range metrics {
    domain := m.Name  // blip uses Name field for domain
    if collected[domain] == nil {
        collected[domain] = make(map[string]bool)
    }
    for metricName := range m.Values {
        collected[domain][metricName] = true
    }
}
```

### Validation Logic

```go
for domain, requiredMetrics := range metricsByDomain {
    for _, metric := range requiredMetrics {
        if isPattern(metric) {
            // For patterns, verify domain returned some metrics
            if len(collected[domain]) == 0 {
                t.Errorf("Domain %s returned no metrics (needed for pattern %s)",
                    domain, metric)
            }
            continue
        }

        // For specific metrics, validate exact match
        if !collected[domain][metric] {
            t.Errorf("Missing metric: %s/%s", domain, metric)
        }
    }
}
```

## Wildcard Metric Handling

### The Problem

Some views use patterns like `com_*` or `^innodb_` instead of specific metric names. These can't be validated against specific collected metric names.

### Solution

**Pattern detection**:
```go
func isPattern(metric string) bool {
    return strings.Contains(metric, "*") || strings.HasPrefix(metric, "^")
}
```

**Validation strategy**:
- For patterns: Verify the domain returned ANY metrics (proves collector can access the domain)
- For specific metrics: Verify the exact metric name exists in collected results

**How it works**: When the collector sees wildcards, it automatically sets `all: "yes"` in domain options, telling blip to fetch all metrics from that domain. We validate that this actually returned data.

## Error Handling and Helpers

### Connection Management

- Connect once at test function level
- Pass `*sql.DB` to subtests via closure
- Reuse `connectMySQL(t)` and `getTestConfig()` from existing integration tests

### Error Messages

Make them actionable:
- "Missing metric: status.global/innodb_rows_read in view 'innodb' - metric may not exist in this MySQL version"
- "Domain status.global returned no metrics for pattern 'com_*' in view 'commands'"

### Empty View Handling

```go
if len(metricsByDomain) == 0 {
    t.Skip("View has no metrics to validate")
}
```

## Test Scope

### What We Test

- Metrics can be collected from MySQL
- All required metrics for each view exist
- Collector can successfully prepare and collect

### What We Don't Test

- View rendering (GetHeader/GetData formatting) - already covered by unit tests
- Metric value correctness - that's domain-specific logic
- Multiple collection cycles - single collect proves existence

### MySQL Version Testing

- Tests run against single MySQL version per execution
- GitHub Actions matrix handles multiple versions in parallel CI jobs
- Environment variables configure target MySQL (MYSQL_HOST, MYSQL_PORT, etc.)

## Benefits

1. **Catches deprecations early**: CI fails immediately when MySQL removes a status variable
2. **Clear failure messages**: Know exactly which view and which metric failed
3. **Version matrix ready**: Already works with existing GitHub Actions MySQL version matrix
4. **Low maintenance**: Adding new views requires no test code changes
5. **Fast execution**: Single collection per view, reused connection
6. **Focused scope**: Tests metric existence only, not formatting or business logic

## Future Extensions

- Add Percona-specific test with empty skip list
- Test against MariaDB with separate skip list
- Add test that validates skip list views actually fail on standard MySQL
- Generate compatibility matrix documentation from test results
