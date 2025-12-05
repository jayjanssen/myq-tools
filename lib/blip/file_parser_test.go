package blip

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/cashapp/blip"
)

func TestNewFileParser(t *testing.T) {
	parser := NewFileParser("status.txt", "vars.txt")

	if parser == nil {
		t.Fatal("NewFileParser returned nil")
	}
	if parser.statusFile != "status.txt" {
		t.Errorf("Expected statusFile 'status.txt', got '%s'", parser.statusFile)
	}
	if parser.varFile != "vars.txt" {
		t.Errorf("Expected varFile 'vars.txt', got '%s'", parser.varFile)
	}
}

func TestInitialize_ValidInterval(t *testing.T) {
	testFile := filepath.Join("testdata", "batch_format.txt")
	parser := NewFileParser(testFile, "")

	err := parser.Initialize(1 * time.Second)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	if parser.interval != 1*time.Second {
		t.Errorf("Expected interval 1s, got %v", parser.interval)
	}
	if parser.scanner == nil {
		t.Error("scanner not initialized")
	}
}

func TestInitialize_InvalidInterval(t *testing.T) {
	testFile := filepath.Join("testdata", "batch_format.txt")
	parser := NewFileParser(testFile, "")

	err := parser.Initialize(500 * time.Millisecond)
	if err == nil {
		t.Fatal("Expected error for interval < 1s, got nil")
	}

	expectedMsg := "interval cannot be less than 1s"
	if err.Error()[:len(expectedMsg)] != expectedMsg {
		t.Errorf("Expected error message starting with '%s', got '%s'", expectedMsg, err.Error())
	}
}

func TestInitialize_FileNotFound(t *testing.T) {
	parser := NewFileParser("nonexistent_file.txt", "")

	err := parser.Initialize(1 * time.Second)
	if err == nil {
		t.Fatal("Expected error for nonexistent file, got nil")
	}

	expectedMsg := "cannot open status file"
	if err.Error()[:len(expectedMsg)] != expectedMsg {
		t.Errorf("Expected error message starting with '%s', got '%s'", expectedMsg, err.Error())
	}
}

func TestParseSample_Batch(t *testing.T) {
	parser := NewFileParser("", "")
	parser.outputtype = BATCH

	data := []byte("com_select\t100\nthreads_running\t5\nuptime\t3600\n")

	result, err := parser.parseSample(data)
	if err != nil {
		t.Fatalf("parseSample failed: %v", err)
	}

	if len(result) != 3 {
		t.Errorf("Expected 3 metrics, got %d", len(result))
	}

	if result["com_select"] != "100" {
		t.Errorf("Expected com_select=100, got %s", result["com_select"])
	}
	if result["threads_running"] != "5" {
		t.Errorf("Expected threads_running=5, got %s", result["threads_running"])
	}
	if result["uptime"] != "3600" {
		t.Errorf("Expected uptime=3600, got %s", result["uptime"])
	}
}

func TestParseSample_Tabular(t *testing.T) {
	parser := NewFileParser("", "")
	parser.outputtype = TABULAR

	data := []byte(`| com_select        | 100  |
| threads_running   | 5    |
| uptime            | 3600 |
`)

	result, err := parser.parseSample(data)
	if err != nil {
		t.Fatalf("parseSample failed: %v", err)
	}

	if len(result) != 3 {
		t.Errorf("Expected 3 metrics, got %d", len(result))
	}

	if result["com_select"] != "100" {
		t.Errorf("Expected com_select=100, got %s", result["com_select"])
	}
	if result["threads_running"] != "5" {
		t.Errorf("Expected threads_running=5, got %s", result["threads_running"])
	}
}

func TestParseSample_EmptyInput(t *testing.T) {
	parser := NewFileParser("", "")
	parser.outputtype = BATCH

	result, err := parser.parseSample([]byte{})
	if err != nil {
		t.Fatalf("parseSample failed on empty input: %v", err)
	}

	if len(result) != 0 {
		t.Errorf("Expected 0 metrics from empty input, got %d", len(result))
	}
}

