package myqlib

import (
  "testing"
  "encoding/json"
  "fmt"
)

func TestJSON(t *testing.T) {
	
	metrics := map[string]MySQLMetricDef{
		"connections" : MySQLMetricDef{"cons", Counter},
		"threads_running" : MySQLMetricDef{"trun", Gauge},
	}
	// MySQLMetricDef( 'Connections', counter ),

	res, _ := json.MarshalIndent(metrics, "", "  ")
	// res, _ := json.Marshal(metrics)
	fmt.Println(string(res))

	var newmetrics map[string]MySQLMetricDef
	json.Unmarshal(res, &newmetrics)

	nres, _ := json.MarshalIndent(newmetrics, "", "  ")
	fmt.Println(string(nres))

	return
}
