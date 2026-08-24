package agents

import (
	"context"
	"errors"
	"net"
	"strings"

	"github.com/mmtaee/ocserv-dashboard/backend/internal/models"
)

var ErrInvalidAddress = errors.New("invalid agent address")

type Usecase struct {
	repository Repository
}

func New(repository Repository) *Usecase {
	return &Usecase{repository: repository}
}

func (u *Usecase) List(ctx context.Context) ([]models.OcservAgent, error) {
	return u.repository.List(ctx)
}

func (u *Usecase) Get(ctx context.Context, id uint) (*models.OcservAgent, error) {
	return u.repository.GetByID(ctx, id)
}

func (u *Usecase) Create(ctx context.Context, input CreateInput) (*models.OcservAgent, error) {
	agent, err := buildAgent(input)
	if err != nil {
		return nil, err
	}
	if err := u.repository.Create(ctx, agent); err != nil {
		return nil, err
	}
	return agent, nil
}

func (u *Usecase) Update(ctx context.Context, id uint, input UpdateInput) (*models.OcservAgent, error) {
	agent, err := buildAgent(input)
	if err != nil {
		return nil, err
	}
	stored, err := u.repository.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	stored.Name, stored.AddressType, stored.Address, stored.Token = agent.Name, agent.AddressType, agent.Address, agent.Token
	if err := u.repository.Update(ctx, stored); err != nil {
		return nil, err
	}
	return stored, nil
}

func (u *Usecase) Delete(ctx context.Context, id uint) error {
	return u.repository.Delete(ctx, id)
}

func buildAgent(input CreateInput) (*models.OcservAgent, error) {
	name := strings.TrimSpace(input.Name)
	address := strings.TrimSpace(input.Address)
	token := strings.TrimSpace(input.Token)
	if name == "" || token == "" || !validAddress(input.AddressType, address) {
		return nil, ErrInvalidAddress
	}
	if input.AddressType == models.AgentAddressTypeIP {
		address = net.ParseIP(address).String()
	} else {
		address = strings.ToLower(strings.TrimSuffix(address, "."))
	}
	return &models.OcservAgent{Name: name, AddressType: input.AddressType, Address: address, Token: token}, nil
}

func validAddress(addressType models.AgentAddressType, address string) bool {
	switch addressType {
	case models.AgentAddressTypeIP:
		return net.ParseIP(address) != nil
	case models.AgentAddressTypeDomain:
		return validDomain(address)
	default:
		return false
	}
}

func validDomain(value string) bool {
	value = strings.TrimSuffix(value, ".")
	if len(value) == 0 || len(value) > 253 || net.ParseIP(value) != nil {
		return false
	}
	labels := strings.Split(value, ".")
	if len(labels) < 2 {
		return false
	}
	for _, label := range labels {
		if len(label) == 0 || len(label) > 63 || label[0] == '-' || label[len(label)-1] == '-' {
			return false
		}
		for _, char := range label {
			if (char < 'a' || char > 'z') && (char < 'A' || char > 'Z') && (char < '0' || char > '9') && char != '-' {
				return false
			}
		}
	}
	return true
}
