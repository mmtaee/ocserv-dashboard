package usecase

import (
	"errors"

	"github.com/mmtaee/ocserv-dashboard/core/models"
	"github.com/mmtaee/ocserv-dashboard/dashboard/api/internal/repository"
	"github.com/mmtaee/ocserv-dashboard/core/pkg/logger"
)

type OcservGroupUseCase interface {
	ListGroups(adminID uint, role string) ([]models.OcservGroupResponse, error)
	GetGroup(id uint, adminID uint, role string) (*models.OcservGroupResponse, error)
	CreateGroup(name string, config *models.OcservGroupConfig, ownerAdminID uint, role string) (*models.OcservGroup, error)
	UpdateGroup(id uint, config *models.OcservGroupConfig, adminID uint, role string) (*models.OcservGroup, error)
	DeleteGroup(id uint, adminID uint, role string) error
	GetGroupsLookup(adminID uint, role string) ([]string, error)
}

type ocservGroupUseCase struct {
	groupRepo repository.OcservGroupRepository
	userRepo  repository.OcservUserRepository
}

func NewOcservGroupUseCase(groupRepo repository.OcservGroupRepository, userRepo repository.OcservUserRepository) OcservGroupUseCase {
	return &ocservGroupUseCase{groupRepo: groupRepo, userRepo: userRepo}
}

func (uc *ocservGroupUseCase) ListGroups(adminID uint, role string) ([]models.OcservGroupResponse, error) {
	return uc.groupRepo.FindAll(adminID, role)
}

func (uc *ocservGroupUseCase) GetGroup(id uint, adminID uint, role string) (*models.OcservGroupResponse, error) {
	group, err := uc.groupRepo.FindByID(id, adminID, role)
	if err != nil {
		return nil, errors.New("6001")
	}
	return group, nil
}

func (uc *ocservGroupUseCase) CreateGroup(name string, config *models.OcservGroupConfig, ownerAdminID uint, role string) (*models.OcservGroup, error) {
	_, err := uc.groupRepo.FindByName(name, ownerAdminID, role)
	if err == nil {
		return nil, errors.New("6002")
	}

	group := &models.OcservGroup{
		Name:         name,
		OwnerAdminID: ownerAdminID,
		Config:       config,
	}

	if err := uc.groupRepo.Create(group); err != nil {
		return nil, errors.New("5001")
	}

	return group, nil
}

func (uc *ocservGroupUseCase) UpdateGroup(id uint, config *models.OcservGroupConfig, adminID uint, role string) (*models.OcservGroup, error) {
	groupResp, err := uc.groupRepo.FindByID(id, adminID, role)
	if err != nil {
		return nil, errors.New("6001")
	}

	groupResp.Config = config

	if err := uc.groupRepo.Update(groupResp.OcservGroup); err != nil {
		return nil, errors.New("5001")
	}

	return groupResp.OcservGroup, nil
}

func (uc *ocservGroupUseCase) DeleteGroup(id uint, adminID uint, role string) error {
	groupResp, err := uc.groupRepo.FindByID(id, adminID, role)
	if err != nil {
		return errors.New("6001")
	}

	if err := uc.groupRepo.Delete(id, adminID, role); err != nil {
		return errors.New("5001")
	}

	// Update users in deleted group to "defaults" in goroutine
	go func(groupName string, groupOwnerID uint) {
		_, err := uc.userRepo.UpdateUsersByDeleteGroup(groupOwnerID, groupName)
		if err != nil {
			logger.Error("Failed to update users for removed group %s: %v", groupName, err)
			return
		}
	}(groupResp.Name, groupResp.OwnerAdminID)

	return nil
}

func (uc *ocservGroupUseCase) GetGroupsLookup(adminID uint, role string) ([]string, error) {
	names, err := uc.groupRepo.GetGroupsLookup(adminID, role)
	if err != nil {
		return nil, errors.New("5001")
	}
	// Prepend "defaults"
	names = append([]string{"defaults"}, names...)
	return names, nil
}
