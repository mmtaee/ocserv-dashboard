package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/mmtaee/ocserv-dashboard/core/models"
	"github.com/mmtaee/ocserv-dashboard/core/ocserv/group"
	"github.com/mmtaee/ocserv-dashboard/core/ocserv/user"
	"github.com/mmtaee/ocserv-dashboard/core/pkg/database"
	"gorm.io/gorm"
)

type BackupRepository struct {
	db                    *gorm.DB
	commonOcservGroupRepo group.OcservGroupInterface
	commonOcservUserRepo  user.OcservUserInterface
}

type BackupRepositoryInterface interface {
	OcservGroupBackup(ctx context.Context, writer io.Writer, defaultGroup *models.OcservGroupConfig) error
	OcservGroupRestore(ctx context.Context, owner string, users *[]models.OcservGroup) ([]string, []string, error)
	OcservUserBackup(ctx context.Context, writer io.Writer) error
	OcservUserRestore(ctx context.Context, owner string, users *[]models.OcservUser) ([]string, []string, error)
}

func NewBackupRepository() *BackupRepository {
	return &BackupRepository{
		db:                    database.GetConnection(),
		commonOcservGroupRepo: group.NewOcservGroup(),
		commonOcservUserRepo:  user.NewOcservUser(),
	}
}

func (b *BackupRepository) OcservGroupBackup(ctx context.Context, writer io.Writer, defaultGroup *models.OcservGroupConfig) error {
	// Start root object
	if _, err := writer.Write([]byte("{")); err != nil {
		return err
	}

	// Write default_group
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

	// Start groups array
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
		var group models.OcservGroup

		if err = b.db.ScanRows(rows, &group); err != nil {
			return err
		}

		if !first {
			if _, err = writer.Write([]byte(",")); err != nil {
				return err
			}
		}
		first = false

		groupBytes, err := json.Marshal(group)
		if err != nil {
			return err
		}

		if _, err = writer.Write(groupBytes); err != nil {
			return err
		}
	}

	// Close array + object
	if _, err := writer.Write([]byte("]}")); err != nil {
		return err
	}

	return nil
}

func (b *BackupRepository) OcservGroupRestore(ctx context.Context, owner string, groups *[]models.OcservGroup) ([]string, []string, error) {
	names := make([]string, 0, len(*groups))
	for _, u := range *groups {
		names = append(names, u.Name)
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

	for _, u := range *groups {
		if _, found := existingMap[u.Name]; !found {
			toInsert = append(toInsert, u)
			insertedNames = append(insertedNames, u.Name)
		}
	}

	if len(toInsert) == 0 {
		return insertedNames, dbExisting, nil
	}

	var wg sync.WaitGroup
	errCh := make(chan error, len(toInsert))
	sem := make(chan struct{}, 10) // limit concurrency

	for _, g := range toInsert {
		wg.Add(1)

		go func(g models.OcservGroup) {
			defer wg.Done()

			sem <- struct{}{}
			defer func() { <-sem }()

			if g.Owner == "" {
				g.Owner = owner
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
		return insertedNames, dbExisting, fmt.Errorf("%s", strings.Join(errs, "; "))
	}

	return insertedNames, dbExisting, nil
}

func (b *BackupRepository) OcservUserBackup(ctx context.Context, writer io.Writer) error {
	rows, err := b.db.WithContext(ctx).
		Model(&models.OcservUser{}).
		Rows()
	if err != nil {
		return err
	}
	defer rows.Close()

	// Start JSON array
	if _, err = writer.Write([]byte("[")); err != nil {
		return err
	}

	first := true

	for rows.Next() {
		var user models.OcservUser

		if err = b.db.ScanRows(rows, &user); err != nil {
			return err
		}

		cert, certErr := b.commonOcservUserRepo.CertificateBackup(user.Username)
		if certErr != nil {
			return certErr
		}
		user.Certificate = cert

		if !first {
			if _, err = writer.Write([]byte(",")); err != nil {
				return err
			}
		}
		first = false

		userBytes, err := json.Marshal(user)
		if err != nil {
			return err
		}

		if _, err = writer.Write(userBytes); err != nil {
			return err
		}
	}

	// Close JSON array
	if _, err = writer.Write([]byte("]")); err != nil {
		return err
	}

	return nil
}

func (b *BackupRepository) OcservUserRestore(ctx context.Context, owner string, users *[]models.OcservUser) ([]string, []string, error) {
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
		return insertedNames, dbExisting, nil
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

			if u.Owner == "" {
				u.Owner = owner
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
		return insertedNames, dbExisting, fmt.Errorf("%s", strings.Join(errs, "; "))
	}

	return insertedNames, dbExisting, nil
}
