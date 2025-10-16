package processor

import (
	"context"
	"github.com/mmtaee/ocserv-users-management/common/models"
	"github.com/mmtaee/ocserv-users-management/common/pkg/database"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"
)

type UserStats struct {
	Username string
	RX       int
	TX       int
}

type Totals struct {
	TotalRx int
	TotalTx int
}

func CalculateUserStats(ctx context.Context, stream <-chan string) {
	for s := range stream {
		u, err := extractUser(s)
		if err != nil {

			continue
		}

		log.Println("user found:", u)

		if err = save(ctx, u); err != nil {
			log.Printf("failed to save user %v: %v", u, err)
			continue
		}

		log.Printf("processed user: %v successfully", u)

		select {
		case <-time.After(500 * time.Millisecond):
		case <-ctx.Done():
			log.Println("stopping: context cancelled")
			return
		}
	}
}

func save(ctx context.Context, u UserStats) error {
	db := database.GetConnection()
	db = db.WithContext(ctx)

	var user models.OcservUser

	err := db.Where("username = ? ", u.Username).First(&user).Error
	if err != nil {
		log.Println(err)
		return err
	}

	traffic := models.OcservUserTrafficStatistics{
		OcUserID: user.ID,
		Rx:       u.RX,
		Tx:       u.TX,
	}

	err = db.Create(&traffic).Error
	if err != nil {
		log.Println(err)
		return err
	}

	user.Rx += u.RX
	user.Tx += u.TX

	var trafficSizeBytes = user.TrafficSize * (1 << 30)

	totalMonthStats, err := getCurrentMonthTotals(db, user.ID)
	if err != nil {
		log.Println(err)
		return err
	}

	switch user.TrafficType {
	case models.TotallyTransmit:
		user.IsLocked = user.Tx >= trafficSizeBytes

	case models.TotallyReceive:
		user.IsLocked = user.Rx >= trafficSizeBytes

	case models.MonthlyTransmit:
		user.IsLocked = totalMonthStats.TotalTx >= trafficSizeBytes

	case models.MonthlyReceive:
		user.IsLocked = totalMonthStats.TotalRx >= trafficSizeBytes

	default:
		log.Printf("free traffic type")
	}

	now := time.Now()
	if user.IsLocked {
		_, err = ocservUserRepo.Lock(user.Username)
		if err != nil {
			log.Println(err)
		}
		user.DeactivatedAt = &now
	}
	err = db.Save(&user).Error
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func getCurrentMonthTotals(db *gorm.DB, userID uint) (Totals, error) {
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

func extractUser(text string) (UserStats, error) {
	var (
		username string
		stats    UserStats
	)

	if strings.Contains(text, "server shutdown complete") {
		log.Println("Ocserv server shutdown abnormally")
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
