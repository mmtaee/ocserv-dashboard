package units

import (
	"context"
	"errors"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/mmtaee/ocserv-dashboard/backend/config"
	"github.com/mmtaee/ocserv-dashboard/backend/internal/models"
	"github.com/mmtaee/ocserv-dashboard/backend/internal/repository"
	adminapi "github.com/mmtaee/ocserv-dashboard/backend/internal/services/admin_api"
	agentsettings "github.com/mmtaee/ocserv-dashboard/backend/internal/usecase/admin_api/agent_settings"
	"github.com/mmtaee/ocserv-dashboard/backend/internal/usecase/admin_api/agents"
	authusecase "github.com/mmtaee/ocserv-dashboard/backend/internal/usecase/admin_api/auth"
	systemusecase "github.com/mmtaee/ocserv-dashboard/backend/internal/usecase/admin_api/system"
	"github.com/mmtaee/ocserv-dashboard/backend/pkg/crypto"
	"github.com/mmtaee/ocserv-dashboard/backend/pkg/request"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type loginSystemRepository struct {
	repository.SystemRepositoryInterface
}

func (loginSystemRepository) System(context.Context) (*models.System, error) {
	return &models.System{}, nil
}

type loginUserRepository struct {
	repository.UserRepositoryInterface
	user *models.User
}

func (r loginUserRepository) GetByUsername(context.Context, string) (*models.User, error) {
	return r.user, nil
}

func (loginUserRepository) UpdateLastLogin(context.Context, *models.User) error { return nil }

type loginSessionRepository struct {
	session *models.UserToken
}

func (r *loginSessionRepository) Create(_ context.Context, session *models.UserToken) error {
	r.session = session
	return nil
}

type loginCaptcha struct{}

func (loginCaptcha) SetSecretKey(string) {}
func (loginCaptcha) Verify(string)       {}
func (loginCaptcha) IsValid() bool       { return true }

func TestLoginCreatesOpaqueDatabaseSession(t *testing.T) {
	config.Init(false, "", 0)
	passwords := crypto.NewCustomPassword()
	password := passwords.CreatePassword("secret")
	sessions := &loginSessionRepository{}
	usecase := systemusecase.New(
		loginSystemRepository{},
		loginUserRepository{user: &models.User{ID: 7, Username: "admin", Password: password.Hash, Salt: password.Salt, Superadmin: true}},
		sessions,
		loginCaptcha{},
		passwords,
		systemusecase.Options{},
	)

	result, err := usecase.Login(context.Background(), systemusecase.LoginData{
		Username: "admin", Password: "secret", RememberMe: true,
	}, "browser/1.0")
	require.NoError(t, err)
	require.NotEmpty(t, result.Token)
	require.NotNil(t, sessions.session)
	require.Equal(t, uint(7), sessions.session.UserID)
	require.Equal(t, "browser/1.0", sessions.session.UserAgent)
	require.Equal(t, crypto.HashToken(result.Token), sessions.session.Token)
	require.NotEqual(t, result.Token, sessions.session.Token)
}

type sessionRepository struct {
	deleted []uint
	items   []models.UserToken
}

func (r *sessionRepository) DeleteByID(_ context.Context, id uint) error {
	r.deleted = append(r.deleted, id)
	return nil
}

func (r *sessionRepository) List(context.Context, *request.Pagination) ([]models.UserToken, int64, error) {
	return r.items, int64(len(r.items)), nil
}

func TestLogoutAndSessionManagementUseSessionIDs(t *testing.T) {
	repository := &sessionRepository{items: []models.UserToken{{
		ID: 4, UserID: 7, UserAgent: "browser", User: models.User{ID: 7, Username: "admin"},
	}}}
	usecase := authusecase.New(repository)
	require.NoError(t, usecase.Logout(context.Background(), 4))
	require.NoError(t, usecase.Revoke(context.Background(), 5))
	sessions, total, err := usecase.Sessions(context.Background(), &request.Pagination{Page: 1, PageSize: 50})
	require.NoError(t, err)
	require.Equal(t, []uint{4, 5}, repository.deleted)
	require.Equal(t, int64(1), total)
	require.Equal(t, "admin", sessions[0].Username)
	require.Equal(t, "browser", sessions[0].UserAgent)
}

type agentRepository struct {
	items map[uint]*models.OcservAgent
	next  uint
}

func newAgentRepository() *agentRepository {
	return &agentRepository{items: make(map[uint]*models.OcservAgent), next: 1}
}

func (r *agentRepository) List(context.Context) ([]models.OcservAgent, error) {
	result := make([]models.OcservAgent, 0, len(r.items))
	for _, item := range r.items {
		result = append(result, *item)
	}
	return result, nil
}

func (r *agentRepository) GetByID(_ context.Context, id uint) (*models.OcservAgent, error) {
	item, ok := r.items[id]
	if !ok {
		return nil, errors.New("not found")
	}
	copy := *item
	return &copy, nil
}

func (r *agentRepository) Create(_ context.Context, agent *models.OcservAgent) error {
	agent.ID = r.next
	r.next++
	copy := *agent
	r.items[agent.ID] = &copy
	return nil
}

func (r *agentRepository) Update(_ context.Context, agent *models.OcservAgent) error {
	copy := *agent
	r.items[agent.ID] = &copy
	return nil
}

func (r *agentRepository) Delete(_ context.Context, id uint) error {
	delete(r.items, id)
	return nil
}

func TestMasterAgentCRUDUsesManuallyProvidedToken(t *testing.T) {
	repository := newAgentRepository()
	usecase := agents.New(repository)
	created, err := usecase.Create(context.Background(), agents.CreateInput{
		Name: "edge", AddressType: models.AgentAddressTypeDomain, Address: "VPN.Example.com", Token: "copied-agent-token",
	})
	require.NoError(t, err)
	require.Equal(t, "vpn.example.com", created.Address)
	require.Equal(t, "copied-agent-token", created.Token)

	updated, err := usecase.Update(context.Background(), created.ID, agents.UpdateInput{
		Name: "edge-2", AddressType: models.AgentAddressTypeIP, Address: "192.0.2.10", Token: "replacement-from-agent",
	})
	require.NoError(t, err)
	require.Equal(t, "replacement-from-agent", updated.Token)
	require.NoError(t, usecase.Delete(context.Background(), created.ID))
	_, err = usecase.Get(context.Background(), created.ID)
	require.Error(t, err)
}

func TestMasterAgentRejectsAddressTypeMismatch(t *testing.T) {
	usecase := agents.New(newAgentRepository())
	_, err := usecase.Create(context.Background(), agents.CreateInput{
		Name: "bad", AddressType: models.AgentAddressTypeIP, Address: "vpn.example.com", Token: "manual",
	})
	require.ErrorIs(t, err, agents.ErrInvalidAddress)
}

type localAgentTokenRepository struct {
	current *models.AgentToken
}

func (r *localAgentTokenRepository) Get(context.Context) (*models.AgentToken, error) {
	if r.current == nil {
		return nil, gorm.ErrRecordNotFound
	}
	copy := *r.current
	return &copy, nil
}

func (r *localAgentTokenRepository) Create(_ context.Context, token string) (*models.AgentToken, error) {
	if r.current != nil {
		return nil, gorm.ErrDuplicatedKey
	}
	r.current = &models.AgentToken{ID: 1, Token: token}
	copy := *r.current
	return &copy, nil
}

func (r *localAgentTokenRepository) Replace(_ context.Context, token string) (*models.AgentToken, error) {
	r.current = &models.AgentToken{ID: 1, Token: token}
	copy := *r.current
	return &copy, nil
}

func (r *localAgentTokenRepository) Delete(context.Context) error {
	r.current = nil
	return nil
}

func TestAgentTokenCreateGetRenewAndRemove(t *testing.T) {
	repository := &localAgentTokenRepository{}
	values := []string{"first", "renewed", "after-remove"}
	index := 0
	usecase := agentsettings.New(repository, func() (string, error) {
		value := values[index]
		index++
		return value, nil
	})

	_, err := usecase.Get(context.Background())
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
	first, err := usecase.Create(context.Background())
	require.NoError(t, err)
	second, err := usecase.Get(context.Background())
	require.NoError(t, err)
	require.Equal(t, "first", first.Token)
	require.Equal(t, first.Token, second.Token)
	_, err = usecase.Create(context.Background())
	require.ErrorIs(t, err, agentsettings.ErrTokenExists)

	renewed, err := usecase.Renew(context.Background())
	require.NoError(t, err)
	require.Equal(t, "renewed", renewed.Token)
	require.NoError(t, usecase.Remove(context.Background()))
	_, err = usecase.Get(context.Background())
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
	afterRemove, err := usecase.Create(context.Background())
	require.NoError(t, err)
	require.Equal(t, "after-remove", afterRemove.Token)
	require.ErrorIs(t, agentsettings.RequireAgentNode(false), agentsettings.ErrAgentNodeRequired)
	require.NoError(t, agentsettings.RequireAgentNode(true))
}

func TestAgentModeRoutesAreConditional(t *testing.T) {
	for _, test := range []struct {
		name             string
		agentNode        string
		wantMasterAgents bool
	}{
		{name: "master", agentNode: "false", wantMasterAgents: true},
		{name: "agent", agentNode: "true", wantMasterAgents: false},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Setenv("AGENT_NODE", test.agentNode)
			config.Init(false, "", 0)
			service, err := adminapi.New(false, false)
			require.NoError(t, err)
			e := echo.New()
			service.Register(e.Group(""))
			paths := make(map[string]bool)
			for _, route := range e.Router().Routes() {
				paths[route.Method+" "+route.Path] = true
			}
			require.False(t, paths["GET /agent/settings/token"])
			require.Equal(t, test.wantMasterAgents, paths["GET /ocserv/agents"])
		})
	}
}
