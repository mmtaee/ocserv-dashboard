package models

type IPBan struct {
	IP       string `json:"IP"`
	Since    string `json:"Since"`
	SinceAlt string `json:"_Since"` // maps to "_Since" in JSON
	Score    int    `json:"Score"`
}

type Iroute struct {
	ID       int      `json:"ID"`
	Username string   `json:"Username"`
	VHost    string   `json:"vhost"`
	Device   string   `json:"Device"`
	IP       string   `json:"IP"`
	IRoutes  []string `json:"iRoutes"`
}

type OnlineUserSession struct {
	ID               int    `json:"ID" validate:"required"`
	Username         string `json:"Username"`
	Group            string `json:"Groupname"`
	AverageRX        string `json:"Average RX"`
	AverageTX        string `json:"Average TX"`
	LastConnectedAt  string `json:"_Last connected at"`
	IPv4             string `json:"IPv4" validate:"required"`
	VHost            string `json:"vhost" validate:"required"`
	Device           string `json:"Device" validate:"required"`
	SessionStartedAt string `json:"Session started at" validate:"required"`
}

type ServerVersion struct {
	OcservVersion string `json:"ocserv_version"`
	OcctlVersion  string `json:"occtl_version"`
}

type OcservInfo struct {
	Version *ServerVersion `json:"version" validate:"required"`
	Status  string         `json:"status" validate:"required"`
}

type IPBanPoints struct {
	IP    string `json:"IP"`
	Since string `json:"Since"`
	Until string `json:"_Since"`
	Score int    `json:"Score"`
}

type IRoute struct {
	ID       string `json:"ID"`
	Username string `json:"Username"`
	Vhost    string `json:"vhost"`
	Device   string `json:"Device"`
	IP       string `json:"IP"`
	IRoute   string `json:"iRoutes"`
}

type OcservStatusGeneralInfo struct {
	ServerPID           int    `json:"Server PID"`
	SecModPID           int    `json:"Sec-mod PID"`
	SecModInstanceCount int    `json:"Sec-mod instance count"`
	Status              string `json:"Status"`
	UpSince             string `json:"Up since"`
	UpSinceDuration     string `json:"_Up since"`
	ActiveSessions      int    `json:"Active sessions"`
	TotalSessions       int    `json:"Total sessions"`
	TotalAuthFailures   int    `json:"Total authentication failures"`
	IPsInBanList        int    `json:"IPs in ban list"`
	MedianLatency       string `json:"Median latency"`
	STDEVLatency        string `json:"STDEV latency"`
	RawMedianLatency    int64  `json:"raw_median_latency"`
	RawSTDEVLatency     int64  `json:"raw_stdev_latency"`
	RawUpSince          int64  `json:"raw_up_since"`
	Uptime              int64  `json:"uptime"`
}

type OcservStatusCurrentStats struct {
	LastStatsReset           string `json:"Last stats reset"`
	LastStatsResetDuration   string `json:"_Last stats reset"`
	SessionsHandled          int    `json:"Sessions handled"`
	TimedOutSessions         int    `json:"Timed out sessions"`
	TimedOutIdleSessions     int    `json:"Timed out (idle) sessions"`
	ClosedDueToErrorSessions int    `json:"Closed due to error sessions"`
	AuthenticationFailures   int    `json:"Authentication failures"`
	AverageAuthTime          string `json:"Average auth time"`
	MaxAuthTime              string `json:"Max auth time"`
	AverageSessionTime       string `json:"Average session time"`
	MaxSessionTime           string `json:"Max session time"`
	RX                       string `json:"RX"`
	TX                       string `json:"TX"`
	RawRX                    int64  `json:"raw_rx"`
	RawTX                    int64  `json:"raw_tx"`
	RawAvgAuthTime           int64  `json:"raw_avg_auth_time"`
	RawMaxAuthTime           int64  `json:"raw_max_auth_time"`
	RawAvgSessionTime        int64  `json:"raw_avg_session_time"`
	RawMaxSessionTime        int64  `json:"raw_max_session_time"`
	RawLastStatsReset        int64  `json:"raw_last_stats_reset"`
}

type OcservStatusResponse struct {
	GeneralInfo  OcservStatusGeneralInfo  `json:"general_info"`
	CurrentStats OcservStatusCurrentStats `json:"current_stats"`
}

func ParseOcservServerStatus(flat map[string]interface{}) OcservStatusResponse {
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

	return OcservStatusResponse{
		GeneralInfo: OcservStatusGeneralInfo{
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
			RawMedianLatency:    getInt64("raw_median_latency"),
			RawSTDEVLatency:     getInt64("raw_stdev_latency"),
			RawUpSince:          getInt64("raw_up_since"),
			Uptime:              getInt64("uptime"),
		},
		CurrentStats: OcservStatusCurrentStats{
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
			RawRX:                    getInt64("raw_rx"),
			RawTX:                    getInt64("raw_tx"),
			RawAvgAuthTime:           getInt64("raw_avg_auth_time"),
			RawMaxAuthTime:           getInt64("raw_max_auth_time"),
			RawAvgSessionTime:        getInt64("raw_avg_session_time"),
			RawMaxSessionTime:        getInt64("raw_max_session_time"),
			RawLastStatsReset:        getInt64("raw_last_stats_reset"),
		},
	}
}
