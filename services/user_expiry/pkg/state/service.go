package state

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

const stateFile = "cron_journal/cron_state.txt"

var stateMu sync.Mutex

type CronState struct {
	DailyLastRun   time.Time
	MonthlyLastRun time.Time
}

func NewCronState() *CronState {
	return LoadStateOrDefault()
}

func ensureStateFile() error {
	if _, err := os.Stat(stateFile); os.IsNotExist(err) {
		// Create file with default values
		defaultContent := "daily_last_run=0\nmonthly_last_run=0\n"
		return os.WriteFile(stateFile, []byte(defaultContent), 0644)
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

	data, err := os.ReadFile(stateFile)
	if err != nil {
		fmt.Println("Failed to read state:", err)
		return &CronState{}
	}

	state := &CronState{}
	lines := strings.Split(string(data), "\n")

	parse := func(s string) time.Time {
		if s == "0" || s == "" {
			return time.Time{}
		}
		t, err := time.Parse("2006-01-02", s)
		if err != nil {
			return time.Time{}
		}
		return t
	}

	for _, l := range lines {
		if strings.HasPrefix(l, "daily_last_run=") {
			state.DailyLastRun = parse(strings.TrimPrefix(l, "daily_last_run="))
		}
		if strings.HasPrefix(l, "monthly_last_run=") {
			state.MonthlyLastRun = parse(strings.TrimPrefix(l, "monthly_last_run="))
		}
	}

	return state
}

func (s *CronState) Save() error {
	stateMu.Lock()
	defer stateMu.Unlock()

	content := fmt.Sprintf(
		"daily_last_run=%s\nmonthly_last_run=%s\n",
		formatTime(s.DailyLastRun),
		formatTime(s.MonthlyLastRun),
	)
	return os.WriteFile(stateFile, []byte(content), 0644)
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return "0"
	}
	return t.Format("2006-01-02")
}
