package routing

import (
	"github.com/labstack/echo/v4"
	"github.com/mmtaee/ocserv-dashboard/common/pkg/logger"
	backupRoutes "github.com/mmtaee/ocserv-dashboard/api/internal/services/backup"
	customerRoutes "github.com/mmtaee/ocserv-dashboard/api/internal/services/customer"
	homeRoutes "github.com/mmtaee/ocserv-dashboard/api/internal/services/home"
	occtlRoutes "github.com/mmtaee/ocserv-dashboard/api/internal/services/occtl"
	ocservGroupRoutes "github.com/mmtaee/ocserv-dashboard/api/internal/services/ocserv_group"
	ocservUserRoutes "github.com/mmtaee/ocserv-dashboard/api/internal/services/ocserv_user"
	reportRoutes "github.com/mmtaee/ocserv-dashboard/api/internal/services/report"
	systemRoutes "github.com/mmtaee/ocserv-dashboard/api/internal/services/system"
	systemdRoutes "github.com/mmtaee/ocserv-dashboard/api/internal/services/systemd"
	telegramRoutes "github.com/mmtaee/ocserv-dashboard/api/internal/services/telegram"
)

func Register(e *echo.Echo) {
	group := e.Group("/api")

	systemRoutes.Routes(group)
	ocservGroupRoutes.Routes(group)
	ocservUserRoutes.Routes(group)
	occtlRoutes.Routes(group)
	homeRoutes.Routes(group)

	// backup
	backupRoutes.Routes(group)

	// customers
	customerRoutes.Routes(group)

	// reports
	reportRoutes.Routes(group)

	// systemd
	systemdRoutes.Routes(group)

	// telegram
	telegramRoutes.InitI18n()
	if err := telegramRoutes.EnsureReceiptDir(); err != nil {
		logger.Warn("telegram receipts directory: %v", err)
	}
	telegramRoutes.Routes(group)
}
