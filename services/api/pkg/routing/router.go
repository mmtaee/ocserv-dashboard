package routing

import (
	"context"
	"errors"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	LabstackLog "github.com/labstack/gommon/log"
	"github.com/mmtaee/ocserv-users-management/api/internal/providers/routing"
	"github.com/mmtaee/ocserv-users-management/api/pkg/routing/middlewares"
	"github.com/mmtaee/ocserv-users-management/common/pkg/config"
	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/renderer"
	"github.com/olekukonko/tablewriter/tw"
	echoSwagger "github.com/swaggo/echo-swagger"
	"log"
	"net/http"
	"os"
	"slices"
	"sort"
	"strings"
	"time"
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

	e.Pre(middleware.RemoveTrailingSlash())
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
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
		e.Logger.SetLevel(LabstackLog.WARN)
		e.HideBanner = true
	}

	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"status": "Healthy",
		})
	})

	if cfg.Debug {
		e.GET("/swagger/*", echoSwagger.WrapHandler)
	}

	e.Use(middleware.GzipWithConfig(middleware.GzipConfig{
		Skipper: func(c echo.Context) bool {
			if strings.Contains(c.Request().URL.Path, "swagger") {
				return true
			}
			return false
		},
	}))

	if err := e.Start(server); err != nil && !errors.Is(err, http.ErrServerClosed) {
		e.Logger.Fatal("shutting down the server", err)
	}
}

func Shutdown(ctx context.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}
	log.Println("server shutdown complete")
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
