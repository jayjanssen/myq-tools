package myqlib

func DefaultViews() map[string]View {
	return map[string]View{
    "cttf": NormalView{
      help: "Connections, Threads, Tables, and Files",
      cols: []Col{
        GroupCol{
          "Connects", "Connection related metrics", []Col{
            RateCol{"cons", "connections", "Connections per second", 5, 0 },
            RateCol{ "acns", "aborted_connects", "Aborted connections per second", 5, 0 },
            RateCol{ "acls", "aborted_clients", "Aborted Clients (those with existing connections) per second", 5, 0 },
          },
        },
        GroupCol{
          "Threads", "Thread related metrics", []Col{
            GaugeCol{ "conn", "threads_connected", "Threads Connected", 5, 0 },
            GaugeCol{ "run", "threads_running", "Threads running", 5, 0 },
            GaugeCol{ "cache", "threads_cached", "Threads Cached", 5, 0 },
            RateCol{ "crtd", "threads_created", "Threads Created per second", 5, 0 },
          },
        },
        GroupCol{
          "Pool", "Thread Pool metrics", []Col{
            GaugeCol{ "tot", "threadpool_threads", "Threadpool Threads", 5, 0 },
            GaugeCol{ "idle", "threadpool_idle_threads", "Threadpool Idle Threads", 5, 0 },
          },
        },
        GroupCol{
          "Tables", "Table metrics", []Col{
            GaugeCol{ "open", "open_tables", "Open Tables", 5, 0 },
            GaugeCol{ "opns", "opened_tables", "Opened Tables per Second", 5, 0 },
          },
        },
        GroupCol{
          "Defs", "Table Definition Metrics", []Col{
            GaugeCol{ "open", "open_table_definitions", "Open Table Definitions", 5, 0 },
            GaugeCol{ "opns", "opened_table_definitions", "Opened Table Definitions per Second", 5, 0 },
          },
        },
        GroupCol{
          "Files", "File Metrics", []Col{
            GaugeCol{ "open", "open_files", "Open Files", 5, 0 },
            RateCol{ "opns", "opened_files", "Opened Files per Second", 5, 0 },
          },
        },
        
        
      },
    },
  }
}