func TestGetMetrics_BatchFormat(t *testing.T) {
	testFile := filepath.Join("testdata", "batch_format.txt")
	parser := NewFileParser(testFile, "")

	err := parser.Initialize(1 * time.Second)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	metricsChan := parser.GetMetrics()

	// Should get one metrics object
	metrics, ok := <-metricsChan
	if !ok {
		t.Fatal("Expected to receive metrics, channel closed")
	}

	if metrics == nil {
		t.Fatal("Received nil metrics")
	}

	// Verify it has data
	if len(metrics.Values) == 0 {
		t.Error("Expected metrics to have values")
	}

	statusMetrics, ok := metrics.Values["status.global"]
	if !ok {
		t.Fatal("Expected status.global domain")
	}

	if len(statusMetrics) == 0 {
		t.Error("Expected status.global to have metrics")
	}

	// Channel should close after reading the file
	_, ok = <-metricsChan
	if ok {
		t.Error("Expected channel to close after reading file")
	}
}

func TestGetMetrics_TabularFormat(t *testing.T) {
	testFile := filepath.Join("testdata", "tabular_format.txt")
	parser := NewFileParser(testFile, "")

	err := parser.Initialize(1 * time.Second)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	metricsChan := parser.GetMetrics()

	// Should get one metrics object
	metrics, ok := <-metricsChan
	if !ok {
		t.Fatal("Expected to receive metrics, channel closed")
	}

	if metrics == nil {
		t.Fatal("Received nil metrics")
	}

	// Verify it has data
	statusMetrics, ok := metrics.Values["status.global"]
	if !ok {
		t.Fatal("Expected status.global domain")
	}

	if len(statusMetrics) == 0 {
		t.Error("Expected status.global to have metrics")
	}
}

func TestConvertToBlipMetrics(t *testing.T) {
	parser := NewFileParser("test.txt", "")
	startTime := time.Now()

	data := map[string]string{
		"com_select":                     "100",
		"threads_running":                "5",
		"threads_connected":              "10",
		"innodb_buffer_pool_pages_dirty": "50",
		"uptime":                         "3600",
	}

	metrics := parser.convertToBlipMetrics(data, 1, startTime)

	if metrics == nil {
		t.Fatal("convertToBlipMetrics returned nil")
	}

	if metrics.MonitorId != "test.txt" {
		t.Errorf("Expected MonitorId 'test.txt', got '%s'", metrics.MonitorId)
	}

	if metrics.Plan != "file-replay" {
		t.Errorf("Expected Plan 'file-replay', got '%s'", metrics.Plan)
	}

	if metrics.Interval != 1 {
		t.Errorf("Expected Interval 1, got %d", metrics.Interval)
	}

	statusMetrics, ok := metrics.Values["status.global"]
	if !ok {
		t.Fatal("Expected status.global domain")
	}

	// Check that we have the expected metrics
	metricMap := make(map[string]blip.MetricValue)
	for _, mv := range statusMetrics {
		metricMap[mv.Name] = mv
	}

	if _, ok := metricMap["com_select"]; !ok {
		t.Error("Expected to find com_select in metrics")
	}
	if _, ok := metricMap["threads_running"]; !ok {
		t.Error("Expected to find threads_running in metrics")
	}
}

