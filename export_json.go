package main

import (
  "encoding/json"
  "fmt"
  "./myqlib"
)

func main() {
  views := myqlib.DefaultViews()
	
	res, _ := json.MarshalIndent(views, "", "  ")
	// res, _ := json.Marshal(metrics)
	fmt.Println(string(res))

  // var newmetrics map[string]MySQLMetricDef
  // json.Unmarshal(res, &newmetrics)

  // nres, _ := json.MarshalIndent(newmetrics, "", "  ")
  // fmt.Println(string(nres))
}
