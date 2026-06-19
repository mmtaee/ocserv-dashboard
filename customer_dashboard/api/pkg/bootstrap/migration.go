package bootstrap

import "github.com/mmtaee/ocserv-dashboard/core/migrations"

func RunMigrations() {
	migrations.Migrate()
}
