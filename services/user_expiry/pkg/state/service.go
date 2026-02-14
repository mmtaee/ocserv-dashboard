package state

import (
	"fmt"
	"github.com/mmtaee/ocserv-users-management/common/pkg/logger"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const stateFile = "cron_journal/cron_state.txt"

var stateMu sync.Mutex

type CronState struct {
	DailyLastRun              time.Time
	MonthlyLastRun            time.Time
	DeleteInactiveUserLastRun time.Time
}

func NewCronState() *CronState {
	return LoadStateOrDefault()
}

func ensureStateFile() error {
	absPath, err := filepath.Abs(stateFile)
	if err != nil {
		logger.Error("Failed to get absolute path for state file %s: %v", stateFile, err)
	} else {
		logger.Info("Start finding state file in path %s", absPath)
	}

	// Ensure directory exists
	dir := filepath.Dir(absPath)
	if err = os.MkdirAll(dir, 0755); err != nil {
		logger.Error("Failed to create state directory %s: %v", dir, err)
		return err
	}

	// Create the file if it does not exist
	if _, err = os.Stat(absPath); os.IsNotExist(err) {
		defaultContent := "daily_last_run=0\nmonthly_last_run=0\n"
		if err = os.WriteFile(absPath, []byte(defaultContent), 0644); err != nil {
			logger.Error("Failed to write state file %s: %v", absPath, err)
			return err
		}
		logger.Info("Created new state file at %s", absPath)
	}

	return nil
}

func LoadStateOrDefault() *CronState {
	stateMu.Lock()
	defer stateMu.Unlock()

	if err := ensureStateFile(); err != nil {
		fmt.Println("Failed to create state file:", err)
		return &CronState{}
	}
	logger.Info("Loading cron jobs state file")

	data, err := os.ReadFile(stateFile)
	if err != nil {
		fmt.Println("Failed to read state:", err)
		return &CronState{}
	}

	state := &CronState{}
	lines := strings.Split(string(data), "\n")

	parse := func(s string) time.Time {
		defaultTime := time.Now().UTC().AddDate(0, -2, 0).Truncate(24 * time.Hour)
		if s == "0" || s == "" {
			// Fallback: 2 days ago for daily, 2 months ago for monthly
			return defaultTime
		}
		t, err := time.Parse("2006-01-02", s)
		if err != nil {
			logger.Error("Failed to parse date: %v", err)
			return time.Time{}
		}
		return t.UTC()
	}

	for _, l := range lines {
		if strings.HasPrefix(l, "daily_last_run=") {
			state.DailyLastRun = parse(strings.TrimPrefix(l, "daily_last_run="))
		}
		if strings.HasPrefix(l, "monthly_last_run=") {
			state.MonthlyLastRun = parse(strings.TrimPrefix(l, "monthly_last_run="))
		}
		if strings.HasPrefix(l, "delete_inactive_user_last_run=") {
			state.DeleteInactiveUserLastRun = parse(strings.TrimPrefix(l, "delete_inactive_user_last_run="))
		}
	}
	return state
}

func (s *CronState) Save() error {
	stateMu.Lock()
	defer stateMu.Unlock()

	content := fmt.Sprintf(
		"daily_last_run=%s\nmonthly_last_run=%s\ndelete_inactive_user_last_run=%s\n",
		formatTime(s.DailyLastRun),
		formatTime(s.MonthlyLastRun),
		formatTime(s.DeleteInactiveUserLastRun),
	)
	return os.WriteFile(stateFile, []byte(content), 0644)
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return "0"
	}
	return t.Format("2006-01-02")
}
