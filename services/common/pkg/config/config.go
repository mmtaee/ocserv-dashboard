package config

import (
	"github.com/mmtaee/ocserv-users-management/common/pkg/logger"
	"os"
	"strings"
)

type Config struct {
	Debug        bool
	Host         string
	Port         int
	SecretKey    string
	JWTSecret    string
	AllowOrigins []string
}

var cfg *Config

func Init(debug bool, host string, port int, ignore ...bool) {
	secretKey := os.Getenv("SECRET_KEY")
	if secretKey == "" {
		secretKey = "SECRET_KEY122456"
	}

	allowOrigins := os.Getenv("ALLOW_ORIGINS")
	if allowOrigins == "" {
		logger.Warn("Warning: ALLOW_ORIGINS environment variable not set")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		logger.Warn("Warning: JWT_SECRET environment variable not set, Default value set to secret")
		jwtSecret = "secret1234"
	}

	cfg = &Config{
		Debug:        debug,
		Host:         host,
		Port:         port,
		SecretKey:    secretKey,
		JWTSecret:    jwtSecret,
		AllowOrigins: strings.Split(allowOrigins, ","),
	}
}

func Get() *Config {
	return cfg
}
