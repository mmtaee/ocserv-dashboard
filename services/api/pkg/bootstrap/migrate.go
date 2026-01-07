package bootstrap

import "github.com/mmtaee/ocserv-users-management/api/internal/migrations"

func Migrate() {
	migrations.Migrate()
}
