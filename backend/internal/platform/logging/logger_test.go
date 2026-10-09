package logger

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"
)

func TestRequestLogIncludesStructuredFieldsAndLevel(t *testing.T) {
	previous := Log
	defer func() { Log = previous }()

	ctx, cancel := context.WithCancel(context.Background())
	var output bytes.Buffer
	Log = &Logger{logChan: make(chan LogMessage, 1), done: make(chan struct{}), output: &output}
	go Log.worker(ctx)

	Request("customer-api", "192.0.2.1", "GET /api/customers/summary", 401, 12*time.Millisecond)
	cancel()
	Wait()

	got := output.String()
	for _, want := range []string{"[WARNING]", "[service=customer-api]", "[client_ip=192.0.2.1]", "[status=401]", "[latency=12ms]", "GET /api/customers/summary", ColorYellow, ColorReset} {
		if !strings.Contains(got, want) {
			t.Fatalf("log output %q does not contain %q", got, want)
		}
	}
}

func TestServiceForFile(t *testing.T) {
	for file, want := range map[string]string{
		"/backend/internal/usecase/admin_api/users/usecase.go":     "admin-api",
		"/backend/internal/services/customer_api/controller.go":    "customer-api",
		"/backend/internal/services/telegram_bot/bot/manager.go":   "telegram-bot",
		"/backend/internal/usecase/worker/logprocessor/usecase.go": "worker",
		"/backend/internal/platform/database/postgres.go":          "backend",
	} {
		if got := serviceForFile(file); got != want {
			t.Errorf("serviceForFile(%q) = %q, want %q", file, got, want)
		}
	}
}
