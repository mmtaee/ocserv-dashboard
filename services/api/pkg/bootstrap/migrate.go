package bootstrap

import (
	"github.com/mmtaee/ocserv-users-management/api/internal/models"
	commonModels "github.com/mmtaee/ocserv-users-management/common/models"
	"github.com/mmtaee/ocserv-users-management/common/pkg/database"
	"github.com/mmtaee/ocserv-users-management/common/pkg/logger"
)

var tables = []interface{}{
	&models.System{},
	&models.User{},
	&models.UserToken{},
	&commonModels.OcservGroup{},
	&commonModels.OcservUser{},
	&commonModels.OcservUserTrafficStatistics{},
}

func Migrate() {
	logger.Info("starting migrations...")
	engine := database.GetConnection()
	err := engine.AutoMigrate(tables...)
	if err != nil {
		logger.Fatal("error in AutoMigrate: %v", err)
	}
	logger.Info("migration complete")
}
