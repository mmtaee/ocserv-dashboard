package routing

import (
	"context"
	"errors"
	"fmt"
	"github.com/mmtaee/ocserv-dashboard/core/pkg/config"
	"github.com/mmtaee/ocserv-dashboard/dashboard/api/internal/providers/routing"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"log/slog"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	"github.com/mmtaee/ocserv-dashboard/dashboard/api/pkg/middlewares"
	"golang.org/x/time/rate"
)

type MultiApp struct {
	Client string
	Admin  string
}

func Serve() {
	e := echo.New()

	// Logger using slog
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	if config.AppConfig.Debug {
		logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
	}
	slog.SetDefault(logger)

	// Middlewares
	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogStatus:   true,
		LogURI:      true,
		LogMethod:   true,
		LogRemoteIP: true,
		LogLatency:  true,
		LogValuesFunc: func(c *echo.Context, v middleware.RequestLoggerValues) error {
			if v.Error != nil {
				logger.Error("request error",
					"uri", v.URI,
					"status", v.Status,
					"error", v.Error,
				)
			} else {
				logger.Info("request",
					"uri", v.URI,
					"status", v.Status,
					"method", v.Method,
					"ip", v.RemoteIP,
					"latency", v.Latency,
				)
			}
			return nil
		},
	}))
	e.Use(middleware.Recover())

	var allowOrigins []string
	if config.AppConfig.Debug {
		allowOrigins = []string{"*"}
	} else {
		allowOrigins = config.AppConfig.AllowOrigins
	}
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: allowOrigins,
		AllowMethods: []string{http.MethodGet, http.MethodHead, http.MethodPut, http.MethodPatch, http.MethodPost, http.MethodDelete},
	}))

	e.Use(middleware.RemoveTrailingSlash())
	e.Use(middleware.RequestID())
	e.Use(middleware.Gzip())
	e.Use(middlewares.TimeoutMiddleware(10 * time.Second))

	// Rate Limiting (in-memory only)
	rateStore := middlewares.NewSystemStore(rate.Limit(1), 5) // 1 request per second, burst 5
	e.Use(middlewares.RateLimitMiddleware(rateStore))

	e.GET("/health", func(c *echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"status": "Healthy",
		})
	})

	routing.Register(e)

	if config.AppConfig.Debug {
		VerboseLog(e, fmt.Sprintf("%s:%d", config.AppConfig.Host, config.AppConfig.Port))
	}

	// Start server with graceful shutdown using StartConfig
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	sc := echo.StartConfig{
		Address:    fmt.Sprintf("%s:%d", config.AppConfig.Host, config.AppConfig.Port),
		HideBanner: !config.AppConfig.Debug,
		HidePort:   !config.AppConfig.Debug,
	}

	if err := sc.Start(ctx, e); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Error("failed to start server", "error", err)
		os.Exit(1)
	}

	logger.Info("server stopped gracefully")
}
