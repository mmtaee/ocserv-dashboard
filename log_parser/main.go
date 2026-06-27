package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/mmtaee/ocserv-dashboard/core/pkg/config"
	"github.com/mmtaee/ocserv-dashboard/core/pkg/database"
	"github.com/mmtaee/ocserv-dashboard/core/pkg/logger"
	"github.com/mmtaee/ocserv-dashboard/log_stream/internal/readers"
	"github.com/mmtaee/ocserv-dashboard/log_stream/internal/sse"
	"github.com/mmtaee/ocserv-dashboard/log_stream/internal/stats"
)

var (
	debug      bool
	host       string
	port       int
	dockerMode bool
)

func main() {
	flag.BoolVar(&debug, "d", false, "debug mode")
	flag.StringVar(&host, "h", "0.0.0.0", "Server Host")
	flag.IntVar(&port, "p", 8080, "Server Port")
	flag.BoolVar(&dockerMode, "docker-mode", false, "Docker Mode")
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	service := "ocserv"

	logger.Init(ctx, 100)

	if err := godotenv.Load(); err != nil {
		logger.Warn("Error loading .env file, using system environment")
	}

	config.Init(debug, host, port)
	cfg := config.Get()

	database.Connect()

	streamChan := make(chan string, 1000)
	lineLogChan := make(chan string, 1000)
	broadcastChan := make(chan string, 1000)

	if !dockerMode {
		logger.Info("Systemd Mode")
		go func() {
			if err := readers.SystemdStreamLogs(ctx, service, streamChan); err != nil {
				logger.Error("Systemd Stream Logs Error: %v", err)
			}
		}()
	} else {
		logger.Info("Docker Mode")
		go func() {
			if err := readers.DockerStreamLogs(ctx, service, streamChan); err != nil {
				logger.Error("Docker Stream Logs Error: %v", err)
			}
		}()
	}

	statService := stats.NewStatService(ctx, lineLogChan, dockerMode)
	go func() {
		statService.CalculateUserStats()
	}()

	sseServer := sse.NewSSEServer()
	sseServer.StartBroadcast(broadcastChan)

	go func() {
		server := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
		http.HandleFunc("/logs", sseServer.SSEHandler())

		logger.Info("Starting server on %s", server)
		if err := http.ListenAndServe(server, nil); err != nil {
			logger.Error("Error starting server: %v", err)
		}
	}()

	go func() {
		start(ctx, streamChan, broadcastChan, lineLogChan)
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		logger.Warn("Received shutdown signal %s", sig)
		cancel()
	}()

	<-ctx.Done()
	logger.Info("Log stream service shutting down successfully")
}

func start(ctx context.Context, streamText <-chan string, broadcaster, lineLogChan chan<- string) {
	for {
		select {
		case <-ctx.Done():
			return

		case line, ok := <-streamText:
			if !ok {
				return
			}

			// UI live logs are useful, but accounting must not be delayed by them.
			select {
			case broadcaster <- line:
			default:
			}

			// Accounting must preserve log order. Do not send this through a goroutine.
			select {
			case lineLogChan <- line:
			case <-ctx.Done():
				return
			}
		}
	}
}
