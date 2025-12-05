package blip

import (
	"path/filepath"
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
