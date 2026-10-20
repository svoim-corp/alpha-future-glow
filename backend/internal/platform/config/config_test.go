package config

import "testing"

func TestLoad(t *testing.T) {
	t.Setenv("DATABASE_URL", "")
	if _, err := Load(); err == nil {
		t.Fatal("ожидалась ошибка при пустом DATABASE_URL")
	}
	t.Setenv("DATABASE_URL", "postgres://x")
	t.Setenv("APP_PROFILE", "demo")
	c, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if c.HTTPAddr != ":8080" || !c.IsDemo() || c.MocksBaseURL != "http://localhost:8090" {
		t.Fatalf("неожиданная конфигурация: %+v", c)
	}
}
