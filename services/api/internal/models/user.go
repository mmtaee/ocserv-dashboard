package models

import (
	"github.com/oklog/ulid/v2"
	"gorm.io/gorm"
	"time"
)

type UserRole string

const (
	RoleSuperAdmin UserRole = "super-admin"
	RoleAdmin      UserRole = "admin"
	RoleStaff      UserRole = "staff"
)

type User struct {
	ID       uint   `json:"-" gorm:"primaryKey;autoIncrement" validate:"required"`
	UID      string `json:"uid" gorm:"type:varchar(26);not null;uniqueIndex" validate:"required"`
	Username string `json:"username" gorm:"type:varchar(16);not null;uniqueIndex"  validate:"required"`
	Password string `json:"-" gorm:"type:varchar(64); not null"`

	Role UserRole `gorm:"type:varchar(16);not null;index"`
	
	// Hierarchy
	AdminID *uint  `json:"admin_id" gorm:"index"` // NULL for superadmin
	Admin   *User  `json:"-" gorm:"foreignKey:AdminID"`
	Staff   []User `json:"staff" gorm:"foreignKey:AdminID"`

	Salt      string      `json:"-" gorm:"type:varchar(8);not null"`
	LastLogin *time.Time  `json:"last_login"  validate:"required"`
	CreatedAt time.Time   `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time   `json:"updated_at" gorm:"autoUpdateTime"`
	Token     []UserToken `json:"-"`
}

type UserToken struct {
	ID        uint      `json:"-" gorm:"primaryKey;autoIncrement"`
	UserID    uint      `json:"-" gorm:"index"`
	UID       string    `json:"uid" gorm:"type:varchar(26);not null;uniqueIndex"`
	Token     string    `json:"token" gorm:"type:varchar(128)"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	ExpireAt  time.Time `json:"expire_at"`
	User      User      `json:"user"`
}

type Permission struct {
	ID      uint   `gorm:"primaryKey;autoIncrement"`
	UserID  uint   `gorm:"not null;index"`                  // staff user
	Service string `gorm:"type:varchar(64);not null;index"` // e.g., "user_management", "ocserv_group"
	Action  string `gorm:"type:varchar(1);not null"`        // "C", "U", "D"

	CreatedAt time.Time
	UpdatedAt time.Time

	User User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}

type UsersLookup struct {
	UID      string `json:"uid" validate:"required"`
	Username string `json:"username" validate:"required"`
}

func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	u.UID = ulid.Make().String()
	return
}

func (t *UserToken) BeforeCreate(tx *gorm.DB) (err error) {
	if t.UID == "" {
		t.UID = ulid.Make().String()
	}
	return
}
