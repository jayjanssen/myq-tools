package viewer

import (
	"fmt"
	"testing"
	"time"

	"github.com/cashapp/blip"
	myqblip "github.com/jayjanssen/myq-tools/lib/blip"
)

func getTestPercentCol() PercentCol {
	rc := PercentCol{}
	rc.Name = "dirt"
	rc.Description = "Buffer pool percent dirty"
	rc.Type = "Percent"
	rc.Numerator, _ = ParseSourceKey("status/innodb_buffer_pool_pages_dirty")
	rc.Denominator, _ = ParseSourceKey("status/innodb_buffer_pool_pages_total")
	rc.Length = 4
	rc.Units = PERCENT
	rc.Precision = 0

	return rc
}

func TestPercentColImplementsViewer(t *testing.T) {
	var _ Viewer = getTestPercentCol()
}

func getTestPercentCache(numerator, denominator string) *myqblip.MetricCache {
	cache := myqblip.NewMetricCache(false)

	var metrics []blip.MetricValue

	if numerator != "" {
		if val, err := parseFloat(numerator); err == nil {
			metrics = append(metrics, blip.MetricValue{
				Name:  "innodb_buffer_pool_pages_dirty",
				Value: val,
				Type:  blip.GAUGE,
			})
		}
	}

	if denominator != "" {
		if val, err := parseFloat(denominator); err == nil {
			metrics = append(metrics, blip.MetricValue{
				Name:  "innodb_buffer_pool_pages_total",
				Value: val,
				Type:  blip.GAUGE,
			})
		}
	}

	if len(metrics) > 0 {
		cache.Update(&blip.Metrics{
			Begin: time.Now(),
			End:   time.Now(),
			Values: map[string][]blip.MetricValue{
				"status.global": metrics,
			},
		})
	}

	return cache
}

func TestPercentColgetPercent(t *testing.T) {
	col := getTestPercentCol()
	cache := getTestPercentCache(`86716`, `15999992`)

	percent, err := col.getPercent(cache)
	if err != nil {
		t.Error(err)
	}
	if fmt.Sprintf("%.5f", percent) != `0.54198` {
		t.Errorf(`unexpected percent: '%s'`, fmt.Sprintf("%.5f", percent))
	}

	data := col.GetData(cache)
	if data[0] != `  1%` {
		t.Errorf(`unexpected data: '%s'`, data)
	}

	cache = getTestPercentCache(`86716`, `notanum`)
	_, err = col.getPercent(cache)
	if err == nil {
		t.Error(`expected denominator error`)
	}

	cache = getTestPercentCache(`notanum`, `15999992`)
	_, err = col.getPercent(cache)
	if err == nil {
		t.Error(`expected numerator error`)
	}

}
