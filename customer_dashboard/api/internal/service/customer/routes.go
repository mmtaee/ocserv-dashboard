package customer

import (
	"github.com/labstack/echo/v5"
	"github.com/mmtaee/ocserv-dashboard/customer_dashboard/api/internal/repository"
	"github.com/mmtaee/ocserv-dashboard/customer_dashboard/api/internal/usecase"
	"github.com/mmtaee/ocserv-dashboard/customer_dashboard/api/pkg/infra"
	"github.com/mmtaee/ocserv-dashboard/customer_dashboard/api/pkg/middlewares"
	"github.com/mmtaee/ocserv-dashboard/core/pkg/ocserv/occtl"
	"github.com/mmtaee/ocserv-dashboard/core/pkg/ocserv/user"
)

// RegisterRoutes registers all customer routes
func RegisterRoutes(e *echo.Echo) {
	ocservUserRepo := repository.NewOcservUserRepository(infra.DB, user.NewOcservUser())
	systemRepo := repository.NewSystemRepository(infra.DB)
	occtlRepo := repository.NewOcctlRepository(occtl.NewOcservOcctl())
	ocservUser := user.NewOcservUser()

	customerUC := usecase.NewCustomerUseCase(systemRepo, ocservUserRepo, occtlRepo, ocservUser)

	customerController := NewController(customerUC)

	// Protected customer routes
	protected := e.Group("/api/v1/customer")
	protected.Use(middlewares.CustomerAuthMiddleware())
	protected.GET("/user", customerController.GetUser)
	protected.GET("/online", customerController.IsOnline)
	protected.GET("/sessions/online", customerController.OnlineSessions)
	protected.POST("/sessions/terminate-all", customerController.TerminateAllSessions)
	protected.POST("/sessions/terminate/:id", customerController.TerminateSession)
	protected.POST("/password", customerController.UpdatePassword)
	protected.POST("/certificate/create", customerController.CreateCertificate)
	protected.GET("/certificate", customerController.DownloadCertificate)
	protected.GET("/statistics", customerController.UserStatistics)
	protected.GET("/session-logs", customerController.SessionLogs)
	protected.GET("/summary", customerController.Summary)
	protected.POST("/disconnect-sessions", customerController.DisconnectSessions)
	protected.GET("/setup/cisco", customerController.CiscoSetup)

	// Public route for Cisco setup certificate
	e.GET("/api/v1/customer/setup/cisco/certificate/:token", customerController.DownloadCiscoSetupCertificate)
}
