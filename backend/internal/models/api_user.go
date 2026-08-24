package models

import (
	"time"
)

type User struct {
	ID         uint        `json:"id" gorm:"primaryKey;autoIncrement" validate:"required"`
	Username   string      `json:"username" gorm:"type:varchar(16);not null;uniqueIndex"  validate:"required"`
	Password   string      `json:"-" gorm:"type:varchar(64); not null"`
	Superadmin bool        `json:"superadmin" gorm:"type:bool;default(false)" validate:"required"`
	Salt       string      `json:"-" gorm:"type:varchar(8);not null"`
	LastLogin  *time.Time  `json:"last_login"  validate:"required"`
	CreatedAt  time.Time   `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt  time.Time   `json:"updated_at" gorm:"autoUpdateTime"`
	Token      []UserToken `json:"-"`
}

type UserToken struct {
	ID        uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID    uint      `json:"-" gorm:"index"`
	Token     string    `json:"token" gorm:"type:text"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	ExpireAt  time.Time `json:"expire_at"`
	User      User      `json:"user"`
}

type UsersLookup struct {
	ID       uint   `json:"id" validate:"required"`
	Username string `json:"username" validate:"required"`
}
