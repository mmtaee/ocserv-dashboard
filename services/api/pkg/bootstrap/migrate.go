package bootstrap

import (
	"fmt"
	"github.com/mmtaee/ocserv-users-management/api/internal/models"
	commonModels "github.com/mmtaee/ocserv-users-management/common/models"
	"github.com/mmtaee/ocserv-users-management/common/pkg/database"
	"log"
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
	log.Println("starting migrations...")
	engine := database.GetConnection()
	err := engine.AutoMigrate(tables...)
	if err != nil {
		log.Fatalln(fmt.Sprintf("error sync tables: %v", err))
	}
	log.Println("migrating tables successfully")
}
