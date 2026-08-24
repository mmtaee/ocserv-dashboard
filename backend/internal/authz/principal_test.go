package authz

import (
	"errors"
	"testing"
)

func TestPrincipalOwnership(t *testing.T) {
	normal := Principal{UserID: 7, Username: "staff"}
	if err := normal.RequireOwner(7); err != nil {
		t.Fatalf("owner should be allowed: %v", err)
	}
	if err := normal.RequireOwner(8); !errors.Is(err, ErrForbidden) {
		t.Fatalf("another user's account should be forbidden, got %v", err)
	}
	superadmin := Principal{UserID: 1, Username: "admin", Superadmin: true}
	if err := superadmin.RequireOwner(8); err != nil {
		t.Fatalf("superadmin should be allowed: %v", err)
	}
}

func TestPrincipalSuperadmin(t *testing.T) {
	if err := (Principal{UserID: 7}).RequireSuperadmin(); !errors.Is(err, ErrForbidden) {
		t.Fatalf("normal user should be forbidden, got %v", err)
	}
	if err := (Principal{UserID: 1, Superadmin: true}).RequireSuperadmin(); err != nil {
		t.Fatalf("superadmin should be allowed: %v", err)
	}
}
