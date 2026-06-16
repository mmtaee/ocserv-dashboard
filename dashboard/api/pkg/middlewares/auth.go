package middlewares

import (
	"github.com/labstack/echo/v5"
	"github.com/mmtaee/ocserv-dashboard/core/models"
	"github.com/mmtaee/ocserv-dashboard/dashboard/api/internal/repository"
	"github.com/mmtaee/ocserv-dashboard/dashboard/api/pkg/auth"
	"github.com/mmtaee/ocserv-dashboard/dashboard/api/pkg/infra"
	"github.com/mmtaee/ocserv-dashboard/dashboard/api/pkg/request"
	"log"
	"strings"
)

// AuthMiddleware creates a middleware to check JWT access token and if admin is suspended
// Usage: e.GET("/protected", handler, middlewares.AuthMiddleware())
func AuthMiddleware() echo.MiddlewareFunc {
	req := &request.Request{}
	adminRepo := repository.NewAdminRepository(infra.DB)

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				log.Printf("No Authorization header\n")
				return req.Unauthorized(c, nil, "missing authorization header")
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				log.Printf("Invalid Authorization header\n")
				return req.Unauthorized(c, nil, "invalid authorization header format")
			}

			tokenString := parts[1]
			claims, err := auth.ValidateAdministratorToken(tokenString)
			if err != nil {
				log.Printf("error accoured in AuthMiddleware validate token %v", err)
				return req.Unauthorized(c, err, "invalid or expired token")
			}

			// Load admin from DB to check suspension status
			admin, err := adminRepo.FindByID(claims.ID)
			if err != nil {
				return req.ResponseWithCode(c, 4003, err)
			}

			if admin.IsSuspended {
				return req.ResponseWithCode(c, 4005, nil)
			}

			// Set user information to echo context
			c.Set("id", claims.ID)
			c.Set("role", claims.Role)

			return next(c)
		}
	}
}

// SuperAdminMiddleware creates a middleware to check if the admin is a super user
// Usage: e.GET("/super-admin", handler, middlewares.AuthMiddleware(), middlewares.SuperAdminMiddleware())
func SuperAdminMiddleware() echo.MiddlewareFunc {
	req := &request.Request{}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			role := c.Get("role").(string)
			if role != models.AdminRoleSuper {
				return req.ResponseWithCode(c, 4001, nil)
			}

			return next(c)
		}
	}
}