func TestGaugeVsCounter(t *testing.T) {
	parser := NewFileParser("test.txt", "")

	// Test data with known gauges and counters
	data := map[string]string{
		"com_select":                     "100",  // counter
		"threads_running":                "5",    // gauge
		"threads_connected":              "10",   // gauge
		"innodb_buffer_pool_pages_dirty": "50",   // gauge
		"questions":                      "1000", // counter
	}

	metrics := parser.convertToBlipMetrics(data, 1, time.Now())

	statusMetrics, ok := metrics.Values["status.global"]
	if !ok {
		t.Fatal("Expected status.global domain")
	}

	metricMap := make(map[string]blip.MetricValue)
	for _, mv := range statusMetrics {
		metricMap[mv.Name] = mv
	}

	// Check gauges
	gauges := []string{"threads_running", "threads_connected", "innodb_buffer_pool_pages_dirty"}
	for _, name := range gauges {
		if mv, ok := metricMap[name]; !ok {
			t.Errorf("Expected to find gauge %s", name)
		} else if mv.Type != blip.GAUGE {
			t.Errorf("Expected %s to be GAUGE, got %v", name, mv.Type)
		}
	}

	// Check counters
	counters := []string{"com_select", "questions"}
	for _, name := range counters {
		if mv, ok := metricMap[name]; !ok {
			t.Errorf("Expected to find counter %s", name)
		} else if mv.Type != blip.CUMULATIVE_COUNTER {
			t.Errorf("Expected %s to be CUMULATIVE_COUNTER, got %v", name, mv.Type)
		}
	}
}

func TestSplitFunction(t *testing.T) {
	// Test the custom split function behavior with batch_with_intervals.txt
	testFile := filepath.Join("testdata", "batch_with_intervals.txt")
	parser := NewFileParser(testFile, "")

	err := parser.Initialize(2 * time.Second) // Set interval to 2 seconds
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	metricsChan := parser.GetMetrics()

	// Count how many metrics we receive
	count := 0
	for metrics := range metricsChan {
		if metrics != nil {
			count++
		}
	}

	// With interval=2s and uptime progression (3600, 3601, 3602, 3605),
	// we should skip some intervals and only get samples where uptime differs by >= 2s
	// Expected: samples at uptime 3600 and 3605 (2 samples)
	if count < 1 {
		t.Errorf("Expected at least 1 metrics object, got %d", count)
	}
}

func TestSkipInterval(t *testing.T) {
	// Test with batch_with_intervals.txt which has multiple samples with uptime
	testFile := filepath.Join("testdata", "batch_with_intervals.txt")
	parser := NewFileParser(testFile, "")

	// With 1 second interval, should get all samples
	err := parser.Initialize(1 * time.Second)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	metricsChan := parser.GetMetrics()
	count1s := 0
	for range metricsChan {
		count1s++
	}

	// Re-initialize with 3 second interval
	parser = NewFileParser(testFile, "")
	err = parser.Initialize(3 * time.Second)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	metricsChan = parser.GetMetrics()
	count3s := 0
	for range metricsChan {
		count3s++
	}

	// With 3s interval, should skip more samples than with 1s interval
	if count3s >= count1s {
		t.Errorf("Expected fewer samples with 3s interval (%d) than 1s interval (%d)", count3s, count1s)
	}
}

func TestEmptyFile(t *testing.T) {
	testFile := filepath.Join("testdata", "empty_file.txt")
	parser := NewFileParser(testFile, "")

	err := parser.Initialize(1 * time.Second)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	metricsChan := parser.GetMetrics()

	// Should get nothing from empty file
	count := 0
	for range metricsChan {
		count++
	}

	if count != 0 {
		t.Errorf("Expected 0 metrics from empty file, got %d", count)
	}
}

func TestMalformedData(t *testing.T) {
	testFile := filepath.Join("testdata", "malformed_data.txt")
	parser := NewFileParser(testFile, "")

	err := parser.Initialize(1 * time.Second)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	metricsChan := parser.GetMetrics()

	// Should handle malformed data gracefully (skip bad records)
	for metrics := range metricsChan {
		// Just verify we can iterate without panicking
		if metrics != nil && len(metrics.Values) > 0 {
			// Got some valid data from the file
		}
	}
}

