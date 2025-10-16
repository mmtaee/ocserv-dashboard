package home

func ParseServerStatus(flat map[string]interface{}) ServerStatusResponse {
	// Helper funcs
	getStr := func(key string) string {
		if v, ok := flat[key].(string); ok {
			return v
		}
		return ""
	}
	getInt := func(key string) int {
		if v, ok := flat[key].(float64); ok {
			return int(v)
		}
		return 0
	}
	getInt64 := func(key string) int64 {
		if v, ok := flat[key].(float64); ok {
			return int64(v)
		}
		return 0
	}

	return ServerStatusResponse{
		GeneralInfo: GeneralInfo{
			ServerPID:           getInt("Server PID"),
			SecModPID:           getInt("Sec-mod PID"),
			SecModInstanceCount: getInt("Sec-mod instance count"),
			Status:              getStr("Status"),
			UpSince:             getStr("Up since"),
			UpSinceDuration:     getStr("_Up since"),
			ActiveSessions:      getInt("Active sessions"),
			TotalSessions:       getInt("Total sessions"),
			TotalAuthFailures:   getInt("Total authentication failures"),
			IPsInBanList:        getInt("IPs in ban list"),
			MedianLatency:       getStr("Median latency"),
			STDEVLatency:        getStr("STDEV latency"),

			RawMedianLatency: getInt64("raw_median_latency"),
			RawSTDEVLatency:  getInt64("raw_stdev_latency"),
			RawUpSince:       getInt64("raw_up_since"),
			Uptime:           getInt64("uptime"),
		},

		CurrentStats: CurrentStats{
			LastStatsReset:           getStr("Last stats reset"),
			LastStatsResetDuration:   getStr("_Last stats reset"),
			SessionsHandled:          getInt("Sessions handled"),
			TimedOutSessions:         getInt("Timed out sessions"),
			TimedOutIdleSessions:     getInt("Timed out (idle) sessions"),
			ClosedDueToErrorSessions: getInt("Closed due to error sessions"),
			AuthenticationFailures:   getInt("Authentication failures"),
			AverageAuthTime:          getStr("Average auth time"),
			MaxAuthTime:              getStr("Max auth time"),
			AverageSessionTime:       getStr("Average session time"),
			MaxSessionTime:           getStr("Max session time"),
			RX:                       getStr("RX"),
			TX:                       getStr("TX"),

			RawRX:             getInt64("raw_rx"),
			RawTX:             getInt64("raw_tx"),
			RawAvgAuthTime:    getInt64("raw_avg_auth_time"),
			RawMaxAuthTime:    getInt64("raw_max_auth_time"),
			RawAvgSessionTime: getInt64("raw_avg_session_time"),
			RawMaxSessionTime: getInt64("raw_max_session_time"),
			RawLastStatsReset: getInt64("raw_last_stats_reset"),
		},
	}
}
