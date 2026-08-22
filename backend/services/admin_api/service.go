package adminapi

import (
	"github.com/labstack/echo/v5"
	"github.com/mmtaee/ocserv-dashboard/backend/internal/repository"
	backupusecase "github.com/mmtaee/ocserv-dashboard/backend/internal/usecase/backup"
	occtlusecase "github.com/mmtaee/ocserv-dashboard/backend/internal/usecase/occtl"
	groupusecase "github.com/mmtaee/ocserv-dashboard/backend/internal/usecase/ocservgroup"
	userusecase "github.com/mmtaee/ocserv-dashboard/backend/internal/usecase/ocservuser"
	reportusecase "github.com/mmtaee/ocserv-dashboard/backend/internal/usecase/report"
	systemusecase "github.com/mmtaee/ocserv-dashboard/backend/internal/usecase/system"
	systemdusecase "github.com/mmtaee/ocserv-dashboard/backend/internal/usecase/systemd"
	telegramusecase "github.com/mmtaee/ocserv-dashboard/backend/internal/usecase/telegram"
	adminuserusecase "github.com/mmtaee/ocserv-dashboard/backend/internal/usecase/user"
	"github.com/mmtaee/ocserv-dashboard/backend/pkg/captcha"
	"github.com/mmtaee/ocserv-dashboard/backend/pkg/crypto"
	"github.com/mmtaee/ocserv-dashboard/backend/services/admin_api/internal/backup"
	"github.com/mmtaee/ocserv-dashboard/backend/services/admin_api/internal/home"
	"github.com/mmtaee/ocserv-dashboard/backend/services/admin_api/internal/occtl"
	"github.com/mmtaee/ocserv-dashboard/backend/services/admin_api/internal/ocserv_group"
	"github.com/mmtaee/ocserv-dashboard/backend/services/admin_api/internal/ocserv_user"
	"github.com/mmtaee/ocserv-dashboard/backend/services/admin_api/internal/report"
	"github.com/mmtaee/ocserv-dashboard/backend/services/admin_api/internal/system"
	"github.com/mmtaee/ocserv-dashboard/backend/services/admin_api/internal/systemd"
	"github.com/mmtaee/ocserv-dashboard/backend/services/admin_api/internal/telegram"
)

type Service struct {
	backup         *backup.Controller
	home           *home.Controller
	occtl          *occtl.Controller
	ocservGroup    *ocserv_group.Controller
	ocservUser     *ocserv_user.Controller
	report         *report.Controller
	system         *system.Controller
	systemd        *systemd.Controller
	telegram       *telegram.Controller
	telegramRoutes bool
}

func New(telegramRoutes bool) *Service {
	occtlUC := occtlusecase.New(repository.NewOcctlRepository())
	ocservUserUC := userusecase.New(repository.NewtOcservUserRepository())
	ocservGroupUC := groupusecase.New(repository.NewOcservGroupRepository())
	reportUC := reportusecase.New(repository.NewtReportRepository())
	telegramUC := telegramusecase.New(repository.NewTelegramRepository())
	return &Service{
		backup:         backup.New(ocservUserUC, ocservGroupUC, backupusecase.New(repository.NewBackupRepository())),
		home:           home.New(occtlUC, ocservUserUC, reportUC, telegramUC),
		occtl:          occtl.New(occtlUC),
		ocservGroup:    ocserv_group.New(ocservGroupUC, ocservUserUC),
		ocservUser:     ocserv_user.New(ocservUserUC, occtlUC, reportUC),
		report:         report.New(reportUC, occtlUC),
		system:         system.New(systemusecase.New(repository.NewSystemRepository()), adminuserusecase.New(repository.NewUserRepository()), captcha.NewGoogleVerifier(), crypto.NewCustomPassword()),
		systemd:        systemd.New(systemdusecase.New(repository.NewSystemdRepository("ocserv"))),
		telegram:       telegram.New(telegramUC, ocservUserUC),
		telegramRoutes: telegramRoutes,
	}
}

func (s *Service) Register(group *echo.Group) {
	system.Routes(group, s.system)
	ocserv_group.Routes(group, s.ocservGroup)
	ocserv_user.Routes(group, s.ocservUser)
	occtl.Routes(group, s.occtl)
	home.Routes(group, s.home)
	backup.Routes(group, s.backup)
	report.Routes(group, s.report)
	systemd.Routes(group, s.systemd)
	if s.telegramRoutes {
		telegram.Routes(group, s.telegram)
	}
}
