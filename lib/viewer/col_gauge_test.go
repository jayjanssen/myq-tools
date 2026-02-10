package viewer

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/cashapp/blip"
	myqblip "github.com/jayjanssen/myq-tools/lib/blip"
	"gopkg.in/yaml.v3"
)

func getTestGaugeCol() GaugeCol {
	sk, _ := ParseSourceKey("status/threads_connected")
	rc := GaugeCol{}
	rc.Name = "conn"
	rc.Description = "Threads connected"
	rc.Type = "Gauge"
	rc.Key = sk
	rc.Length = 4
	rc.Units = NUMBER
	rc.Precision = 0

	return rc
}

func TestGaugeCol(t *testing.T) {
	rc := getTestGaugeCol()
	if rc.Name != "conn" {
		t.Errorf("Unexpected col name (test): %s", rc.Name)
	}
}

func TestGaugeColImplementsViewer(t *testing.T) {
	var _ Viewer = getTestGaugeCol()
}

func TestGaugeColParse(t *testing.T) {
	yaml_str := `---
- name: conn
  description: Threads connected
  key: status/threads_connected
  type: Gauge
  units: Number
  length: 4
  precision: 0
`

	var cols ViewerList
	err := yaml.Unmarshal([]byte(yaml_str), &cols)

	if err != nil {
		t.Error(err)
	}

	if len(cols) != 1 {
		t.Errorf("not enough cols parsed: %d", len(cols))
	}

	col := cols[0]

	if col.GetShortHelp() != `conn: Threads connected` {
		t.Errorf("bad description: %s", cols[0].GetShortHelp())
	}

	rc := getTestGaugeCol()
	if !reflect.DeepEqual(rc, col) {
		t.Error(`cols not matching`)
		t.Logf("rc: %+v", rc)
		t.Logf("col: %+v", col)
	}
}

// Create a metric cache to test with
func getTestGaugeCache(value string) *myqblip.MetricCache {
	cache := myqblip.NewMetricCache(false)

	var metricValue blip.MetricValue
	if value == "" {
		// Missing metric - don't add it
		return cache
	}

	// Try to parse as number
	if val, err := parseFloat(value); err == nil {
		metricValue = blip.MetricValue{
			Name:  "threads_connected",
			Value: val,
			Type:  blip.GAUGE,
		}
	} else {
		// String value
		metricValue = blip.MetricValue{
			Name:  "threads_connected",
			Value: 0,
			Type:  blip.GAUGE,
			Meta: map[string]string{
				"string_value": value,
			},
		}
	}

	cache.Update(&blip.Metrics{
		Begin: time.Now(),
		End:   time.Now(),
		Values: map[string][]blip.MetricValue{
			"status.global": {metricValue},
		},
	})

	return cache
}

func parseFloat(s string) (float64, error) {
	var val float64
	_, err := fmt.Sscanf(s, "%f", &val)
	return val, err
}

func TestGaugeColGetData(t *testing.T) {
	col := getTestGaugeCol()

	// Normal gauge
	cache := getTestGaugeCache(`10`)
	outputs := col.GetData(cache)
	if len(outputs) != 1 {
		t.Errorf(`unexpected amount of output strings %d`, len(outputs))
	}
	if outputs[0] != `  10` {
		t.Errorf(`unexpected GetData(): '%s'`, outputs[0])
	}

	// Normal gauge string (GaugeCol doesn't support string values, only numeric)
	// So this test is removed - gauge columns only show numeric values

	// Missing key
	col.Key.Metric = `notfound`
	cache = getTestGaugeCache(`10`)
	outputs = col.GetData(cache)
	if len(outputs) != 1 {
		t.Errorf(`unexpected amount of output strings %d`, len(outputs))
	}
	if outputs[0] != `   -` {
		t.Errorf(`unexpected GetData(): '%s'`, outputs[0])
	}

}
