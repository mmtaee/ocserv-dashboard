package routing

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	LabstackLog "github.com/labstack/gommon/log"
	"github.com/mmtaee/ocserv-dashboard/api/internal/providers/routing"
	"github.com/mmtaee/ocserv-dashboard/api/pkg/routing/middlewares"
	"github.com/mmtaee/ocserv-dashboard/core/pkg/config"
	"github.com/mmtaee/ocserv-dashboard/core/pkg/logger"
	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/renderer"
	"github.com/olekukonko/tablewriter/tw"
)

var (
	e            *echo.Echo
	allowMethods = []string{
		http.MethodGet,
		http.MethodPost,
		http.MethodDelete,
		http.MethodPatch,
		http.MethodPut,
		http.MethodHead,
		http.MethodOptions,
	}
)

func Serve(cfg *config.Config) {
	server := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)

	e = echo.New()

	//e.Logger = NewLoggerWrapper(logger.GetLogger())

	e.Pre(middleware.RemoveTrailingSlash())
	e.Use(middlewares.RequestLoggerMiddleware())
	e.Use(middleware.RecoverWithConfig(middleware.RecoverConfig{
		LogErrorFunc: func(c echo.Context, err error, stack []byte) error {
			logger.Error("panic recovered: %v\n%s", err, stack)
			return nil
		},
	}))
	e.Use(middlewares.TimeoutMiddleware(10 * time.Second))

	if cfg.Debug {
		e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
			AllowOrigins: []string{"*"},                                                                             // allow all origins
			AllowHeaders: []string{"*"},                                                                             // allow all headers
			AllowMethods: []string{echo.GET, echo.HEAD, echo.PUT, echo.PATCH, echo.POST, echo.DELETE, echo.OPTIONS}, // allow all methods
		}))
	} else {
		e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
			AllowOrigins: cfg.AllowOrigins,
			AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
			AllowMethods: allowMethods,
		}))
	}

	routing.Register(e)

	if cfg.Debug {
		e.Debug = true
		e.Logger.SetLevel(LabstackLog.DEBUG)
		verboseLog(server)
	} else {
		e.Logger.SetLevel(LabstackLog.INFO)
		e.HideBanner = true
		e.HidePort = true
	}

	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"status": "Healthy",
		})
	})

	e.Use(middleware.GzipWithConfig(middleware.GzipConfig{
		Skipper: func(c echo.Context) bool {
			path := c.Path()

			switch {
			case strings.HasPrefix(path, "/api/v1/ocserv/users/backup"):
				return true
			case strings.HasPrefix(path, "/api/v1/ocserv/groups/backup"):
				return true
			}

			return false
		},
	}))

	err := e.Start(server)
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("shutting down the server: %v\n", err) // use fmt, not Fatal
		os.Exit(1)
	}

	logger.Info("Starting server at %s", server)
}

func Shutdown(ctx context.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}
	logger.Info("server shutdown complete")
}

func verboseLog(service string) {
	paths := e.Routes()
	sort.SliceStable(paths, func(i, j int) bool {
		return paths[i].Path < paths[j].Path
	})

	table := tablewriter.NewTable(
		os.Stdout,
		tablewriter.WithRenderer(
			renderer.NewBlueprint(
				tw.Rendition{
					Settings: tw.Settings{Separators: tw.Separators{BetweenRows: tw.On}},
				},
			),
		),
	)
	table.Header([]string{"Method", "Url", "Handler"})

	for _, route := range e.Routes() {
		if slices.Contains(allowMethods, route.Method) {
			err := table.Append(
				[]string{
					route.Method,
					fmt.Sprintf("http://%s%s/", service, route.Path),
					strings.Split(strings.TrimSuffix(route.Name, "-fm"), ".")[2],
				},
			)
			if err != nil {
				return
			}
		}
	}

	err := table.Render()
	if err != nil {
		return
	}
}
