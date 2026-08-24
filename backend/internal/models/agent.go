package models

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

type AgentAddressType string

const (
	AgentAddressTypeIP     AgentAddressType = "ip"
	AgentAddressTypeDomain AgentAddressType = "domain"
)

func (value AgentAddressType) IsValid() bool {
	return value == AgentAddressTypeIP || value == AgentAddressTypeDomain
}

type OcservAgent struct {
	ID          uint             `json:"id" gorm:"primaryKey;autoIncrement"`
	Name        string           `json:"name" gorm:"type:varchar(128);not null" validate:"required,max=128"`
	AddressType AgentAddressType `json:"address_type" gorm:"type:varchar(16);not null" validate:"required,oneof=ip domain"`
	Address     string           `json:"address" gorm:"type:varchar(255);not null;uniqueIndex" validate:"required,max=255"`
	Token       string           `json:"token" gorm:"type:varchar(512);not null" validate:"required,max=512"`
	CreatedAt   time.Time        `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time        `json:"updated_at" gorm:"autoUpdateTime"`
}

func (agent *OcservAgent) BeforeSave(_ *gorm.DB) error {
	if !agent.AddressType.IsValid() {
		return errors.New("invalid agent address type")
	}
	return nil
}

type AgentToken struct {
	ID        uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	Token     string    `json:"token" gorm:"type:varchar(512);not null"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}
