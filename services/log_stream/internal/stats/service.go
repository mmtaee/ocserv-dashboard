package stats

import (
	"context"
	"errors"
	"github.com/mmtaee/ocserv-users-management/common/models"
	"github.com/mmtaee/ocserv-users-management/common/ocserv/occtl"
	"github.com/mmtaee/ocserv-users-management/common/ocserv/user"
	"github.com/mmtaee/ocserv-users-management/common/pkg/database"
	"github.com/mmtaee/ocserv-users-management/common/pkg/logger"
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
}

func NewStatService(ctx context.Context, stream chan string) *StatService {
	return &StatService{
		ctx:             ctx,
		stream:          stream,
		ocservUserRepo:  user.NewOcservUser(),
		ocservOcctlRepo: occtl.NewOcservOcctl(),
	}
}

func (s *StatService) CalculateUserStats() {
	for {
		select {
		case <-s.ctx.Done():
			logger.Warn("stopping: context cancelled")
			return

		case msg, ok := <-s.stream:
			if !ok {
				logger.Warn("stream closed, exiting ...")
				return
			}

			cleanMsg := strings.TrimSpace(msg) // remove whitespace/newlines and normalize case

			if strings.Contains(cleanMsg, "user disconnected") {
				u, err := s.extractUser(cleanMsg)
				if err != nil {
					logger.Error("Error extracting user msg (%q): %v", cleanMsg, err)
					continue
				}

				if err = s.save(s.ctx, u); err != nil {
					logger.Error("Error saving user msg (%v): %v", u, err)
					continue
				}

				logger.Info("Processed user: %v successfully", u)
			}
		}
	}
}

func (s *StatService) save(ctx context.Context, u UserStats) error {
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

	switch ocUser.TrafficType {
	case models.TotallyTransmit:
		ocUser.IsLocked = ocUser.Tx >= trafficSizeBytes

	case models.TotallyReceive:
		ocUser.IsLocked = ocUser.Rx >= trafficSizeBytes

	case models.MonthlyTransmit:
		ocUser.IsLocked = totalMonthStats.TotalTx >= trafficSizeBytes

	case models.MonthlyReceive:
		ocUser.IsLocked = totalMonthStats.TotalRx >= trafficSizeBytes

	default:
		logger.Error("Unknown traffic type: %v", ocUser.TrafficType)
	}

	now := time.Now()
	if ocUser.IsLocked {
		_, err = s.ocservUserRepo.Lock(ocUser.Username)
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

func (s *StatService) extractUser(text string) (UserStats, error) {
	var (
		username string
		stats    UserStats
	)

	if strings.Contains(text, "server shutdown complete") {
		logger.Error("Ocserv server shutdown abnormally")
		p, _ := os.FindProcess(os.Getpid())
		_ = p.Signal(syscall.SIGTERM)
		return stats, errors.New("shutdown signal sent")
	}

	re := regexp.MustCompile(`main\[(.*?)\].*rx:\s*(\d+),\s*tx:\s*(\d+)`)
	match := re.FindStringSubmatch(text)
	if len(match) > 0 {
		username = match[1]
		stats.RX, _ = strconv.Atoi(match[2])
		stats.TX, _ = strconv.Atoi(match[3])
		stats.Username = username
		return stats, nil
	}
	return stats, errors.New("no user found")

}
