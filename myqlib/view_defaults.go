package myqlib

import( 
  "fmt"
  "bytes"
  "time"
)

// Time Columns
var (
  Timestamp_col FuncCol = FuncCol{ "time", 
    "Time data was printed", 8, 0,
    func(b *bytes.Buffer, state MyqState, c Col) {
      b.WriteString( time.Now().Format("15:04:05"))
    },
  }
  Runtime_col FuncCol = FuncCol{ "time", 
    "Interval since data started", 8, 0,
    func(b *bytes.Buffer, state MyqState, c Col) {
      runtime := time.Duration( state.Cur[`uptime`].(int64) - state.FirstUptime) * time.Second
      b.WriteString( fmt.Sprintf( fmt.Sprint( `%`, c.Width(), `s`),
        runtime.String() ))
    },
  }
)

func DefaultViews() map[string]View {
	return map[string]View{
		"cttf": NormalView{
			help: "Connections, Threads, Tables, and Files",
			cols: []Col{
				GroupCol{
					"Connects", "Connection related metrics", []Col{
						RateCol{"cons", "connections", "Connections per second", 4, 0},
						RateCol{"acns", "aborted_connects", "Aborted connections per second", 4, 0},
						RateCol{"acls", "aborted_clients", "Aborted Clients (those with existing connections) per second", 4, 0},
					},
				},
				GroupCol{
					"Threads", "Thread related metrics", []Col{
						GaugeCol{"conn", "threads_connected", "Threads Connected", 4, 0},
						GaugeCol{"run", "threads_running", "Threads running", 4, 0},
						GaugeCol{"cach", "threads_cached", "Threads Cached", 4, 0},
						RateCol{"crtd", "threads_created", "Threads Created per second", 4, 0},
					},
				},
				GroupCol{
					"Pool", "Thread Pool metrics", []Col{
						GaugeCol{"tot", "threadpool_threads", "Threadpool Threads", 4, 0},
						GaugeCol{"idle", "threadpool_idle_threads", "Threadpool Idle Threads", 4, 0},
					},
				},
				GroupCol{
					"Tables", "Table metrics", []Col{
						GaugeCol{"open", "open_tables", "Open Tables", 4, 0},
						GaugeCol{"opns", "opened_tables", "Opened Tables per Second", 4, 0},
					},
				},
				GroupCol{
					"Defs", "Table Definition Metrics", []Col{
						GaugeCol{"open", "open_table_definitions", "Open Table Definitions", 4, 0},
						GaugeCol{"opns", "opened_table_definitions", "Opened Table Definitions per Second", 4, 0},
					},
				},
				GroupCol{
					"Files", "File Metrics", []Col{
						GaugeCol{"open", "open_files", "Open Files", 4, 0},
						RateCol{"opns", "opened_files", "Opened Files per Second", 4, 0},
					},
				},
			},
		},
	}
}