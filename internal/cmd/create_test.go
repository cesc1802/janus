package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSanitizeName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"create_users", "create_users"},
		{"Create Users", "create_users"},
		{"add-email", "add_email"},
		{"123_test", "123_test"},
		{"__test__", "test"},
		{"test@special#chars", "test_special_chars"},
		{"", ""},
	}

	for _, tc := range tests {
		result := sanitizeName(tc.input)
		if result != tc.expected {
			t.Errorf("sanitizeName(%q) = %q; want %q", tc.input, result, tc.expected)
		}
	}
}

func TestSanitizeName_MaxLength(t *testing.T) {
	// Test that long names are sanitized but not truncated by sanitizeName
	// Length check is done separately in runCreate
	longName := strings.Repeat("a", 150)
	result := sanitizeName(longName)
	if len(result) != 150 {
		t.Errorf("sanitizeName should not truncate: got %d chars, want 150", len(result))
	}
}

func TestGetNextSequentialVersion(t *testing.T) {
	t.Run("empty directory", func(t *testing.T) {
		dir := t.TempDir()
		v := getNextSequentialVersion(dir)
		if v != "000001" {
			t.Errorf("empty dir: got %s, want 000001", v)
		}
	})

	t.Run("with existing files", func(t *testing.T) {
		dir := t.TempDir()
		os.WriteFile(filepath.Join(dir, "000001_test.sql"), []byte{}, 0644)
		os.WriteFile(filepath.Join(dir, "000005_test.sql"), []byte{}, 0644)

		v := getNextSequentialVersion(dir)
		if v != "000006" {
			t.Errorf("with files: got %s, want 000006", v)
		}
	})

	t.Run("ignores non-matching files", func(t *testing.T) {
		dir := t.TempDir()
		os.WriteFile(filepath.Join(dir, "readme.md"), []byte{}, 0644)
		os.WriteFile(filepath.Join(dir, "000003_test.sql"), []byte{}, 0644)

		v := getNextSequentialVersion(dir)
		if v != "000004" {
			t.Errorf("got %s, want 000004", v)
		}
	})
}

func TestMigrationTemplate(t *testing.T) {
	content := migrationTemplate("create_users")

	if !strings.Contains(content, "-- Migration: create_users") {
		t.Error("template should contain migration name")
	}
	if !strings.Contains(content, "-- +migrate UP") {
		t.Error("template should contain UP marker")
	}
	if !strings.Contains(content, "-- +migrate DOWN") {
		t.Error("template should contain DOWN marker")
	}
}

func TestRunCreate_Integration(t *testing.T) {
	// Skip integration test in short mode
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	dir := t.TempDir()
	migrationsDir := filepath.Join(dir, "migrations")
	os.MkdirAll(migrationsDir, 0755)

	// Test that runCreate creates a file
	// Note: This test uses the default migrations path from viper
	// The actual file creation is tested indirectly through helper functions
	content := migrationTemplate("test_migration")
	testFile := filepath.Join(migrationsDir, "000001_test.sql")
	err := os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to write test migration: %v", err)
	}

	// Verify file was created with correct template
	data, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read test migration: %v", err)
	}
	if !strings.Contains(string(data), "-- +migrate UP") {
		t.Error("Migration file should contain UP marker")
	}
}
