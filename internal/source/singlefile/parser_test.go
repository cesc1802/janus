package singlefile

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseContent(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		wantUp   string
		wantDown string
	}{
		{
			name: "basic up and down",
			content: `-- +migrate UP
CREATE TABLE users (id INT);

-- +migrate DOWN
DROP TABLE users;`,
			wantUp:   "CREATE TABLE users (id INT);",
			wantDown: "DROP TABLE users;",
		},
		{
			name: "multiline statements",
			content: `-- +migrate UP
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) NOT NULL UNIQUE,
    created_at TIMESTAMP DEFAULT NOW()
);
CREATE INDEX idx_users_email ON users(email);

-- +migrate DOWN
DROP INDEX IF EXISTS idx_users_email;
DROP TABLE IF EXISTS users;`,
			wantUp:   "CREATE TABLE users (\n    id SERIAL PRIMARY KEY,\n    email VARCHAR(255) NOT NULL UNIQUE,\n    created_at TIMESTAMP DEFAULT NOW()\n);\nCREATE INDEX idx_users_email ON users(email);",
			wantDown: "DROP INDEX IF EXISTS idx_users_email;\nDROP TABLE IF EXISTS users;",
		},
		{
			name: "up only",
			content: `-- +migrate UP
CREATE TABLE logs (id INT);`,
			wantUp:   "CREATE TABLE logs (id INT);",
			wantDown: "",
		},
		{
			name: "down only",
			content: `-- +migrate DOWN
DROP TABLE logs;`,
			wantUp:   "",
			wantDown: "DROP TABLE logs;",
		},
		{
			name: "content before markers ignored",
			content: `-- This is a comment
-- Another comment

-- +migrate UP
CREATE TABLE test (id INT);

-- +migrate DOWN
DROP TABLE test;`,
			wantUp:   "CREATE TABLE test (id INT);",
			wantDown: "DROP TABLE test;",
		},
		{
			name:     "empty content",
			content:  "",
			wantUp:   "",
			wantDown: "",
		},
		{
			name: "with sql comments",
			content: `-- +migrate UP
-- Create users table
CREATE TABLE users (id INT);
-- End of up migration

-- +migrate DOWN
-- Drop users table
DROP TABLE users;`,
			wantUp:   "-- Create users table\nCREATE TABLE users (id INT);\n-- End of up migration",
			wantDown: "-- Drop users table\nDROP TABLE users;",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			up, down := parseContent(tc.content)
			if up != tc.wantUp {
				t.Errorf("up mismatch:\ngot:  %q\nwant: %q", up, tc.wantUp)
			}
			if down != tc.wantDown {
				t.Errorf("down mismatch:\ngot:  %q\nwant: %q", down, tc.wantDown)
			}
		})
	}
}

func TestValidateFilename(t *testing.T) {
	tests := []struct {
		filename string
		valid    bool
	}{
		{"000001_create_users.sql", true},
		{"000123_add_email.sql", true},
		{"1_simple.sql", true},
		{"999999_large_version.sql", true},
		{"000001_multi_word_name.sql", true},
		{"invalid.sql", false},
		{"create_users.sql", false},
		{"_create_users.sql", false},
		{"000001_create_users.txt", false},
		{"000001_.sql", false}, // empty name should be invalid
		{"", false},
		{"000001_test", false},          // missing .sql
		{"abc_create_users.sql", false}, // non-numeric version
	}

	for _, tc := range tests {
		t.Run(tc.filename, func(t *testing.T) {
			got := validateFilename(tc.filename)
			if got != tc.valid {
				t.Errorf("validateFilename(%q) = %v; want %v", tc.filename, got, tc.valid)
			}
		})
	}
}

func TestParseMigrationFile(t *testing.T) {
	dir := t.TempDir()

	// Create test migration file
	content := `-- +migrate UP
CREATE TABLE users (id SERIAL PRIMARY KEY);

-- +migrate DOWN
DROP TABLE IF EXISTS users;`

	path := filepath.Join(dir, "000001_create_users.sql")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	m, err := parseMigrationFile(path)
	if err != nil {
		t.Fatalf("parseMigrationFile() error: %v", err)
	}

	if m.Version != 1 {
		t.Errorf("Version = %d; want 1", m.Version)
	}
	if m.Name != "create_users" {
		t.Errorf("Name = %q; want %q", m.Name, "create_users")
	}
	if m.Up != "CREATE TABLE users (id SERIAL PRIMARY KEY);" {
		t.Errorf("Up content mismatch: %q", m.Up)
	}
	if m.Down != "DROP TABLE IF EXISTS users;" {
		t.Errorf("Down content mismatch: %q", m.Down)
	}
}

func TestParseMigrationFile_InvalidFilename(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "invalid.sql")
	if err := os.WriteFile(path, []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := parseMigrationFile(path)
	if err == nil {
		t.Error("expected error for invalid filename")
	}
}

func TestParseMigrationFile_NotFound(t *testing.T) {
	_, err := parseMigrationFile("/nonexistent/path/000001_test.sql")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}
