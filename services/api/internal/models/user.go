package models

import (
	"github.com/oklog/ulid/v2"
	"gorm.io/gorm"
	"time"
)

type User struct {
	ID        uint        `json:"-" gorm:"primaryKey;autoIncrement" validate:"required"`
	UID       string      `json:"uid" gorm:"type:varchar(26);not null;unique" validate:"required"`
	Username  string      `json:"username" gorm:"type:varchar(16);not null;unique"  validate:"required"`
	Password  string      `json:"-" gorm:"type:varchar(64); not null"`
	IsAdmin   bool        `json:"is_admin" gorm:"type:bool;default(false)"  validate:"required"`
	Salt      string      `json:"-" gorm:"type:varchar(8);not null"`
	LastLogin *time.Time  `json:"last_login"  validate:"required"`
	CreatedAt time.Time   `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time   `json:"updated_at" gorm:"autoUpdateTime"`
	Token     []UserToken `json:"-"`
}

type UserToken struct {
	ID        uint      `json:"-" gorm:"primaryKey;autoIncrement"`
	UserID    uint      `json:"-" gorm:"index"`
	UID       string    `json:"uid" gorm:"type:varchar(26);not null;unique"`
	Token     string    `json:"token" gorm:"type:varchar(128)"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	ExpireAt  time.Time `json:"expire_at"`
	User      User      `json:"user"`
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
