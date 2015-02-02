package myqlib

import (
	`bytes`
	`fmt`
	`sort`
	`strconv`
	`strings`
	`time`
)

// Time Columns
var (
	Timestamp_col Col = NewFuncCol(`time`, `Time data was printed`, 8,
		func(state *MyqState, c Col) (chan string){
			ch := make( chan string, 1 ); defer close( ch )
			ch <- fit_string(time.Now().Format(`15:04:05`), c.Width())
			return ch
		})

	Runtime_col Col = NewFuncCol(`time`, `Interval since data started`, 8,
		func(state *MyqState, c Col) (chan string) {
			ch := make( chan string, 1 ); defer close( ch )
			runtime := time.Duration(state.Cur.getI(`uptime`)-state.FirstUptime) * time.Second
			ch <- fit_string(runtime.String(), c.Width())
			return ch
		})
)

func DefaultViews() map[string]View {
	return map[string]View{
		`cttf`: NewNormalView(`Connections, Threads, Tables, and Files`,
			NewGroupCol(`Connects`, `Collection related metrics`,
				NewRateCol(`cons`, `Connections per second`, 4, `connections`, 0, NumberUnits),
				NewRateCol(`acns`, `Aborted connections per second`, 4, `aborted_connects`, 0, NumberUnits),
				NewRateCol(`acls`, `Aborted Clients (those with existing connections) per second`, 4, `aborted_clients`, 0, NumberUnits),
			),
			NewGroupCol(`Threads`, `Thread related metrics`,
				NewGaugeCol(`conn`, `Threads Connected`, 4, `threads_connected`, 0, NumberUnits),
				NewGaugeCol(`run`, `Threads running`, 4, `threads_running`, 0, NumberUnits),
				NewGaugeCol(`cach`, `Threads Cached`, 4, `threads_cached`, 0, NumberUnits),
				NewRateCol(`crtd`, `Threads Created per second`, 4, `threads_created`, 0, NumberUnits),
				NewRateCol(`slow`, `Threads that were slow to launch per second`, 4, `slow_launch_threads`, 0, NumberUnits),
			),
			NewGroupCol(`Pool`, `Thread Pool metrics`,
				NewGaugeCol(`tot`, `Threadpool Threads`, 4, `threadpool_threads`, 0, NumberUnits),
				NewGaugeCol(`idle`, `Threadpool Idle Threads`, 4, `threadpool_idle_threads`, 0, NumberUnits),
			),
			NewGroupCol(`Tables`, `Table metrics`,
				NewGaugeCol(`open`, `Open Tables`, 4, `open_tables`, 0, NumberUnits),
				NewRateCol(`opns`, `Opened Tables per Second`, 4, `opened_tables`, 0, NumberUnits),
				NewRateCol(`immd`, `Immediate Table locks`, 4, `table_locks_immediate`, 0, NumberUnits),
				NewRateCol(`wait`, `Table locks Waited`, 4, `table_locks_waited`, 0, NumberUnits),
			),
			NewGroupCol(`Defs`, `Table Definition Metrics`,
				NewGaugeCol(`open`, `Open Table Definitions`, 4, `open_table_definitions`, 0, NumberUnits),
				NewGaugeCol(`opns`, `Opened Table Definitions per Second`, 4, `opened_table_definitions`, 0, NumberUnits),
			),
			NewGroupCol(`Files`, `File Metrics`,
				NewGaugeCol(`open`, `Open Files`, 4, `open_files`, 0, NumberUnits),
				NewRateCol(`opns`, `Opened Files per Second`, 4, `opened_files`, 0, NumberUnits),
			),
		),
		`coms`: NewNormalView(`MySQL Commands`,
			NewRateCol(`sel`, `Selects per second`, 5, `com_select`, 0, NumberUnits),
			NewRateSumCol(`dml`, `Inserts, Updates + Deletes (and other various DML) / Second`, 5, 0, NumberUnits, `com_insert.*`, `com_update.*`, `com_delete.*`, `com_load`, `com_replace.*`, `com_truncate`),
			NewRateSumCol(`ddl`, `Data Definition commands / Second`, 5, 0, NumberUnits, `com_alter.*`, `com_create.*`, `com_drop.*`, `com_rename_table`),
			NewRateSumCol(`admin`, `Admin commands / Second`, 5, 0, NumberUnits, `com_admin.*`),
			NewRateSumCol(`show`, `SHOW commands / Second`, 5, 0, NumberUnits, `com_show.*`),
			NewRateSumCol(`show`, `SET commands / Second`, 5, 0, NumberUnits, `com_set.*`),
			NewRateSumCol(`lock`, `LOCK commands / Second`, 5, 0, NumberUnits, `com_lock.*`, `com_unlock.*`),
			NewRateSumCol(`trx`, `Transactional commands / Second`, 5, 0, NumberUnits, `com_begin`, `com_commit`, `com_rollback.*`, `com_savepoint`),
			NewRateSumCol(`xa`, `XA commands / Second`, 5, 0, NumberUnits, `com_xa.*`),
			NewRateSumCol(`prep`, `Prepared Statement commands / Second`, 5, 0, NumberUnits, `Com_stmt.*`, `Com_.*_sql`),
		),
		`throughput`: NewNormalView(`MySQL Server Throughput`,
			NewGroupCol(`Throughput`, `Bytes in/out of the server`,
				NewDiffCol(`recv`, `Data received since last sample`, 6, `bytes_received`, 0, MemoryUnits),
				NewRateCol(`recv/s`, `Bytes received / sec`, 6, `bytes_received`, 0, MemoryUnits),
				NewDiffCol(`sent`, `Data sent since last sample`, 6, `bytes_sent`, 0, MemoryUnits),
				NewRateCol(`sent/s`, `Bytes sent / sec`, 6, `bytes_sent`, 0, MemoryUnits),
			),
		),
		`query`: NewNormalView(`Query types and sorts`,
			NewRateCol(`slow`, `Slow queries per second`, 4, `slow_queries`, 0, NumberUnits),
			NewGroupCol(`Selects`, `Select Types`,
				NewRateCol(`fjn`, `Full Joins / sec`, 5, `select_full_join`, 0, NumberUnits),
				NewRateCol(`frj`, `Full Range Joins / sec`, 5, `select_full_range_join`, 0, NumberUnits),
				NewRateCol(`rang`, `Range / sec`, 5, `select_range`, 0, NumberUnits),
				NewRateCol(`rchk`, `Range Check / sec`, 5, `select_range_check`, 0, NumberUnits),
				NewRateCol(`scan`, `Scan / sec`, 5, `select_scan`, 0, NumberUnits),
			),
			NewGroupCol(`Sorts`, `Sort Types`,
				NewRateCol(`pass`, `Merge Passes / sec`, 5, `sort_merge_passes`, 0, NumberUnits),
				NewRateCol(`rang`, `Range / sec`, 5, `sort_range`, 0, NumberUnits),
				NewRateCol(`rows`, `Rows / sec`, 5, `sort_rows`, 0, NumberUnits),
				NewRateCol(`scan`, `Scan / sec`, 5, `sort_scan`, 0, NumberUnits),
			),
		),
		`temp`: NewNormalView(`Internal Temporary Tables`,
			NewRateCol(`tmps`, `Temporary Tables / second`, 5, `created_tmp_tables`, 0, NumberUnits),
			NewRateCol(`disk`, `On Disk Temp Tables / second`, 5, `created_tmp_disk_tables`, 0, NumberUnits),
			NewRateCol(`files`, `Temp Files / second`, 5, `created_tmp_files`, 0, NumberUnits),
		),
		`handler`: NewNormalView(`Storage Engine Handler metrics`,
			NewGroupCol(`Reads`, `Handler read stats`,
				NewRateCol(`rfst`, `Read First / s`, 5, `handler_read_first`, 0, NumberUnits),
				NewRateCol(`rkey`, `Read Key / s`, 5, `handler_read_key`, 0, NumberUnits),
				NewRateCol(`rnex`, `Read Next / s`, 5, `handler_read_next`, 0, NumberUnits),
				NewRateCol(`rprv`, `Read Prev / s`, 5, `handler_read_prev`, 0, NumberUnits),
				NewRateCol(`rrd`, `Random reads / s`, 5, `handler_read_rnd`, 0, NumberUnits),
				NewRateCol(`rrdn`, `Read First / s`, 5, `handler_read_rnd_next`, 0, NumberUnits),
			),
			NewGroupCol(`Other`, `Other handler stats`,
				NewRateCol(`ins`, `Inserts / s`, 5, `handler_write`, 0, NumberUnits),
				NewRateCol(`upd`, `Updates / s`, 5, `handler_update`, 0, NumberUnits),
				NewRateCol(`del`, `Deletes / s`, 5, `handler_delete`, 0, NumberUnits),
				NewRateCol(`cmt`, `Commits / s`, 5, `handler_commit`, 0, NumberUnits),
				NewRateCol(`rbk`, `Rollbacks / s`, 5, `handler_rollback`, 0, NumberUnits),
				NewRateCol(`disc`, `Discovers / s`, 5, `handler_discover`, 0, NumberUnits),
			),
		),
		`innodb`: NewNormalView(`Innodb metrics`,
			NewGroupCol(`Row Ops`, `Row-level operations`,
				NewRateCol(`read`, `Reads / s`, 5, `innodb_rows_read`, 0, NumberUnits),
				NewRateSumCol(`dml`, `Inserts, Updates + Deletes / Second`, 5, 0, NumberUnits, `innodb_rows_inserted`, `innodb_rows_updated`, `innodb_rows_deleted`),
			),
			NewGroupCol(`Buffer Pool`, `Buffer Pool Stats`,
				NewGaugeCol(`data`, `Data Buffered`, 5, `innodb_buffer_pool_bytes_data`, 0, MemoryUnits),
				NewPercentCol(`dirt`, `Buffer pool %dirty`, 4, `innodb_buffer_pool_pages_dirty`, `innodb_buffer_pool_pages_total`, 0),
				NewRateCol(`rreq`, `Read Requests (Logical) / s`, 5, `innodb_buffer_pool_read_requests`, 0, NumberUnits),
				NewRateCol(`read`, `Reads (Physical) / s`, 4, `innodb_buffer_pool_reads`, 0, NumberUnits),
				NewRateCol(`wreq`, `Write Requests / s`, 5, `innodb_buffer_pool_write_requests`, 0, NumberUnits),
				NewRateCol(`write`, `Writes (Physical) / s`, 4, `innodb_buffer_pool_pages_flushed`, 0, NumberUnits),
			),
			NewGroupCol(`Log`, `Log Information`,
				NewGaugeCol(`Chkpt`, `Checkpoint age`, 5, `innodb_checkpoint_age`, 0, MemoryUnits),
				NewPercentCol(`%`, `% of Checkpoint age target`, 4, `innodb_checkpoint_age`, `innodb_checkpoint_max_age`, 0),
				NewRateCol(`lsn`, `Log growth (LSN)`, 5, `innodb_lsn_current`, 0, MemoryUnits),
			),
			NewGroupCol(`Data`, `Data Operations`,
				NewRateCol(`read`, `Bytes Read / s`, 5, `innodb_data_read`, 0, MemoryUnits),
				NewRateCol(`writes`, `Bytes Written / s`, 5, `innodb_data_written`, 0, MemoryUnits),
			),
			NewGaugeCol(`Hist`, `History List Length`, 5, `innodb_history_list_length`, 0, NumberUnits),
		),
		`innodb_buffer_pool`: NewNormalView(`Innodb Buffer Pool stats`,
			NewGroupCol(`Buffer Pool Pages`, `Innodb Buffer Pool Pages stats`,
				NewGaugeCol(`data`, `BP data pages`, 4, `innodb_buffer_pool_pages_data`, 0, NumberUnits),
				NewGaugeCol(`old`, `BP old pages`, 4, `innodb_buffer_pool_pages_old`, 0, NumberUnits),
				NewGaugeCol(`dirty`, `BP dirty pages`, 4, `innodb_buffer_pool_pages_dirty`, 0, NumberUnits),
				NewGaugeCol(`free`, `BP free pages`, 4, `innodb_buffer_pool_pages_free`, 0, NumberUnits),
				NewGaugeCol(`latched`, `BP latched pages`, 4, `innodb_buffer_pool_pages_latched`, 0, NumberUnits),
				NewGaugeCol(`misc`, `BP misc pages`, 4, `innodb_buffer_pool_pages_misc`, 0, NumberUnits),
				NewGaugeCol(`total`, `BP total pages`, 4, `innodb_buffer_pool_pages_total`, 0, NumberUnits),
			),
			NewGroupCol(`Read Ahead`, `Read ahead stats`,
				NewRateCol(`Reads`, `Read-ahead operations`, 4, `innodb_buffer_pool_read_ahead`, 0, NumberUnits),
				NewRateCol(`Evicted`, `Read-ahead evictions`, 4, `innodb_buffer_pool_read_ahead_evicted`, 0, NumberUnits),
			),
			NewGroupCol(`Reads`, `Read stats`,
				NewRateCol(`reqs`, `Read requests`, 4, `innodb_buffer_pool_read_requests`, 0, NumberUnits),
				NewRateCol(`phys`, `Physical Reads`, 4, `innodb_buffer_pool_reads`, 0, NumberUnits),
			),
			NewRateCol(`wait`, `Page waits`, 4, `innodb_buffer_pool_wait_free`, 0, NumberUnits),
			NewGroupCol(`Writes`, `Write stats`,
				NewRateCol(`reqs`, `Write requests`, 4, `innodb_buffer_pool_write_requests`, 0, NumberUnits),
				NewRateCol(`phys`, `Physical Writes`, 4, `innodb_buffer_pool_pages_flushed`, 0, NumberUnits),
				NewRateCol(`lruf`, `LRU flushed`, 4, `innodb_buffer_pool_pages_lru_flushed`, 0, NumberUnits),
			),
			NewGroupCol(`Midpoint`, `Midpoint Insertion stats`,
				NewRateCol(`old`, `Old pages inserted`, 4, `innodb_buffer_pool_pages_made_not_young`, 0, NumberUnits),
				NewRateCol(`new`, `New pages inserted`, 4, `innodb_buffer_pool_pages_made_young`, 0, NumberUnits),
			),
		),
		`innodb_flush`: NewNormalView(`Innodb flushing metrics`,
			NewGroupCol(`Pages`, `Checkpoint info`,
				NewPercentCol(`dirt`, `Buffer pool %dirty`, 4, `innodb_buffer_pool_pages_dirty`, `innodb_buffer_pool_pages_total`, 0),
				NewRateCol(`flush`, `All pages flushed`, 5, `innodb_buffer_pool_pages_flushed`, 0, NumberUnits),
				NewRateCol(`lruf`, `LRU flushes`, 5, `innodb_buffer_pool_pages_lru_flushed`, 0, NumberUnits),
			),
			NewGroupCol(`Checkpoint`, `Checkpoint info`,
				NewGaugeCol(`age`, `Checkpoint Age`, 5, `innodb_checkpoint_age`, 0, MemoryUnits),
				NewPercentCol(`max`, `Percent of checkpoint age out of max`, 5, `innodb_checkpoint_age`, `innodb_checkpoint_max_age`, 0),
			),
			NewGroupCol(`Data`, `Data stats`,
				NewRateCol(`pages`, `Pages written`, 5, `innodb_pages_written`, 0, NumberUnits),
				NewRateCol(`wops`, `Write operations`, 5, `innodb_data_writes`, 0, NumberUnits),
				NewRateCol(`bytes`, `Write data`, 5, `innodb_data_written`, 0, MemoryUnits),
			),
			NewGroupCol(`Log`, `Log Sequence Number stats`,
				NewRateCol(`lsn`, `Log growth (LSN)`, 5, `innodb_lsn_current`, 0, MemoryUnits),
				NewRateCol(`chkpt`, `Log checkpoints`, 5, `innodb_lsn_last_checkpoint`, 0, MemoryUnits),
			),
		),
		`wsrep`: NewExtraHeaderView(`Galera Wsrep statistics`,
			func(state *MyqState) (chan string){
				ch := make( chan string, 1 )
				defer close(ch)
				ch <- fmt.Sprintf("%s / %s (idx: %d) / %s %s", 
					state.Cur.getStr(`V_wsrep_cluster_name`),
					state.Cur.getStr(`V_wsrep_node_name`), 
					state.Cur.getI(`wsrep_local_index`),
					state.Cur.getStr(`wsrep_provider_name`), 
					state.Cur.getStr(`wsrep_provider_version`))
				return ch
			},
			NewGroupCol(`Cluster`, `Cluster-wide stats (at least according to this node)`,
				NewStringCol(`P`, `Primary (P) or Non-primary (N)`, 1, `wsrep_cluster_status`),
				NewRightmostCol(`cnf`, `Cluster configuration id (increments every time a node joins/leaves the cluster)`, 3, `wsrep_cluster_conf_id`),
				NewGaugeCol(`#`, `Cluster size`, 2, `wsrep_cluster_size`, 0, NumberUnits),
			),
			NewGroupCol(`Node`, `Node's specific state`,
				NewStringCol(`state`, `State of this node`, 4, `wsrep_local_state_comment`),
			),
			NewFuncCol(`laten`, `Average replication latency`, 5, func( state *MyqState, c Col) (chan string) {
				ch := make( chan string, 1 ); defer close( ch )
				vals := strings.Split(state.Cur.getStr(`wsrep_evs_repl_latency`), `/`)

				// Expecting 5 vals here, filler if not
				if len(vals) != 5 {
					ch <- column_filler(c)
				} else {
					if avg, err := strconv.ParseFloat(vals[1], 64); err == nil {
						cv := collapse_number(avg, c.Width(), 2, SecondUnits)
						ch <- fmt.Sprintf(fmt.Sprint(`%`, c.Width(), `s`), cv)
					} else {
						ch <- column_filler(c)
					}
				}
				return ch
			}),
			NewGroupCol(`Outbound`, `Sent replication events`,
				NewRateCol(`msgs`, `Replicated messages (usually transactions) per second`, 4, `wsrep_replicated`, 0, NumberUnits),
				NewRateCol(`data`, `Replicated bytes per second`, 4, `wsrep_replicated_bytes`, 0, MemoryUnits),
				NewGaugeCol(`queue`, `Outbound replication queue`, 3, `wsrep_local_send_queue`, 0, NumberUnits),
			),
			NewGroupCol(`Inbound`, `Received replication events`,
				NewRateCol(`msgs`, `Received messages (usually transactions) per second`, 4, `wsrep_received`, 0, NumberUnits),
				NewRateCol(`data`, `Received bytes per second`, 4, `wsrep_received_bytes`, 0, MemoryUnits),
				NewGaugeCol(`queue`, `Received replication apply queue`, 3, `wsrep_local_recv_queue`, 0, NumberUnits),
			),
			NewGroupCol(`FlowC`, `Flow control stats`,
				NewDiffCol(`paused`, `Flow control paused (could be from anywhere in the cluster)`, 5, `wsrep_flow_control_paused_ns`, 0, NanoSecondUnits),
				NewDiffCol(`snt`, `Flow control sent messages (could be starting or stopping FC)`, 3, `wsrep_flow_control_sent`, 0, NumberUnits),
			),
			NewGroupCol(`Conflcts`, `Galera replication conflicts (on this node)`,
				NewDiffCol(`lcf`, `Local certification failures since last sample`, 3, `wsrep_local_cert_failures`, 0, NumberUnits),
				NewDiffCol(`bfa`, `Brute force aborts since last sample`, 3, `wsrep_local_bf_aborts`, 0, NumberUnits),
			),
			NewGroupCol(`Gcache`, `Galera cache (gcache) information`,
				NewCurDiffCol(`ist`, `Gcached transactions`, 5, `wsrep_last_committed`, `wsrep_local_cached_downto`, 0, NumberUnits),
				NewGaugeCol(`idx`, `Certification index size (keys)`, 4, `wsrep_cert_index_size`, 0, NumberUnits),
			),
			NewGroupCol(`Apply`, `Theoretical and actual apply efficiency`,
				NewPercentCol(`%ef`, `Percent of threads being used`, 4, `wsrep_apply_window`, `V_wsrep_slave_threads`, 0),
			),
		),
		`qcache`: NewNormalView(`Query cache stats`,
			NewStringCol(`type`, `Query cache type`, 6, `V_query_cache_type`),
			NewRateSumCol(`sel`, `Total Selects + Qcache Hits per second`, 4, 0, NumberUnits, `com_select`, `qcache_hits`),
			NewRateCol(`hits`, `Query cache hits per second`, 4, `qcache_hits`, 0, NumberUnits),
			NewRateCol(`ins`, `Query inserts per second (new entries to the cache)`, 4, `qcache_inserts`, 0, NumberUnits),
			NewRateCol(`notc`, `Queries not cached per second (either can not be cached, or SELECT SQL_NO_CACHE)`, 4, `qcache_not_cached`, 0, NumberUnits),
			NewGaugeCol(`tot`, `Total queries in the cache`, 4, `qcache_queries_in_cache`, 0, NumberUnits),
			NewRateCol(`lowm`, `Low memory prunes (cache entries removed due to memory limit)`, 4, `qcache_lowmem_prunes`, 0, NumberUnits),
			NewPercentCol(`%free`, `Percent of cache memory free`, 5, `qcache_free_blocks`, `qcache_total_blocks`, 0),
		),
		`myisam`: NewNormalView(`MyISAM stats`,
			NewGroupCol(`Key Buffer`, `Key Buffer Stats`,
				NewGaugeCol(`used`, `Current Key Buffer blocks unused`, 6, `key_blocks_unused`, 0, NumberUnits),
				NewGaugeCol(`maxu`, `Maxiumum Key Buffer blocks used`, 6, `key_blocks_used`, 0, NumberUnits),
			),
			NewGroupCol(`I/O`, `MyISAM Key Buffer IO Stats (not data)`,
				NewRateCol(`logr`, `Logical read requests`, 5, `key_read_requests`, 0, NumberUnits),
				NewRateCol(`phyr`, `Physical reads (cache misses)`, 5, `key_reads`, 0, NumberUnits),
				NewRateCol(`logw`, `Logical write requests`, 5, `key_write_requests`, 0, NumberUnits),
				NewRateCol(`phyw`, `Physical writes`, 5, `key_writes`, 0, NumberUnits),
			),
		),
		`commands`: NewNormalView(`Sorted list of all commands run in a given interval`,
			NewFuncCol(`Counts`, `All commands tracked by the Com_* counters`, 4, func(state *MyqState, c Col) (chan string) {
				var all_diffs []float64
				diff_variables := map[float64][]string{}

				// Get the rate for every ^com* variable
				for _, variable := range expand_variables([]string{`^com.*`}, state.Cur) {
					diff := calculate_diff(state.Cur.getF(variable), state.Prev.getF(variable))

					// Skip those without activity
					if diff <= 0 {
						continue
					}

					// Create the [] slice for a rate we haven't seen yet
					if _, ok := diff_variables[diff]; ok == false {
						diff_variables[diff] = make([]string, 0)
						all_diffs = append(all_diffs, diff) // record the diff the first time
					}

					// Push the variable name onto the rate slice
					diff_variables[diff] = append(diff_variables[diff], variable)
				}

				// Sort all the rates so we can iterate through them from big to small
				sort.Sort(sort.Reverse(sort.Float64Slice(all_diffs)))

				// Each rate
				ch := make( chan string )
				go func() {
					defer close(ch)
					for _, diff := range all_diffs {
						var out bytes.Buffer
						out.WriteString( fit_string(collapse_number(diff, c.Width(), 0, NumberUnits), c.Width()))
						out.WriteString(fmt.Sprintf(" %v", diff_variables[diff]))
						ch <- out.String()
					}
				}()
				return ch
			}),
		),
	}
}
