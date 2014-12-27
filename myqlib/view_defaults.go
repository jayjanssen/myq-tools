package myqlib

func DefaultViews() map[string]View {
	return map[string]View{
    "cttf": NormalView{
      help: "Connections, Threads, Tables, and Files",
      cols: []Col{
        RateCol{
          name: "cons",
          variable_name: "connections",
          help: "Connections per second",
          width: 5,
          precision: 0,
        },
        GaugeCol{
          name: "trun",
          variable_name: "threads_running",
          help: "Threads running",
          width: 5,
          precision: 0,
        },
      },
    },
  }
}