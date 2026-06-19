package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/mmtaee/ocserv-dashboard/core/models"
	"github.com/mmtaee/ocserv-dashboard/core/pkg/ocserv/group"
	"github.com/mmtaee/ocserv-dashboard/core/pkg/ocserv/user"
	"gorm.io/gorm"
	"io"
	"strings"
	"sync"
)

type BackupRepository interface {
	OcservGroupBackup(ctx context.Context, writer io.Writer, defaultGroup *models.OcservGroupConfig) error
	OcservGroupRestore(ctx context.Context, ownerAdminID uint, groups *[]models.OcservGroup) (*[]string, *[]string, error)
	OcservUserBackup(ctx context.Context, writer io.Writer) error
	OcservUserRestore(ctx context.Context, ownerAdminID uint, users *[]models.OcservUser) (*[]string, *[]string, error)
}

type backupRepository struct {
	db                    *gorm.DB
	commonOcservGroupRepo group.OcservGroupInterface
	commonOcservUserRepo  user.OcservUserInterface
}

func NewBackupRepository(db *gorm.DB) BackupRepository {
	return &backupRepository{
		db:                    db,
		commonOcservGroupRepo: group.NewOcservGroup(),
		commonOcservUserRepo:  user.NewOcservUser(),
	}
}

func (b *backupRepository) OcservGroupBackup(ctx context.Context, writer io.Writer, defaultGroup *models.OcservGroupConfig) error {
	if _, err := writer.Write([]byte("{")); err != nil {
		return err
	}

	if _, err := writer.Write([]byte(`"default_group":`)); err != nil {
		return err
	}

	defaultBytes, err := json.Marshal(defaultGroup)
	if err != nil {
		return err
	}

	if _, err = writer.Write(defaultBytes); err != nil {
		return err
	}

	if _, err = writer.Write([]byte(`,"groups":[`)); err != nil {
		return err
	}

	rows, err := b.db.WithContext(ctx).
		Model(&models.OcservGroup{}).
		Rows()
	if err != nil {
		return err
	}
	defer rows.Close()

	first := true

	for rows.Next() {
		var g models.OcservGroup

		if err = b.db.ScanRows(rows, &g); err != nil {
			return err
		}

		if !first {
			if _, err = writer.Write([]byte(",")); err != nil {
				return err
			}
		}
		first = false

		groupBytes, err := json.Marshal(g)
		if err != nil {
			return err
		}

		if _, err = writer.Write(groupBytes); err != nil {
			return err
		}
	}

	if _, err := writer.Write([]byte("]}")); err != nil {
		return err
	}

	return nil
}

func (b *backupRepository) OcservGroupRestore(ctx context.Context, ownerAdminID uint, groups *[]models.OcservGroup) (*[]string, *[]string, error) {
	names := make([]string, 0, len(*groups))
	for _, g := range *groups {
		names = append(names, g.Name)
	}

	var dbExisting []string

	err := b.db.WithContext(ctx).
		Model(&models.OcservGroup{}).
		Select("name").
		Where("name IN ?", names).
		Scan(&dbExisting).Error
	if err != nil {
		return nil, nil, err
	}

	existingMap := make(map[string]struct{}, len(dbExisting))
	for _, name := range dbExisting {
		existingMap[name] = struct{}{}
	}

	var toInsert []models.OcservGroup
	var insertedNames []string

	for _, g := range *groups {
		if _, found := existingMap[g.Name]; !found {
			toInsert = append(toInsert, g)
			insertedNames = append(insertedNames, g.Name)
		}
	}

	if len(toInsert) == 0 {
		return &insertedNames, &dbExisting, nil
	}

	var wg sync.WaitGroup
	errCh := make(chan error, len(toInsert))
	sem := make(chan struct{}, 10)

	for _, g := range toInsert {
		wg.Add(1)

		go func(g models.OcservGroup) {
			defer wg.Done()

			sem <- struct{}{}
			defer func() { <-sem }()

			if g.OwnerAdminID == 0 {
				g.OwnerAdminID = ownerAdminID
			}

			txErr := b.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
				res := tx.Create(&g)
				if res.Error != nil {
					return res.Error
				}
				if res.RowsAffected == 0 {
					return nil
				}

				if err = b.commonOcservGroupRepo.Create(g.Name, g.Config); err != nil {
					return err
				}

				return nil
			})

			if txErr != nil {
				errCh <- fmt.Errorf("group %s: %w", g.Name, txErr)
			}
		}(g)
	}

	wg.Wait()
	close(errCh)

	var errs []string
	for e := range errCh {
		errs = append(errs, e.Error())
	}

	if len(errs) > 0 {
		return &insertedNames, &dbExisting, fmt.Errorf("%s", strings.Join(errs, "; "))
	}

	return &insertedNames, &dbExisting, nil
}

