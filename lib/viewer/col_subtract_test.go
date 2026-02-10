package viewer

import (
	"testing"
	"time"

	"github.com/cashapp/blip"
	myqblip "github.com/jayjanssen/myq-tools/lib/blip"
)

func TestSubtractCol_MissingMetric(t *testing.T) {
	cache := myqblip.NewMetricCache(false)

	col := SubtractCol{
		colNum: colNum{
			defaultCol: defaultCol{
				Name:   "test",
				Length: 5,
			},
			Precision: 0,
		},
		Bigger: SourceKey{
			Domain: "status.global",
			Metric: "bigger_metric",
		},
		Smaller: SourceKey{
			Domain: "status.global",
			Metric: "smaller_metric",
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

func TestSubtractCol_MissingBigger(t *testing.T) {
	cache := myqblip.NewMetricCache(false)

	// Add only smaller metric
	cache.Update(&blip.Metrics{
		Begin: time.Now(),
		End:   time.Now(),
		Values: map[string][]blip.MetricValue{
			"status.global": {
				{Name: "smaller_metric", Value: 50, Type: blip.GAUGE},
			},
		},
	})

	col := SubtractCol{
		colNum: colNum{
			defaultCol: defaultCol{
				Name:   "test",
				Length: 5,
			},
			Precision: 0,
		},
		Bigger: SourceKey{
			Domain: "status.global",
			Metric: "bigger_metric",
		},
		Smaller: SourceKey{
			Domain: "status.global",
			Metric: "smaller_metric",
		},
	}

	result := col.GetData(cache)

	if len(result) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(result))
	}

	// Should display "-" for missing bigger metric
	if result[0] != "    -" {
		t.Errorf("Expected '-', got '%s'", result[0])
	}
}

func TestSubtractCol_MissingSmaller(t *testing.T) {
	cache := myqblip.NewMetricCache(false)

	// Add only bigger metric
	cache.Update(&blip.Metrics{
		Begin: time.Now(),
		End:   time.Now(),
		Values: map[string][]blip.MetricValue{
			"status.global": {
				{Name: "bigger_metric", Value: 100, Type: blip.GAUGE},
			},
		},
	})

	col := SubtractCol{
		colNum: colNum{
			defaultCol: defaultCol{
				Name:   "test",
				Length: 5,
			},
			Precision: 0,
		},
		Bigger: SourceKey{
			Domain: "status.global",
			Metric: "bigger_metric",
		},
		Smaller: SourceKey{
			Domain: "status.global",
			Metric: "smaller_metric",
		},
	}

	result := col.GetData(cache)

	if len(result) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(result))
	}

	// Should display "-" for missing smaller metric
	if result[0] != "    -" {
		t.Errorf("Expected '-', got '%s'", result[0])
	}
}

func TestSubtractCol_ExistingMetrics(t *testing.T) {
	cache := myqblip.NewMetricCache(false)

	// Add both metrics
	cache.Update(&blip.Metrics{
		Begin: time.Now(),
		End:   time.Now(),
		Values: map[string][]blip.MetricValue{
			"status.global": {
				{Name: "bigger_metric", Value: 100, Type: blip.GAUGE},
				{Name: "smaller_metric", Value: 30, Type: blip.GAUGE},
			},
		},
	})

	col := SubtractCol{
		colNum: colNum{
			defaultCol: defaultCol{
				Name:   "test",
				Length: 5,
			},
			Precision: 0,
		},
		Bigger: SourceKey{
			Domain: "status.global",
			Metric: "bigger_metric",
		},
		Smaller: SourceKey{
			Domain: "status.global",
			Metric: "smaller_metric",
		},
	}

	result := col.GetData(cache)

	if len(result) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(result))
	}

	// Should display the subtraction (100 - 30 = 70)
	if result[0] != "   70" {
		t.Errorf("Expected '   70', got '%s'", result[0])
	}
}
