package home

import (
	"github.com/mmtaee/ocserv-dashboard/api/internal/repository"
	"github.com/mmtaee/ocserv-dashboard/core/models"
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

type OcservStatusResponse struct {
	GeneralInfo  GeneralInfo  `json:"general_info"`
	CurrentStats CurrentStats `json:"current_stats"`
}

type GetHomeUser struct {
	Total  int64                      `json:"total" validate:"omitempty"`
	Online []models.OnlineUserSession `json:"online_users_session" validate:"omitempty"`
}

type GetHomeResponse struct {
	Statistics       *[]models.DailyTraffic       `json:"statistics" validate:"omitempty"`
	Users            GetHomeUser                  `json:"users" validate:"omitempty"`
	IPBans           *[]models.IPBanPoints        `json:"ip_bans" validate:"omitempty"`
	TopBandwidthUser repository.TopBandwidthUsers `json:"top_bandwidth_user" validate:"omitempty"`
	TotalBandwidth   repository.TotalBandwidths   `json:"total_bandwidth" validate:"omitempty"`
	// TelegramService is a lightweight snapshot for the admin dashboard (no outbound Telegram calls).
	TelegramService *TelegramServiceStatus `json:"telegram_service,omitempty" validate:"omitempty"`
	//IRoutes    *[]models.Iroute       `json:"iroutes" validate:"omitempty"` // has bug on version 1.2.4
}

// TelegramServiceStatus summarizes Telegram integration for the home screen.
type TelegramServiceStatus struct {
	Enabled     bool   `json:"enabled"`
	HasBotToken bool   `json:"has_bot_token"`
	BotUsername string `json:"bot_username,omitempty"`
}

type CPU struct {
	AvgPercent float64 `json:"avg_percent"`
	UsedUnits  float64 `json:"used_units"`
	Total      int     `json:"total"`
}

type RAM struct {
	Used        float64 `json:"used"`
	Total       float64 `json:"total"`
	UsedPercent float64 `json:"used_percent"`
}

type Swap struct {
	Used        float64 `json:"used"`
	Total       float64 `json:"total"`
	UsedPercent float64 `json:"used_percent"`
}

type Disk struct {
	Used        float64 `json:"used"`
	Total       float64 `json:"total"`
	UsedPercent float64 `json:"used_percent"`
}

type DockerStats struct {
	Name string `json:"name" validate:"required"`
	CPU  CPU    `json:"cpu" validate:"omitempty"`
	RAM  RAM    `json:"ram" validate:"omitempty"`
}

type DockerService struct {
	Postgres    DockerStats `json:"postgres" validate:"required"`
	Ocserv      DockerStats `json:"ocserv" validate:"required"`
	LogStream   DockerStats `json:"log_stream" validate:"required"`
	UserExpiry  DockerStats `json:"user_expiry" validate:"required"`
	TelegramBot DockerStats `json:"telegram_bot" validate:"omitempty"`
	Web         DockerStats `json:"web" validate:"required"`
}

type ServerStatusResponse struct {
	CPU  CPU  `json:"cpu"`
	RAM  RAM  `json:"ram"`
	Swap Swap `json:"swap"`
	Disk Disk `json:"disk"`
}