func (b *backupRepository) OcservUserBackup(ctx context.Context, writer io.Writer) error {
	rows, err := b.db.WithContext(ctx).
		Model(&models.OcservUser{}).
		Rows()
	if err != nil {
		return err
	}
	defer rows.Close()

	if _, err = writer.Write([]byte("[")); err != nil {
		return err
	}

	first := true

	for rows.Next() {
		var u models.OcservUser

		if err = b.db.ScanRows(rows, &u); err != nil {
			return err
		}

		cert, certErr := b.commonOcservUserRepo.CertificateBackup(u.Username)
		if certErr != nil {
			return certErr
		}
		u.Certificate = cert

		if !first {
			if _, err = writer.Write([]byte(",")); err != nil {
				return err
			}
		}
		first = false

		userBytes, err := json.Marshal(u)
		if err != nil {
			return err
		}

		if _, err = writer.Write(userBytes); err != nil {
			return err
		}
	}

	if _, err = writer.Write([]byte("]")); err != nil {
		return err
	}

	return nil
}

func (b *backupRepository) OcservUserRestore(ctx context.Context, ownerAdminID uint, users *[]models.OcservUser) (*[]string, *[]string, error) {
	usernames := make([]string, 0, len(*users))
	for _, u := range *users {
		usernames = append(usernames, u.Username)
	}

	var dbExisting []string

	err := b.db.WithContext(ctx).
		Model(&models.OcservUser{}).
		Select("username").
		Where("username IN ?", usernames).
		Scan(&dbExisting).Error
	if err != nil {
		return nil, nil, err
	}

	existingMap := make(map[string]struct{}, len(dbExisting))
	for _, name := range dbExisting {
		existingMap[name] = struct{}{}
	}

	var toInsert []models.OcservUser
	var insertedNames []string

	for _, u := range *users {
		if _, found := existingMap[u.Username]; !found {
			toInsert = append(toInsert, u)
			insertedNames = append(insertedNames, u.Username)
		}
	}

	if len(toInsert) == 0 {
		return &insertedNames, &dbExisting, nil
	}

	var wg sync.WaitGroup
	errCh := make(chan error, len(toInsert))
	sem := make(chan struct{}, 10)

	for _, u := range toInsert {
		wg.Add(1)

		go func(u models.OcservUser) {
			defer wg.Done()

			sem <- struct{}{}
			defer func() { <-sem }()

			if u.OwnerAdminID == 0 {
				u.OwnerAdminID = ownerAdminID
			}

			txErr := b.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
				res := tx.Create(&u)
				if res.Error != nil {
					return res.Error
				}
				if res.RowsAffected == 0 {
					return nil
				}

				if err = b.commonOcservUserRepo.Create(u.Group, u.Username, u.Password, u.Config); err != nil {
					return err
				}

				if u.Certificate != nil {
					if err = b.commonOcservUserRepo.RestoreCertificateBackup(u.Username, u.Certificate); err != nil {
						_, _ = b.commonOcservUserRepo.Delete(u.Username)
						return err
					}
				}

				return nil
			})

			if txErr != nil {
				errCh <- fmt.Errorf("user %s: %w", u.Username, txErr)
			}
		}(u)
	}

	wg.Wait()
	close(errCh)

	var errs []string
	for e := range errCh {
		errs = append(errs, e.Error())
	}

	if len(errs) > 0 {
		return &insertedNames, &dbExisting, fmt.Errorf("%s", strings.Join(errs, "; "))
	}

	return &insertedNames, &dbExisting, nil
}
