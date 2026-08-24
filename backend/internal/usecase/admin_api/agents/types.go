package agents

import (
	"context"

	"github.com/mmtaee/ocserv-dashboard/backend/internal/models"
)

type Repository interface {
	List(ctx context.Context) ([]models.OcservAgent, error)
	GetByID(ctx context.Context, id uint) (*models.OcservAgent, error)
	Create(ctx context.Context, agent *models.OcservAgent) error
	Update(ctx context.Context, agent *models.OcservAgent) error
	Delete(ctx context.Context, id uint) error
}

type CreateInput struct {
	Name        string                  `json:"name" validate:"required,max=128"`
	AddressType models.AgentAddressType `json:"address_type" validate:"required,oneof=ip domain"`
	Address     string                  `json:"address" validate:"required,max=255"`
	Token       string                  `json:"token" validate:"required,max=512"`
}

type UpdateInput = CreateInput
