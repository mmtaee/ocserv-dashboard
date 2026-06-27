package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	occtlDocker "github.com/mmtaee/ocserv-dashboard/core/occtl_docker"
	"github.com/mmtaee/ocserv-dashboard/core/ocserv/occtl"
	"github.com/mmtaee/ocserv-dashboard/core/ocserv/user"
	"github.com/mmtaee/ocserv-dashboard/core/pkg/logger"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

type Webhook struct {
	occtlHandler      occtl.OcservOcctlInterface
	ocservUserHandler user.OcservUserInterface
}

var (
	occtlHandler      occtl.OcservOcctlInterface
	ocservUserHandler user.OcservUserInterface
)

func init() {
	occtlHandler = occtl.NewOcservOcctl()
	ocservUserHandler = user.NewOcservUser()
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/webhook/", webhookHandler)

	server := &http.Server{
		Addr:    "0.0.0.0:8888",
		Handler: mux,
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		logger.Info("Webhook server listening on: %s ", server.Addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatal("Failed to start webhook server: %v", err)
		}
	}()

	<-stop
	logger.Warn("Shutting down webhook server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Fatal("Failed to shutdown webhook server: %v", err)
	}

	logger.Info("Webhook server shutdown successfully")
}

// webhookHandler extracts action from path and handles requests
func webhookHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	payload := occtlDocker.WebhookPayload{}
	defer r.Body.Close()
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&payload); err != nil {
		http.Error(w, "Invalid payload: "+err.Error(), http.StatusBadRequest)
		return
	}

	if payload.Username == "" {
		http.Error(w, "Username is required", http.StatusBadRequest)
		return
	}

	// Extract action from path: /webhook/<action>
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 2 {
		http.Error(w, "Action not specified in URL path", http.StatusBadRequest)
		return
	}
	action := strings.ToLower(parts[1])

	logger.Info("Received webhook action: %s for username %s", action, payload.Username)

	switch action {
	case "disconnect":
		msg, err := occtlHandler.DisconnectUser(payload.Username)
		if err != nil {
			http.Error(w, "Failed to disconnect user: "+err.Error(), http.StatusBadRequest)
			return
		}
		_, _ = fmt.Fprintf(w, "User %s disconnected successfully. message: %s", payload.Username, msg)

	case "lock":
		msg, err := ocservUserHandler.Lock(payload.Username)
		if err != nil {
			http.Error(w, "Failed to lock user: "+err.Error(), http.StatusBadRequest)
			return
		}
		_, _ = fmt.Fprintf(w, "User %s locked successfully. message: %s", payload.Username, msg)

	case "unlock":
		msg, err := ocservUserHandler.UnLock(payload.Username)
		if err != nil {
			http.Error(w, "Failed to unlock user: "+err.Error(), http.StatusBadRequest)
			return
		}
		_, _ = fmt.Fprintf(w, "User %s unlocked successfully. message: %s", payload.Username, msg)

	default:
		http.Error(w, "Unknown action: "+action, http.StatusBadRequest)
	}
}
