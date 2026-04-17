package config

import (
	"path/filepath"
	"testing"
	"time"
)

func TestLoadConfigWithoutConfigFileUsesEnvAndDefaults(t *testing.T) {
	t.Setenv("APP_ENV", "missing-env")
	t.Setenv("APP_CONFIG_FILE", "")
	t.Setenv("SERVER_PORT", "9191")
	t.Setenv("DATABASE_HOST", "env-only-db")
	t.Setenv("JWT_ACCESS_EXPIRE", "20m")

	if err := LoadConfig(); err != nil {
		t.Fatalf("LoadConfig returned error without config file: %v", err)
	}

	cfg := Get()
	if cfg.Server.Port != 9191 {
		t.Fatalf("expected server port 9191, got %d", cfg.Server.Port)
	}
	if cfg.Database.Host != "env-only-db" {
		t.Fatalf("expected database host env-only-db, got %q", cfg.Database.Host)
	}
	if cfg.JWT.AccessExpire != 20*time.Minute {
		t.Fatalf("expected access expire 20m, got %s", cfg.JWT.AccessExpire)
	}
	if cfg.Server.Mode != "debug" {
		t.Fatalf("expected default server mode debug, got %q", cfg.Server.Mode)
	}
}

func TestLoadConfigFromYAML(t *testing.T) {
	t.Setenv("DATABASE_PASSWORD", "")
	t.Setenv("JWT_ACCESS_EXPIRE", "")
	t.Setenv("JWT_ACCESS_TTL", "")

	filePath := filepath.Join("testdata", "config.test.yaml")
	if err := loadConfigFromFile(filePath, false); err != nil {
		t.Fatalf("loadConfigFromFile returned error: %v", err)
	}

	cfg := Get()
	if cfg.Server.Port != 9090 {
		t.Fatalf("expected server port 9090, got %d", cfg.Server.Port)
	}
	if cfg.Database.Host != "yaml-db" {
		t.Fatalf("expected database host yaml-db, got %q", cfg.Database.Host)
	}
	if cfg.JWT.AccessExpire != 45*time.Minute {
		t.Fatalf("expected access expire 45m, got %s", cfg.JWT.AccessExpire)
	}
}

func TestLoadConfigFromJSONWithEnvOverride(t *testing.T) {
	t.Setenv("DATABASE_PASSWORD", "override-password")
	t.Setenv("JWT_ACCESS_EXPIRE", "3600")

	filePath := filepath.Join("testdata", "config.test.json")
	if err := loadConfigFromFile(filePath, false); err != nil {
		t.Fatalf("loadConfigFromFile returned error: %v", err)
	}

	cfg := Get()
	if cfg.Database.Password != "override-password" {
		t.Fatalf("expected env override password, got %q", cfg.Database.Password)
	}
	if cfg.JWT.AccessExpire != time.Hour {
		t.Fatalf("expected access expire 1h, got %s", cfg.JWT.AccessExpire)
	}
}

func TestLoadConfigFromEnvFile(t *testing.T) {
	t.Setenv("DATABASE_PASSWORD", "")
	t.Setenv("JWT_ACCESS_EXPIRE", "")
	t.Setenv("JWT_ACCESS_TTL", "")

	filePath := filepath.Join("testdata", "config.test.env")
	if err := loadConfigFromFile(filePath, false); err != nil {
		t.Fatalf("loadConfigFromFile returned error: %v", err)
	}

	cfg := Get()
	if cfg.Server.Port != 7070 {
		t.Fatalf("expected server port 7070, got %d", cfg.Server.Port)
	}
	if cfg.Database.Name != "env_blog" {
		t.Fatalf("expected database name env_blog, got %q", cfg.Database.Name)
	}
	if cfg.CORS.AllowCredentials != true {
		t.Fatal("expected CORS allow credentials to be true")
	}
	if cfg.JWT.RefreshExpire != 48*time.Hour {
		t.Fatalf("expected refresh expire 48h, got %s", cfg.JWT.RefreshExpire)
	}
}
