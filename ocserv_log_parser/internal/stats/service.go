package stats

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/mmtaee/ocserv-dashboard/core/models"
	"github.com/mmtaee/ocserv-dashboard/core/pkg/database"
	"github.com/mmtaee/ocserv-dashboard/core/pkg/logger"
	"github.com/mmtaee/ocserv-dashboard/core/pkg/ocserv/occtl"
	"github.com/mmtaee/ocserv-dashboard/core/pkg/ocserv/user"
	"gorm.io/gorm"
)

type StatService struct {
	ctx                 context.Context
	stream              <-chan string
	ocservUserHandler   user.OcservUserInterface
	ocservOcctlHandler  occtl.OcservOcctlInterface
	// TODO: occtlDockerRepo (for webhook-based docker mode) is temporarily disabled
	// occtlDockerRepo     occtlDocker.OcservOcctlUsersDocker
	sessionStats        map[string]UserStats
	pendingMainSessions map[string][]pendingMainSession
	workerSessionIDs    map[string]string
}

type pendingMainSession struct {
	Endpoint  string
	CreatedAt time.Time
}

func NewStatService(ctx context.Context, stream chan string) *StatService {
	s := &StatService{
		ctx:                 ctx,
		stream:              stream,
		sessionStats:        make(map[string]UserStats),
		pendingMainSessions: make(map[string][]pendingMainSession),
		workerSessionIDs:    make(map[string]string),
	}
	// TODO: Docker mode (webhook-based occtl/lock operations) is temporarily disabled
	// For now, only non-docker mode (direct occtl/ocpasswd calls) is supported
	s.ocservUserHandler = user.NewOcservUser()
	s.ocservOcctlHandler = occtl.NewOcservOcctl()

	return s
}

func (s *StatService) CalculateUserStats() {
	for {
		select {
		case <-s.ctx.Done():
			logger.Warn("Stopping: context cancelled")
			return

		case line, ok := <-s.stream:
			if !ok {
				logger.Warn("Stream closed, exiting...")
				return
			}

			cleanLine := strings.TrimSpace(line) // remove whitespace/newlines and normalize case

			if strings.Contains(cleanLine, "server shutdown complete") {
				logger.Error("Ocserv server shutdown abnormally")
				p, _ := os.FindProcess(os.Getpid())
				_ = p.Signal(syscall.SIGTERM)
				return
			}

			if !strings.Contains(cleanLine, "worker[") && !strings.Contains(cleanLine, "main[") {
				continue
			}

			s.trackSessionIdentity(cleanLine)

			if strings.Contains(cleanLine, "sent periodic stats") {
				stats, err := s.getPeriodicStat(cleanLine)
				if err != nil {
					logger.Error("Failed to parse periodic RxTx stats: %v", err)
				}

				if stats != nil {
					if err = s.saveRxTxDelta(s.ctx, stats, false); err != nil {
						logger.Error("Failed to save periodic RxTx stats: %v", err)
					}
				}
			}

			if strings.Contains(cleanLine, "user disconnected") {
				stats, err := s.getDisconnectStat(cleanLine)
				if err != nil {
					logger.Error("Failed to parse disconnect RxTx stats: %v", err)
				}

				if stats != nil {
					if err = s.saveRxTxDelta(s.ctx, stats, true); err != nil {
						logger.Error("Failed to save disconnect RxTx stats: %v", err)
					}
				}

				// replace main word with worker to extract user session log
				cleanLine = strings.Replace(cleanLine, "main[", "worker[", 1)
			}

			logger.Info("Starting get user session from line: %s", cleanLine)

			sessionLog := s.getUserSessionLog(cleanLine)
			if sessionLog == nil {
				continue
			}

			if err := s.saveSessionLog(s.ctx, sessionLog); err != nil {
				logger.Error("Error saving session msg (%v): %v", sessionLog.Username, err)
				continue
			}
		}
	}
}

