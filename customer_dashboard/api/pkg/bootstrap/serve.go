package bootstrap

import (
	"github.com/mmtaee/ocserv-dashboard/customer_dashboard/api/config"
	"github.com/mmtaee/ocserv-dashboard/customer_dashboard/api/pkg/infra"
	"github.com/mmtaee/ocserv-dashboard/customer_dashboard/api/pkg/routing"
)

func Serve() {
	config.Init()
	infra.InitInfra()
	RunMigrations()
	routing.Serve()
}
