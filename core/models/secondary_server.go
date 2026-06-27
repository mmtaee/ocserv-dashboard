package models

import "time"

type SecondaryServer struct {
	ID        uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	Title     string    `json:"title" gorm:"type:varchar(255);not null" validate:"required"`
	IP        string    `json:"ip" gorm:"type:varchar(255);not null" validate:"required"`
	Port      int       `json:"port" gorm:"not null" validate:"required"`
	Token     string    `json:"token" gorm:"type:text;not null" validate:"required"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}
