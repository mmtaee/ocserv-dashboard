package adminapi

import (
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/mmtaee/ocserv-dashboard/backend/config"
	ocservgroupconfig "github.com/mmtaee/ocserv-dashboard/backend/internal/ocserv/group"
	ocservaccount "github.com/mmtaee/ocserv-dashboard/backend/internal/ocserv/user"
	platformocserv "github.com/mmtaee/ocserv-dashboard/backend/internal/platform/ocserv"
	telegramclient "github.com/mmtaee/ocserv-dashboard/backend/internal/platform/telegram"
	"github.com/mmtaee/ocserv-dashboard/backend/internal/repository"
	agentsettingscontroller "github.com/mmtaee/ocserv-dashboard/backend/internal/services/admin_api/agent_settings"
	authcontroller "github.com/mmtaee/ocserv-dashboard/backend/internal/services/admin_api/auth"
	backupcontroller "github.com/mmtaee/ocserv-dashboard/backend/internal/services/admin_api/backup"
	dashboardcontroller "github.com/mmtaee/ocserv-dashboard/backend/internal/services/admin_api/dashboard"
	occtlcontroller "github.com/mmtaee/ocserv-dashboard/backend/internal/services/admin_api/occtl"
	agentcontroller "github.com/mmtaee/ocserv-dashboard/backend/internal/services/admin_api/ocserv_agent"
	groupcontroller "github.com/mmtaee/ocserv-dashboard/backend/internal/services/admin_api/ocserv_group"
	usercontroller "github.com/mmtaee/ocserv-dashboard/backend/internal/services/admin_api/ocserv_user"
	reportcontroller "github.com/mmtaee/ocserv-dashboard/backend/internal/services/admin_api/reports"
	runtimecontroller "github.com/mmtaee/ocserv-dashboard/backend/internal/services/admin_api/runtime"
	systemcontroller "github.com/mmtaee/ocserv-dashboard/backend/internal/services/admin_api/system"
	telegramcontroller "github.com/mmtaee/ocserv-dashboard/backend/internal/services/admin_api/telegram"
	agentsettingsusecase "github.com/mmtaee/ocserv-dashboard/backend/internal/usecase/admin_api/agent_settings"
	agentusecase "github.com/mmtaee/ocserv-dashboard/backend/internal/usecase/admin_api/agents"
	authusecase "github.com/mmtaee/ocserv-dashboard/backend/internal/usecase/admin_api/auth"
	backupusecase "github.com/mmtaee/ocserv-dashboard/backend/internal/usecase/admin_api/backup"
	dashboardusecase "github.com/mmtaee/ocserv-dashboard/backend/internal/usecase/admin_api/dashboard"
	groupusecase "github.com/mmtaee/ocserv-dashboard/backend/internal/usecase/admin_api/groups"
	occtlusecase "github.com/mmtaee/ocserv-dashboard/backend/internal/usecase/admin_api/occtl"
	reportusecase "github.com/mmtaee/ocserv-dashboard/backend/internal/usecase/admin_api/reports"
	systemusecase "github.com/mmtaee/ocserv-dashboard/backend/internal/usecase/admin_api/system"
	telegramusecase "github.com/mmtaee/ocserv-dashboard/backend/internal/usecase/admin_api/telegram"
	userusecase "github.com/mmtaee/ocserv-dashboard/backend/internal/usecase/admin_api/users"
	runtimeusecase "github.com/mmtaee/ocserv-dashboard/backend/internal/usecase/system"
	"github.com/mmtaee/ocserv-dashboard/backend/pkg/captcha"
	"github.com/mmtaee/ocserv-dashboard/backend/pkg/crypto"
	"github.com/mmtaee/ocserv-dashboard/backend/pkg/middlewares"
)

type Service struct {
	agentNode      bool
	agentSettings  *agentsettingscontroller.Controller
	agents         *agentcontroller.Controller
	auth           *authcontroller.Controller
	authenticate   echo.MiddlewareFunc
	backup         *backupcontroller.Controller
	dashboard      *dashboardcontroller.Controller
	occtl          *occtlcontroller.Controller
	groups         *groupcontroller.Controller
	users          *usercontroller.Controller
	reports        *reportcontroller.Controller
	system         *systemcontroller.Controller
	runtime        *runtimecontroller.Controller
	telegram       *telegramcontroller.Controller
	telegramRoutes bool
}

// New constructs the Admin API dependency graph.
func New(telegramRoutes, dockerMode bool) (*Service, error) {
	cfg := config.Get()
	telegramRuntimeEnabled := strings.EqualFold(strings.TrimSpace(os.Getenv("TELEGRAM_BOT_ENABLED")), "true")
	userRepository := repository.NewUserRepository()
	sessionRepository := repository.NewUserTokenRepository()
	accountStore := ocservaccount.NewOcservUser()
	occtlUC := occtlusecase.New(platformocserv.NewClient())
	reportUC := reportusecase.New(repository.NewtReportRepository(), occtlUC)
	ocservUserUC := userusecase.New(repository.NewtOcservUserRepository(), accountStore, occtlUC, reportUC)
	ocservGroupUC := groupusecase.New(repository.NewOcservGroupRepository(), ocservUserUC, ocservgroupconfig.NewOcservGroup(), occtlUC)
	telegramUC := telegramusecase.New(repository.NewTelegramRepository(), ocservUserUC, telegramclient.NewClient(&http.Client{Timeout: 8 * time.Second}))
	dashboardUC := dashboardusecase.New(occtlUC, reportUC, telegramUC, telegramRuntimeEnabled)
	runtimeService, err := runtimecontroller.NewRuntime(
		dockerMode,
		strings.EqualFold(strings.TrimSpace(os.Getenv("SYSTEMD")), "true"),
	)
	if err != nil {
		return nil, err
	}
	runtimeUC := runtimeusecase.New(runtimeService, runtimecontroller.NewConfigFile(runtimecontroller.DefaultOcservConfigPath))
	var agentSettings *agentsettingscontroller.Controller
	var agents *agentcontroller.Controller
	if cfg.AgentNode {
		agentSettings = agentsettingscontroller.New(agentsettingsusecase.New(repository.NewAgentTokenRepository()))
	} else {
		agents = agentcontroller.New(agentusecase.New(repository.NewOcservAgentRepository()))
	}

	return &Service{
		agentNode:     cfg.AgentNode,
		agentSettings: agentSettings,
		agents:        agents,
		auth:          authcontroller.New(authusecase.New(sessionRepository)),
		authenticate:  middlewares.AuthMiddleware(sessionRepository),
		backup:        backupcontroller.New(backupusecase.New(repository.NewBackupRepository(), ocservGroupUC, ocservUserUC, accountStore)),
		dashboard:     dashboardcontroller.New(dashboardUC),
		occtl:         occtlcontroller.New(occtlUC),
		groups:        groupcontroller.New(ocservGroupUC),
		users:         usercontroller.New(ocservUserUC),
		reports:       reportcontroller.New(reportUC),
		system: systemcontroller.New(systemusecase.New(
			repository.NewSystemRepository(), userRepository, sessionRepository, captcha.NewGoogleVerifier(), crypto.NewCustomPassword(),
			systemusecase.Options{
				SecretKey: cfg.SecretKey, CurrentRelease: os.Getenv("CURRENT_RELEASE"), TelegramEnabled: telegramRuntimeEnabled,
				ReleaseTimeout: 5 * time.Second,
			},
		)),
		runtime:        runtimecontroller.New(runtimeUC),
		telegram:       telegramcontroller.New(telegramUC),
		telegramRoutes: telegramRoutes,
	}, nil
}
