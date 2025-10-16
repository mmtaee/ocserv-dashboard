package home

import (
	"github.com/mmtaee/ocserv-users-management/api/internal/repository"
	"github.com/mmtaee/ocserv-users-management/common/models"
)

type GeneralInfo struct {
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

	// raw fields if you want
	RawMedianLatency int64 `json:"raw_median_latency"`
	RawSTDEVLatency  int64 `json:"raw_stdev_latency"`
	RawUpSince       int64 `json:"raw_up_since"`
	Uptime           int64 `json:"uptime"`
}

type CurrentStats struct {
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

	// raw fields if you want
	RawRX             int64 `json:"raw_rx"`
	RawTX             int64 `json:"raw_tx"`
	RawAvgAuthTime    int64 `json:"raw_avg_auth_time"`
	RawMaxAuthTime    int64 `json:"raw_max_auth_time"`
	RawAvgSessionTime int64 `json:"raw_avg_session_time"`
	RawMaxSessionTime int64 `json:"raw_max_session_time"`
	RawLastStatsReset int64 `json:"raw_last_stats_reset"`
}

type ServerStatusResponse struct {
	GeneralInfo  GeneralInfo  `json:"general_info"`
	CurrentStats CurrentStats `json:"current_stats"`
}

type GetHomeUser struct {
	Total  int64                       `json:"total" validate:"omitempty"`
	Online *[]models.OnlineUserSession `json:"online_users_session" validate:"omitempty"`
}

type GetHomeResponse struct {
	ServerStatus     ServerStatusResponse         `json:"server_status" validate:"required"`
	Statistics       *[]models.DailyTraffic       `json:"statistics" validate:"omitempty"`
	Users            GetHomeUser                  `json:"users" validate:"omitempty"`
	IPBans           *[]models.IPBanPoints        `json:"ip_bans" validate:"omitempty"`
	TopBandwidthUser repository.TopBandwidthUsers `json:"top_bandwidth_user" validate:"omitempty"`
	TotalBandwidth   repository.TotalBandwidths   `json:"total_bandwidth" validate:"omitempty"`
	//IRoutes    *[]models.Iroute       `json:"iroutes" validate:"omitempty"` // has bug on version 1.2.4
}
