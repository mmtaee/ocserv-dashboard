package units

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/mmtaee/ocserv-dashboard/backend/internal/models"
	"github.com/mmtaee/ocserv-dashboard/backend/internal/repository"
	customerusecase "github.com/mmtaee/ocserv-dashboard/backend/internal/usecase/customer_api"
	"github.com/stretchr/testify/require"
)

type customerSystemRepository struct{}

func (customerSystemRepository) System(context.Context) (*models.System, error) {
	return &models.System{ClientProfileConnectionName: "VPN", ClientProfileServerAddress: "vpn.example.com", ClientProfileServerPort: 443}, nil
}

type customerUserRepository struct {
	user *models.OcservUser
}

func (r customerUserRepository) GetByUsername(context.Context, string) (*models.OcservUser, error) {
	if r.user == nil {
		return nil, errors.New("not found")
	}
	return r.user, nil
}

func (customerUserRepository) TotalBandwidthUserDateRange(context.Context, uint, *time.Time, *time.Time) (repository.TotalBandwidths, error) {
	return repository.TotalBandwidths{RX: 10, TX: 20}, nil
}

func (customerUserRepository) CertificatePathByUsername(context.Context, string) (string, error) {
	return "/tmp/user.p12", nil
}

func (customerUserRepository) CreateCertificate(context.Context, string) error { return nil }

type customerOcctlRepository struct{}

func (customerOcctlRepository) Disconnect(string) (string, error) { return "", nil }

func TestCustomerSummaryRejectsInvalidPassword(t *testing.T) {
	users := customerUserRepository{user: &models.OcservUser{Username: "alice", Password: "correct"}}
	usecase := customerusecase.New(customerSystemRepository{}, users, customerOcctlRepository{}, "secret")

	_, err := usecase.Summary(context.Background(), customerusecase.Credentials{Username: "alice", Password: "wrong"})

	require.ErrorIs(t, err, customerusecase.ErrInvalidCredentials)
}

func TestCustomerSummaryMapsUsage(t *testing.T) {
	users := customerUserRepository{user: &models.OcservUser{Username: "alice", Password: "correct", Owner: "admin"}}
	usecase := customerusecase.New(customerSystemRepository{}, users, customerOcctlRepository{}, "secret")

	result, err := usecase.Summary(context.Background(), customerusecase.Credentials{Username: "alice", Password: "correct"})

	require.NoError(t, err)
	require.Equal(t, "alice", result.OcservUser.Username)
	require.Equal(t, "admin", result.OcservUser.Owner)
	require.Equal(t, float64(10), result.Usage.Bandwidths.RX)
	require.Equal(t, float64(20), result.Usage.Bandwidths.TX)
}
