package viewer

import (
	"testing"
	"time"

	"github.com/cashapp/blip"
	myqblip "github.com/jayjanssen/myq-tools/lib/blip"
)

func getTestView() View {
	view := View{}
	view.Name = "Test View"
	view.Description = "My Test View"
	view.Groups = make([]GroupCol, 1)
	view.Groups[0] = getTestGroupCol()

	return view
}

func TestViewImplementsViewer(t *testing.T) {
	view := getTestView()
	var _ Viewer = view
}

// Create a metric cache to test with
func getTestViewCache() *myqblip.MetricCache {
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

func TestViewGetHeader(t *testing.T) {
	view := getTestView()
	cache := getTestViewCache()

	lines := view.GetHeader(cache)

	expectedLines := []string{
		`         Connects `,
		`    time cons conn`,
	}

	if len(lines) != len(expectedLines) {
		t.Errorf(`unexpected # of lines: %d`, len(lines))
	}
	for i, expected := range expectedLines {
		if lines[i] != expected {
			t.Errorf(`unexpected line %d output: '%s' (expected '%s')`, i, lines[i], expected)
		}
	}
}

func TestViewGetData(t *testing.T) {
	view := getTestView()
	cache := getTestViewCache()

	lines := view.GetData(cache)

	expectedLines := []string{
		`      0s    5    4`,
	}

	if len(lines) != len(expectedLines) {
		t.Errorf(`unexpected # of lines: %d`, len(lines))
	}
	for i, expected := range expectedLines {
		if lines[i] != expected {
			t.Errorf(`unexpected line %d output: '%s' (expected '%s')`, i, lines[i], expected)
		}
	}
}
