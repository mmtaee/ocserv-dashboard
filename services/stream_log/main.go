package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/mmtaee/ocserv-users-management/common/pkg/config"
	"github.com/mmtaee/ocserv-users-management/common/pkg/database"
	"github.com/mmtaee/ocserv-users-management/stream_log/internal/processor"
	"github.com/mmtaee/ocserv-users-management/stream_log/internal/stream"
	"github.com/mmtaee/ocserv-users-management/stream_log/internal/web"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

var (
	debug   bool
	host    string
	port    int
	systemd bool
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment")
	}

	flag.BoolVar(&debug, "d", false, "debug mode")
	flag.StringVar(&host, "h", "0.0.0.0", "Server Host")
	flag.IntVar(&port, "p", 8080, "Server Port")
	flag.BoolVar(&systemd, "systemd", false, "Systemd Mode")
	flag.Parse()

	if debug {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
	}

	service := "ocserv"
	stringStream := make(chan string, 100)
	broadcaster := make(chan string, 1)
	ctx, cancel := context.WithCancel(context.Background())

	config.Init(debug, host, port)
	cfg := config.Get()

	database.Connect()

	processor.Init()
	go processor.Processor(ctx, stringStream, broadcaster)
	go processor.CalculateUserStats(ctx, stringStream)
	go processor.UserExpiryCron(ctx)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		log.Printf("\nReceived signal: %s\n", sig)
		cancel()
	}()

	if systemd {
		log.Println("Running on host – using systemd logs")
		go func() {
			if err := stream.SystemdStreamLogs(ctx, service, stringStream); err != nil {
				log.Printf("Systemd log error: %v\n", err)
			}
		}()
	} else {
		log.Println("Running in Docker – using Docker logs")
		go func() {
			if err := stream.DockerStreamLogs(ctx, service, stringStream); err != nil {
				log.Println(err)
			}
		}()
	}

	sseServer := web.NewSSEServer()
	sseServer.StartBroadcast(broadcaster)

	go func() {
		server := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
		http.HandleFunc("/logs", sseServer.SSEHandler())
		log.Println("Starting server on ", server)
		if err := http.ListenAndServe(server, nil); err != nil {
			log.Fatalf("ListenAndServe failed: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("Service shutting down successfully")

}
