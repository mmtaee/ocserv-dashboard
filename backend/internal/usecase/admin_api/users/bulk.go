package ocservuser

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/mmtaee/ocserv-dashboard/backend/internal/authz"
	"github.com/mmtaee/ocserv-dashboard/backend/internal/models"
	"github.com/mmtaee/ocserv-dashboard/backend/internal/platform/logging"
	"github.com/mmtaee/ocserv-dashboard/backend/internal/repository"
)

var ErrNoBulkChanges = errors.New("bulk update has no changes")

type BulkAccountStore interface {
	Create(group, username, password string, config *models.OcservUserConfig) error
	Lock(username string) (string, error)
	UnLock(username string) (string, error)
	Delete(username string) (string, error)
	CertificateBackup(username string) (*models.OcservUserCertificateBackup, error)
	RestoreCertificateBackup(username string, certificate *models.OcservUserCertificateBackup) error
}

type BulkRuntime interface {
	Reload() (string, error)
	Disconnect(username string) (string, error)
	Terminate(username string) (string, error)
}

type BulkUsecase struct {
	repository repository.OcservUserBulkRepository
	accounts   BulkAccountStore
	runtime    BulkRuntime
}

func NewBulk(repo repository.OcservUserBulkRepository, accounts BulkAccountStore, runtime BulkRuntime) *BulkUsecase {
	return &BulkUsecase{repository: repo, accounts: accounts, runtime: runtime}
}

func (u *Usecase) BulkUpdate(ctx context.Context, principal authz.Principal, input BulkUpdateRequest) (*BulkUsersResponse, error) {
	return u.bulk.Update(ctx, principal, input)
}

func (u *Usecase) BulkDelete(ctx context.Context, principal authz.Principal, input BulkIDsRequest) (*BulkDeleteResponse, error) {
	return u.bulk.Delete(ctx, principal, input)
}

func (u *Usecase) BulkSetEnabled(ctx context.Context, principal authz.Principal, input BulkStatusRequest) (*BulkUsersResponse, error) {
	return u.bulk.SetEnabled(ctx, principal, input)
}

func (u *Usecase) BulkSetGroup(ctx context.Context, principal authz.Principal, input BulkGroupRequest) (*BulkUsersResponse, error) {
	return u.bulk.SetGroup(ctx, principal, input)
}

func (u *BulkUsecase) Update(ctx context.Context, principal authz.Principal, input BulkUpdateRequest) (*BulkUsersResponse, error) {
	ids := make([]uint, 0, len(input.Users))
	changes := make(map[uint]UpdateOcservUserData, len(input.Users))
	for _, item := range input.Users {
		if !hasUserUpdates(item.Changes) {
			return nil, fmt.Errorf("user id %d: %w", item.ID, ErrNoBulkChanges)
		}
		ids = append(ids, item.ID)
		changes[item.ID] = item.Changes
	}
	if err := validateBulkIDs(ids); err != nil {
		return nil, err
	}

	var updated []models.OcservUser
	undo := make([]func(), 0, len(ids))
	err := u.repository.WithTransaction(ctx, func(tx repository.OcservUserBulkTx) error {
		users, err := loadAuthorizedBulkUsers(ctx, tx, principal, ids)
		if err != nil {
			return err
		}
		byID := usersByID(users)
		updated = make([]models.OcservUser, 0, len(ids))
		for _, id := range ids {
			current := byID[id]
			previous := *current
			if err := applyUserUpdate(current, changes[id]); err != nil {
				return fmt.Errorf("update ocserv user %d: %w", id, err)
			}
			if _, err := tx.Update(ctx, current); err != nil {
				return err
			}
			if err := u.accounts.Create(current.Group, current.Username, current.Password, current.Config); err != nil {
				u.restoreAccount(previous)
				return fmt.Errorf("update ocserv user %d: %w", id, err)
			}
			undo = append(undo, func() { u.restoreAccount(previous) })
			updated = append(updated, *current)
		}
		return nil
	})
	if err != nil {
		u.rollbackExternal(undo)
		return nil, err
	}
	u.reload()
	return &BulkUsersResponse{Count: len(updated), Users: updated}, nil
}

