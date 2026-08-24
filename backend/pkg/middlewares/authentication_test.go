package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/mmtaee/ocserv-dashboard/backend/config"
	"github.com/mmtaee/ocserv-dashboard/backend/pkg/crypto"
)

func TestAuthMiddlewareBuildsPrincipalFromClaims(t *testing.T) {
	config.Init(false, "", 0)
	token, err := crypto.GenerateAccessToken(7, "staff", time.Now().Add(time.Hour).Unix(), false)
	if err != nil {
		t.Fatal(err)
	}

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderAuthorization, "Bearer "+token)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	handler := AuthMiddleware()(func(c *echo.Context) error {
		principal, err := Principal(c)
		if err != nil {
			return err
		}
		if principal.UserID != 7 || principal.Username != "staff" || principal.Superadmin {
			t.Fatalf("unexpected principal: %+v", principal)
		}
		return c.NoContent(http.StatusNoContent)
	})
	if err := handler(c); err != nil {
		t.Fatal(err)
	}
}

func TestSuperadminPermission(t *testing.T) {
	config.Init(false, "", 0)
	for _, tc := range []struct {
		name       string
		superadmin bool
		wantCode   int
	}{
		{name: "normal user", superadmin: false, wantCode: http.StatusForbidden},
		{name: "superadmin", superadmin: true, wantCode: http.StatusNoContent},
	} {
		t.Run(tc.name, func(t *testing.T) {
			token, err := crypto.GenerateAccessToken(7, "staff", time.Now().Add(time.Hour).Unix(), tc.superadmin)
			if err != nil {
				t.Fatal(err)
			}
			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, "/", nil)
			req.Header.Set(echo.HeaderAuthorization, "Bearer "+token)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			handler := AuthMiddleware()(SuperadminPermission()(func(c *echo.Context) error {
				return c.NoContent(http.StatusNoContent)
			}))
			err = handler(c)
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
