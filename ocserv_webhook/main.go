package main

import (
	"github.com/mmtaee/ocserv-dashboard/ocserv_webhook/cmd"
	"github.com/mmtaee/ocserv-dashboard/core/pkg/logger"
)

func main() {
	if err := cmd.Execute(); err != nil {
		logger.Fatal("Error: %v", err)
	}
}
