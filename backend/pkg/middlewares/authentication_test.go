package middlewares

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/mmtaee/ocserv-dashboard/backend/internal/models"
)

type testAuthenticator struct {
	superadmin bool
}

func (a testAuthenticator) FindActiveByToken(context.Context, string) (*models.UserToken, error) {
	return &models.UserToken{ID: 11, User: models.User{ID: 7, Username: "staff", Superadmin: a.superadmin}}, nil
}

func TestAuthMiddlewareBuildsPrincipalFromDatabaseSession(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderAuthorization, "Bearer token")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	handler := AuthMiddleware(testAuthenticator{})(func(c *echo.Context) error {
		principal, err := Principal(c)
		if err != nil {
			return err
		}
		if principal.SessionID != 11 || principal.UserID != 7 || principal.Username != "staff" || principal.Superadmin {
			t.Fatalf("unexpected principal: %+v", principal)
		}
		return c.NoContent(http.StatusNoContent)
	})
	if err := handler(c); err != nil {
		t.Fatal(err)
	}
}

func TestSuperadminPermission(t *testing.T) {
	for _, tc := range []struct {
		name       string
		superadmin bool
		wantCode   int
	}{
		{name: "normal user", superadmin: false, wantCode: http.StatusForbidden},
		{name: "superadmin", superadmin: true, wantCode: http.StatusNoContent},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, "/", nil)
			req.Header.Set(echo.HeaderAuthorization, "Bearer token")
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			handler := AuthMiddleware(testAuthenticator{superadmin: tc.superadmin})(SuperadminPermission()(func(c *echo.Context) error {
				return c.NoContent(http.StatusNoContent)
			}))
			err := handler(c)
			if httpErr, ok := err.(*echo.HTTPError); ok {
				if httpErr.Code != tc.wantCode {
					t.Fatalf("got %d, want %d", httpErr.Code, tc.wantCode)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if rec.Code != tc.wantCode {
				t.Fatalf("got %d, want %d", rec.Code, tc.wantCode)
			}
		})
	}
}
