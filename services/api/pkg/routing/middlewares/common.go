package middlewares

import (
	"github.com/labstack/echo/v4"
	apiModels "github.com/mmtaee/ocserv-users-management/api/internal/models"
	"net/http"
	"strings"
)

type Unauthorized struct {
	Error string `json:"error"`
}

type PermissionDenied struct {
	Error string `json:"error"`
}

type TooManyRequests struct {
	Error string `json:"error"`
}

func UnauthorizedError(c echo.Context, msg string) error {
	return c.JSON(http.StatusUnauthorized, Unauthorized{Error: msg})
}

func PermissionDeniedError(c echo.Context, msg string) error {
	return c.JSON(http.StatusForbidden, PermissionDenied{Error: msg})
}

func TooManyRequestsError(c echo.Context, msg string) error {
	return c.JSON(http.StatusTooManyRequests, TooManyRequests{Error: msg})
}

func actionFromMethod(method string) (apiModels.UserAction, bool) {
	switch method {
	case http.MethodGet:
		return apiModels.ActionGet, true
	case http.MethodPost:
		return apiModels.ActionPost, true
	case http.MethodDelete:
		return apiModels.ActionDelete, true
	case http.MethodPatch, http.MethodPut:
		return apiModels.ActionPatch, true
	default:
		return "", false
	}
}

func serviceParentWildcard(service string) string {
	if idx := strings.Index(service, "."); idx != -1 {
		return service[:idx] + ".*"
	}
	return service + ".*"
}
