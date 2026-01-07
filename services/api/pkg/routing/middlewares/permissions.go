package middlewares

import (
	"github.com/labstack/echo/v4"
	apiModels "github.com/mmtaee/ocserv-users-management/api/internal/models"
	"github.com/mmtaee/ocserv-users-management/common/pkg/database"
)

func AdminPermission() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			role := c.Get("role").(apiModels.UserRole)
			if role == apiModels.RoleSuperAdmin || role == apiModels.RoleAdmin {
				return next(c)
			}
			return PermissionDeniedError(c, "Super Admin or Admin permission required")
		}
	}
}

func SuperAdminPermission() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			role := c.Get("role").(apiModels.UserRole)
			if role == apiModels.RoleSuperAdmin {
				return next(c)
			}
			return PermissionDeniedError(c, "Super Admin permission required")
		}
	}
}

// StaffPermissionMiddleware returns a middleware that checks if the user
// has permission to perform a given action on a given service.
// usage:    StaffPermissionMiddleware("ocserv-user", "R"),
func StaffPermissionMiddleware(service string, action string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			userID := c.Get("userUID").(string)
			roleVal := c.Get("role")
			role, _ := roleVal.(apiModels.UserRole)

			if role == apiModels.RoleSuperAdmin {
				return next(c)
			}

			db := database.GetConnection()
			var count int64
			if err := db.Model(&apiModels.Permission{}).
				Where("uid = ? AND service = ? AND action = ?", userID, service, action).
				Count(&count).Error; err != nil {
				return echo.NewHTTPError(500, "failed to check permission")
			}

			if count == 0 {
				return echo.NewHTTPError(403, "permission denied")
			}
			return next(c)
		}
	}
}
