package agent_settings

import (
	"context"

	"github.com/mmtaee/ocserv-dashboard/backend/internal/models"
	"github.com/mmtaee/ocserv-dashboard/backend/pkg/crypto"
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
	token, err := u.generate()
	if err != nil {
		return nil, err
	}
	return u.repository.GetOrCreate(ctx, token)
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
