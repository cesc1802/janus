package migrator

import (
	"fmt"

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

	return &Migrator{
		m:            m,
		env:          env,
		envName:      envName,
		sourceDriver: srcDriver,
	}, nil
}

// Close releases resources
func (mg *Migrator) Close() error {
	sourceErr, dbErr := mg.m.Close()
	if sourceErr != nil {
		return sourceErr
	}
	return dbErr
}

// Up applies pending migrations
// steps=0 means apply all, steps>0 means apply N migrations
func (mg *Migrator) Up(steps int) error {
	if steps > 0 {
		return mg.m.Steps(steps)
	}
	return mg.m.Up()
}

// Down rolls back migrations
// steps=0 means rollback 1 (safety default), steps>0 means rollback N
func (mg *Migrator) Down(steps int) error {
	if steps > 0 {
		return mg.m.Steps(-steps)
	}
	// Default: rollback 1 migration for safety
	return mg.m.Steps(-1)
}

// Force sets migration version without running actual migration
// Use this to fix dirty state
func (mg *Migrator) Force(version int) error {
	return mg.m.Force(version)
}

// Goto migrates to a specific version (up or down)
func (mg *Migrator) Goto(version uint) error {
	return mg.m.Migrate(version)
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
