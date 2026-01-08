package middlewares

import (
	"errors"
	"github.com/labstack/echo/v4"
	apiModels "github.com/mmtaee/ocserv-users-management/api/internal/models"
	"github.com/mmtaee/ocserv-users-management/common/pkg/database"
	"gorm.io/gorm"
)

func SuperAdminOrAdminPermission() echo.MiddlewareFunc {
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

// StaffPermissionMiddleware returns a middleware that checks
// if the user has permission to perform a given action on a given service.
// Usage: StaffPermissionMiddleware(apiModels.OcservUsersCRUDService)
func StaffPermissionMiddleware(UserService apiModels.UserService) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// -----------------------
			// Safe type assertions
			// -----------------------
			userIDI := c.Get("userUID")
			roleI := c.Get("role")

			userID, ok := userIDI.(string)
			if !ok || userID == "" {
				return PermissionDeniedError(c, "invalid user context")
			}

			role, ok := roleI.(apiModels.UserRole)
			if !ok {
				return PermissionDeniedError(c, "invalid role context")
			}

			// -----------------------
			// Admin bypass
			// -----------------------
			if role == apiModels.RoleSuperAdmin || role == apiModels.RoleAdmin {
				return next(c)
			}

			// -----------------------
			// Resolve action from HTTP method
			// -----------------------
			action, ok := actionFromMethod(c.Request().Method)
			if !ok {
				return PermissionDeniedError(c, "unsupported HTTP method")
			}

			// -----------------------
			// Wildcard handling
			// -----------------------
			service := string(UserService)
			parentService := serviceParentWildcard(service) // e.g. "ocserv-groups.*"

			// -----------------------
			// Permission check (optimized)
			// -----------------------
			db := database.GetConnection()
			var exists bool

			err := db.Model(&apiModels.Permission{}).
				Select("1").
				Where(`
					uid = ?
					AND (
						(service = ? AND action = ?)          -- exact match
						OR (service = ? AND action = '*')     -- service-level wildcard action
						OR (service = ?)                       -- full service wildcard (ocserv-groups.*)
						OR (service = '*' AND action = '*')   -- global admin (optional)
					)
				`,
					userID,
					service, action,
					service,
					parentService,
				).
				Limit(1).
				Scan(&exists).Error

			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) || !exists {
					return PermissionDeniedError(c, "You do not have permission to access this route")
				}
				return PermissionDeniedError(c, "permission check failed")
			}

			if !exists {
				return PermissionDeniedError(c, "You do not have permission to access this route")
			}

			return next(c)
		}
	}
}
