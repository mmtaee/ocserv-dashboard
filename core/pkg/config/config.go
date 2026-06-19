package config

import (
	"strconv"

	"github.com/mmtaee/ocserv-dashboard/core/pkg/logger"
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
	DB           PostgresConfig
	Telegram     TelegramConfig
}

type PostgresConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type TelegramConfig struct {
	APIBase     string
	ReceiptsDir string
}

var cfg *Config

func Init() {
	debugStr := os.Getenv("DEBUG")
	debug := debugStr == "true" || debugStr == "1"

	host := os.Getenv("HOST")
	if host == "" {
		host = "0.0.0.0"
	}

	portStr := os.Getenv("PORT")
	port := 8080
	if portStr != "" {
		if p, err := strconv.Atoi(portStr); err == nil {
			port = p
		}
	}

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
		DB:           loadDatabaseEnv(),
		Telegram:     loadTelegramEnv(),
	}
}

func loadTelegramEnv() TelegramConfig {
	apiBase := getEnv("TELEGRAM_API_BASE", "https://api.telegram.org")
	receiptsDir := getEnv("TELEGRAM_RECEIPTS_DIR", "/opt/ocserv_dashboard/uploads/receipts")
	
	return TelegramConfig{
		APIBase:     apiBase,
		ReceiptsDir: receiptsDir,
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
