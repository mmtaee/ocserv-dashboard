package models

import (
	"time"
)

const (
	AdminRoleSuper = "super"
	AdminRoleAdmin = "admin"
)

type Administrator struct {
	ID         uint               `json:"id" gorm:"primaryKey;autoIncrement"`
	Username   string             `json:"username" gorm:"type:varchar(255);not null;uniqueIndex" validate:"required"`
	Password   string             `json:"-" gorm:"type:varchar(255);not null" validate:"required"`
	Role       string             `json:"role" gorm:"type:varchar(32);not null;default:'admin'" validate:"required,oneof=super admin"`
	LastLogin  *time.Time         `json:"last_login"`
	CreatedAt  time.Time          `json:"created_at" gorm:"autoCreateTime" validate:"required"`
	UpdatedAt  time.Time          `json:"updated_at" gorm:"autoUpdateTime" validate:"omitempty"`
	Tokens     []AdministratorToken `json:"-" gorm:"foreignKey:AdministratorID"`
}


type AdministratorToken struct {
	ID               uint         `json:"-" gorm:"primaryKey;autoIncrement"`
	AdministratorID  uint         `gorm:"index;not null" validate:"required"`
	Token            string       `json:"token" gorm:"type:text;not null" validate:"required"`
	CreatedAt        time.Time    `json:"created_at" gorm:"autoCreateTime"`
	ExpireAt         time.Time    `json:"expire_at"`
	Administrator    Administrator `json:"-" gorm:"foreignKey:AdministratorID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}