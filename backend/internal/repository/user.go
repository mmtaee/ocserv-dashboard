package repository

import (
	"context"
	"errors"
	"github.com/mmtaee/ocserv-dashboard/backend/internal/models"
	"github.com/mmtaee/ocserv-dashboard/backend/internal/platform/database"
	"github.com/mmtaee/ocserv-dashboard/backend/pkg/crypto"
	"github.com/mmtaee/ocserv-dashboard/backend/pkg/request"
	"gorm.io/gorm"
	"time"
)

type UserRepository struct {
	db *gorm.DB
}

type UserCRUD interface {
	GetByUsername(ctx context.Context, username string) (*models.User, error)
	GetByID(ctx context.Context, id uint) (*models.User, error)
	CreateUser(ctx context.Context, user *models.User) (*models.User, error)
	DeleteUser(ctx context.Context, id uint) error
	EnsureSuperadmin(ctx context.Context, user *models.User) (*models.User, error)
}

func (r *UserRepository) EnsureSuperadmin(ctx context.Context, user *models.User) (*models.User, error) {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var existing models.User
		err := tx.Where("username = ?", user.Username).First(&existing).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return tx.Create(user).Error
		}
		if err != nil {
			return err
		}
		if err := tx.Model(&existing).Updates(map[string]interface{}{
			"password": user.Password, "salt": user.Salt, "superadmin": true,
		}).Error; err != nil {
			return err
		}
		*user = existing
		user.Password = ""
		user.Salt = ""
		user.Superadmin = true
		return nil
	})
	return user, err
}

type UserAuth interface {
	CreateToken(ctx context.Context, user *models.User, rememberMe bool) (string, error)
	ChangePassword(ctx context.Context, id uint, password, salt string) error
	UpdateLastLogin(ctx context.Context, user *models.User) error
}

type UserQuery interface {
	Users(ctx context.Context, pagination *request.Pagination) ([]models.User, int64, error)
	UsersLookup(ctx context.Context) (*[]models.UsersLookup, error)
}

type UserRepositoryInterface interface {
	UserCRUD
	UserAuth
	UserQuery
}

func NewUserRepository() *UserRepository {
	return &UserRepository{
		db: database.GetConnection(),
	}
}

func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) CreateToken(ctx context.Context, user *models.User, rememberMe bool) (string, error) {
	expire := time.Now().Add(24 * time.Hour)
	if rememberMe {
		expire = expire.AddDate(0, 1, 0)
	}

	access, err := crypto.GenerateAccessToken(user.ID, user.Username, expire.Unix(), user.Superadmin)
	if err != nil {
		return "", err
	}

	err = r.db.WithContext(ctx).Create(
		&models.UserToken{
			UserID:   user.ID,
			Token:    access,
			ExpireAt: expire,
		},
	).Error
	if err != nil {
		return "", err
	}
	return access, nil
}

func (r *UserRepository) CreateUser(ctx context.Context, user *models.User) (*models.User, error) {
	err := r.db.WithContext(ctx).Create(user).Error
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserRepository) Users(ctx context.Context, pagination *request.Pagination) ([]models.User, int64, error) {
	var totalRecords int64

	whereFilters := "superadmin = false"

	if err := r.db.WithContext(ctx).Model(&models.User{}).Where(whereFilters).Count(&totalRecords).Error; err != nil {
		return nil, 0, err
	}

	var staffs []models.User
	txPaginator := request.Paginator(ctx, r.db, pagination)
	err := txPaginator.Model(&staffs).Where(whereFilters).Find(&staffs).Error
	if err != nil {
		return nil, 0, err
	}
	return staffs, totalRecords, nil
}

func (r *UserRepository) ChangePassword(ctx context.Context, id uint, password, salt string) error {
	var user models.User

	err := r.db.WithContext(ctx).Model(&user).Where("id = ?", id).Updates(
		map[string]interface{}{
			"password": password,
			"salt":     salt,
		},
	).Error
	if err != nil {
		return err
	}

	return nil
}

func (r *UserRepository) DeleteUser(ctx context.Context, id uint) error {
	var user models.User
	err := r.db.WithContext(ctx).Where("id = ? AND superadmin = ?", id, false).First(&user).Error
	if err != nil {
		return err
	}

	err = r.db.WithContext(ctx).Delete(&user).Error
	if err != nil {
		return err
	}
	return nil
}

func (r *UserRepository) GetByID(ctx context.Context, id uint) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) UpdateLastLogin(ctx context.Context, user *models.User) error {
	err := r.db.WithContext(ctx).
		Model(&models.User{}).
		Where("id = ?", user.ID).
		Updates(map[string]interface{}{
			"last_login": user.LastLogin,
		}).Error
	return err
}

func (r *UserRepository) UsersLookup(ctx context.Context) (*[]models.UsersLookup, error) {
	var users []models.UsersLookup
	err := r.db.Model(&models.User{}).WithContext(ctx).Scan(&users).Error
	if err != nil {
		return nil, err
	}
	return &users, nil
}
