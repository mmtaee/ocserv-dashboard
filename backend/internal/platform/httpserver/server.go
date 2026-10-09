package httpserver

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	"github.com/mmtaee/ocserv-dashboard/backend/config"
	_ "github.com/mmtaee/ocserv-dashboard/backend/docs"
	appmiddleware "github.com/mmtaee/ocserv-dashboard/backend/pkg/middlewares"
	echoSwagger "github.com/swaggo/echo-swagger/v2"
)

type RouteRegistrar interface {
	Register(group *echo.Group)
}

type namedRegistrar interface {
	ServiceName() string
}

type Server struct {
	http            *http.Server
	shutdownTimeout time.Duration
}

func New(cfg *config.Config, registrars ...RouteRegistrar) *Server {
	e := echo.New()
	e.Pre(middleware.RemoveTrailingSlash())
	e.Use(middleware.Recover())
	e.Use(appmiddleware.TimeoutMiddleware(10 * time.Second))

	origins := cfg.AllowOrigins
	if cfg.Debug {
		origins = []string{"*"}
	}
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: origins,
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
		AllowMethods: []string{http.MethodGet, http.MethodHead, http.MethodPatch, http.MethodPost, http.MethodDelete, http.MethodOptions},
	}))
	e.Use(middleware.GzipWithConfig(middleware.GzipConfig{Skipper: func(c *echo.Context) bool {
		path := c.Path()
		return strings.HasPrefix(path, "/swagger") ||
			strings.HasPrefix(path, "/api/v1/ocserv/users/backup") ||
			strings.HasPrefix(path, "/api/v1/ocserv/groups/backup")
	}}))
	e.GET("/swagger/*", echoSwagger.WrapHandler, appmiddleware.RequestLoggerMiddleware("http-server"))

	api := e.Group("/api")
	for _, registrar := range registrars {
		serviceGroup := api.Group("")
		serviceName := "backend-api"
		if named, ok := registrar.(namedRegistrar); ok {
			serviceName = named.ServiceName()
		}
		serviceGroup.Use(appmiddleware.RequestLoggerMiddleware(serviceName))
		registrar.Register(serviceGroup)
	}
	e.GET("/health", func(c *echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "Healthy"})
	}, appmiddleware.RequestLoggerMiddleware("http-server"))

	return &Server{
		http:            &http.Server{Addr: fmt.Sprintf("%s:%d", cfg.Host, cfg.Port), Handler: e, ReadHeaderTimeout: 10 * time.Second},
		shutdownTimeout: 10 * time.Second,
	}
}

func (s *Server) Run(ctx context.Context) error {
	errCh := make(chan error, 1)
	go func() {
		errCh <- s.http.ListenAndServe()
	}()

	select {
	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return fmt.Errorf("http server: %w", err)
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), s.shutdownTimeout)
		defer cancel()
		if err := s.http.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("shutdown http server: %w", err)
		}
		err := <-errCh
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("http server: %w", err)
		}
		return nil
	}
}
