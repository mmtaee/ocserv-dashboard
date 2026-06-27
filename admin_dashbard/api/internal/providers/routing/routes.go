package routing

import (
	"os"

	"github.com/labstack/echo/v4"
	"github.com/mmtaee/ocserv-dashboard/api/internal/repository"
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
	"github.com/mmtaee/ocserv-dashboard/api/internal/usecase"
	"github.com/mmtaee/ocserv-dashboard/core/pkg/config"
)

func Register(e *echo.Echo) {
	group := e.Group("/api")

	systemRoutes.Routes(group)
	ocservGroupRoutes.Routes(group)
	ocservUserRoutes.Routes(group)

	// home
	occtlRepo := repository.NewOcctlRepository()
	ocservUserRepo := repository.NewtOcservUserRepository()
	reportRepo := repository.NewtReportRepository()
	telegramRepo := repository.NewTelegramRepository()
	homeUC := usecase.NewHomeUsecase(occtlRepo, ocservUserRepo, reportRepo, telegramRepo)
	homeCtl := homeRoutes.New(homeUC)
	homeRoutes.Routes(group, homeCtl)

	// backup
	ocservGroupRepo := repository.NewOcservGroupRepository()
	backupRepo := repository.NewBackupRepository()
	backupUC := usecase.NewBackupUsecase(ocservUserRepo, ocservGroupRepo, backupRepo)
	backupCtl := backupRoutes.New(backupUC)
	backupRoutes.Routes(group, backupCtl)

	// occtl
	occtlUC := usecase.NewOcctlUsecase(occtlRepo)
	occtlCtl := occtlRoutes.New(occtlUC)
	occtlRoutes.Routes(group, occtlCtl)

	// customers
	customerRoutes.Routes(group)

	// reports
	reportRoutes.Routes(group)

	// systemd
	systemdRoutes.Routes(group)

	// telegram
	if os.Getenv("TELEGRAM_BOT_ENABLED") == "true" || config.Get().Debug {
		telegramRoutes.Routes(group)
	}
}
