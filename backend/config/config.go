package config

import (
	"github.com/mmtaee/ocserv-dashboard/backend/internal/platform/logging"
	"os"
	"strings"
)

type Config struct {
	Debug        bool
	Host         string
	Port         int
	SecretKey    string
	AgentNode    bool
	AllowOrigins []string
	DB           PostgresConfig
}

type PostgresConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

var cfg *Config

func Init(debug bool, host string, port int) {
	secretKey := os.Getenv("SECRET_KEY")
	if secretKey == "" {
		secretKey = "SECRET_KEY122456"
	}

	allowOrigins := os.Getenv("ALLOW_ORIGINS")
	if allowOrigins == "" {
		logger.Warn("Warning: ALLOW_ORIGINS environment variable not set")
	}

	cfg = &Config{
		Debug:        debug,
		Host:         host,
		Port:         port,
		SecretKey:    secretKey,
		AgentNode:    strings.EqualFold(strings.TrimSpace(os.Getenv("AGENT_NODE")), "true"),
		AllowOrigins: strings.Split(allowOrigins, ","),
		DB:           loadDatabaseEnv(),
	}
}

func loadDatabaseEnv() PostgresConfig {
	host := getEnv("POSTGRES_HOST", "localhost")
	port := getEnv("POSTGRES_PORT", "5432")
	user := getEnv("POSTGRES_USER", "ocserv")
	password := getEnv("POSTGRES_PASSWORD", "ocserv-passwd")
	dbName := getEnv("POSTGRES_DB", "ocserv_db")
	sslMode := getEnv("POSTGRES_SSLMODE", "disable")

	return PostgresConfig{
		Host:     host,
		Port:     port,
		User:     user,
		Password: password,
		DBName:   dbName,
		SSLMode:  sslMode,
	}
}

func Get() *Config {
	return cfg
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
