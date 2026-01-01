package singlefile

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/golang-migrate/migrate/v4/source"
)

func init() {
	source.Register("singlefile", &Driver{})
}

// Driver implements source.Driver for single-file up/down migrations
type Driver struct {
	path       string
	migrations map[uint]Migration
	versions   []uint // sorted ascending
}

// Open parses the URL and initializes the driver
// URL format: singlefile://path/to/migrations
func (d *Driver) Open(url string) (source.Driver, error) {
	path := strings.TrimPrefix(url, "singlefile://")

	// Validate path exists and is a directory
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("migrations path: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("migrations path is not a directory: %s", path)
	}

	driver := &Driver{
		path:       path,
		migrations: make(map[uint]Migration),
	}

	if err := driver.scanMigrations(); err != nil {
		return nil, err
	}

	return driver, nil
}

// Close releases resources
func (d *Driver) Close() error {
	return nil
}

// First returns the lowest version
func (d *Driver) First() (uint, error) {
	if len(d.versions) == 0 {
		return 0, os.ErrNotExist
	}
	return d.versions[0], nil
}

// Prev returns the previous version before the given version
func (d *Driver) Prev(version uint) (uint, error) {
	for i := len(d.versions) - 1; i >= 0; i-- {
		if d.versions[i] < version {
			return d.versions[i], nil
		}
	}
	return 0, os.ErrNotExist
}

// Next returns the next version after the given version
func (d *Driver) Next(version uint) (uint, error) {
	for _, v := range d.versions {
		if v > version {
			return v, nil
		}
	}
	return 0, os.ErrNotExist
}

// ReadUp returns the UP migration content for a version
func (d *Driver) ReadUp(version uint) (io.ReadCloser, string, error) {
	m, ok := d.migrations[version]
	if !ok {
		return nil, "", os.ErrNotExist
	}
	if m.Up == "" {
		return nil, "", os.ErrNotExist
	}
	return io.NopCloser(strings.NewReader(m.Up)), m.Name, nil
}

// ReadDown returns the DOWN migration content for a version
func (d *Driver) ReadDown(version uint) (io.ReadCloser, string, error) {
	m, ok := d.migrations[version]
	if !ok {
		return nil, "", os.ErrNotExist
	}
	if m.Down == "" {
		return nil, "", os.ErrNotExist
	}
	return io.NopCloser(strings.NewReader(m.Down)), m.Name, nil
}

// scanMigrations reads all .sql files from the migrations directory
func (d *Driver) scanMigrations() error {
	entries, err := os.ReadDir(d.path)
	if err != nil {
		return fmt.Errorf("read migrations dir: %w", err)
	}

	absPath, err := filepath.Abs(d.path)
	if err != nil {
		return fmt.Errorf("resolve migrations path: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}

		// Skip files that don't match migration pattern
		if !validateFilename(entry.Name()) {
			continue
		}

		// Security: prevent path traversal by validating resolved path stays within migrations dir
		filePath := filepath.Join(d.path, entry.Name())
		absFilePath, err := filepath.Abs(filePath)
		if err != nil {
			continue
		}
		if !strings.HasPrefix(absFilePath, absPath+string(filepath.Separator)) {
			continue
		}

		m, err := parseMigrationFile(filePath)
		if err != nil {
			return err
		}

		// Check for duplicate versions
		if _, exists := d.migrations[m.Version]; exists {
			return fmt.Errorf("duplicate migration version: %d", m.Version)
		}

		d.migrations[m.Version] = m
		d.versions = append(d.versions, m.Version)
	}

	sort.Slice(d.versions, func(i, j int) bool {
		return d.versions[i] < d.versions[j]
	})

	return nil
}

// NewWithPath creates a driver directly from a filesystem path
// This is useful for programmatic access without URL parsing
func NewWithPath(path string) (source.Driver, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("migrations path: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("migrations path is not a directory: %s", path)
	}

	d := &Driver{
		path:       path,
		migrations: make(map[uint]Migration),
	}
	if err := d.scanMigrations(); err != nil {
		return nil, err
	}
	return d, nil
}

// GetMigrations returns all parsed migrations (for debugging/status)
func (d *Driver) GetMigrations() map[uint]Migration {
	return d.migrations
}

// GetVersions returns sorted list of versions
func (d *Driver) GetVersions() []uint {
	return d.versions
}
