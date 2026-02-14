package models

import (
	"errors"
	"gorm.io/gorm"
)

type System struct {
	ID                      uint   `json:"_" gorm:"primaryKey"`
	GoogleCaptchaSecretKey  string `json:"google_captcha_secret" gorm:"type:text"`
	GoogleCaptchaSiteKey    string `json:"google_captcha_site_key" gorm:"type:text"`
	AutoDeleteInactiveUsers bool   `json:"auto_delete_inactive_users" gorm:"type:boolean;default:false"`
	KeepInactiveUserDays    int    `json:"keep_inactive_user_days" gorm:"default:30"`
}

func (s *System) BeforeCreate(tx *gorm.DB) error {
	var count int64
	if err := tx.Model(&System{}).Count(&count).Error; err != nil {
		return err
	}

	if count > 0 {
		return errors.New("system config already exists")
	}
	return nil
}