func (u *BulkUsecase) Delete(ctx context.Context, principal authz.Principal, input BulkIDsRequest) (*BulkDeleteResponse, error) {
	if err := validateBulkIDs(input.IDs); err != nil {
		return nil, err
	}

	undo := make([]func(), 0, len(input.IDs))
	deleted := make([]models.OcservUser, 0, len(input.IDs))
	err := u.repository.WithTransaction(ctx, func(tx repository.OcservUserBulkTx) error {
		users, err := loadAuthorizedBulkUsers(ctx, tx, principal, input.IDs)
		if err != nil {
			return err
		}
		byID := usersByID(users)
		for _, id := range input.IDs {
			account := *byID[id]
			certificate, err := u.accounts.CertificateBackup(account.Username)
			if err != nil {
				return fmt.Errorf("backup certificate for user %d: %w", id, err)
			}
			if _, err := u.accounts.Delete(account.Username); err != nil {
				u.restoreDeletedAccount(account, certificate)
				return fmt.Errorf("delete ocserv user %d: %w", id, err)
			}
			undo = append(undo, func() { u.restoreDeletedAccount(account, certificate) })
			if _, err := tx.Delete(ctx, id); err != nil {
				return err
			}
			deleted = append(deleted, account)
		}
		return nil
	})
	if err != nil {
		u.rollbackExternal(undo)
		return nil, err
	}
	u.reload()
	for _, account := range deleted {
		if _, err := u.runtime.Terminate(account.Username); err != nil {
			logger.Warn("failed to terminate deleted ocserv user %s: %v", account.Username, err)
		}
	}
	return &BulkDeleteResponse{Count: len(deleted)}, nil
}

func (u *BulkUsecase) SetEnabled(ctx context.Context, principal authz.Principal, input BulkStatusRequest) (*BulkUsersResponse, error) {
	if input.Enabled == nil {
		return nil, errors.New("enabled is required")
	}
	if err := validateBulkIDs(input.IDs); err != nil {
		return nil, err
	}

	undo := make([]func(), 0, len(input.IDs))
	var updated []models.OcservUser
	err := u.repository.WithTransaction(ctx, func(tx repository.OcservUserBulkTx) error {
		users, err := loadAuthorizedBulkUsers(ctx, tx, principal, input.IDs)
		if err != nil {
			return err
		}
		byID := usersByID(users)
		updated = make([]models.OcservUser, 0, len(input.IDs))
		for _, id := range input.IDs {
			account := byID[id]
			locked := !*input.Enabled
			if account.IsLocked == locked {
				updated = append(updated, *account)
				continue
			}
			if err := u.setExternalLocked(account.Username, locked); err != nil {
				_ = u.setExternalLocked(account.Username, !locked)
				return fmt.Errorf("set ocserv user %d enabled=%t: %w", id, *input.Enabled, err)
			}
			undo = append(undo, func() { _ = u.setExternalLocked(account.Username, !locked) })
			if locked {
				err = tx.Lock(ctx, id)
			} else {
				err = tx.UnLock(ctx, id)
			}
			if err != nil {
				return err
			}
			account.IsLocked = locked
			updated = append(updated, *account)
		}
		return nil
	})
	if err != nil {
		u.rollbackExternal(undo)
		return nil, err
	}
	if !*input.Enabled {
		for _, account := range updated {
			if !account.IsLocked {
				continue
			}
			if _, err := u.runtime.Disconnect(account.Username); err != nil {
				logger.Warn("failed to disconnect disabled ocserv user %s: %v", account.Username, err)
			}
		}
	}
	return &BulkUsersResponse{Count: len(updated), Users: updated}, nil
}