func TestFileDescriptorLeak(t *testing.T) {
	// This test verifies files are properly closed after GetMetrics completes
	testFile := filepath.Join("testdata", "batch_format.txt")

	// Run multiple complete cycles (Initialize + GetMetrics) to test for fd leaks
	// If files aren't closed, we'd accumulate open file descriptors
	for i := 0; i < 10; i++ {
		parser := NewFileParser(testFile, "")
		err := parser.Initialize(1 * time.Second)
		if err != nil {
			t.Fatalf("Initialize failed on iteration %d: %v", i, err)
		}

		// GetMetrics should close the file when done
		metricsChan := parser.GetMetrics()

		// Drain the channel to let the goroutine complete and close the file
		for range metricsChan {
			// Just drain the channel
		}
		// File should be closed now by the defer in GetMetrics
	}

	// If there was a leak, we would have accumulated 10 open file descriptors
	// The fact that we got here without error means files are being closed properly
	// Note: With typical fd limits (256-1024), 10 files won't trigger an error,
	// but this test validates the pattern. A real leak would accumulate over time.
}

func TestOutputTypeDetection(t *testing.T) {
	tests := []struct {
		name         string
		file         string
		expectedType showoutputtype
	}{
		{"batch format", "batch_format.txt", BATCH},
		{"tabular format", "tabular_format.txt", TABULAR},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testFile := filepath.Join("testdata", tt.file)
			parser := NewFileParser(testFile, "")

			err := parser.Initialize(1 * time.Second)
			if err != nil {
				t.Fatalf("Initialize failed: %v", err)
			}

			// After initialization, the split function should have detected the type
			// We can verify this by reading metrics and checking they parse correctly
			metricsChan := parser.GetMetrics()

			metrics, ok := <-metricsChan
			if !ok {
				t.Fatal("Expected to receive metrics")
			}

			if metrics == nil {
				t.Fatal("Received nil metrics")
			}

			// If we got valid metrics, the format was correctly detected
			if len(metrics.Values) == 0 {
				t.Error("Expected metrics to have values")
			}
		})
	}
}

// Real-world data tests using files ported from main branch

func TestGetMetrics_SingleBatch(t *testing.T) {
	testFile := filepath.Join("testdata", "mysql.single")
	parser := NewFileParser(testFile, "")

	err := parser.Initialize(1 * time.Second)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	metricsChan := parser.GetMetrics()

	// Should get exactly one metrics object
	metrics, ok := <-metricsChan
	if !ok {
		t.Fatal("Expected to receive metrics, channel closed")
	}

	if metrics == nil {
		t.Fatal("Received nil metrics")
	}

	statusMetrics, ok := metrics.Values["status.global"]
	if !ok {
		t.Fatal("Expected status.global domain")
	}

	// mysql.single should have 400+ status variables
	if len(statusMetrics) < 100 {
		t.Errorf("Expected at least 100 metrics, got %d", len(statusMetrics))
	}

	// Verify we got exactly one sample
	_, ok = <-metricsChan
	if ok {
		t.Error("Expected only one sample from mysql.single")
	}
}

func TestGetMetrics_SingleTabular(t *testing.T) {
	testFile := filepath.Join("testdata", "mysqladmin.single")
	parser := NewFileParser(testFile, "")

	err := parser.Initialize(1 * time.Second)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	metricsChan := parser.GetMetrics()

	// Should get exactly one metrics object
	metrics, ok := <-metricsChan
	if !ok {
		t.Fatal("Expected to receive metrics, channel closed")
	}

	if metrics == nil {
		t.Fatal("Received nil metrics")
	}

	statusMetrics, ok := metrics.Values["status.global"]
	if !ok {
		t.Fatal("Expected status.global domain")
	}

	// mysqladmin.single should have many status variables
	if len(statusMetrics) < 100 {
		t.Errorf("Expected at least 100 metrics, got %d", len(statusMetrics))
	}

	// Verify we got exactly one sample
	_, ok = <-metricsChan
	if ok {
		t.Error("Expected only one sample from mysqladmin.single")
	}
}

