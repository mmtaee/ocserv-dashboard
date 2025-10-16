package bootstrap

import (
	"context"
	"fmt"
	"github.com/mmtaee/ocserv-users-management/api/pkg/routing"
	"github.com/mmtaee/ocserv-users-management/common/pkg/config"
	"github.com/mmtaee/ocserv-users-management/common/pkg/database"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func Serve(debug bool, host string, port int) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("panic recovered: %v", r)
		}
	}()

	config.Init(debug, host, port)
	cfg := config.Get()

	database.Connect()
	Migrate()

	defer database.CloseConnection()

	go routing.Serve(cfg)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-quit
		log.Println("Second signal received, forcing exit")
		os.Exit(1)
	}()

	sig := <-quit
	fmt.Println()
	log.Printf("signal %v received", sig)
	log.Println("initiating shutdown...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	routing.Shutdown(ctx)
	database.CloseConnection()

	log.Println("api service shutdown complete")
}
