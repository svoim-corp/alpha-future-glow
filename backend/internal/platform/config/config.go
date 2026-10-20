// Package config читает настройки сервера из переменных окружения.
package config

import (
	"errors"
	"os"
)

type Config struct {
	HTTPAddr     string // HTTP_ADDR, по умолчанию ":8080"
	Profile      string // APP_PROFILE (dev | demo | prod), по умолчанию "dev"
	DatabaseURL  string // DATABASE_URL, обязателен
	MocksBaseURL string // MOCKS_BASE_URL, по умолчанию "http://localhost:8090"
}

// Load читает конфигурацию из окружения.
func Load() (Config, error) {
	c := Config{
		HTTPAddr:     env("HTTP_ADDR", ":8080"),
		Profile:      env("APP_PROFILE", "dev"),
		DatabaseURL:  os.Getenv("DATABASE_URL"),
		MocksBaseURL: env("MOCKS_BASE_URL", "http://localhost:8090"),
	}
	if c.DatabaseURL == "" {
		return c, errors.New("DATABASE_URL не задан")
	}
	return c, nil
}

// IsDemo — включён ли demo-профиль (сброс демо-данных и т.п.).
func (c Config) IsDemo() bool { return c.Profile == "demo" }

// IsDev — локальная разработка (gin в debug-режиме).
func (c Config) IsDev() bool { return c.Profile == "dev" }

func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
