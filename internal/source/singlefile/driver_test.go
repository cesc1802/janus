package singlefile

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/golang-migrate/migrate/v4/source"
)

func TestDriver_Interface(t *testing.T) {
	// Verify Driver implements source.Driver interface
	var _ source.Driver = (*Driver)(nil)
}

func TestDriver_Open(t *testing.T) {
	dir := t.TempDir()

	// Create test migration
	content := `-- +migrate UP
CREATE TABLE test (id INT);

-- +migrate DOWN
DROP TABLE test;`
	if err := os.WriteFile(filepath.Join(dir, "000001_test.sql"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	d := &Driver{}
	driver, err := d.Open("singlefile://" + dir)
	if err != nil {
		t.Fatalf("Open() error: %v", err)
	}
	defer driver.Close()

	if driver == nil {
		t.Fatal("Open() returned nil driver")
	}
}

func TestDriver_Open_InvalidPath(t *testing.T) {
	d := &Driver{}
	_, err := d.Open("singlefile:///nonexistent/path")
	if err == nil {
		t.Error("expected error for nonexistent path")
	}
}

func TestDriver_Open_NotDirectory(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "file.txt")
	if err := os.WriteFile(file, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	d := &Driver{}
	_, err := d.Open("singlefile://" + file)
	if err == nil {
		t.Error("expected error for file path")
	}
}

func TestNewWithPath(t *testing.T) {
	dir := t.TempDir()

	content := `-- +migrate UP
CREATE TABLE test (id INT);
-- +migrate DOWN
DROP TABLE test;`
	if err := os.WriteFile(filepath.Join(dir, "000001_test.sql"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	d, err := NewWithPath(dir)
	if err != nil {
		t.Fatalf("NewWithPath() error: %v", err)
	}

	if d == nil {
		t.Fatal("NewWithPath() returned nil")
	}
}

func TestDriver_First(t *testing.T) {
	dir := t.TempDir()

	// Create migrations
	for _, file := range []string{"000002_second.sql", "000001_first.sql", "000003_third.sql"} {
		content := "-- +migrate UP\nCREATE TABLE t;\n-- +migrate DOWN\nDROP TABLE t;"
		if err := os.WriteFile(filepath.Join(dir, file), []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}

	d, _ := NewWithPath(dir)
	first, err := d.First()
	if err != nil {
		t.Fatalf("First() error: %v", err)
	}
	if first != 1 {
		t.Errorf("First() = %d; want 1", first)
	}
}

func TestDriver_First_Empty(t *testing.T) {
	dir := t.TempDir()
	d, _ := NewWithPath(dir)

	_, err := d.First()
	if !os.IsNotExist(err) {
		t.Errorf("First() error = %v; want os.ErrNotExist", err)
	}
}

func TestDriver_Next(t *testing.T) {
	dir := t.TempDir()

	for _, file := range []string{"000001_first.sql", "000002_second.sql", "000003_third.sql"} {
		content := "-- +migrate UP\nCREATE TABLE t;\n-- +migrate DOWN\nDROP TABLE t;"
		if err := os.WriteFile(filepath.Join(dir, file), []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}

	d, _ := NewWithPath(dir)

	tests := []struct {
		version  uint
		wantNext uint
		wantErr  bool
	}{
		{1, 2, false},
		{2, 3, false},
		{3, 0, true}, // no next after last
		{0, 1, false},
	}

	for _, tc := range tests {
		next, err := d.Next(tc.version)
		if tc.wantErr {
			if !os.IsNotExist(err) {
				t.Errorf("Next(%d) error = %v; want os.ErrNotExist", tc.version, err)
			}
		} else {
			if err != nil {
				t.Errorf("Next(%d) error: %v", tc.version, err)
			}
			if next != tc.wantNext {
				t.Errorf("Next(%d) = %d; want %d", tc.version, next, tc.wantNext)
			}
		}
	}
}

func TestDriver_Prev(t *testing.T) {
	dir := t.TempDir()

	for _, file := range []string{"000001_first.sql", "000002_second.sql", "000003_third.sql"} {
		content := "-- +migrate UP\nCREATE TABLE t;\n-- +migrate DOWN\nDROP TABLE t;"
		if err := os.WriteFile(filepath.Join(dir, file), []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}

	d, _ := NewWithPath(dir)

	tests := []struct {
		version  uint
		wantPrev uint
		wantErr  bool
	}{
		{3, 2, false},
		{2, 1, false},
		{1, 0, true}, // no prev before first
		{4, 3, false},
	}

	for _, tc := range tests {
		prev, err := d.Prev(tc.version)
		if tc.wantErr {
			if !os.IsNotExist(err) {
				t.Errorf("Prev(%d) error = %v; want os.ErrNotExist", tc.version, err)
			}
		} else {
			if err != nil {
				t.Errorf("Prev(%d) error: %v", tc.version, err)
			}
			if prev != tc.wantPrev {
				t.Errorf("Prev(%d) = %d; want %d", tc.version, prev, tc.wantPrev)
			}
		}
	}
}

func TestDriver_ReadUp(t *testing.T) {
	dir := t.TempDir()

	content := `-- +migrate UP
CREATE TABLE users (id INT);

-- +migrate DOWN
DROP TABLE users;`
	if err := os.WriteFile(filepath.Join(dir, "000001_create_users.sql"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	d, _ := NewWithPath(dir)
	reader, name, err := d.ReadUp(1)
	if err != nil {
		t.Fatalf("ReadUp() error: %v", err)
	}
	defer reader.Close()

	upContent, _ := io.ReadAll(reader)
	if string(upContent) != "CREATE TABLE users (id INT);" {
		t.Errorf("ReadUp content = %q; want CREATE TABLE users (id INT);", upContent)
	}
	if name != "create_users" {
		t.Errorf("ReadUp name = %q; want create_users", name)
	}
}

func TestDriver_ReadUp_NotFound(t *testing.T) {
	dir := t.TempDir()
	d, _ := NewWithPath(dir)

	_, _, err := d.ReadUp(999)
	if !os.IsNotExist(err) {
		t.Errorf("ReadUp(999) error = %v; want os.ErrNotExist", err)
	}
}

func TestDriver_ReadUp_NoUpSection(t *testing.T) {
	dir := t.TempDir()

	content := `-- +migrate DOWN
DROP TABLE users;`
	if err := os.WriteFile(filepath.Join(dir, "000001_down_only.sql"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	d, _ := NewWithPath(dir)
	_, _, err := d.ReadUp(1)
	if !os.IsNotExist(err) {
		t.Errorf("ReadUp() error = %v; want os.ErrNotExist", err)
	}
}

func TestDriver_ReadDown(t *testing.T) {
	dir := t.TempDir()

	content := `-- +migrate UP
CREATE TABLE users (id INT);

-- +migrate DOWN
DROP TABLE users;`
	if err := os.WriteFile(filepath.Join(dir, "000001_create_users.sql"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	d, _ := NewWithPath(dir)
	reader, name, err := d.ReadDown(1)
	if err != nil {
		t.Fatalf("ReadDown() error: %v", err)
	}
	defer reader.Close()

	downContent, _ := io.ReadAll(reader)
	if string(downContent) != "DROP TABLE users;" {
		t.Errorf("ReadDown content = %q; want DROP TABLE users;", downContent)
	}
	if name != "create_users" {
		t.Errorf("ReadDown name = %q; want create_users", name)
	}
}

func TestDriver_DuplicateVersion(t *testing.T) {
	dir := t.TempDir()

	content := "-- +migrate UP\nCREATE TABLE t;\n-- +migrate DOWN\nDROP TABLE t;"
	if err := os.WriteFile(filepath.Join(dir, "000001_first.sql"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "000001_duplicate.sql"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := NewWithPath(dir)
	if err == nil {
		t.Error("expected error for duplicate version")
	}
}

func TestDriver_SkipsNonMigrationFiles(t *testing.T) {
	dir := t.TempDir()

	// Create valid migration
	content := "-- +migrate UP\nCREATE TABLE t;\n-- +migrate DOWN\nDROP TABLE t;"
	if err := os.WriteFile(filepath.Join(dir, "000001_valid.sql"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	// Create non-migration files that should be skipped
	if err := os.WriteFile(filepath.Join(dir, "readme.md"), []byte("readme"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "invalid.sql"), []byte("sql"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(dir, "subdir"), 0755); err != nil {
		t.Fatal(err)
	}

	d, err := NewWithPath(dir)
	if err != nil {
		t.Fatalf("NewWithPath() error: %v", err)
	}

	// Should only have one migration
	driver := d.(*Driver)
	if len(driver.GetVersions()) != 1 {
		t.Errorf("GetVersions() len = %d; want 1", len(driver.GetVersions()))
	}
}

func TestDriver_GetMigrations(t *testing.T) {
	dir := t.TempDir()

	content := "-- +migrate UP\nCREATE TABLE t;\n-- +migrate DOWN\nDROP TABLE t;"
	if err := os.WriteFile(filepath.Join(dir, "000001_test.sql"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	d, _ := NewWithPath(dir)
	driver := d.(*Driver)

	migrations := driver.GetMigrations()
	if len(migrations) != 1 {
		t.Errorf("GetMigrations() len = %d; want 1", len(migrations))
	}
	if _, ok := migrations[1]; !ok {
		t.Error("GetMigrations() missing version 1")
	}
}

func TestDriver_GetVersions(t *testing.T) {
	dir := t.TempDir()

	content := "-- +migrate UP\nCREATE TABLE t;\n-- +migrate DOWN\nDROP TABLE t;"
	for _, file := range []string{"000003_third.sql", "000001_first.sql", "000002_second.sql"} {
		if err := os.WriteFile(filepath.Join(dir, file), []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}

	d, _ := NewWithPath(dir)
	driver := d.(*Driver)

	versions := driver.GetVersions()
	if len(versions) != 3 {
		t.Fatalf("GetVersions() len = %d; want 3", len(versions))
	}

	// Should be sorted
	want := []uint{1, 2, 3}
	for i, v := range versions {
		if v != want[i] {
			t.Errorf("GetVersions()[%d] = %d; want %d", i, v, want[i])
		}
	}
}
