package utils

import (
	"github.com/mmtaee/ocserv-users-management/api/internal/models"
	"gorm.io/gorm"
)

func CanPerform(db *gorm.DB, userID uint, service string, action string) bool {
	var count int64
	db.Model(&models.Permission{}).Where("user_id = ? AND service = ? AND action = ?", userID, service, action).Count(&count)
	return count > 0
}
