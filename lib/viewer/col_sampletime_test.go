package viewer

import (
	"testing"
	"time"

	"github.com/cashapp/blip"
	myqblip "github.com/jayjanssen/myq-tools/lib/blip"
)

// Create a metric cache to test with
func getTestSampleTimeCache() *myqblip.MetricCache {
	cache := myqblip.NewMetricCache(false)

	cache.Update(&blip.Metrics{
		Begin: time.Now(),
		End:   time.Now(),
		Values: map[string][]blip.MetricValue{
			"status.global": {
				{Name: "threads_connected", Value: 10, Type: blip.GAUGE},
			},
		},
	})

	return cache
}

func TestSampleTimeColGetHeader(t *testing.T) {
	tc := NewSampleTimeCol()
	cache := getTestSampleTimeCache()

	h := tc.GetHeader(cache)
	if len(h) != 1 {
		t.Errorf(`got wrong number of header lines: %d`, len(h))
	}

	if h[0] != `    time` {
		t.Errorf(`got wrong time header: '%s'`, h[0])
	}
}

func TestTimeColGetData(t *testing.T) {
	tc := NewSampleTimeCol()
	cache := getTestSampleTimeCache()

	h := tc.GetData(cache)
	if len(h) != 1 {
		t.Errorf(`got wrong number of data lines: %d`, len(h))
	}

	// In file mode, time should start at 0s
	if h[0] != `      0s` {
		t.Errorf(`got wrong time data: '%s'`, h[0])
	}
}