func (s *StatService) getUserSessionLog(cleanLine string) *models.OcservUserSessionLog {
	workerRe := regexp.MustCompile(`worker\[(?P<user>[^\]]+)\]:\s*(?P<rest>.*)`)
	ipRe := regexp.MustCompile(`^(?P<ip>\d+\.\d+\.\d+\.\d+)(?::\d+)?\s+(?P<rest>.*)$`)
	var username, ip, msg string

	// Step 1: extract worker
	if m := workerRe.FindStringSubmatch(cleanLine); m != nil {
		username = m[1]
		msg = m[2]
	} else {
		logger.Error("No worker found in line: %s", cleanLine)
		return nil
	}

	// Step 2: extract IP
	if m := ipRe.FindStringSubmatch(msg); m != nil {
		ip = m[1]
		msg = m[2]
	}

	// Step 3: detect event
	var event string

	switch {
	case strings.Contains(msg, "User-agent"):
		event = models.EventUseragent
	case strings.Contains(msg, "DTLS handshake completed"):
		event = models.EventHandshake
	case strings.Contains(msg, "sent periodic stats"):
		event = models.EventPeriodicStats
	case strings.Contains(msg, "user disconnected"):
		event = models.EventDisconnect
	default:
		return nil
	}

	return &models.OcservUserSessionLog{
		Username: username,
		IP:        ip,
		Event:     event,
		Message:   msg,
	}
}

func (s *StatService) getPeriodicStat(cleanLine string) (*UserStats, error) {
	reTxRx := regexp.MustCompile(`worker\[([^\]]+)\]:\s*(\S+)\s+sent periodic stats\s+\((?:in|rx):\s*(\d+),\s*(?:out|tx):\s*(\d+)\)`)
	matchRxTx := reTxRx.FindStringSubmatch(cleanLine)
	if len(matchRxTx) <= 4 {
		return nil, nil
	}

	rx, err := strconv.Atoi(matchRxTx[3])
	if err != nil {
		return nil, err
	}

	tx, err := strconv.Atoi(matchRxTx[4])
	if err != nil {
		return nil, err
	}

	ip := normalizeSessionIP(matchRxTx[2])

	return &UserStats{
		Username:  matchRxTx[1],
		IP:        ip,
		SessionID: s.workerSessionID(matchRxTx[1], ip, cleanLine),
		RX:        rx,
		TX:        tx,
	}, nil
}

func (s *StatService) getDisconnectStat(cleanLine string) (*UserStats, error) {
	reTxRx := regexp.MustCompile(`main\[([^\]]+)\]:(\S+)\s+user disconnected.*rx:\s*(\d+),\s*tx:\s*(\d+)`)
	matchRxTx := reTxRx.FindStringSubmatch(cleanLine)
	if len(matchRxTx) == 5 {
		rx, err := strconv.Atoi(matchRxTx[3])
		if err != nil {
			return nil, err
		}

		tx, err := strconv.Atoi(matchRxTx[4])
		if err != nil {
			return nil, err
		}

		return &UserStats{
			Username:  matchRxTx[1],
			IP:        normalizeSessionIP(matchRxTx[2]),
			SessionID: matchRxTx[2],
			RX:        rx,
			TX:        tx,
		}, nil
	}

	fallbackRe := regexp.MustCompile(`main\[([^\]]+)\].*rx:\s*(\d+),\s*tx:\s*(\d+)`)
	fallbackMatch := fallbackRe.FindStringSubmatch(cleanLine)
	if len(fallbackMatch) != 4 {
		return nil, nil
	}

	rx, err := strconv.Atoi(fallbackMatch[2])
	if err != nil {
		return nil, err
	}

	tx, err := strconv.Atoi(fallbackMatch[3])
	if err != nil {
		return nil, err
	}

	return &UserStats{
		Username:  fallbackMatch[1],
		SessionID: processIDFromLine(cleanLine),
		RX:        rx,
		TX:        tx,
	}, nil
}

func normalizeSessionIP(endpoint string) string {
	endpoint = strings.TrimSpace(endpoint)
	if strings.Count(endpoint, ":") == 1 {
		parts := strings.Split(endpoint, ":")
		return parts[0]
	}
	return endpoint
}

func processIDFromLine(line string) string {
	re := regexp.MustCompile(`(?:^|\s)ocserv\[(\d+)\]:`)
	match := re.FindStringSubmatch(line)
	if len(match) != 2 {
		return ""
	}

	return match[1]
}

