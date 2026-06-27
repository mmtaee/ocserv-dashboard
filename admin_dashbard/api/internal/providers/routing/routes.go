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
	"github.com/mmtaee/ocserv-dashboard/api/pkg/captcha"
	"github.com/mmtaee/ocserv-dashboard/api/pkg/crypto"
)

func Register(e *echo.Echo) {
	group := e.Group("/api")

	// Repositories
	occtlRepo := repository.NewOcctlRepository()
	ocservUserRepo := repository.NewtOcservUserRepository()
	ocservGroupRepo := repository.NewOcservGroupRepository()
	reportRepo := repository.NewtReportRepository()
	telegramRepo := repository.NewTelegramRepository()
	systemRepo := repository.NewSystemRepository()
	userRepo := repository.NewUserRepository()
	systemdRepo := repository.NewSystemdRepository("ocserv")

	// System
	systemUC := usecase.NewSystemUsecase(
		systemRepo,
		userRepo,
		captcha.NewGoogleVerifier(),
		crypto.NewCustomPassword(),
	)
	systemCtl := systemRoutes.New(systemUC)
	systemRoutes.Routes(group, systemCtl)

	// home
	homeUC := usecase.NewHomeUsecase(occtlRepo, ocservUserRepo, reportRepo, telegramRepo)
	homeCtl := homeRoutes.New(homeUC)
	homeRoutes.Routes(group, homeCtl)

	// backup
	backupRepo := repository.NewBackupRepository()
	backupUC := usecase.NewBackupUsecase(ocservUserRepo, ocservGroupRepo, backupRepo)
	backupCtl := backupRoutes.New(backupUC)
	backupRoutes.Routes(group, backupCtl)

	// occtl
	occtlUC := usecase.NewOcctlUsecase(occtlRepo)
	occtlCtl := occtlRoutes.New(occtlUC)
	occtlRoutes.Routes(group, occtlCtl)

	// ocserv group
	ocservGroupUC := usecase.NewOcservGroupUsecase(ocservGroupRepo, ocservUserRepo)
	ocservGroupCtl := ocservGroupRoutes.New(ocservGroupUC)
	ocservGroupRoutes.Routes(group, ocservGroupCtl)

	// ocserv user
	ocservUserUC := usecase.NewOcservUserUsecase(ocservUserRepo, occtlRepo, reportRepo)
	ocservUserCtl := ocservUserRoutes.New(ocservUserUC)
	ocservUserRoutes.Routes(group, ocservUserCtl)

	// customers
	customerRoutes.Routes(group)

	// reports
	reportUC := usecase.NewReportUsecase(reportRepo, occtlRepo)
	reportCtl := reportRoutes.New(reportUC)
	reportRoutes.Routes(group, reportCtl)

	// systemd
	systemdUC := usecase.NewSystemdUsecase(systemdRepo)
	systemdCtl := systemdRoutes.New(systemdUC)
	systemdRoutes.Routes(group, systemdCtl)

	// telegram
	if os.Getenv("TELEGRAM_BOT_ENABLED") == "true" || config.Get().Debug {
		telegramUC := usecase.NewTelegramUsecase(telegramRepo, ocservUserRepo)
		telegramCtl := telegramRoutes.New(telegramUC)
		telegramRoutes.Routes(group, telegramCtl)
	}
}