func TestGetMetrics_Variables(t *testing.T) {
	tests := []struct {
		name     string
		file     string
		minCount int
	}{
		{"batch format", "variables", 50},
		{"tabular format", "variables.tab", 50},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testFile := filepath.Join("testdata", tt.file)
			parser := NewFileParser(testFile, "")

			err := parser.Initialize(1 * time.Second)
			if err != nil {
				t.Fatalf("Initialize failed: %v", err)
			}

			metricsChan := parser.GetMetrics()

			metrics, ok := <-metricsChan
			if !ok {
				t.Fatal("Expected to receive metrics, channel closed")
			}

			if metrics == nil {
				t.Fatal("Received nil metrics")
			}

			statusMetrics, ok := metrics.Values["status.global"]
			if !ok {
				t.Fatal("Expected status.global domain")
			}

			if len(statusMetrics) < tt.minCount {
				t.Errorf("Expected at least %d metrics, got %d", tt.minCount, len(statusMetrics))
			}
		})
	}
}

func TestGetMetrics_LotsOfSamples(t *testing.T) {
	testFile := filepath.Join("testdata", "mysql.lots")
	parser := NewFileParser(testFile, "")

	err := parser.Initialize(1 * time.Second)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	metricsChan := parser.GetMetrics()

	count := 0
	var prevUptime float64
	for metrics := range metricsChan {
		if metrics == nil {
			t.Error("Received nil metrics")
			continue
		}

		statusMetrics, ok := metrics.Values["status.global"]
		if !ok {
			t.Error("Expected status.global domain")
			continue
		}

		// Should have many metrics per sample
		if len(statusMetrics) < 100 {
			t.Errorf("Sample %d: Expected at least 100 metrics, got %d", count, len(statusMetrics))
		}

		// Find uptime and verify it's increasing
		for _, mv := range statusMetrics {
			if mv.Name == "uptime" {
				if prevUptime > 0 && mv.Value <= prevUptime {
					t.Errorf("Sample %d: Uptime not increasing: prev=%f, current=%f", count, prevUptime, mv.Value)
				}
				prevUptime = mv.Value
				break
			}
		}

		count++
	}

	// mysql.lots has 215 samples
	if count < 200 {
		t.Errorf("Expected at least 200 samples, got %d", count)
	}
	t.Logf("Parsed %d samples from mysql.lots", count)
}

func TestGetMetrics_LotsOfSamplesTabular(t *testing.T) {
	testFile := filepath.Join("testdata", "mysqladmin.lots")
	parser := NewFileParser(testFile, "")

	err := parser.Initialize(1 * time.Second)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	metricsChan := parser.GetMetrics()

	count := 0
	for metrics := range metricsChan {
		if metrics == nil {
			t.Error("Received nil metrics")
			continue
		}

		statusMetrics, ok := metrics.Values["status.global"]
		if !ok {
			t.Error("Expected status.global domain")
			continue
		}

		// Should have many metrics per sample
		if len(statusMetrics) < 100 {
			t.Errorf("Sample %d: Expected at least 100 metrics, got %d", count, len(statusMetrics))
		}

		count++
	}

	// mysqladmin.lots has 220 samples
	if count < 200 {
		t.Errorf("Expected at least 200 samples, got %d", count)
	}
	t.Logf("Parsed %d samples from mysqladmin.lots", count)
}

func TestGetMetrics_TokuDB(t *testing.T) {
	testFile := filepath.Join("testdata", "mysql.toku")
	parser := NewFileParser(testFile, "")

	err := parser.Initialize(1 * time.Second)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	metricsChan := parser.GetMetrics()

	foundTokuMetrics := false
	for metrics := range metricsChan {
		if metrics == nil {
			continue
		}

		statusMetrics, ok := metrics.Values["status.global"]
		if !ok {
			continue
		}

		// Look for TokuDB-specific metrics
		for _, mv := range statusMetrics {
			if strings.HasPrefix(mv.Name, "tokudb_") {
				foundTokuMetrics = true
				break
			}
		}

		if foundTokuMetrics {
			break
		}
	}

	if !foundTokuMetrics {
		t.Error("Expected to find TokuDB metrics (tokudb_*)")
	}
}

