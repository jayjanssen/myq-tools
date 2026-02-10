package viewer

import (
	"reflect"
	"testing"
	"time"

	"github.com/cashapp/blip"
	myqblip "github.com/jayjanssen/myq-tools/lib/blip"
	"gopkg.in/yaml.v3"
)

func getTestRateSumCol() RateSumCol {
	sk1, _ := ParseSourceKey("status/com_set_option")
	sk2, _ := ParseSourceKey("status/com_set_password")
	sk3, _ := ParseSourceKey("status/com_set_resource_group")
	rc := RateSumCol{}
	rc.Name = "set"
	rc.Description = "SET commands per second"
	rc.Type = "RateSum"
	rc.Keys = []SourceKey{sk1, sk2, sk3}
	rc.Length = 5
	rc.Units = NUMBER
	rc.Precision = 0

	return rc
}

func TestRateSumCol(t *testing.T) {
	rc := getTestRateSumCol()
	if rc.Name != "set" {
		t.Errorf("Unexpected col name (set): %s", rc.Name)
	}
}

func TestRateSumColImplementsViewer(t *testing.T) {
	var _ Viewer = getTestRateSumCol()
}

func TestRateSumColParse(t *testing.T) {
	yaml_str := `---
- name: set
  description: SET commands per second
  keys: 
    - status/com_set_option
    - status/com_set_password
    - status/com_set_resource_group
  type: RateSum
  units: Number
  length: 5
  precision: 0
`

	var cols ViewerList
	err := yaml.Unmarshal([]byte(yaml_str), &cols)

	if err != nil {
		t.Fatal(err)
	}

	if len(cols) != 1 {
		t.Fatalf("not enough cols parsed: %d", len(cols))
	}

	col := cols[0]

	if col.GetShortHelp() != `set: SET commands per second` {
		t.Errorf("bad description: %s", cols[0].GetShortHelp())
	}

	rc := getTestRateSumCol()
	if !reflect.DeepEqual(rc, col) {
		t.Error(`cols not matching`)
		t.Logf("rc: %+v", rc)
		t.Logf("col: %+v", col)
	}
}

// Create a metric cache to test with
func getTestRateSumCache(con_prev, con_cur string) *myqblip.MetricCache {
	cache := myqblip.NewMetricCache(false)

	var prevMetrics, curMetrics []blip.MetricValue

	if con_prev != "" {
		if val, err := parseFloat(con_prev); err == nil {
			prevMetrics = []blip.MetricValue{
				{Name: "com_set_option", Value: val, Type: blip.CUMULATIVE_COUNTER},
				{Name: "com_set_password", Value: val, Type: blip.CUMULATIVE_COUNTER},
				{Name: "com_set_resource_group", Value: val, Type: blip.CUMULATIVE_COUNTER},
			}
		}
	}

	if con_cur != "" {
		if val, err := parseFloat(con_cur); err == nil {
			curMetrics = []blip.MetricValue{
				{Name: "com_set_option", Value: val, Type: blip.CUMULATIVE_COUNTER},
				{Name: "com_set_password", Value: val, Type: blip.CUMULATIVE_COUNTER},
				{Name: "com_set_resource_group", Value: val, Type: blip.CUMULATIVE_COUNTER},
			}
		}
	}

	if len(prevMetrics) > 0 {
		cache.Update(&blip.Metrics{
			Begin: time.Now(),
			End:   time.Now(),
			Values: map[string][]blip.MetricValue{
				"status.global": prevMetrics,
			},
		})
	}

	if len(curMetrics) > 0 {
		cache.Update(&blip.Metrics{
			Begin: time.Now().Add(1 * time.Second),
			End:   time.Now().Add(1 * time.Second),
			Values: map[string][]blip.MetricValue{
				"status.global": curMetrics,
			},
		})
	}

	return cache
}

func TestRateSumColgetRate(t *testing.T) {
	col := getTestRateSumCol()

	// Normal rate - 3 metrics, each increased by 5, so total rate is 15
	cache := getTestRateSumCache(`10`, `15`)
	rate, err := col.getRate(cache)
	if err != nil {
		t.Error(err)
	}
	// Rate should be approximately 15 (3 metrics * 5 diff each)
	if rate < 14.9 || rate > 15.1 {
		t.Fatalf(`unexpected rate: %f (expected ~15)`, rate)
	}
	outputs := col.GetData(cache)
	if len(outputs) != 1 {
		t.Errorf(`unexpected amount of output strings %d`, len(outputs))
	}
	// Output formatting may vary, check it's reasonable
	if len(outputs[0]) != 5 {
		t.Errorf(`unexpected output length: %d`, len(outputs[0]))
	}

	// Blank prev rate - should use current values as rate
	cache = getTestRateSumCache(``, `15`)
	rate, err = col.getRate(cache)
	if err != nil {
		t.Error(err)
	}
	// Should be approximately 45 (3 metrics * 15 each)
	if rate < 44.9 || rate > 45.1 {
		t.Errorf(`unexpected rate: %f (expected ~45)`, rate)
	}

}
