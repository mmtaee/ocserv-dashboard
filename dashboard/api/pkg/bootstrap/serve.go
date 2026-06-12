package bootstrap

import (
	"github.com/mmtaee/ocserv-dashboard/dashboard/api/config"
	"github.com/mmtaee/ocserv-dashboard/dashboard/api/pkg/infra"
	"github.com/mmtaee/ocserv-dashboard/dashboard/api/pkg/routing"
)

func Serve() {
	config.Init()
	infra.InitInfra()
	RunMigrations()
	routing.Serve()
}
