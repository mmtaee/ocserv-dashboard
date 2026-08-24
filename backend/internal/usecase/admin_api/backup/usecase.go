package backup

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/mmtaee/ocserv-dashboard/backend/internal/models"
	groupusecase "github.com/mmtaee/ocserv-dashboard/backend/internal/usecase/admin_api/groups"
)

type Usecase struct {
	repository   Repository
	groups       GroupManager
	users        UserManager
	certificates CertificateStore
}

func New(repo Repository, groups GroupManager, users UserManager, certificates CertificateStore) *Usecase {
	return &Usecase{repository: repo, groups: groups, users: users, certificates: certificates}
}

func (u *Usecase) GroupBackup(ctx context.Context, writer io.Writer) error {
	defaults, err := u.groups.DefaultGroup()
	if err != nil {
		return err
	}
	groups, err := u.repository.Groups(ctx)
	if err != nil {
		return err
	}
	return json.NewEncoder(writer).Encode(GroupFile{DefaultGroup: defaults, Groups: groups})
}

func (u *Usecase) RestoreGroups(ctx context.Context, owner string, file GroupFile) (*RestoreResult, error) {
	if owner == "" {
		return nil, errors.New("admin or staff username not found")
	}
	if err := u.groups.UpdateDefaultGroup(file.DefaultGroup); err != nil {
		return nil, err
	}
	result := &RestoreResult{}
	if len(file.Groups) == 0 {
		return result, nil
	}
	names := make([]string, 0, len(file.Groups))
	for _, item := range file.Groups {
		names = append(names, item.Name)
	}
	existing, err := u.repository.ExistingGroupNames(ctx, names)
	if err != nil {
		return nil, err
	}
	existingSet := stringSet(existing)
	inserted := make([]string, 0, len(file.Groups))
	for _, item := range file.Groups {
		if _, found := existingSet[item.Name]; found {
			continue
		}
		if item.Owner == "" {
			item.Owner = owner
		}
		if _, err := u.groups.Create(ctx, item.Owner, groupusecase.CreateInput{Name: item.Name, Config: item.Config}); err != nil {
			return nil, fmt.Errorf("group %s: %w", item.Name, err)
		}
		inserted = append(inserted, item.Name)
	}
	result.Inserted, result.Existing = &inserted, &existing
	return result, nil
}

func (u *Usecase) UserBackup(ctx context.Context, writer io.Writer) error {
	users, err := u.repository.Users(ctx)
	if err != nil {
		return err
	}
	for i := range users {
		certificate, err := u.certificates.CertificateBackup(users[i].Username)
		if err != nil {
			return err
		}
		users[i].Certificate = certificate
	}
	return json.NewEncoder(writer).Encode(users)
}

func (u *Usecase) RestoreUsers(ctx context.Context, owner string, users []models.OcservUser) (*RestoreResult, error) {
	if owner == "" {
		return nil, errors.New("admin or staff username not found")
	}
	if len(users) == 0 {
		return &RestoreResult{}, nil
	}
	usernames := make([]string, 0, len(users))
	for _, account := range users {
		usernames = append(usernames, account.Username)
	}
	existing, err := u.repository.ExistingUsernames(ctx, usernames)
	if err != nil {
		return nil, err
	}
	existingSet := stringSet(existing)
	inserted := make([]string, 0, len(users))
	for _, account := range users {
		if _, found := existingSet[account.Username]; found {
			continue
		}
		if account.Owner == "" {
			account.Owner = owner
		}
		if _, err := u.users.Create(ctx, &account); err != nil {
			return nil, fmt.Errorf("user %s: %w", account.Username, err)
		}
		if account.Certificate != nil {
			if err := u.certificates.RestoreCertificateBackup(account.Username, account.Certificate); err != nil {
				_ = u.users.DeleteUser(ctx, account.ID)
				return nil, fmt.Errorf("user %s certificate: %w", account.Username, err)
			}
		}
		inserted = append(inserted, account.Username)
	}
	return &RestoreResult{Inserted: &inserted, Existing: &existing}, nil
}

func stringSet(values []string) map[string]struct{} {
	set := make(map[string]struct{}, len(values))
	for _, value := range values {
		set[value] = struct{}{}
	}
	return set
}
