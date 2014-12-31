package myqlib

import( 
  "fmt"
  "bytes"
  "time"
)

// Time Columns
var (
  Timestamp_col FuncCol = FuncCol{ 
    DefaultCol{"time", "Time data was printed", 8}, 0,
    func(b *bytes.Buffer, state MyqState, c Col) {
      b.WriteString( time.Now().Format("15:04:05"))
    },
  }
  Runtime_col FuncCol = FuncCol{ 
    DefaultCol{"time", "Interval since data started", 8}, 0,
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
          DefaultCol{"Connects", "Connection related metrics", 0}, 
          []Col{
						RateCol{DefaultCol{"cons", "Connections per second", 4}, "connections", 0, NumberUnits},
						RateCol{DefaultCol{"acns", "Aborted connections per second", 4}, "aborted_connects", 0, NumberUnits},
						RateCol{DefaultCol{"acls", "Aborted Clients (those with existing connections) per second", 4}, "aborted_clients", 0, NumberUnits},
					},
				},
				GroupCol{
					DefaultCol{"Threads", "Thread related metrics", 0},
          []Col{
						GaugeCol{DefaultCol{"conn", "Threads Connected", 4},  "threads_connected", 0, NumberUnits},
						GaugeCol{DefaultCol{"run", "Threads running", 4}, "threads_running", 0, NumberUnits},
						GaugeCol{DefaultCol{"cach", "Threads Cached", 4}, "threads_cached", 0, NumberUnits},
						RateCol{DefaultCol{"crtd", "Threads Created per second", 4}, "threads_created", 0, NumberUnits},
					},
				},
				GroupCol{
					DefaultCol{"Pool", "Thread Pool metrics", 0},
          []Col{
						GaugeCol{DefaultCol{"tot", "Threadpool Threads", 4}, "threadpool_threads", 0, NumberUnits},
						GaugeCol{DefaultCol{"idle", "Threadpool Idle Threads", 4}, "threadpool_idle_threads", 0, NumberUnits},
					},
				},
				GroupCol{
					DefaultCol{"Tables", "Table metrics", 0},
          []Col{
						GaugeCol{DefaultCol{"open", "Open Tables", 4}, "open_tables", 0, NumberUnits},
						GaugeCol{DefaultCol{"opns", "Opened Tables per Second", 4}, "opened_tables", 0, NumberUnits},
					},
				},
				GroupCol{
					DefaultCol{"Defs", "Table Definition Metrics", 0},
          []Col{
						GaugeCol{DefaultCol{"open", "Open Table Definitions", 4}, "open_table_definitions", 0, NumberUnits},
						GaugeCol{DefaultCol{"opns", "Opened Table Definitions per Second", 4}, "opened_table_definitions", 0, NumberUnits},
					},
				},
				GroupCol{
					DefaultCol{"Files", "File Metrics", 0},
          []Col{
						GaugeCol{DefaultCol{"open", "Open Files", 4}, "open_files", 0, NumberUnits},
						RateCol{DefaultCol{"opns", "Opened Files per Second", 4}, "opened_files", 0, NumberUnits},
					},
				},
			},
		},
    "coms": NormalView{
      help: "MySQL Commands",
      cols: []Col{
        RateCol{DefaultCol{"sel", "Selects per second", 4}, "com_select", 0, NumberUnits},
      },
    },
    "throughput": NormalView{
      help: "MySQL Server Throughput",
      cols: []Col{
        GroupCol {
          DefaultCol{"Throughput", "Bytes in/out of the server",0},
          []Col{
						RateCol{DefaultCol{"recv", "Bytes received / sec", 5}, "bytes_received", 0, MemoryUnits},
						RateCol{DefaultCol{"sent", "Bytes sent / sec", 5}, "bytes_sent", 0, MemoryUnits},
          },
        },
      },
    },
    "query": NormalView{
      help: "Query types and sorts",
      cols: []Col{
        RateCol{DefaultCol{"slow", "Slow queries per second", 4}, "slow_queries", 0, NumberUnits},
        GroupCol { DefaultCol{"Selects", "Select Types",0},
          []Col{
						RateCol{DefaultCol{"fjn", "Full Joins / sec", 5}, "select_full_join", 0, NumberUnits},
						RateCol{DefaultCol{"frj", "Full Range Joins / sec", 5}, "select_full_range_join", 0, NumberUnits},
						RateCol{DefaultCol{"rang", "Range / sec", 5}, "select_range", 0, NumberUnits},
						RateCol{DefaultCol{"rchk", "Range Check / sec", 5}, "select_range_check", 0, NumberUnits},
						RateCol{DefaultCol{"scan", "Scan / sec", 5}, "select_scan", 0, NumberUnits},
          },
        },
        GroupCol { DefaultCol{"Sorts", "Sort Types",0},
          []Col{
						RateCol{DefaultCol{"pass", "Merge Passes / sec", 5}, "sort_merge_passes", 0, NumberUnits},
						RateCol{DefaultCol{"rang", "Range / sec", 5}, "sort_range", 0, NumberUnits},
						RateCol{DefaultCol{"rows", "Rows / sec", 5}, "sort_rows", 0, NumberUnits},
						RateCol{DefaultCol{"scan", "Scan / sec", 5}, "sort_scan", 0, NumberUnits},
          },
        },
      },
    },
    "temp": NormalView{
      help: "Internal Temporary Tables",
      cols: []Col{
				RateCol{DefaultCol{"tmps", "Temporary Tables / second", 5}, "created_tmp_tables", 0, NumberUnits},
				RateCol{DefaultCol{"disk", "On Disk Temp Tables / second", 5}, "created_tmp_disk_tables", 0, NumberUnits},
				RateCol{DefaultCol{"files", "Temp Files / second", 5}, "created_tmp_files", 0, NumberUnits},
      },
    },
    "handler": NormalView{
      help: "Storage Engine Handler metrics",
      cols: []Col{
        GroupCol { DefaultCol{"Reads", "Handler read stats",0},
          []Col{
    				RateCol{DefaultCol{"rfst", "Read First / s", 5}, "handler_read_first", 0, NumberUnits},
    				RateCol{DefaultCol{"rkey", "Read Key / s", 5}, "handler_read_key", 0, NumberUnits},
    				RateCol{DefaultCol{"rnex", "Read Next / s", 5}, "handler_read_next", 0, NumberUnits},
    				RateCol{DefaultCol{"rprv", "Read Prev / s", 5}, "handler_read_prev", 0, NumberUnits},
    				RateCol{DefaultCol{"rrd", "Random reads / s", 5}, "handler_read_rnd", 0, NumberUnits},
    				RateCol{DefaultCol{"rrdn", "Read First / s", 5}, "handler_read_rnd_next", 0, NumberUnits},
          },
        },
        GroupCol { DefaultCol{"Other", "Other handler stats",0},
          []Col{
    				RateCol{DefaultCol{"ins", "Inserts / s", 5}, "handler_write", 0, NumberUnits},
    				RateCol{DefaultCol{"upd", "Updates / s", 5}, "handler_update", 0, NumberUnits},
    				RateCol{DefaultCol{"del", "Deletes / s", 5}, "handler_delete", 0, NumberUnits},
    				RateCol{DefaultCol{"cmt", "Commits / s", 5}, "handler_commit", 0, NumberUnits},
    				RateCol{DefaultCol{"rbk", "Rollbacks / s", 5}, "handler_rollback", 0, NumberUnits},
    				RateCol{DefaultCol{"disc", "Discovers / s", 5}, "handler_discover", 0, NumberUnits},
          },
        },
      },
    },
    "innodb": NormalView{
      help: "Innodb metrics",
      cols: []Col{
        GroupCol { DefaultCol{"Row Operations", "Row-level operations",0},
          []Col{
    				RateCol{DefaultCol{"read", "Reads / s", 5}, "innodb_rows_read", 0, NumberUnits},
    				RateCol{DefaultCol{"ins", "Inserts / s", 5}, "innodb_rows_inserted", 0, NumberUnits},
    				RateCol{DefaultCol{"upd", "Updates / s", 5}, "innodb_rows_updated", 0, NumberUnits},
    				RateCol{DefaultCol{"del", "Deletes / s", 5}, "innodb_rows_deleted", 0, NumberUnits},
          },
        },
        GroupCol { DefaultCol{"Buffer Pool (pages)", "Page-level operations",0},
          []Col{
    				RateCol{DefaultCol{"logr", "Read Requests (Logical) / s", 5}, "innodb_buffer_pool_read_requests", 0, NumberUnits},
    				RateCol{DefaultCol{"phyr", "Reads (Physical) / s", 5}, "innodb_buffer_pool_reads", 0, NumberUnits},
    				RateCol{DefaultCol{"logw", "Write Requests / s", 5}, "innodb_buffer_pool_write_requests", 0, NumberUnits},
    				RateCol{DefaultCol{"phyw", "Writes (Physical) / s", 5}, "innodb_buffer_pool_pages_flushed", 0, NumberUnits},
    				PercentCol{DefaultCol{`%dirt`, `Buffer pool %dirty`, 5}, "innodb_buffer_pool_pages_dirty", "innodb_buffer_pool_pages_total", 0},
          },
        },
      },
    },
    "innodb_buffer_pool": NormalView{
      help: "Innodb Buffer Pool stats",
      cols: []Col{
        GroupCol { DefaultCol{"Row Operations", "Row-level operations",0},
          []Col{
    				RateCol{DefaultCol{"read", "Reads / s", 5}, "innodb_rows_read", 0, NumberUnits},
    				RateCol{DefaultCol{"ins", "Inserts / s", 5}, "innodb_rows_inserted", 0, NumberUnits},
    				RateCol{DefaultCol{"upd", "Updates / s", 5}, "innodb_rows_updated", 0, NumberUnits},
    				RateCol{DefaultCol{"del", "Deletes / s", 5}, "innodb_rows_deleted", 0, NumberUnits},
          },
        },
        GroupCol { DefaultCol{"Buffer Pool (pages)", "Page-level operations",0},
          []Col{
    				RateCol{DefaultCol{"logr", "Read Requests (Logical) / s", 5}, "innodb_buffer_pool_read_requests", 0, NumberUnits},
    				RateCol{DefaultCol{"phyr", "Reads (Physical) / s", 5}, "innodb_buffer_pool_reads", 0, NumberUnits},
    				RateCol{DefaultCol{"logw", "Write Requests / s", 5}, "innodb_buffer_pool_write_requests", 0, NumberUnits},
    				RateCol{DefaultCol{"phyw", "Writes (Physical) / s", 5}, "innodb_buffer_pool_pages_flushed", 0, NumberUnits},
    				PercentCol{DefaultCol{`%dirt`, `Buffer pool %dirty`, 5}, "innodb_buffer_pool_pages_dirty", "innodb_buffer_pool_pages_total", 0},
          },
        },
      },
    },
	}
}