package config

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Env             string
	Port            string
	DatabaseURL     string
	JWTSecret       []byte
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
	LogLevel        slog.Level
	AllowedOrigins  []string
}

func Load() (*Config, error) {
	cfg := &Config{
		Env:             envOr("APP_ENV", "development"),
		Port:            envOr("PORT", "8080"),
		DatabaseURL:     os.Getenv("DATABASE_URL"),
		AccessTokenTTL:  10 * time.Minute,
		RefreshTokenTTL: 30 * 24 * time.Hour,
		LogLevel:        slog.LevelInfo,
	}

	if cfg.DatabaseURL == "" {
		return nil, errors.New("DATABASE_URL is required")
	}

	secret := os.Getenv("JWT_SECRET")
	if len(secret) < 32 {
		return nil, errors.New("JWT_SECRET must be at least 32 bytes")
	}
	cfg.JWTSecret = []byte(secret)

	if v := os.Getenv("ACCESS_TOKEN_TTL_MINUTES"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n <= 0 {
			return nil, fmt.Errorf("ACCESS_TOKEN_TTL_MINUTES invalid: %q", v)
		}
		cfg.AccessTokenTTL = time.Duration(n) * time.Minute
	}

	if v := os.Getenv("REFRESH_TOKEN_TTL_DAYS"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n <= 0 {
			return nil, fmt.Errorf("REFRESH_TOKEN_TTL_DAYS invalid: %q", v)
		}
		cfg.RefreshTokenTTL = time.Duration(n) * 24 * time.Hour
	}

	if v := os.Getenv("LOG_LEVEL"); v != "" {
		var lvl slog.Level
		if err := lvl.UnmarshalText([]byte(v)); err != nil {
			return nil, fmt.Errorf("LOG_LEVEL invalid: %q", v)
		}
		cfg.LogLevel = lvl
	}

	if v := os.Getenv("ALLOWED_ORIGINS"); v != "" {
		for _, o := range strings.Split(v, ",") {
			if o = strings.TrimSpace(o); o != "" {
				cfg.AllowedOrigins = append(cfg.AllowedOrigins, o)
			}
		}
	}

	return cfg, nil
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}