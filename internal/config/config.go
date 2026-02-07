package config

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
	"net/url"
)

type Config struct {
	HTTPAddr string
	LogLevel slog.Level

	DatabaseURL string
}

func buildDatabaseURL(user string, pass string, host string, port string, name string) string {
	u := &url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(user, pass),
		Host:  host + ":" + port,
		Path:  "/" + name,
	}
	q := u.Query()
	q.Set("sslmode", "disable")
	u.RawQuery = q.Encode()
	return u.String()
}

func Load() (Config, error) {
	cfg := Config{
		HTTPAddr: getEnv("HTTP_ADDR", ":8080"),
		LogLevel: parseLogLevel(getEnv("LOG_LEVEL", "info")),
		DatabaseURL: buildDatabaseURL(
			getEnv("DB_USERNAME", "miniapi"),
			getEnv("DB_PASSWORD", "miniapi"),
			getEnv("DB_HOST", "db"),
			getEnv("DB_PORT", "5432"),
			getEnv("DB_NAME", "miniapi"),
		),
	}

	if strings.TrimSpace(cfg.HTTPAddr) == "" {
		return Config{}, fmt.Errorf("HTTP_ADDR must not be empty")
	}
	if strings.TrimSpace(cfg.DatabaseURL) == "" {
		return Config{}, fmt.Errorf("DATABASE_URL must not be empty")
	}

	return cfg, nil
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
