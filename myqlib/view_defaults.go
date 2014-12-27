package myqlib

func DefaultMyqViews() map[string]MyqView {
	return map[string]MyqView{
    "cttf": MyqView{[]colDef{
      colDef{"threads_running", "trun", "%5s", "%5d", Gauge, None},
    }},
  }
}