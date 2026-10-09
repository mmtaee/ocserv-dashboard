package units

import (
	"context"
	"testing"

	"github.com/mmtaee/ocserv-dashboard/backend/internal/models"
	adminsystem "github.com/mmtaee/ocserv-dashboard/backend/internal/usecase/admin_api/system"
	"github.com/stretchr/testify/require"
)

type adminSystemRepository struct {
	updated *models.System
}

func (r *adminSystemRepository) System(context.Context) (*models.System, error) {
	return &models.System{}, nil
}

func (r *adminSystemRepository) SystemUpdate(_ context.Context, settings *models.System) (*models.System, error) {
	r.updated = settings
	return settings, nil
}

func TestAdminSystemUpdateMarksFirstInit(t *testing.T) {
	repository := &adminSystemRepository{}
	usecase := adminsystem.New(repository, nil, nil, nil, nil, adminsystem.Options{})

	siteKey := "site"
	secretKey := "secret"
	autoDelete := false
	keepInactiveDays := 30
	serverAddress := "vpn.example.com"
	serverPort := 443
	connectionName := "VPN"

	response, err := usecase.Update(context.Background(), 1, adminsystem.PatchSystemUpdateData{
		GoogleCaptchaSiteKey:        &siteKey,
		GoogleCaptchaSecretKey:      &secretKey,
		AutoDeleteInactiveUsers:     &autoDelete,
		KeepInactiveUserDays:        &keepInactiveDays,
		ClientProfileServerAddress:  &serverAddress,
		ClientProfileServerPort:     &serverPort,
		ClientProfileConnectionName: &connectionName,
	})

	require.NoError(t, err)
	require.NotNil(t, repository.updated)
	require.True(t, repository.updated.FirstInit)
	require.True(t, response.FirstInit)
}