func (u *BulkUsecase) SetGroup(ctx context.Context, principal authz.Principal, input BulkGroupRequest) (*BulkUsersResponse, error) {
	if err := validateBulkIDs(input.IDs); err != nil {
		return nil, err
	}
	group := strings.TrimSpace(input.Group)
	if group == "" {
		group = "defaults"
	}

	undo := make([]func(), 0, len(input.IDs))
	var updated []models.OcservUser
	err := u.repository.WithTransaction(ctx, func(tx repository.OcservUserBulkTx) error {
		users, err := loadAuthorizedBulkUsers(ctx, tx, principal, input.IDs)
		if err != nil {
			return err
		}
		byID := usersByID(users)
		updated = make([]models.OcservUser, 0, len(input.IDs))
		for _, id := range input.IDs {
			account := byID[id]
			previous := *account
			account.Group = group
			if _, err := tx.Update(ctx, account); err != nil {
				return err
			}
			if err := u.accounts.Create(account.Group, account.Username, account.Password, account.Config); err != nil {
				u.restoreAccount(previous)
				return fmt.Errorf("assign group to ocserv user %d: %w", id, err)
			}
			undo = append(undo, func() { u.restoreAccount(previous) })
			updated = append(updated, *account)
		}
		return nil
	})
	if err != nil {
		u.rollbackExternal(undo)
		return nil, err
	}
	u.reload()
	return &BulkUsersResponse{Count: len(updated), Users: updated}, nil
}

func loadAuthorizedBulkUsers(ctx context.Context, tx repository.OcservUserBulkTx, principal authz.Principal, ids []uint) ([]models.OcservUser, error) {
	users, err := tx.GetByIDsForUpdate(ctx, ids)
	if err != nil {
		return nil, err
	}
	found := make(map[uint]struct{}, len(users))
	for i := range users {
		found[users[i].ID] = struct{}{}
		if err := principal.RequireOwner(users[i].OwnerID); err != nil {
			return nil, fmt.Errorf("ocserv user id %d: %w", users[i].ID, err)
		}
	}
	missing := make([]uint, 0)
	for _, id := range ids {
		if _, ok := found[id]; !ok {
			missing = append(missing, id)
		}
	}
	if len(missing) > 0 {
		return nil, fmt.Errorf("ocserv users not found: %v", missing)
	}
	return users, nil
}

func validateBulkIDs(ids []uint) error {
	if len(ids) == 0 {
		return errors.New("at least one user id is required")
	}
	seen := make(map[uint]struct{}, len(ids))
	for _, id := range ids {
		if id == 0 {
			return errors.New("user ids must be greater than zero")
		}
		if _, exists := seen[id]; exists {
			return fmt.Errorf("duplicate user id: %d", id)
		}
		seen[id] = struct{}{}
	}
	return nil
}

func usersByID(users []models.OcservUser) map[uint]*models.OcservUser {
	result := make(map[uint]*models.OcservUser, len(users))
	for i := range users {
		result[users[i].ID] = &users[i]
	}
	return result
}

func hasUserUpdates(input UpdateOcservUserData) bool {
	return input.Group != nil || input.Password != nil || input.ExpireAt != nil || input.ExpiryMode != nil ||
		input.ExpireDaysAfterFirstConnection != nil || input.ResetFirstConnection || input.Unlimited ||
		input.TrafficType != nil || input.TrafficSize != nil || input.Description != nil || input.Config != nil
}

func (u *BulkUsecase) setExternalLocked(username string, locked bool) error {
	if locked {
		_, err := u.accounts.Lock(username)
		return err
	}
	_, err := u.accounts.UnLock(username)
	return err
}

func (u *BulkUsecase) restoreAccount(account models.OcservUser) {
	if err := u.accounts.Create(account.Group, account.Username, account.Password, account.Config); err != nil {
		logger.Error("failed to rollback ocserv user %s: %v", account.Username, err)
		return
	}
	if account.IsLocked {
		if _, err := u.accounts.Lock(account.Username); err != nil {
			logger.Error("failed to restore lock for ocserv user %s: %v", account.Username, err)
		}
	}
}

func (u *BulkUsecase) restoreDeletedAccount(account models.OcservUser, certificate *models.OcservUserCertificateBackup) {
	u.restoreAccount(account)
	if err := u.accounts.RestoreCertificateBackup(account.Username, certificate); err != nil {
		logger.Error("failed to restore certificate for ocserv user %s: %v", account.Username, err)
	}
}

func (u *BulkUsecase) rollbackExternal(actions []func()) {
	for index := len(actions) - 1; index >= 0; index-- {
		actions[index]()
	}
	u.reload()
}

func (u *BulkUsecase) reload() {
	if _, err := u.runtime.Reload(); err != nil {
		logger.Warn("failed to reload ocserv after bulk operation: %v", err)
	}
}
