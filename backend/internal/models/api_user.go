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
	Tokens     []UserToken `json:"-"`
}

type UserToken struct {
	ID        uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID    uint      `json:"user_id" gorm:"not null;index"`
	Token     string    `json:"-" gorm:"type:varchar(64);not null;uniqueIndex"`
	UserAgent string    `json:"user_agent" gorm:"type:varchar(512);not null;default:''"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	ExpireAt  time.Time `json:"expire_at"`
	User      User      `json:"user" gorm:"foreignKey:UserID"`
}

type UsersLookup struct {
	ID       uint   `json:"id" validate:"required"`
	Username string `json:"username" validate:"required"`
}
