# Missing Metrics Warning Design

**Goal:** Warn users at startup when metrics require manual enabling in MySQL, so they understand why columns show `-` instead of data.

## Approach

- **Timing:** At startup, after collector is prepared but before main display loop
- **Detection:** Domain-based convention - metrics from `innodb` domain (INNODB_METRICS table) require enabling
- **Method:** Probe collection once, compare requested vs. collected metrics
- **Output:** Single-line warning to stderr per domain with missing metrics

## Implementation

### New Types in `lib/blip/collector.go`

```go
// MissingMetricsWarning describes metrics that were requested but not collected
type MissingMetricsWarning struct {
    Domain  string
    Metrics []string
    Hint    string  // Short fix hint
}

// domainWarnings maps domains to their fix hints
var domainWarnings = map[string]string{
    "innodb": "SET GLOBAL innodb_monitor_enable = '<metric_name>';",
}
```

### New Method on Collector

```go
// CheckMissingMetrics compares requested metrics against collected results
// and returns warnings for domains that have known fix hints
func (c *Collector) CheckMissingMetrics(collected []*Metrics) []MissingMetricsWarning
```

The method:
1. Iterates through domains that were requested in `Prepare()`
2. For domains in `domainWarnings` map, checks if metrics were returned
3. Returns `MissingMetricsWarning` for each domain with missing metrics

### Changes to `myq-status/main.go`

After `collector.Prepare()` and before main loop:

```go
// Probe collection to check for missing metrics
metrics, err := collector.Collect()
if err == nil && len(metrics) > 0 {
    cache.Update(metrics[0])

    for _, w := range collector.CheckMissingMetrics(metrics) {
        fmt.Fprintf(os.Stderr, "Warning: %s metrics %v require: %s\n",
            w.Domain, w.Metrics, w.Hint)
    }
}
```

## Output Format

```
Warning: innodb metrics [log_lsn_checkpoint_age, log_lsn_current] require: SET GLOBAL innodb_monitor_enable = '<metric_name>';
```

## Future Extensibility

To add warnings for another domain, add an entry to `domainWarnings` map:

```go
var domainWarnings = map[string]string{
    "innodb": "SET GLOBAL innodb_monitor_enable = '<metric_name>';",
    "newdomain": "Fix hint for new domain",
}
```
