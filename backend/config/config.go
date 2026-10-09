package config

import (
	"errors"
	"os"
	"strconv"
	"strings"

	"github.com/mmtaee/ocserv-dashboard/backend/internal/platform/logging"
)

type Config struct {
	Debug              bool
	Host               string
	Port               int
	SecretKey          string
	AgentNode          bool
	TelegramEnabled    bool
	CustomerAPIEnabled bool
	AllowOrigins       []string
	DB                 PostgresConfig
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
	secretKey := strings.TrimSpace(os.Getenv("SECRET_KEY"))

	allowOrigins := os.Getenv("ALLOW_ORIGINS")
	if allowOrigins == "" {
		logger.Warn("Warning: ALLOW_ORIGINS environment variable not set")
	}

	cfg = &Config{
		Debug:              debug,
		Host:               host,
		Port:               port,
		SecretKey:          secretKey,
		AgentNode:          strings.EqualFold(strings.TrimSpace(os.Getenv("AGENT_NODE")), "true"),
		TelegramEnabled:    getBoolEnv("TELEGRAM_BOT_ENABLED", false),
		CustomerAPIEnabled: getBoolEnv("CUSTOMER_API_ENABLED", true),
		AllowOrigins:       strings.Split(allowOrigins, ","),
		DB:                 loadDatabaseEnv(),
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

func Validate() error {
	if cfg == nil || len(cfg.SecretKey) < 32 {
		return errors.New("SECRET_KEY must be set to a random value of at least 32 characters")
	}
	return nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getBoolEnv(key string, fallback bool) bool {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}
	return parsed
}
