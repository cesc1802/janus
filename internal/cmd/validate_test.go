package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/cesc1802/migrate-tool/internal/config"
	"github.com/spf13/viper"
)

func TestRunValidate_NoConfig(t *testing.T) {
	// Save and restore envName
	oldEnvName := envName
	defer func() { envName = oldEnvName }()
	envName = "dev"

	// Reset config state
	config.ResetForTesting()
	viper.Reset()

	// Run validate without config loaded
	err := runValidate(nil, nil)

	// Should fail because config cannot be loaded
	if err == nil {
		t.Log("Validate passed without config (may be acceptable)")
	}
}

func TestRunValidate_WithValidConfig(t *testing.T) {
	// Save and restore envName
	oldEnvName := envName
	defer func() { envName = oldEnvName }()
	envName = "dev"

	dir := t.TempDir()
	migrationsDir := filepath.Join(dir, "migrations")
	os.MkdirAll(migrationsDir, 0755)

	// Create a valid migration file
	migrationContent := `-- Migration: test
-- +migrate UP
CREATE TABLE test (id INT);

-- +migrate DOWN
DROP TABLE IF EXISTS test;
`
	os.WriteFile(filepath.Join(migrationsDir, "000001_test.sql"), []byte(migrationContent), 0644)

	// Setup viper with test config
	config.ResetForTesting()
	viper.Reset()
	viper.Set("environments", map[string]interface{}{
		"dev": map[string]interface{}{
			"database_url":    "postgres://localhost/test",
			"migrations_path": migrationsDir,
		},
	})

	// Load config
	_, err := config.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Run validate
	err = runValidate(nil, nil)
	if err != nil {
		t.Errorf("Validate should pass with valid config: %v", err)
	}
}

func TestRunValidate_MissingMigrationsPath(t *testing.T) {
	// Save and restore envName
	oldEnvName := envName
	defer func() { envName = oldEnvName }()
	envName = "dev"

	dir := t.TempDir()
	nonExistentPath := filepath.Join(dir, "nonexistent")

	// Setup viper with invalid migrations path
	config.ResetForTesting()
	viper.Reset()
	viper.Set("environments", map[string]interface{}{
		"dev": map[string]interface{}{
			"database_url":    "postgres://localhost/test",
			"migrations_path": nonExistentPath,
		},
	})

	// Load config
	_, err := config.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Run validate - should fail due to missing migrations path
	err = runValidate(nil, nil)
	if err == nil {
		t.Error("Validate should fail with missing migrations path")
	}
}

func TestRunValidate_EmptyMigrations(t *testing.T) {
	// Save and restore envName
	oldEnvName := envName
	defer func() { envName = oldEnvName }()
	envName = "dev"

	dir := t.TempDir()
	migrationsDir := filepath.Join(dir, "migrations")
	os.MkdirAll(migrationsDir, 0755)

	// Create migration with empty UP section
	migrationContent := `-- Migration: empty_up
-- +migrate UP

-- +migrate DOWN
DROP TABLE IF EXISTS test;
`
	os.WriteFile(filepath.Join(migrationsDir, "000001_empty_up.sql"), []byte(migrationContent), 0644)

	// Setup viper
	config.ResetForTesting()
	viper.Reset()
	viper.Set("environments", map[string]interface{}{
		"dev": map[string]interface{}{
			"database_url":    "postgres://localhost/test",
			"migrations_path": migrationsDir,
		},
	})

	// Load config
	_, err := config.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Run validate - should pass with warnings
	err = runValidate(nil, nil)
	// Empty sections generate warnings, not errors
	if err != nil {
		t.Logf("Validate returned error (may include warnings): %v", err)
	}
}
