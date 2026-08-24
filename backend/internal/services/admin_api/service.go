package adminapi

import (
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/mmtaee/ocserv-dashboard/backend/config"
	ocservgroupconfig "github.com/mmtaee/ocserv-dashboard/backend/internal/ocserv/group"
	ocservaccount "github.com/mmtaee/ocserv-dashboard/backend/internal/ocserv/user"
	platformocserv "github.com/mmtaee/ocserv-dashboard/backend/internal/platform/ocserv"
	platformsystemd "github.com/mmtaee/ocserv-dashboard/backend/internal/platform/systemd"
	telegramclient "github.com/mmtaee/ocserv-dashboard/backend/internal/platform/telegram"
	"github.com/mmtaee/ocserv-dashboard/backend/internal/repository"
	backupcontroller "github.com/mmtaee/ocserv-dashboard/backend/internal/services/admin_api/backup"
	dashboardcontroller "github.com/mmtaee/ocserv-dashboard/backend/internal/services/admin_api/dashboard"
	occtlcontroller "github.com/mmtaee/ocserv-dashboard/backend/internal/services/admin_api/occtl"
	groupcontroller "github.com/mmtaee/ocserv-dashboard/backend/internal/services/admin_api/ocserv_group"
	usercontroller "github.com/mmtaee/ocserv-dashboard/backend/internal/services/admin_api/ocserv_user"
	reportcontroller "github.com/mmtaee/ocserv-dashboard/backend/internal/services/admin_api/reports"
	systemcontroller "github.com/mmtaee/ocserv-dashboard/backend/internal/services/admin_api/system"
	systemdcontroller "github.com/mmtaee/ocserv-dashboard/backend/internal/services/admin_api/systemd"
	telegramcontroller "github.com/mmtaee/ocserv-dashboard/backend/internal/services/admin_api/telegram"
	backupusecase "github.com/mmtaee/ocserv-dashboard/backend/internal/usecase/admin_api/backup"
	dashboardusecase "github.com/mmtaee/ocserv-dashboard/backend/internal/usecase/admin_api/dashboard"
	groupusecase "github.com/mmtaee/ocserv-dashboard/backend/internal/usecase/admin_api/groups"
	occtlusecase "github.com/mmtaee/ocserv-dashboard/backend/internal/usecase/admin_api/occtl"
	reportusecase "github.com/mmtaee/ocserv-dashboard/backend/internal/usecase/admin_api/reports"
	systemusecase "github.com/mmtaee/ocserv-dashboard/backend/internal/usecase/admin_api/system"
	systemdusecase "github.com/mmtaee/ocserv-dashboard/backend/internal/usecase/admin_api/systemd"
	telegramusecase "github.com/mmtaee/ocserv-dashboard/backend/internal/usecase/admin_api/telegram"
	userusecase "github.com/mmtaee/ocserv-dashboard/backend/internal/usecase/admin_api/users"
	"github.com/mmtaee/ocserv-dashboard/backend/pkg/captcha"
	"github.com/mmtaee/ocserv-dashboard/backend/pkg/crypto"
)

type Service struct {
	backup         *backupcontroller.Controller
	dashboard      *dashboardcontroller.Controller
	occtl          *occtlcontroller.Controller
	groups         *groupcontroller.Controller
	users          *usercontroller.Controller
	reports        *reportcontroller.Controller
	system         *systemcontroller.Controller
	systemd        *systemdcontroller.Controller
	telegram       *telegramcontroller.Controller
	telegramRoutes bool
}

// New constructs the Admin API dependency graph.
func New(telegramRoutes bool) *Service {
	telegramRuntimeEnabled := strings.EqualFold(strings.TrimSpace(os.Getenv("TELEGRAM_BOT_ENABLED")), "true")
	accountStore := ocservaccount.NewOcservUser()
	occtlUC := occtlusecase.New(platformocserv.NewClient())
	reportUC := reportusecase.New(repository.NewtReportRepository(), occtlUC)
	ocservUserUC := userusecase.New(repository.NewtOcservUserRepository(), accountStore, occtlUC, reportUC)
	ocservGroupUC := groupusecase.New(repository.NewOcservGroupRepository(), ocservUserUC, ocservgroupconfig.NewOcservGroup(), occtlUC)
	telegramUC := telegramusecase.New(repository.NewTelegramRepository(), ocservUserUC, telegramclient.NewClient(&http.Client{Timeout: 8 * time.Second}))
	dashboardUC := dashboardusecase.New(occtlUC, reportUC, telegramUC, telegramRuntimeEnabled)

	return &Service{
		backup:    backupcontroller.New(backupusecase.New(repository.NewBackupRepository(), ocservGroupUC, ocservUserUC, accountStore)),
		dashboard: dashboardcontroller.New(dashboardUC),
		occtl:     occtlcontroller.New(occtlUC),
		groups:    groupcontroller.New(ocservGroupUC),
		users:     usercontroller.New(ocservUserUC),
		reports:   reportcontroller.New(reportUC),
		system: systemcontroller.New(systemusecase.New(
			repository.NewSystemRepository(), repository.NewUserRepository(), captcha.NewGoogleVerifier(), crypto.NewCustomPassword(),
			systemusecase.Options{
				SecretKey: config.Get().SecretKey, CurrentRelease: os.Getenv("CURRENT_RELEASE"), TelegramEnabled: telegramRuntimeEnabled,
				ReleaseTimeout: 5 * time.Second,
			},
		)),
		systemd: systemdcontroller.New(systemdusecase.New(
			platformsystemd.NewClient("ocserv"),
			strings.EqualFold(strings.TrimSpace(os.Getenv("SYSTEMD")), "true"),
		)),
		telegram:       telegramcontroller.New(telegramUC),
		telegramRoutes: telegramRoutes,
	}
}
