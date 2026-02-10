package viewer

import (
	"testing"
	"time"

	"github.com/cashapp/blip"
	myqblip "github.com/jayjanssen/myq-tools/lib/blip"
)

func getTestGroupCol() GroupCol {
	gc := GroupCol{}
	gc.Name = "Connects"
	gc.Description = "Connection related metrics"
	gc.Type = "Group"

	gc.Cols = make(ViewerList, 2)
	gc.Cols[0] = getTestRateCol()
	gc.Cols[1] = getTestGaugeCol()

	return gc
}

func TestGroupColImplementsViewer(t *testing.T) {
	gc := getTestGroupCol()
	var _ Viewer = gc
}

// Create a metric cache to test with
func getTestGroupCache() *myqblip.MetricCache {
	cache := myqblip.NewMetricCache(false)

	// Add previous sample
	cache.Update(&blip.Metrics{
		Begin: time.Now(),
		End:   time.Now(),
		Values: map[string][]blip.MetricValue{
			"status.global": {
				{Name: "connections", Value: 10, Type: blip.CUMULATIVE_COUNTER},
				{Name: "threads_connected", Value: 3, Type: blip.GAUGE},
			},
		},
	})

	// Add current sample
	cache.Update(&blip.Metrics{
		Begin: time.Now().Add(1 * time.Second),
		End:   time.Now().Add(1 * time.Second),
		Values: map[string][]blip.MetricValue{
			"status.global": {
				{Name: "connections", Value: 15, Type: blip.CUMULATIVE_COUNTER},
				{Name: "threads_connected", Value: 4, Type: blip.GAUGE},
			},
		},
	})

	return cache
}

func TestGroupColGetHeader(t *testing.T) {
	gc := getTestGroupCol()
	cache := getTestGroupCache()

	lines := gc.GetHeader(cache)
	if len(lines) != 2 {
		t.Errorf(`unexpected # of lines: %d`, len(lines))
	}

	if lines[0] != `Connects ` {
		t.Errorf(`unexpected header first line output: '%s'`, lines[0])
	}
	if lines[1] != `cons conn` {
		t.Errorf(`unexpected header second line output: '%s'`, lines[1])
	}

}

func TestGroupColGetData(t *testing.T) {
	gc := getTestGroupCol()
	cache := getTestGroupCache()

	lines := gc.GetData(cache)
	if len(lines) != 1 {
		t.Errorf(`unexpected # of lines: %d`, len(lines))
	}

	// Rate should be ~5 (15-10), gauge should be 4
	if lines[0] != `   5    4` {
		t.Errorf(`unexpected GetData output: '%s'`, lines[0])
	}
}
