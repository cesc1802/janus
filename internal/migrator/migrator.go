package migrator

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/source"

	"github.com/cesc1802/janus/internal/config"
	"github.com/cesc1802/janus/internal/source/singlefile"

	// Database drivers
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite3"
)

// Migrator wraps golang-migrate with our config and source driver
type Migrator struct {
	m            *migrate.Migrate
	env          config.Environment
	envName      string
	sourceDriver source.Driver
	historyStore *HistoryStore
}

// New creates a Migrator for the given environment
func New(envName string) (*Migrator, error) {
	_, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	env, err := config.GetEnv(envName)
	if err != nil {
		return nil, err
	}

	// Create source driver
	srcDriver, err := singlefile.NewWithPath(env.MigrationsPath)
	if err != nil {
		return nil, fmt.Errorf("source driver: %w", err)
	}

	// Create migrate instance
	m, err := migrate.NewWithSourceInstance("singlefile", srcDriver, env.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("migrate instance: %w", err)
	}

	// Create history store
	hs, err := NewHistoryStore(env.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("history store: %w", err)
	}

	if err := hs.EnsureTable(); err != nil {
		return nil, fmt.Errorf("ensure history table: %w", err)
	}

	return &Migrator{
		m:            m,
		env:          env,
		envName:      envName,
		sourceDriver: srcDriver,
		historyStore: hs,
	}, nil
}

// Close releases resources
func (mg *Migrator) Close() error {
	if mg.historyStore != nil {
		_ = mg.historyStore.Close()
	}
	sourceErr, dbErr := mg.m.Close()
	if sourceErr != nil {
		return sourceErr
	}
	return dbErr
}

// Up applies pending migrations
// steps=0 means apply all, steps>0 means apply N migrations
func (mg *Migrator) Up(steps int) error {
	limit := steps
	if limit == 0 {
		limit = -1 // Infinite
	}

	count := 0
	for limit < 0 || count < limit {
		start := time.Now()
		err := mg.m.Steps(1)
		duration := time.Since(start)

		if err != nil {
			if errors.Is(err, os.ErrNotExist) || errors.Is(err, migrate.ErrNoChange) {
				if count == 0 && steps > 0 {
					// If we specifically asked for steps but found none/no change
					return err
				}
				return nil // Done
			}
			// Try to record failure
			// We don't know the version easily if it failed before updating DB??
			// But usually migrate updates version if dirty.
			// Let's rely on checking version after.
			return err
		}

		// Success
		version, _, _ := mg.m.Version()
		// Dictionary lookup for name
		_, name, _ := mg.sourceDriver.ReadUp(version)

		_ = mg.historyStore.Record(HistoryEntry{
			Version:   version,
			Action:    "up",
			Name:      name,
			StartTime: start,
			Duration:  duration,
			Success:   true,
		})

		count++
	}

	return nil
}

// Down rolls back migrations
// steps=0 means rollback 1 (safety default), steps>0 means rollback N
func (mg *Migrator) Down(steps int) error {
	limit := steps
	if limit == 0 {
		limit = 1 // Default safety
	}

	count := 0
	for count < limit {
		// Get info BEFORE rollback
		version, _, err := mg.m.Version()
		if err != nil {
			if count == 0 {
				return err
			}
			return nil // Should not happen usually
		}

		_, name, _ := mg.sourceDriver.ReadUp(version)

		start := time.Now()
		err = mg.m.Steps(-1)
		duration := time.Since(start)

		if err != nil {
			if errors.Is(err, os.ErrNotExist) || errors.Is(err, migrate.ErrNoChange) {
				if count == 0 {
					return err // Nothing to rollback
				}
				return nil
			}
			return err
		}

		// Record the rollback of 'version'
		_ = mg.historyStore.Record(HistoryEntry{
			Version:   version,
			Action:    "down",
			Name:      name,
			StartTime: start,
			Duration:  duration,
			Success:   true,
		})

		count++
	}
	return nil
}

// Force sets migration version without running actual migration
// Use this to fix dirty state
func (mg *Migrator) Force(version int) error {
	return mg.m.Force(version)
}

// Goto migrates to a specific version (up or down)
func (mg *Migrator) Goto(version uint) error {
	// Goto is tricky to wrap because it might take multiple steps in one go.
	// For now, we delegate to migrate.Migrate which calls Steps internally.
	// Hooking history here would require refactoring Goto to calculate steps and call Up/Down.
	// Given the scope, we might skip detailed history for Goto OR implement it properly.
	// Let's implement properly: determine direction and steps diff.

	current, _, err := mg.m.Version()
	if err != nil && err != migrate.ErrNilVersion {
		return err
	}
	if err == migrate.ErrNilVersion {
		current = 0
	}

	if version == current {
		return migrate.ErrNoChange
	}

	if version > current {
		// Up
		// We can't know exactly how many steps if versions are not sequential (missing gaps?)
		// But assuming singlefile sequential:
		// We iterate steps until we reach the target version.

		// Safe bet: just create a manual record that we jumped?
		// Or loop steps=1 until we hit version?
		// Looping until version is safer for history consistency.

		for {
			curr, _, _ := mg.m.Version()
			if curr >= version {
				break
			}
			if err := mg.Up(1); err != nil {
				return err
			}
		}
	} else {
		// Down
		for {
			curr, _, _ := mg.m.Version()
			if curr <= version && (curr != 0 || version == 0) {
				// handled 0 case
				break
			}
			// Special check for 0
			if curr == 0 {
				break
			}

			if err := mg.Down(1); err != nil {
				return err
			}
		}
	}

	return nil
}

// RequiresConfirmation returns whether this env needs user confirmation
func (mg *Migrator) RequiresConfirmation() bool {
	return mg.env.RequireConfirmation
}

// EnvName returns the environment name
func (mg *Migrator) EnvName() string {
	return mg.envName
}

// Source returns the source driver for iteration
func (mg *Migrator) Source() source.Driver {
	return mg.sourceDriver
}

// History returns the history store
func (mg *Migrator) History() *HistoryStore {
	return mg.historyStore
}
