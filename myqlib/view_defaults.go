package myqlib

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Time Columns
var (
	Timestamp_col FuncCol = FuncCol{
		DefaultCol{"time", "Time data was printed", 8},
		func(b *bytes.Buffer, state *MyqState, c Col) {
			b.WriteString(time.Now().Format("15:04:05"))
		},
	}
	Runtime_col FuncCol = FuncCol{
		DefaultCol{"time", "Interval since data started", 8},
		func(b *bytes.Buffer, state *MyqState, c Col) {
			runtime := time.Duration(state.Cur[`uptime`].(int64)-state.FirstUptime) * time.Second
			b.WriteString(fmt.Sprintf(fmt.Sprint(`%`, c.Width(), `s`), runtime.String()))
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
						GaugeCol{DefaultCol{"conn", "Threads Connected", 4}, "threads_connected", 0, NumberUnits},
						GaugeCol{DefaultCol{"run", "Threads running", 4}, "threads_running", 0, NumberUnits},
						GaugeCol{DefaultCol{"cach", "Threads Cached", 4}, "threads_cached", 0, NumberUnits},
						RateCol{DefaultCol{"crtd", "Threads Created per second", 4}, "threads_created", 0, NumberUnits},
						RateCol{DefaultCol{"slow", "Threads that were slow to launch per second", 4}, "slow_launch_threads", 0, NumberUnits},
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
				GroupCol{
					DefaultCol{"Throughput", "Bytes in/out of the server", 0},
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
				GroupCol{DefaultCol{"Selects", "Select Types", 0},
					[]Col{
						RateCol{DefaultCol{"fjn", "Full Joins / sec", 5}, "select_full_join", 0, NumberUnits},
						RateCol{DefaultCol{"frj", "Full Range Joins / sec", 5}, "select_full_range_join", 0, NumberUnits},
						RateCol{DefaultCol{"rang", "Range / sec", 5}, "select_range", 0, NumberUnits},
						RateCol{DefaultCol{"rchk", "Range Check / sec", 5}, "select_range_check", 0, NumberUnits},
						RateCol{DefaultCol{"scan", "Scan / sec", 5}, "select_scan", 0, NumberUnits},
					},
				},
				GroupCol{DefaultCol{"Sorts", "Sort Types", 0},
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
				GroupCol{DefaultCol{"Reads", "Handler read stats", 0},
					[]Col{
						RateCol{DefaultCol{"rfst", "Read First / s", 5}, "handler_read_first", 0, NumberUnits},
						RateCol{DefaultCol{"rkey", "Read Key / s", 5}, "handler_read_key", 0, NumberUnits},
						RateCol{DefaultCol{"rnex", "Read Next / s", 5}, "handler_read_next", 0, NumberUnits},
						RateCol{DefaultCol{"rprv", "Read Prev / s", 5}, "handler_read_prev", 0, NumberUnits},
						RateCol{DefaultCol{"rrd", "Random reads / s", 5}, "handler_read_rnd", 0, NumberUnits},
						RateCol{DefaultCol{"rrdn", "Read First / s", 5}, "handler_read_rnd_next", 0, NumberUnits},
					},
				},
				GroupCol{DefaultCol{"Other", "Other handler stats", 0},
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
				GroupCol{DefaultCol{"Row Operations", "Row-level operations", 0},
					[]Col{
						RateCol{DefaultCol{"read", "Reads / s", 5}, "innodb_rows_read", 0, NumberUnits},
						RateCol{DefaultCol{"ins", "Inserts / s", 5}, "innodb_rows_inserted", 0, NumberUnits},
						RateCol{DefaultCol{"upd", "Updates / s", 5}, "innodb_rows_updated", 0, NumberUnits},
						RateCol{DefaultCol{"del", "Deletes / s", 5}, "innodb_rows_deleted", 0, NumberUnits},
					},
				},
				GroupCol{DefaultCol{"Buffer Pool (pages)", "Page-level operations", 0},
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
				GroupCol{DefaultCol{"Row Operations", "Row-level operations", 0},
					[]Col{
						RateCol{DefaultCol{"read", "Reads / s", 5}, "innodb_rows_read", 0, NumberUnits},
						RateCol{DefaultCol{"ins", "Inserts / s", 5}, "innodb_rows_inserted", 0, NumberUnits},
						RateCol{DefaultCol{"upd", "Updates / s", 5}, "innodb_rows_updated", 0, NumberUnits},
						RateCol{DefaultCol{"del", "Deletes / s", 5}, "innodb_rows_deleted", 0, NumberUnits},
					},
				},
				GroupCol{DefaultCol{"Buffer Pool (pages)", "Page-level operations", 0},
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
		"wsrep": NormalView{
			help: "Galera Wsrep statistics",
			extra_header: func(b *bytes.Buffer, state *MyqState) {
				b.WriteString( fmt.Sprintf( "%s / %s (idx: %d) / %s %s", state.Cur[`V_wsrep_cluster_name`],
					state.Cur[`V_wsrep_node_name`], state.Cur[`wsrep_local_index`], state.Cur[`wsrep_provider_name`],
					state.Cur[`wsrep_provider_version`] ))
			},
			cols: []Col{
				GroupCol{DefaultCol{"Cluster", "Cluster-wide stats (at least according to this node)", 0},
					[]Col{
						StringCol{DefaultCol{"P", "Primary (P) or Non-primary (N)", 1}, "wsrep_cluster_status"},
						// GaugeCol{ DefaultCol{"cnf", , 3}, "wsrep_cluster_conf_id", 0, NumberUnits},
						RightmostCol{DefaultCol{"cnf", "Cluster configuration id (increments every time a node joins/leaves the cluster)", 3}, `wsrep_cluster_conf_id`},
						GaugeCol{DefaultCol{"#", "Cluster size", 2}, "wsrep_cluster_size", 0, NumberUnits},
					},
				},
				GroupCol{DefaultCol{"Node", "Node's specific state", 0},
					[]Col{
						StringCol{DefaultCol{"state", "State of this node", 4}, "wsrep_local_state_comment"},
					},
				},
				FuncCol{DefaultCol{"laten", "Average replication latency", 5},
					func(b *bytes.Buffer, state *MyqState, c Col) {
						vals := strings.Split(state.Cur[`wsrep_evs_repl_latency`].(string), `/`)
						// Expecting 5 vals here, filler if not
						if len(vals) != 5 {
							filler(b, c)
						} else {
							avg, _ := strconv.ParseFloat(vals[1], 64)
							cv := collapse_number(avg, int64(c.Width()), 2, SecondUnits)
							b.WriteString(fmt.Sprintf(fmt.Sprint(`%`, c.Width(), `s`), cv))
						}
					},
				},
				GroupCol{DefaultCol{"Outbound", "Sent replication events", 0},
					[]Col{
						RateCol{DefaultCol{"msgs", "Replicated messages (usually transactions) per second", 4}, "wsrep_replicated", 0, NumberUnits},
						RateCol{DefaultCol{"data", "Replicated bytes per second", 4}, "wsrep_replicated_bytes", 0, MemoryUnits},
						GaugeCol{DefaultCol{"queue", "Outbound replication queue", 3}, "wsrep_local_send_queue", 0, NumberUnits},
					},
				},
				GroupCol{DefaultCol{"Inbound", "Received replication events", 0},
					[]Col{
						RateCol{DefaultCol{"msgs", "Received messages (usually transactions) per second", 4}, "wsrep_received", 0, NumberUnits},
						RateCol{DefaultCol{"data", "Received bytes per second", 4}, "wsrep_received_bytes", 0, MemoryUnits},
						GaugeCol{DefaultCol{"queue", "Received replication apply queue", 3}, "wsrep_local_recv_queue", 0, NumberUnits},
					},
				},
				GroupCol{DefaultCol{"FlowC", "Flow control stats", 0},
					[]Col{
						DiffCol{DefaultCol{"paused", "Flow control paused (could be from anywhere in the cluster)", 5}, "wsrep_flow_control_paused_ns", 0, NanoSecondUnits},
						DiffCol{DefaultCol{"snt", "Flow control sent messages (could be starting or stopping FC)", 3}, "wsrep_flow_control_sent", 0, NumberUnits},
					},
				},
				GroupCol{DefaultCol{"Cnflcts", "Galera replication conflicts (on this node)", 0},
					[]Col{
						DiffCol{DefaultCol{"lcf", "Local certification failures since last sample", 3}, "wsrep_local_cert_failures", 0, NumberUnits},
						DiffCol{DefaultCol{"bfa", "Brute force aborts since last sample", 3}, "wsrep_local_bf_aborts", 0, NumberUnits},
					},
				},
				GroupCol{DefaultCol{"Gcache", "Galera cache (gcache) information", 0},
					[]Col{
						CurDiffCol{DefaultCol{"ist", "Gcached transactions", 5}, `wsrep_last_committed`, `wsrep_local_cached_downto`, 0, NumberUnits},
						GaugeCol{DefaultCol{"idx", "Certification index size (keys)", 4}, "wsrep_cert_index_size", 0, NumberUnits},
					},
				},
				GroupCol{DefaultCol{"Apply", "Theoretical and actual apply efficiency", 0},
					[]Col{
						PercentCol{DefaultCol{`%ef`, `Percent of threads being used`, 4}, "wsrep_apply_window", "V_wsrep_slave_threads", 0},
					},
				},
			},
		},
		"qcache": NormalView{
			help: "Query cache stats",
			cols: []Col{
				StringCol{DefaultCol{"type", "Query cache type", 6}, "V_query_cache_type"},
				RateSumCol{DefaultCol{"sel", "Total Selects + Qcache Hits per second", 4}, []string{"com_select", "qcache_hits"}, 0, NumberUnits},
				RateCol{DefaultCol{"hits", "Query cache hits per second", 4}, "qcache_hits", 0, NumberUnits},
				RateCol{DefaultCol{"ins", "Query inserts per second (new entries to the cache)", 4}, "qcache_inserts", 0, NumberUnits},
				RateCol{DefaultCol{"notc", "Queries not cached per second (either can't be cached, or SELECT SQL_NO_CACHE)", 4}, "qcache_not_cached", 0, NumberUnits},
				GaugeCol{DefaultCol{"tot", "Total queries in the cache", 4}, "qcache_queries_in_cache", 0, NumberUnits},
				RateCol{DefaultCol{"lowm", "Low memory prunes (cache entries removed due to memory limit)", 4}, "qcache_lowmem_prunes", 0, NumberUnits},
				PercentCol{DefaultCol{`%free`, "Percent of cache memory free", 5}, "qcache_free_blocks", "qcache_total_blocks", 0},
			},
		},
	}
}