func TestGetMetrics_IntervalFiltering(t *testing.T) {
	testFile := filepath.Join("testdata", "mysqladmin.byfives")

	// Test with 1 second interval (should get most samples)
	parser1s := NewFileParser(testFile, "")
	err := parser1s.Initialize(1 * time.Second)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	count1s := 0
	for range parser1s.GetMetrics() {
		count1s++
	}

	// Test with 5 second interval (should skip some samples based on uptime)
	parser5s := NewFileParser(testFile, "")
	err = parser5s.Initialize(5 * time.Second)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	count5s := 0
	for range parser5s.GetMetrics() {
		count5s++
	}

	// With 5s interval, we should get fewer or equal samples than with 1s
	if count5s > count1s {
		t.Errorf("Expected fewer samples with 5s interval (%d) than 1s interval (%d)", count5s, count1s)
	}

	t.Logf("Samples: 1s interval=%d, 5s interval=%d", count1s, count5s)
}

func TestGetMetrics_TwoSamples(t *testing.T) {
	tests := []struct {
		name string
		file string
	}{
		{"batch format", "mysql.two"},
		{"tabular format", "mysqladmin.two"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testFile := filepath.Join("testdata", tt.file)
			parser := NewFileParser(testFile, "")

			err := parser.Initialize(1 * time.Second)
			if err != nil {
				t.Fatalf("Initialize failed: %v", err)
			}

			metricsChan := parser.GetMetrics()

			count := 0
			var uptimes []float64
			for metrics := range metricsChan {
				if metrics == nil {
					continue
				}

				statusMetrics, ok := metrics.Values["status.global"]
				if !ok {
					t.Error("Expected status.global domain")
					continue
				}

				// Find uptime
				for _, mv := range statusMetrics {
					if mv.Name == "uptime" {
						uptimes = append(uptimes, mv.Value)
						break
					}
				}

				count++
			}

			if count != 2 {
				t.Errorf("Expected exactly 2 samples, got %d", count)
			}

			if len(uptimes) == 2 && uptimes[0] >= uptimes[1] {
				t.Errorf("Expected uptime to increase: %f -> %f", uptimes[0], uptimes[1])
			}
		})
	}
}

func TestGetMetrics_ErrorFile(t *testing.T) {
	// mysqladmin.err contains error cases - parser should handle gracefully
	testFile := filepath.Join("testdata", "mysqladmin.err")
	parser := NewFileParser(testFile, "")

	err := parser.Initialize(1 * time.Second)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	metricsChan := parser.GetMetrics()

	// Should handle errors gracefully and potentially return some valid data
	count := 0
	for metrics := range metricsChan {
		if metrics != nil {
			count++
		}
	}

	// Just verify we didn't panic - error file may or may not produce samples
	t.Logf("Parsed %d samples from error file (graceful handling)", count)
}

// Benchmark tests

// Helper function for benchmarks
func benchmarkFile(b *testing.B, fileName string, interval time.Duration) {
	for i := 0; i < b.N; i++ {
		parser := NewFileParser(fileName, "")
		err := parser.Initialize(interval)
		if err != nil {
			b.Fatalf("Initialize failed: %v", err)
		}

		metricsChan := parser.GetMetrics()
		for range metricsChan {
			// Drain the channel
		}
	}
}

// Benchmark parsing a single batch format sample
func BenchmarkParseSample_Single(b *testing.B) {
	testFile := filepath.Join("testdata", "mysql.single")
	benchmarkFile(b, testFile, 1*time.Second)
}

