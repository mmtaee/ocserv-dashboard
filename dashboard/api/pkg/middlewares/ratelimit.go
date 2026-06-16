package middlewares

import (
	"net/http"
	"sync"

	"github.com/labstack/echo/v5"
	"github.com/mmtaee/ocserv-dashboard/dashboard/api/pkg/request"
	"golang.org/x/time/rate"
)

// SystemStore implements in-memory rate limiting
type SystemStore struct {
	ips map[string]*rate.Limiter
	mu  *sync.RWMutex
	r   rate.Limit
	b   int
}

func NewSystemStore(r rate.Limit, b int) *SystemStore {
	return &SystemStore{
		ips: make(map[string]*rate.Limiter),
		mu:  &sync.RWMutex{},
		r:   r,
		b:   b,
	}
}

func (s *SystemStore) Allow(ip string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	limiter, exists := s.ips[ip]
	if !exists {
		limiter = rate.NewLimiter(s.r, s.b)
		s.ips[ip] = limiter
	}

	return limiter.Allow()
}

// RateLimitMiddleware creates a middleware for IP-based rate limiting
func RateLimitMiddleware(store *SystemStore) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			ip := c.RealIP()

			if !store.Allow(ip) {
				return c.JSON(http.StatusTooManyRequests, request.MessageResponse{Message: "too many requests"})
			}

			return next(c)
		}
	}
}
