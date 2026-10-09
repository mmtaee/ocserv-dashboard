package units

import (
	"context"
	"testing"
	"time"

	"github.com/mmtaee/ocserv-dashboard/backend/config"
	"github.com/mmtaee/ocserv-dashboard/backend/internal/platform/httpserver"
	"github.com/stretchr/testify/require"
)

func TestHTTPServerStopsOnContextCancellation(t *testing.T) {
	server := httpserver.New(&config.Config{Host: "127.0.0.1", Port: 0, AllowOrigins: []string{"*"}})
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- server.Run(ctx) }()
	cancel()

	select {
	case err := <-done:
		require.NoError(t, err)
	case <-time.After(2 * time.Second):
		t.Fatal("HTTP server did not stop after cancellation")
	}
}
