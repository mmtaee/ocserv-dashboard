package agent_settings

import (
	"context"
	"errors"

	"github.com/mmtaee/ocserv-dashboard/backend/internal/models"
	"github.com/mmtaee/ocserv-dashboard/backend/pkg/crypto"
	"gorm.io/gorm"
)

var (
	ErrAgentNodeRequired = errors.New("agent token commands are available only when AGENT_NODE=true")
	ErrTokenExists       = errors.New("agent token already exists; use agent-token renew to replace it")
)

type Usecase struct {
	repository Repository
	generate   TokenGenerator
}

func New(repository Repository, generators ...TokenGenerator) *Usecase {
	generate := TokenGenerator(crypto.GenerateSecureToken)
	if len(generators) > 0 && generators[0] != nil {
		generate = generators[0]
	}
	return &Usecase{repository: repository, generate: generate}
}

func (u *Usecase) Get(ctx context.Context) (*models.AgentToken, error) {
	return u.repository.Get(ctx)
}

func (u *Usecase) Create(ctx context.Context) (*models.AgentToken, error) {
	if _, err := u.repository.Get(ctx); err == nil {
		return nil, ErrTokenExists
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	token, err := u.generate()
	if err != nil {
		return nil, err
	}
	created, err := u.repository.Create(ctx, token)
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return nil, ErrTokenExists
	}
	return created, err
}

func (u *Usecase) Renew(ctx context.Context) (*models.AgentToken, error) {
	token, err := u.generate()
	if err != nil {
		return nil, err
	}
	return u.repository.Replace(ctx, token)
}

func (u *Usecase) Remove(ctx context.Context) error {
	return u.repository.Delete(ctx)
}

func RequireAgentNode(agentNode bool) error {
	if !agentNode {
		return ErrAgentNodeRequired
	}
	return nil
}
