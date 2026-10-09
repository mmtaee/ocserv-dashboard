package superadmin

import (
	"context"
	"testing"

	"github.com/mmtaee/ocserv-dashboard/backend/config"
	"github.com/mmtaee/ocserv-dashboard/backend/internal/models"
	"github.com/mmtaee/ocserv-dashboard/backend/pkg/crypto"
)

type fakeRepository struct {
	user  *models.User
	calls int
}

func (r *fakeRepository) EnsureSuperadmin(_ context.Context, user *models.User) (*models.User, error) {
	r.calls++
	if r.user == nil {
		r.user = user
	}
	return r.user, nil
}

func TestEnsureIsIdempotentAndCreatesSuperadmin(t *testing.T) {
	config.Init(false, "", 0)
	repository := &fakeRepository{}
	usecase := New(repository, crypto.NewCustomPassword())
	first, err := usecase.Ensure(context.Background(), " Admin ", "strong-password")
	if err != nil {
		t.Fatal(err)
	}
	if first.Username != "admin" || !first.Superadmin {
		t.Fatalf("unexpected user: %+v", first)
	}
	if _, err := usecase.Ensure(context.Background(), "admin", "strong-password"); err != nil {
		t.Fatal(err)
	}
	if repository.calls != 2 {
		t.Fatalf("expected two idempotent ensure calls, got %d", repository.calls)
	}
}