func usernameIPKey(username, ip string) string {
	return fmt.Sprintf("%s|%s", username, ip)
}

func workerIdentityKey(username, ip, pid string) string {
	return fmt.Sprintf("%s|%s|pid:%s", username, ip, pid)
}

func sessionStatsKey(stats *UserStats) string {
	if stats.SessionID != "" {
		return fmt.Sprintf("%s|%s|%s", stats.Username, stats.IP, stats.SessionID)
	}

	return fmt.Sprintf("%s|%s|unknown", stats.Username, stats.IP)
}

func (s *StatService) trackSessionIdentity(line string) {
	mainRe := regexp.MustCompile(`main\[([^\]]+)\]:(\S+)\s+new user session`)
	if match := mainRe.FindStringSubmatch(line); len(match) == 3 {
		username := match[1]
		endpoint := match[2]
		ip := normalizeSessionIP(endpoint)
		key := usernameIPKey(username, ip)

		s.pendingMainSessions[key] = append(s.pendingMainSessions[key], pendingMainSession{
			Endpoint:  endpoint,
			CreatedAt: time.Now(),
		})

		return
	}

	workerRe := regexp.MustCompile(`worker\[([^\]]+)\]:\s*(\S+)`)
	match := workerRe.FindStringSubmatch(line)
	if len(match) != 3 {
		return
	}

	pid := processIDFromLine(line)
	if pid == "" {
		return
	}

	username := match[1]
	ip := normalizeSessionIP(match[2])
	workerKey := workerIdentityKey(username, ip, pid)
	if _, ok := s.workerSessionIDs[workerKey]; ok {
		return
	}

	pendingKey := usernameIPKey(username, ip)
	pending := s.pendingMainSessions[pendingKey]
	if len(pending) == 0 {
		return
	}

	now := time.Now()
	filtered := pending[:0]
	for _, item := range pending {
		if now.Sub(item.CreatedAt) <= 2*time.Minute {
			filtered = append(filtered, item)
		}
	}

	if len(filtered) == 0 {
		delete(s.pendingMainSessions, pendingKey)
		return
	}

	s.workerSessionIDs[workerKey] = filtered[0].Endpoint

	if len(filtered) == 1 {
		delete(s.pendingMainSessions, pendingKey)
		return
	}

	s.pendingMainSessions[pendingKey] = filtered[1:]
}

func (s *StatService) workerSessionID(username, ip, line string) string {
	pid := processIDFromLine(line)
	if pid == "" {
		return ""
	}

	workerKey := workerIdentityKey(username, ip, pid)
	if sessionID, ok := s.workerSessionIDs[workerKey]; ok {
		return sessionID
	}

	return "pid:" + pid
}

func (s *StatService) hasSessionStatsForUserIP(username, ip string) bool {
	prefix := fmt.Sprintf("%s|%s|", username, ip)

	for key := range s.sessionStats {
		if strings.HasPrefix(key, prefix) {
			return true
		}
	}

	return false
}

func (s *StatService) saveRxTxDelta(ctx context.Context, stats *UserStats, final bool) error {
	key := sessionStatsKey(stats)
	lastStats, found := s.sessionStats[key]

	if final && !found && s.hasSessionStatsForUserIP(stats.Username, stats.IP) {
		logger.Warn("Skipping unmatched final RxTx stats for user=%s ip=%s session=%s RX=%d TX=%d", stats.Username, stats.IP, stats.SessionID, stats.RX, stats.TX)
		return nil
	}

	deltaRx := stats.RX
	deltaTx := stats.TX

	if found {
		deltaRx = stats.RX - lastStats.RX
		deltaTx = stats.TX - lastStats.TX
	}

	// A lower value means ocserv started a new session for the same stats key.
	// In that case, the current stat is the first value of the new session.
	if deltaRx < 0 || deltaTx < 0 {
		deltaRx = stats.RX
		deltaTx = stats.TX
	}

	// Keep the latest value, including final disconnect values.
	// This prevents duplicate final/periodic lines from being counted again.
	s.sessionStats[key] = *stats

	if deltaRx == 0 && deltaTx == 0 {
		return nil
	}

	return s.saveRxTx(ctx, &UserStats{
		Username:  stats.Username,
		IP:        stats.IP,
		SessionID: stats.SessionID,
		RX:        deltaRx,
		TX:        deltaTx,
	})
}

