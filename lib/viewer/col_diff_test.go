package viewer

import (
	"testing"
	"time"

	"github.com/cashapp/blip"
	myqblip "github.com/jayjanssen/myq-tools/lib/blip"
)

func TestDiffCol_MissingMetric(t *testing.T) {
	cache := myqblip.NewMetricCache(false)

	col := DiffCol{
		colNum: colNum{
			defaultCol: defaultCol{
				Name:   "test",
				Length: 5,
			},
			Precision: 0,
		},
		Key: SourceKey{
			Domain: "status.global",
			Metric: "nonexistent_metric",
		},
	}

	result := col.GetData(cache)

	if len(result) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(result))
	}

	// Should display "-" for missing metric
	if result[0] != "    -" {
		t.Errorf("Expected '-', got '%s'", result[0])
	}
}

func TestDiffCol_ExistingMetric(t *testing.T) {
	cache := myqblip.NewMetricCache(false)

	// Add first sample
	cache.Update(&blip.Metrics{
		Begin: time.Now(),
		End:   time.Now(),
		Values: map[string][]blip.MetricValue{
			"status.global": {
				{Name: "test_metric", Value: 100, Type: blip.CUMULATIVE_COUNTER},
			},
		},
	})

	// Add second sample
	cache.Update(&blip.Metrics{
		Begin: time.Now(),
		End:   time.Now(),
		Values: map[string][]blip.MetricValue{
			"status.global": {
				{Name: "test_metric", Value: 150, Type: blip.CUMULATIVE_COUNTER},
			},
		},
	})

	col := DiffCol{
		colNum: colNum{
			defaultCol: defaultCol{
				Name:   "test",
				Length: 5,
			},
			Precision: 0,
		},
		Key: SourceKey{
			Domain: "status.global",
			Metric: "test_metric",
		},
	}

	result := col.GetData(cache)

	if len(result) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(result))
	}

	// Should display the diff (150 - 100 = 50)
	if result[0] != "   50" {
		t.Errorf("Expected '   50', got '%s'", result[0])
	}
}
