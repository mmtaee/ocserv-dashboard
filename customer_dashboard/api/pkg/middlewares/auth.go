package middlewares

import (
	"log"
	"strings"

	"github.com/labstack/echo/v5"
	"github.com/mmtaee/ocserv-dashboard/core/models"
	"github.com/mmtaee/ocserv-dashboard/core/pkg/ocserv/user"
	"github.com/mmtaee/ocserv-dashboard/core/pkg/request"
	"github.com/mmtaee/ocserv-dashboard/customer_dashboard/api/internal/repository"
	"github.com/mmtaee/ocserv-dashboard/customer_dashboard/api/pkg/auth"
	"github.com/mmtaee/ocserv-dashboard/customer_dashboard/api/pkg/infra"
)

// CustomerAuthMiddleware creates a middleware to check customer JWT access token
// Usage: e.GET("/customer-protected", handler, middlewares.CustomerAuthMiddleware())
func CustomerAuthMiddleware() echo.MiddlewareFunc {
	req := &request.Request{}
	ocservUserRepo := repository.NewOcservUserRepository(infra.DB, user.NewOcservUser())

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
			claims, err := auth.ValidateCustomerToken(tokenString)
			if err != nil {
				log.Printf("error occurred in CustomerAuthMiddleware validate token %v", err)
				return req.Unauthorized(c, err, "invalid or expired token")
			}

			// Find user by username to check if locked
			ocservUser, err := ocservUserRepo.FindByUsername(claims.Username)
			if err != nil {
				return req.Unauthorized(c, err, "invalid or expired token")
			}
			if ocservUser.IsLocked {
				return req.ResponseWithCode(c, "8001", nil)
			}

			// Set customer username to echo context
			c.Set("username", claims.Username)

			return next(c)
		}
	}
}
