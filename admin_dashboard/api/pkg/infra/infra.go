package infra

import (
	"github.com/mmtaee/ocserv-dashboard/core/pkg/database"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitInfra() {
	database.Connect()
	DB = database.GetConnection()
}
