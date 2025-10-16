package models

import (
	"errors"
	"gorm.io/gorm"
)

type System struct {
	ID                     uint   `json:"_" gorm:"primaryKey"`
	GoogleCaptchaSecretKey string `json:"google_captcha_secret" gorm:"type:text"`
	GoogleCaptchaSiteKey   string `json:"google_captcha_site_key" gorm:"type:text"`
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
