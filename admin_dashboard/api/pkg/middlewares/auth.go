package middlewares

import (
	"github.com/labstack/echo/v5"
	"github.com/mmtaee/ocserv-dashboard/core/models"
	"github.com/mmtaee/ocserv-dashboard/core/pkg/request"
	"github.com/mmtaee/ocserv-dashboard/dashboard/api/internal/repository"
	"github.com/mmtaee/ocserv-dashboard/dashboard/api/pkg/infra"
	"log"
	"strings"
)

// AuthMiddleware creates a middleware to check auth token and if admin is suspended
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
			
			adminToken, err := adminRepo.FindToken(tokenString)
			if err != nil {
				log.Printf("error accoured in AuthMiddleware find token %v", err)
				return req.Unauthorized(c, err, "invalid token")
			}
			
			// Check if admin is suspended
			if adminToken.Administrator.IsSuspended {
				return req.ResponseWithCode(c, 4005, nil)
			}
			
			// Set token on context so we can use it for logout
			c.Set("token", tokenString)
			// Set user information to echo context
			c.Set("id", adminToken.Administrator.ID)
			c.Set("role", adminToken.Administrator.Role)
			
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

// AdminPermissionMiddleware creates a middleware to check if the user is any authenticated admin (super or admin)
// Usage: e.GET("/protected", handler, middlewares.AuthMiddleware(), middlewares.AdminPermissionMiddleware())
func AdminPermissionMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			// Just check that we have an authenticated user (role is set)
			role := c.Get("role")
			if role == nil {
				req := &request.Request{}
				return req.ResponseWithCode(c, 4001, nil)
			}
			return next(c)
		}
	}
}