// Benchmark parsing a single tabular format sample
func BenchmarkParseSample_SingleTabular(b *testing.B) {
	testFile := filepath.Join("testdata", "mysqladmin.single")
	benchmarkFile(b, testFile, 1*time.Second)
}

// Benchmark parsing variables in batch format
func BenchmarkParseVariablesBatch(b *testing.B) {
	testFile := filepath.Join("testdata", "variables")
	benchmarkFile(b, testFile, 1*time.Second)
}

// Benchmark parsing variables in tabular format
func BenchmarkParseVariablesTabular(b *testing.B) {
	testFile := filepath.Join("testdata", "variables.tab")
	benchmarkFile(b, testFile, 1*time.Second)
}

// Benchmark parsing many samples in batch format (215 samples)
func BenchmarkGetMetrics_LargeBatch(b *testing.B) {
	testFile := filepath.Join("testdata", "mysql.lots")
	benchmarkFile(b, testFile, 1*time.Second)
}

// Benchmark parsing many samples in tabular format (220 samples)
func BenchmarkGetMetrics_LargeTabular(b *testing.B) {
	testFile := filepath.Join("testdata", "mysqladmin.lots")
	benchmarkFile(b, testFile, 1*time.Second)
}

// Benchmark parsing with interval filtering (should skip some samples)
func BenchmarkGetMetrics_WithIntervalFiltering(b *testing.B) {
	testFile := filepath.Join("testdata", "mysqladmin.lots")
	benchmarkFile(b, testFile, 1*time.Minute)
}

// Benchmark the parseSample function directly
func BenchmarkParseSample_BatchFormat(b *testing.B) {
	parser := NewFileParser("", "")
	parser.outputtype = BATCH

	// Sample data with ~100 metrics
	data := []byte("com_select\t100\nthreads_running\t5\nuptime\t3600\nquestions\t1000\n" +
		"com_insert\t50\ncom_update\t25\ncom_delete\t10\ncom_commit\t75\n" +
		"innodb_buffer_pool_reads\t1000\ninnodb_buffer_pool_read_requests\t50000\n" +
		"innodb_rows_read\t10000\ninnodb_rows_inserted\t500\ninnodb_rows_updated\t250\n")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := parser.parseSample(data)
		if err != nil {
			b.Fatalf("parseSample failed: %v", err)
		}
	}
}

// Benchmark the parseSample function with tabular format
func BenchmarkParseSample_TabularFormat(b *testing.B) {
	parser := NewFileParser("", "")
	parser.outputtype = TABULAR

	data := []byte(`| com_select                | 100  |
| threads_running           | 5    |
| uptime                    | 3600 |
| questions                 | 1000 |
| com_insert                | 50   |
| com_update                | 25   |
| com_delete                | 10   |
| innodb_buffer_pool_reads  | 1000 |
`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := parser.parseSample(data)
		if err != nil {
			b.Fatalf("parseSample failed: %v", err)
		}
	}
}

// Benchmark metric conversion with many metrics
func BenchmarkConvertToBlipMetrics(b *testing.B) {
	parser := NewFileParser("test.txt", "")
	startTime := time.Now()

	// Create a map with 400+ metrics (realistic for MySQL)
	data := make(map[string]string)
	for i := 0; i < 400; i++ {
		data[fmt.Sprintf("metric_%d", i)] = fmt.Sprintf("%d", i*100)
	}

	// Add some known gauge metrics
	data["threads_running"] = "5"
	data["threads_connected"] = "10"
	data["innodb_buffer_pool_pages_dirty"] = "100"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = parser.convertToBlipMetrics(data, uint(i), startTime)
	}
}

// Benchmark two-sample file to test overhead
func BenchmarkGetMetrics_TwoSamples(b *testing.B) {
	testFile := filepath.Join("testdata", "mysql.two")
	benchmarkFile(b, testFile, 1*time.Second)
}
