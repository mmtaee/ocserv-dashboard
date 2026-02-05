package repository

import (
	"context"
	"fmt"
	"github.com/mmtaee/ocserv-users-management/api/internal/models"
	"github.com/mmtaee/ocserv-users-management/api/pkg/crypto"
	"github.com/mmtaee/ocserv-users-management/api/pkg/request"
	"github.com/mmtaee/ocserv-users-management/common/pkg/database"
	"gorm.io/gorm"
	"time"
)

type UserRepository struct {
	db *gorm.DB
}

type UserCRUD interface {
	GetByUsername(ctx context.Context, username string) (*models.User, error)
	GetByUID(ctx context.Context, uid string) (*models.User, error)
	CreateUser(ctx context.Context, user *models.User) (*models.User, error)
	CreateUserPermission(ctx context.Context, permission []models.Permission) error
	DeleteUser(ctx context.Context, uid string) error
}

type UserAuth interface {
	CreateToken(ctx context.Context, user *models.User, rememberMe bool) (string, error)
	ChangePassword(ctx context.Context, uid, password, salt string) error
	UpdateLastLogin(ctx context.Context, user *models.User) error
}

type UserQuery interface {
	Users(ctx context.Context, pagination *request.Pagination, adminID *uint) ([]models.User, int64, error)
	UsersLookup(ctx context.Context) (*[]models.UsersLookup, error)
}

type UserPermission interface {
	CreateUserPermission(ctx context.Context, permissions []models.Permission) error
	RemoveUserPermission(ctx context.Context, id uint) error
}

type UserRepositoryInterface interface {
	UserCRUD
	UserAuth
	UserQuery
	UserPermission
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

	access, err := crypto.GenerateAccessToken(user, expire.Unix())
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

func (r *UserRepository) Users(ctx context.Context, pagination *request.Pagination, adminID *uint) ([]models.User, int64, error) {
	var totalRecords int64

	roles := []models.UserRole{models.RoleAdmin, models.RoleStaff}
	whereFilters := fmt.Sprintf("role IN %v", roles)
	if adminID != nil {
		whereFilters += fmt.Sprintf(" admin_id = %d", *adminID)
	}

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

func (r *UserRepository) ChangePassword(ctx context.Context, uid, password, salt string) error {
	var user models.User

	err := r.db.WithContext(ctx).Model(&user).Where("uid = ?", uid).Updates(
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

func (r *UserRepository) DeleteUser(ctx context.Context, uid string) error {
	var user models.User
	err := r.db.WithContext(ctx).Where("uid = ? AND is_admin = ?", uid, false).First(&user).Error
	if err != nil {
		return err
	}

	err = r.db.WithContext(ctx).Delete(&user).Error
	if err != nil {
		return err
	}
	return nil
}

func (r *UserRepository) GetByUID(ctx context.Context, uid string) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).Where("uid = ?", uid).First(&user).Error
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
	err := r.db.Model(&models.User{}).WithContext(ctx).Where("role = ?", models.RoleAdmin).Scan(&users).Error
	if err != nil {
		return nil, err
	}
	return &users, nil
}

func (r *UserRepository) RemoveUserPermission(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Where("user_id = ?", id).Delete(&models.Permission{}).Error
}

func (r *UserRepository) CreateUserPermission(ctx context.Context, permissions []models.Permission) error {
	if len(permissions) == 0 {
		return nil
	}

	return r.db.WithContext(ctx).Create(&permissions).Error
}
