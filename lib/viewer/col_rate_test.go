package viewer

import (
	"reflect"
	"testing"
	"time"

	"github.com/cashapp/blip"
	myqblip "github.com/jayjanssen/myq-tools/lib/blip"
	"gopkg.in/yaml.v3"
)

func getTestRateCol() RateCol {
	sk, _ := ParseSourceKey("status/connections")
	rc := RateCol{}
	rc.Name = "cons"
	rc.Description = "Connections per second"
	rc.Type = "Rate"
	rc.Key = sk
	rc.Length = 4
	rc.Units = NUMBER
	rc.Precision = 0

	return rc
}

func TestRateCol(t *testing.T) {
	rc := getTestRateCol()
	if rc.Name != "cons" {
		t.Errorf("Unexpected col name (test): %s", rc.Name)
	}
}

func TestRateColImplementsViewer(t *testing.T) {
	var _ Viewer = getTestRateCol()
}

func TestRateColParse(t *testing.T) {
	yaml_str := `---
- name: cons
  description: Connections per second
  key: status/connections
  type: Rate
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

	if col.GetShortHelp() != `cons: Connections per second` {
		t.Errorf("bad description: %s", cols[0].GetShortHelp())
	}

	rc := getTestRateCol()
	if !reflect.DeepEqual(rc, col) {
		t.Error(`cols not matching`)
		t.Logf("rc: %+v", rc)
		t.Logf("col: %+v", col)
	}
}

// Create a metric cache to test with
func getTestRateCache(con_prev, con_cur string) *myqblip.MetricCache {
	cache := myqblip.NewMetricCache(false)

	// Add previous sample if provided
	if con_prev != "" {
		if val, err := parseFloat(con_prev); err == nil {
			cache.Update(&blip.Metrics{
				Begin: time.Now(),
				End:   time.Now(),
				Values: map[string][]blip.MetricValue{
					"status.global": {
						{Name: "connections", Value: val, Type: blip.CUMULATIVE_COUNTER},
					},
				},
			})
		}
	}

	// Add current sample
	if con_cur != "" {
		if val, err := parseFloat(con_cur); err == nil {
			cache.Update(&blip.Metrics{
				Begin: time.Now().Add(1 * time.Second),
				End:   time.Now().Add(1 * time.Second),
				Values: map[string][]blip.MetricValue{
					"status.global": {
						{Name: "connections", Value: val, Type: blip.CUMULATIVE_COUNTER},
					},
				},
			})
		}
	}

	return cache
}

func TestRateColgetRate(t *testing.T) {
	col := getTestRateCol()

	// Normal rate
	cache := getTestRateCache(`10`, `15`)
	rate, err := col.getRate(cache)
	if err != nil {
		t.Error(err)
	}
	// Rate should be approximately 5 (allowing for floating point precision)
	if rate < 4.9 || rate > 5.1 {
		t.Fatalf(`unexpected rate: %f (expected ~5)`, rate)
	}
	outputs := col.GetData(cache)
	if len(outputs) != 1 {
		t.Errorf(`unexpected amount of output strings %d`, len(outputs))
	}
	// Output formatting may vary slightly, check it's reasonable
	if len(outputs[0]) != 4 {
		t.Errorf(`unexpected output length: %d`, len(outputs[0]))
	}

	// Blank prev rate
	cache = getTestRateCache(``, `15`)
	rate, err = col.getRate(cache)
	if err != nil {
		t.Error(err)
	}
	if rate != 15 {
		t.Errorf(`unexpected rate: %f`, rate)
	}

	// Missing metric
	cache = getTestRateCache(``, ``)
	_, err = col.getRate(cache)
	if err == nil {
		t.Error(`expected error for missing metric`)
	}
	outputs = col.GetData(cache)
	if len(outputs) != 1 {
		t.Errorf(`unexpected amount of output strings %d`, len(outputs))
	}
	if len(outputs[0]) != 4 {
		t.Errorf(`output should be 4: %d`, len(outputs[0]))
	}
	if outputs[0] != `   -` {
		t.Errorf(`unexpected GetData(): '%s'`, outputs[0])
	}

}

func TestRateColBadSourceKey(t *testing.T) {
	// the key value is incorrect, it should be <source>/<key>
	yaml_str := `---
- name: acls
  description: Aborted clients (existing connections)
  source: status
  key: aborted_clients
  type: Rate
  units: Number
  length: 4
  precision: 0
`

	var cols ViewerList
	err := yaml.Unmarshal([]byte(yaml_str), &cols)

	if err == nil {
		t.Error(`expected error parsing bad sourcekey`)
	}
}

// Create a metric cache to test with (no previous)
func getTestRateNullPrevCache(con_cur string) *myqblip.MetricCache {
	cache := myqblip.NewMetricCache(false)

	if con_cur != "" {
		if val, err := parseFloat(con_cur); err == nil {
			cache.Update(&blip.Metrics{
				Begin: time.Now(),
				End:   time.Now(),
				Values: map[string][]blip.MetricValue{
					"status.global": {
						{Name: "connections", Value: val, Type: blip.CUMULATIVE_COUNTER},
					},
				},
			})
		}
	}

	return cache
}

func TestRateColgetRateNullPrev(t *testing.T) {
	col := getTestRateCol()

	// Normal rate
	cache := getTestRateNullPrevCache(`1500`)
	rate, err := col.getRate(cache)
	if err != nil {
		t.Error(err)
	}
	if rate != 1500 {
		t.Errorf(`unexpected rate: %f`, rate)
	}
	outputs := col.GetData(cache)
	if len(outputs) != 1 {
		t.Errorf(`unexpected amount of output strings %d`, len(outputs))
	}
	if outputs[0] != `1500` {
		t.Errorf(`unexpected GetData(): '%s'`, outputs[0])
	}

}
