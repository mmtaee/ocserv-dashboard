package stats

import (
	"context"
	"fmt"
	"github.com/mmtaee/ocserv-dashboard/common/models"
	occtlDocker "github.com/mmtaee/ocserv-dashboard/common/occtl_docker"
	"github.com/mmtaee/ocserv-dashboard/common/ocserv/occtl"
	"github.com/mmtaee/ocserv-dashboard/common/ocserv/user"
	"github.com/mmtaee/ocserv-dashboard/common/pkg/database"
	"github.com/mmtaee/ocserv-dashboard/common/pkg/logger"
	"gorm.io/gorm"
	"os"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"
)

type StatService struct {
	ctx             context.Context
	stream          <-chan string
	ocservUserRepo  user.OcservUserInterface
	ocservOcctlRepo occtl.OcservOcctlInterface
	occtlDockerRepo occtlDocker.OcservOcctlUsersDocker
	dockerMode      bool
	sessionStats    map[string]UserStats
}

func NewStatService(ctx context.Context, stream chan string, dockerMode bool) *StatService {
	s := &StatService{
		ctx:          ctx,
		stream:       stream,
		dockerMode:   dockerMode,
		sessionStats: make(map[string]UserStats),
	}

	if dockerMode {
		s.occtlDockerRepo = occtlDocker.NewOcservOcctlDocker()
	} else {
		s.ocservUserRepo = user.NewOcservUser()
		s.ocservOcctlRepo = occtl.NewOcservOcctl()
	}

	return s
}

func (s *StatService) CalculateUserStats() {
	for {
		select {
		case <-s.ctx.Done():
			logger.Warn("stopping: context cancelled")
			return

		case line, ok := <-s.stream:
			if !ok {
				logger.Warn("stream closed, exiting ...")
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

			logger.Info("starting get user session from line: %s", cleanLine)

			sessionLog := s.getUserSessionLog(cleanLine)
			if sessionLog == nil {
				continue
			}

			if err := s.saveSessionLog(s.ctx, sessionLog); err != nil {
				logger.Error("Error saving session msg (%v): %v", sessionLog.Username, err)
				continue
			}
			//logger.Info("Processed user: %v successfully", sessionLog.Username)
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
		logger.Error("no worker found in line: %s", cleanLine)
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
		IP:       ip,
		Event:    event,
		Message:  msg,
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

	return &UserStats{
		Username: matchRxTx[1],
		IP:       normalizeSessionIP(matchRxTx[2]),
		RX:       rx,
		TX:       tx,
	}, nil
}

func (s *StatService) getDisconnectStat(cleanLine string) (*UserStats, error) {
	reTxRx := regexp.MustCompile(`main\[([^\]]+)\].*rx:\s*(\d+),\s*tx:\s*(\d+)`)
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

	return &UserStats{
		Username: matchRxTx[1],
		IP:       normalizeSessionIP(matchRxTx[2]),
		RX:       rx,
		TX:       tx,
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

func sessionStatsKey(stats *UserStats) string {
	return fmt.Sprintf("%s|%s", stats.Username, stats.IP)
}

func (s *StatService) saveRxTxDelta(ctx context.Context, stats *UserStats, final bool) error {
	key := sessionStatsKey(stats)
	lastStats, found := s.sessionStats[key]

	deltaRx := stats.RX
	deltaTx := stats.TX
	if found {
		deltaRx = stats.RX - lastStats.RX
		deltaTx = stats.TX - lastStats.TX
	}

	// A lower value means ocserv started a new session for the same user/IP.
	// In that case, the current stat is the first value of the new session.
	if deltaRx < 0 || deltaTx < 0 {
		deltaRx = stats.RX
		deltaTx = stats.TX
	}

	// Keep the final value too.
	// ocserv may emit a duplicate periodic stat after the disconnect line.
	// If we delete the memory on disconnect, that duplicate gets counted again.
	s.sessionStats[key] = *stats

	if deltaRx == 0 && deltaTx == 0 {
		return nil
	}

	return s.saveRxTx(ctx, &UserStats{
		Username: stats.Username,
		IP:       stats.IP,
		RX:       deltaRx,
		TX:       deltaTx,
	})
}

func (s *StatService) saveRxTx(ctx context.Context, u *UserStats) error {
	logger.Info("saveRxTx called for user=%s RX=%d TX=%d", u.Username, u.RX, u.TX)

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

	var trafficSizeBytes = ocUser.TrafficSize * (1 << 30)

	totalMonthStats, err := s.getCurrentMonthTotals(db, ocUser.ID)
	if err != nil {
		logger.Error("Error getting current month stats: %v", err)
		return err
	}

	shouldLock := false
	switch ocUser.TrafficType {
	case models.TotallyTransmit:
		shouldLock = ocUser.Tx >= trafficSizeBytes

	case models.TotallyReceive:
		shouldLock = ocUser.Rx >= trafficSizeBytes

	case models.TotallyRxTx:
		shouldLock = ocUser.Rx+ocUser.Tx >= trafficSizeBytes

	case models.MonthlyTransmit:
		shouldLock = totalMonthStats.TotalTx >= trafficSizeBytes

	case models.MonthlyReceive:
		shouldLock = totalMonthStats.TotalRx >= trafficSizeBytes

	case models.MonthlyRxTx:
		shouldLock = totalMonthStats.TotalRx+totalMonthStats.TotalTx >= trafficSizeBytes

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
		var (
			disconnectFunc func(username string) (string, error)
			lockFunc       func(username string) (string, error)
		)
		if s.dockerMode {
			disconnectFunc = s.occtlDockerRepo.DisconnectUser
			lockFunc = s.occtlDockerRepo.Lock
		} else {
			disconnectFunc = s.ocservOcctlRepo.DisconnectUser
			lockFunc = s.ocservUserRepo.Lock
		}

		_, err = disconnectFunc(ocUser.Username)
		if err != nil {
			logger.Error("Error disconnecting user: %v", err)
		}

		_, err = lockFunc(ocUser.Username)
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
		logger.Error("Error updating user stats: %v", err)
		return err
	}

	return nil
}

func (s *StatService) getCurrentMonthTotals(db *gorm.DB, userID uint) (Totals, error) {
	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	endOfMonth := startOfMonth.AddDate(0, 1, 0)

	var result Totals
	err := db.Model(&models.OcservUserTrafficStatistics{}).
		Select("SUM(rx) as total_rx, SUM(tx) as total_tx").
		Where("oc_user_id = ? AND created_at >= ? AND created_at < ?", userID, startOfMonth, endOfMonth).
		Scan(&result).Error

	return result, err
}
