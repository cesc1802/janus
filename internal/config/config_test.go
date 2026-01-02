package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
)

func setupTestConfig(t *testing.T, content string) func() {
	t.Helper()

	// Reset config state for each test
	ResetForTesting()

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "janus.yaml")
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	// Reset viper for each test
	viper.Reset()
	viper.SetConfigFile(configPath)
	if err := viper.ReadInConfig(); err != nil {
		t.Fatalf("failed to read test config: %v", err)
	}

	return func() {
		viper.Reset()
		ResetForTesting()
	}
}

func TestLoad_BasicConfig(t *testing.T) {
	cleanup := setupTestConfig(t, `
environments:
  dev:
    database_url: "postgres://localhost:5432/dev"
    migrations_path: "./migrations"
`)
	defer cleanup()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	if len(cfg.Environments) != 1 {
		t.Errorf("expected 1 environment, got %d", len(cfg.Environments))
	}

	dev, ok := cfg.Environments["dev"]
	if !ok {
		t.Fatal("expected 'dev' environment")
	}
	if dev.DatabaseURL != "postgres://localhost:5432/dev" {
		t.Errorf("unexpected database_url: %s", dev.DatabaseURL)
	}
	if dev.MigrationsPath != "./migrations" {
		t.Errorf("unexpected migrations_path: %s", dev.MigrationsPath)
	}
}

func TestLoad_WithEnvVarExpansion(t *testing.T) {
	_ = os.Setenv("TEST_LOAD_DB_URL", "postgres://test:5432/testdb")
	defer func() { _ = os.Unsetenv("TEST_LOAD_DB_URL") }()

	cleanup := setupTestConfig(t, `
environments:
  test:
    database_url: "${TEST_LOAD_DB_URL}"
    migrations_path: "./migrations"
`)
	defer cleanup()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	env := cfg.Environments["test"]
	if env.DatabaseURL != "postgres://test:5432/testdb" {
		t.Errorf("env var not expanded: got %s", env.DatabaseURL)
	}
}

func TestLoad_AppliesDefaults(t *testing.T) {
	cleanup := setupTestConfig(t, `
defaults:
  migrations_path: "./db/migrations"
  require_confirmation: true
environments:
  dev:
    database_url: "postgres://localhost:5432/dev"
`)
	defer cleanup()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	dev := cfg.Environments["dev"]
	if dev.MigrationsPath != "./db/migrations" {
		t.Errorf("default migrations_path not applied: got %s", dev.MigrationsPath)
	}
	// Note: require_confirmation is not inherited from defaults per current impl
}

func TestLoad_FallbackMigrationsPath(t *testing.T) {
	cleanup := setupTestConfig(t, `
environments:
  dev:
    database_url: "postgres://localhost:5432/dev"
`)
	defer cleanup()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	dev := cfg.Environments["dev"]
	if dev.MigrationsPath != "./migrations" {
		t.Errorf("expected fallback migrations_path './migrations', got %s", dev.MigrationsPath)
	}
}

func TestGetEnv(t *testing.T) {
	cleanup := setupTestConfig(t, `
environments:
  dev:
    database_url: "postgres://localhost:5432/dev"
  prod:
    database_url: "postgres://prod:5432/prod"
    require_confirmation: true
`)
	defer cleanup()

	_, err := Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	dev, err := GetEnv("dev")
	if err != nil {
		t.Fatalf("GetEnv('dev') returned error: %v", err)
	}
	if dev.RequireConfirmation {
		t.Error("dev should not require confirmation")
	}

	prod, err := GetEnv("prod")
	if err != nil {
		t.Fatalf("GetEnv('prod') returned error: %v", err)
	}
	if !prod.RequireConfirmation {
		t.Error("prod should require confirmation")
	}
}

func TestGetEnv_NotFound(t *testing.T) {
	cleanup := setupTestConfig(t, `
environments:
  dev:
    database_url: "postgres://localhost:5432/dev"
`)
	defer cleanup()

	_, err := Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	_, err = GetEnv("nonexistent")
	if err == nil {
		t.Error("GetEnv() should return error for nonexistent environment")
	}
}

func TestGet_BeforeLoad(t *testing.T) {
	// Reset global state
	ResetForTesting()

	if Get() != nil {
		t.Error("Get() should return nil before Load()")
	}
}
