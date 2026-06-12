package config

import (
	"github.com/mmtaee/ocserv-dashboard/core/pkg/config"
)

type Config = config.Config
type PostgresConfig = config.PostgresConfig

var AppConfig *config.Config

func Init() {
	config.Init()
	AppConfig = config.Get()
}
