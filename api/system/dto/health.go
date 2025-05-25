package dto

import "time"

type HealthStatusResponse struct {
	Status     string    `json:"status"`
	Timestamp  time.Time `json:"timestamp"`
	Components struct {
		Database struct {
			Status  string `json:"status"`
			Latency string `json:"latency"`
			Details struct {
				Type       string `json:"type"`
				Version    string `json:"version"`
				Connection struct {
					PoolSize          int    `json:"pool_size"`
					ActiveConnections int    `json:"active_connections"`
					IdleConnections   int    `json:"idle_connections"`
					MaxConnectionAge  string `json:"max_connection_age"`
					ConnectionTimeout string `json:"connection_timeout"`
				} `json:"connection"`
				Operations struct {
					TotalQueries     int64  `json:"total_queries"`
					QueriesPerSecond int    `json:"queries_per_second"`
					AverageQueryTime string `json:"average_query_time"`
					SlowQueries      int    `json:"slow_queries"`
					FailedQueries    int    `json:"failed_queries"`
				} `json:"operations"`
				Collections struct {
					Total          int    `json:"total"`
					Indexes        int    `json:"indexes"`
					TotalDocuments int64  `json:"total_documents"`
					TotalSize      string `json:"total_size"`
				} `json:"collections"`
				Replication struct {
					Status     string `json:"status"`
					ReplicaSet string `json:"replica_set"`
					Members    int    `json:"members"`
					Lag        string `json:"lag"`
				} `json:"replication"`
			} `json:"details"`
		} `json:"database"`
		Redis struct {
			Status  string `json:"status"`
			Latency string `json:"latency"`
			Details struct {
				Version string `json:"version"`
				Mode    string `json:"mode"`
				Nodes   int    `json:"nodes"`
				Memory  struct {
					UsedMemory      string `json:"used_memory"`
					UsedMemoryPeak  string `json:"used_memory_peak"`
					UsedMemoryLua   string `json:"used_memory_lua"`
					MaxMemory       string `json:"maxmemory"`
					MaxMemoryPolicy string `json:"maxmemory_policy"`
				} `json:"memory"`
				Clients struct {
					Connected  int `json:"connected"`
					Blocked    int `json:"blocked"`
					MaxClients int `json:"max_clients"`
				} `json:"clients"`
				Stats struct {
					TotalConnectionsReceived int64   `json:"total_connections_received"`
					TotalCommandsProcessed   int64   `json:"total_commands_processed"`
					InstantaneousOpsPerSec   int     `json:"instantaneous_ops_per_sec"`
					HitRatio                 float64 `json:"hit_ratio"`
					KeyspaceHits             int64   `json:"keyspace_hits"`
					KeyspaceMisses           int64   `json:"keyspace_misses"`
				} `json:"stats"`
				Persistence struct {
					RdbLastSaveTime         time.Time `json:"rdb_last_save_time"`
					RdbChangesSinceLastSave int64     `json:"rdb_changes_since_last_save"`
					AofEnabled              bool      `json:"aof_enabled"`
					AofRewriteInProgress    bool      `json:"aof_rewrite_in_progress"`
				} `json:"persistence"`
			} `json:"details"`
		} `json:"redis"`
		Server struct {
			Status  string `json:"status"`
			Details struct {
				Hostname    string    `json:"hostname"`
				Environment string    `json:"environment"`
				GoVersion   string    `json:"go_version"`
				StartTime   time.Time `json:"start_time"`
				Uptime      string    `json:"uptime"`
				Memory      struct {
					Alloc        string    `json:"alloc"`
					TotalAlloc   string    `json:"total_alloc"`
					Sys          string    `json:"sys"`
					NumGC        uint32    `json:"num_gc"`
					GCPauseTotal string    `json:"gc_pause_total"`
					LastGC       time.Time `json:"last_gc"`
				} `json:"memory"`
				Goroutines struct {
					Total  int `json:"total"`
					Active int `json:"active"`
					Idle   int `json:"idle"`
				} `json:"goroutines"`
				CPU struct {
					NumCPU       int       `json:"num_cpu"`
					NumGoroutine int       `json:"num_goroutine"`
					CPUUsage     float64   `json:"cpu_usage"`
					LoadAverage  []float64 `json:"load_average"`
				} `json:"cpu"`
				Disk struct {
					Total        string  `json:"total"`
					Used         string  `json:"used"`
					Free         string  `json:"free"`
					UsagePercent float64 `json:"usage_percent"`
				} `json:"disk"`
				Network struct {
					TotalConnections     int    `json:"total_connections"`
					ActiveConnections    int    `json:"active_connections"`
					ConnectionsPerSecond int    `json:"connections_per_second"`
					BytesIn              string `json:"bytes_in"`
					BytesOut             string `json:"bytes_out"`
				} `json:"network"`
			} `json:"details"`
		} `json:"server"`
		Services struct {
			Status  string `json:"status"`
			Details struct {
				AuthService struct {
					Status         string `json:"status"`
					Latency        string `json:"latency"`
					ActiveSessions int    `json:"active_sessions"`
					FailedAttempts int    `json:"failed_attempts"`
				} `json:"auth_service"`
				MediaService struct {
					Status         string `json:"status"`
					Latency        string `json:"latency"`
					StorageUsed    string `json:"storage_used"`
					FilesProcessed int64  `json:"files_processed"`
				} `json:"media_service"`
				NotificationService struct {
					Status       string `json:"status"`
					Latency      string `json:"latency"`
					QueueSize    int    `json:"queue_size"`
					MessagesSent int64  `json:"messages_sent"`
				} `json:"notification_service"`
			} `json:"details"`
		} `json:"services"`
		Security struct {
			Status  string `json:"status"`
			Details struct {
				SSL struct {
					Enabled    bool      `json:"enabled"`
					Version    string    `json:"version"`
					CertExpiry time.Time `json:"cert_expiry"`
				} `json:"ssl"`
				RateLimiting struct {
					Enabled         bool   `json:"enabled"`
					CurrentRequests int    `json:"current_requests"`
					MaxRequests     int    `json:"max_requests"`
					Window          string `json:"window"`
				} `json:"rate_limiting"`
				Firewall struct {
					Status          string `json:"status"`
					BlockedIPs      int    `json:"blocked_ips"`
					TotalRequests   int64  `json:"total_requests"`
					BlockedRequests int    `json:"blocked_requests"`
				} `json:"firewall"`
			} `json:"details"`
		} `json:"security"`
	} `json:"components"`
	Alerts []struct {
		Level     string    `json:"level"`
		Message   string    `json:"message"`
		Timestamp time.Time `json:"timestamp"`
		Component string    `json:"component"`
	} `json:"alerts"`
	Metrics struct {
		ResponseTime struct {
			P50 string `json:"p50"`
			P90 string `json:"p90"`
			P99 string `json:"p99"`
		} `json:"response_time"`
		ErrorRate         float64 `json:"error_rate"`
		RequestsPerSecond int     `json:"requests_per_second"`
	} `json:"metrics"`
}
