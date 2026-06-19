package routing

import (
	"github.com/labstack/echo/v5"
	"github.com/mmtaee/ocserv-dashboard/core/pkg/config"
	"github.com/mmtaee/ocserv-dashboard/dashboard/api/internal/service/auth"
	"github.com/mmtaee/ocserv-dashboard/dashboard/api/internal/service/backup"
	"github.com/mmtaee/ocserv-dashboard/dashboard/api/internal/service/occtl"
	"github.com/mmtaee/ocserv-dashboard/dashboard/api/internal/service/ocserv_group"
	"github.com/mmtaee/ocserv-dashboard/dashboard/api/internal/service/ocserv_user"
	"github.com/mmtaee/ocserv-dashboard/dashboard/api/internal/service/report"
	"github.com/mmtaee/ocserv-dashboard/dashboard/api/internal/service/super_admin"
	"github.com/mmtaee/ocserv-dashboard/dashboard/api/internal/service/system"
	"github.com/mmtaee/ocserv-dashboard/dashboard/api/internal/service/systemd"
	"github.com/mmtaee/ocserv-dashboard/dashboard/api/internal/service/telegram"
)

func Register(e *echo.Echo, cfg config.Config) {
	auth.RegisterRoutes(e)
	system.RegisterRoutes(e)
	super_admin.RegisterRoutes(e)
	ocserv_group.RegisterRoutes(e)
	ocserv_user.RegisterRoutes(e)
	report.RegisterRoutes(e)
	occtl.RegisterRoutes(e)
	telegram.RegisterRoutes(e, cfg.Telegram)
	backup.RegisterRoutes(e)
	systemd.RegisterRoutes(e)
}
