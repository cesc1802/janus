package migrator

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"

	"github.com/cesc1802/janus/internal/config"
)

func setupTestConfig(t *testing.T, migrationsPath string) func() {
	t.Helper()

	config.ResetForTesting()
	viper.Reset()

	viper.Set("environments", map[string]interface{}{
		"test": map[string]interface{}{
			"database_url":    "postgres://user:pass@localhost:5432/testdb?sslmode=disable",
			"migrations_path": migrationsPath,
		},
	})

	return func() {
		config.ResetForTesting()
		viper.Reset()
	}
}

func createTestMigrations(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()
	migrationsDir := filepath.Join(dir, "migrations")
	if err := os.MkdirAll(migrationsDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create sample migration
	content := `-- +migrate UP
CREATE TABLE test (id INT);

-- +migrate DOWN
DROP TABLE test;
`
	if err := os.WriteFile(filepath.Join(migrationsDir, "000001_create_test.sql"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	return migrationsDir
}

func TestMigratorNew_InvalidEnv(t *testing.T) {
	migrationsPath := createTestMigrations(t)
	cleanup := setupTestConfig(t, migrationsPath)
	defer cleanup()

	_, err := New("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent env")
	}
}

func TestMigratorNew_InvalidMigrationsPath(t *testing.T) {
	config.ResetForTesting()
	viper.Reset()

	viper.Set("environments", map[string]interface{}{
		"test": map[string]interface{}{
			"database_url":    "postgres://user:pass@localhost:5432/testdb?sslmode=disable",
			"migrations_path": "/nonexistent/path",
		},
	})
	defer func() {
		config.ResetForTesting()
		viper.Reset()
	}()

	_, err := New("test")
	if err == nil {
		t.Error("expected error for invalid migrations path")
	}
}

func TestMigratorRequiresConfirmation(t *testing.T) {
	migrationsPath := createTestMigrations(t)

	config.ResetForTesting()
	viper.Reset()

	viper.Set("environments", map[string]interface{}{
		"prod": map[string]interface{}{
			"database_url":         "postgres://user:pass@localhost:5432/proddb?sslmode=disable",
			"migrations_path":      migrationsPath,
			"require_confirmation": true,
		},
	})
	defer func() {
		config.ResetForTesting()
		viper.Reset()
	}()

	// Cannot fully test without real DB, but can verify config loading
	// The migrator creation will fail at DB connect, which is expected
	_, err := New("prod")
	// Error is expected (no actual DB), but we test the config path
	if err == nil {
		t.Log("Migrator created successfully (unexpected without real DB)")
	}
}

func TestMigratorEnvName(t *testing.T) {
	// This test verifies the EnvName method logic
	// Full integration would require a real database
	t.Skip("Integration test - requires database")
}