func (s *StatService) saveRxTx(ctx context.Context, u *UserStats) error {
	logger.Info("SaveRxTx called for user=%s ip=%s session=%s RX=%d TX=%d", u.Username, u.IP, u.SessionID, u.RX, u.TX)

	db := database.GetConnection()
	db = db.WithContext(ctx)

	var ocUser models.OcservUser

	err := db.Where("username = ? ", u.Username).First(&ocUser).Error
	if err != nil {
		logger.Error("Error finding oc user: %v", err)
		return err
	}

	traffic := models.OcservUserTrafficStatistics{
		OcUserID: ocUser.ID,
		Rx:       u.RX,
		Tx:       u.TX,
	}

	err = db.Create(&traffic).Error
	if err != nil {
		logger.Error("Error creating traffic stats: %v", err)
		return err
	}

	ocUser.Rx += u.RX
	ocUser.Tx += u.TX

	trafficSizeBytes := ocUser.TrafficSize

	totalMonthStats, err := s.getCurrentMonthTotals(db, ocUser.ID, ocUser.UsageResetAt)
	if err != nil {
		logger.Error("Error getting current month stats: %v", err)
		return err
	}

	shouldLock := false
	switch ocUser.TrafficType {
	case models.TotallyTransmit:
		shouldLock = int64(ocUser.Tx) >= trafficSizeBytes

	case models.TotallyReceive:
		shouldLock = int64(ocUser.Rx) >= trafficSizeBytes

	case models.TotallyRxTx:
		shouldLock = int64(ocUser.Rx)+int64(ocUser.Tx) >= trafficSizeBytes

	case models.MonthlyTransmit:
		shouldLock = int64(totalMonthStats.TotalTx) >= trafficSizeBytes

	case models.MonthlyReceive:
		shouldLock = int64(totalMonthStats.TotalRx) >= trafficSizeBytes

	case models.MonthlyRxTx:
		shouldLock = int64(totalMonthStats.TotalRx)+int64(totalMonthStats.TotalTx) >= trafficSizeBytes

	case models.Free:

	default:
		logger.Error("Unknown traffic type: %v", ocUser.TrafficType)
	}
	wasLocked := ocUser.IsLocked
	if shouldLock {
		ocUser.IsLocked = true
	}

	now := time.Now()
	if shouldLock && !wasLocked {
		_, err = s.ocservOcctlHandler.DisconnectUser(ocUser.Username)
		if err != nil {
			logger.Error("Error disconnecting user: %v", err)
		}

		_, err = s.ocservUserHandler.Lock(ocUser.Username)
		if err != nil {
			logger.Error("Error locking user: %v", err)
		}

		ocUser.DeactivatedAt = &now
	}
	err = db.Save(&ocUser).Error
	if err != nil {
		logger.Error("Error updating user stats: %v", err)
		return err
	}
	return nil
}

func (s *StatService) saveSessionLog(ctx context.Context, log *models.OcservUserSessionLog) error {
	db := database.GetConnection()
	db = db.WithContext(ctx)

	err := db.Save(log).Error
	if err != nil {
		logger.Error("Error saving session log: %v", err)
		return err
	}

	return nil
}

func (s *StatService) getCurrentMonthTotals(db *gorm.DB, userID uint, usageResetAt *time.Time) (Totals, error) {
	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	endOfMonth := startOfMonth.AddDate(0, 1, 0)

	startAt := startOfMonth
	if usageResetAt != nil && usageResetAt.After(startAt) {
		startAt = *usageResetAt
	}

	var result Totals
	err := db.Model(&models.OcservUserTrafficStatistics{}).
		Select("COALESCE(SUM(rx), 0) as total_rx, COALESCE(SUM(tx), 0) as total_tx").
		Where("oc_user_id = ? AND created_at >= ? AND created_at < ?", userID, startAt, endOfMonth).
		Scan(&result).Error

	return result, err
}
