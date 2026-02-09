package config

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
	"net/url"
	"path/filepath"
	"strconv"
)

type Config struct {
	HTTPAddr string
	LogLevel slog.Level

	DatabaseURL string
	AutoMigrate bool

	MigrationsPath string
}

func buildDatabaseURL(user string, pass string, host string, port string, name string, sslmode string) string {
	u := &url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(user, pass),
		Host:  host + ":" + port,
		Path:  "/" + name,
	}
	q := u.Query()
	q.Set("sslmode", sslmode)
	u.RawQuery = q.Encode()
	return u.String()
}

func Load() (Config, error) {
	sslmodeRaw := getEnv("DB_SSLMODE", "disable")
	sslmode, err := validateSSLMode(sslmodeRaw)
	if err != nil {
		return Config{}, err
	}
	cfg := Config{
		HTTPAddr: getEnv("HTTP_ADDR", ":8080"),
		LogLevel: parseLogLevel(getEnv("LOG_LEVEL", "info")),
		DatabaseURL: buildDatabaseURL(
			getEnv("DB_USERNAME", "miniapi"),
			getEnv("DB_PASSWORD", "miniapi"),
			getEnv("DB_HOST", "db"),
			getEnv("DB_PORT", "5432"),
			getEnv("DB_NAME", "miniapi"),
                        sslmode,
		),
		AutoMigrate: parseBool(getEnv("AUTO_MIGRATE", "0")),
		MigrationsPath: getEnv("MIGRATIONS_PATH", defaultMigrationsPath()),
	}

	if strings.TrimSpace(cfg.HTTPAddr) == "" {
		return Config{}, fmt.Errorf("HTTP_ADDR must not be empty")
	}
	if strings.TrimSpace(cfg.DatabaseURL) == "" {
		return Config{}, fmt.Errorf("DATABASE_URL must not be empty")
	}

	return cfg, nil
}


func defaultMigrationsPath() string {
	if _, err := os.Stat("/app/migrations"); err == nil {
		return "/app/migrations"
	}
	return filepath.Clean("./migrations")
}

func getEnv(key string, def string) string {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	return v
}

func parseLogLevel(v string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func validateSSLMode(v string) (string, error) {
	v = strings.ToLower(strings.TrimSpace(v))
	if v == "" {
		return "", fmt.Errorf("DB_SSLMODE must not be empty")
	}

	switch v {
	case "disable", "allow", "prefer", "require", "verify-ca", "verify-full":
		return v, nil
	default:
		return "", fmt.Errorf("invalid DB_SSLMODE=%q; allowed: disable|allow|prefer|require|verify-ca|verify-full", v)
	}
}

func parseBool(v string) bool {
	v = strings.TrimSpace(strings.ToLower(v))
	if v == "" {
		return false
	}
	if v == "1" || v == "true" || v == "yes" || v == "y" || v == "on" {
		return true
	}
	if i, err := strconv.Atoi(v); err == nil {
		return i != 0
	}
	return false
}
