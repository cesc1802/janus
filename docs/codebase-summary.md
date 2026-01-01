# Codebase Summary - migrate-tool

## Project Overview

Golang migration CLI tool - Phase 1-5 complete. A cross-platform database migration tool with configuration system, validation, migration execution, status tracking, and utility commands for creating migrations and validating configurations.

**Module:** `github.com/cesc1802/migrate-tool`
**Go Version:** 1.25.1
**Total Files:** 33+ files (with Phase 5 utility commands)

---

## Directory Structure

```
.
├── cmd/
│   └── migrate-tool/
│       └── main.go              # CLI entry point
├── internal/
│   ├── cmd/
│   │   ├── root.go              # Root command & config initialization
│   │   ├── config_show.go       # "config show" command
│   │   ├── up.go                # "up" command - apply migrations
│   │   ├── down.go              # "down" command - rollback migrations
│   │   ├── status.go            # "status" command - show migration status
│   │   ├── history.go           # "history" command - list migrations
│   │   ├── create.go            # "create" command - generate new migrations
│   │   ├── validate.go          # "validate" command - validate config & migrations
│   │   ├── version.go           # "version" command - show version info
│   │   ├── *_test.go            # Command tests
│   │   └── root_test.go         # Root command tests
│   ├── config/
│   │   ├── config.go            # Config struct, Load(), Get(), GetEnv()
│   │   ├── env_expand.go        # ExpandEnvVars() for ${VAR} pattern
│   │   ├── validation.go        # Validate() with go-playground/validator
│   │   └── *_test.go            # Config tests
│   ├── migrator/
│   │   ├── migrator.go          # Migrator struct, New(), Up(), Down(), Close()
│   │   ├── status.go            # Status methods & helpers
│   │   └── *_test.go            # Migrator tests
│   └── source/
│       └── singlefile/
│           ├── parser.go        # Migration file parser
│           ├── driver.go        # source.Driver implementation
│           └── *_test.go        # Parser & driver tests
├── migrations/
│   └── 000001_create_users.sql  # Sample migration
├── Makefile                      # Build & development tasks
├── go.mod                        # Module dependencies
├── migrate-tool.example.yaml     # Configuration template
├── .gitignore                    # Git ignore rules
└── .repomixignore               # Repomix ignore rules
```

---

## Core Components

### 1. CLI Entry Point (cmd/migrate-tool/main.go)
- **Purpose:** Application bootstrap and command execution
- **Key Features:**
  - Version information injection (version, commit, date)
  - Delegates command execution to internal command handler
  - Proper error handling with exit codes

### 2. Root Command Handler (internal/cmd/root.go)
- **Purpose:** Root CLI command configuration and config file handling
- **Key Features:**
  - Cobra-based command framework
  - Viper configuration management (YAML config + environment variables)
  - Persistent flags: `--config` (config file path), `--env` (environment name, default: "dev")
  - Auto-config initialization on startup
  - Helper functions: `GetEnvName()`, `IsConfigLoaded()`

### 3. Configuration System (internal/config/)
**Purpose:** Type-safe config loading, validation, and environment variable expansion

**Files:**
- **config.go:**
  - `Config` struct with Environments (map[string]Environment) & Defaults
  - `Environment` struct: DatabaseURL, MigrationsPath, RequireConfirmation
  - `Load()` - loads from Viper, applies defaults, expands env vars, validates (thread-safe with sync.Once)
  - `Get()` - returns loaded config
  - `GetEnv(name)` - retrieves specific environment config
  - `ResetForTesting()` - test utility

- **env_expand.go:**
  - `ExpandEnvVars(s)` - replaces `${VAR}` patterns with environment variable values
  - Preserves unexpanded patterns if env var not found

- **validation.go:**
  - `Validate(c)` - validates config using go-playground/validator
  - Checks: environments required, min 1, database_url required per env
  - Custom validation: detects unexpanded env vars
  - User-friendly error messages via `formatValidationError()`

### 4. Configuration Command (internal/cmd/config_show.go)
- **Purpose:** Display current configuration with security masking
- **Commands:**
  - `config show` - displays all environments and defaults
  - Password masking in database URLs (e.g., `postgres://user:***@host:port/db`)

### 5. Testing (internal/cmd/root_test.go)
- **Purpose:** Unit tests for root command functionality
- **Test Coverage:**
  - `TestSetVersionInfo` - Version information setting
  - `TestExecute` - Command execution without errors
  - `TestRootCmdExists` - Root command initialization

### 6. Migrator Package (internal/migrator/) - Phase 4
**Purpose:** Core migration execution logic wrapping golang-migrate

**Files:**
- **migrator.go:**
  - `Migrator` struct: wraps golang-migrate.Migrate instance
  - `New(envName)` - creates Migrator with config validation
  - `Up(steps)` - apply pending migrations (steps=0 means all)
  - `Down(steps)` - rollback migrations (default: 1 for safety)
  - `Force(version)` - fix dirty state
  - `Goto(version)` - migrate to specific version
  - `Close()` - cleanup resources
  - Helper methods: `RequiresConfirmation()`, `EnvName()`, `Source()`

- **status.go:**
  - `Status` struct: Version, Dirty, Pending, Applied, Total
  - `Status()` method - get current migration status
  - `countMigrations()` - helper to count pending/applied/total
  - `MigrationInfo` struct - single migration entry
  - `GetMigrationList(version)` - list all migrations with applied status

### 7. Migration Commands (internal/cmd/) - Phase 4
**Purpose:** CLI subcommands for migration operations

**Files:**
- **up.go:**
  - Command: `migrate-tool up [--steps=N] --env=ENV`
  - Flag: `--steps` (default: 0 = apply all)
  - Behavior: Gets status, applies N/all pending migrations, shows result

- **down.go:**
  - Command: `migrate-tool down [--steps=N] --env=ENV`
  - Flag: `--steps` (default: 1 = rollback 1 for safety)
  - Behavior: Gets status, rolls back N migrations, shows result

- **status.go:**
  - Command: `migrate-tool status --env=ENV`
  - Output: Current version, dirty state, applied/total, pending count
  - Warnings: Shows dirty state help text if DB in dirty state

- **history.go:**
  - Command: `migrate-tool history [--limit=N] --env=ENV`
  - Flag: `--limit` (default: 10)
  - Output: List of migrations with [x] for applied, [ ] for pending
  - Pagination: Shows "... and N more" if exceeds limit

**Test Coverage (up_test.go, down_test.go, status_test.go, history_test.go):**
  - Command registration verification
  - Flag existence & defaults
  - Error handling for missing config

### 8. Utility Commands (internal/cmd/) - Phase 5
**Purpose:** Helper commands for migration lifecycle and system information

**Files:**
- **create.go:**
  - Command: `migrate-tool create <name> [--seq]`
  - Name sanitization via regex: replaces spaces/special chars with underscores, converts to lowercase
  - Version generation: sequential (000001, 000002, etc.) or timestamp-based (unix timestamp)
  - Template generation: includes migration name, creation timestamp, UP/DOWN markers, TODO comments
  - Security: path traversal prevention, restrictive file permissions (0600)
  - Validation: max 100 char name length, empty name rejection
  - Config integration: reads migrations_path from config or uses ./migrations default
  - Helper functions: `sanitizeName()`, `getNextSequentialVersion()`, `migrationTemplate()`

- **validate.go:**
  - Command: `migrate-tool validate [--env=ENV]`
  - Config validation: loads config, checks syntax and structure
  - Multi-environment validation: validates all envs or specific env via --env flag
  - Migration inspection: scans files, counts total/migrations, detects duplicates
  - Empty section detection: identifies UP/DOWN sections with no SQL content
  - Output formatting: separates errors (✗), warnings (!), success (✓)
  - Error handling: returns exit code 1 on errors, 0 on success

- **version.go:**
  - Command: `migrate-tool version`
  - Version display: shows compiled version, defaults to "dev" if unset
  - Commit hash: displays short git commit, defaults to "unknown"
  - Build date: shows UTC timestamp, defaults to "unknown"
  - Runtime info: Go version via runtime.Version()
  - Platform info: OS/arch via runtime.GOOS and runtime.GOARCH
  - Build-time variables: injected via ldflags from main.go
  - Helper: SetVersionInfo() to set version variables

**Test Coverage (create_test.go, validate_test.go, version_test.go):**
  - Name sanitization: spaces, special chars, case conversion
  - Sequential versioning: empty dir, existing files, non-matching files
  - Template structure: migration markers, timestamp, name inclusion
  - Config validation: with/without config, invalid paths
  - Empty migration detection: empty UP/DOWN sections, warnings
  - Version output: format, Go version, OS info, dev defaults

---

## Dependencies

### Core CLI Framework
- `github.com/spf13/cobra` v1.10.2 - Command framework
- `github.com/spf13/viper` v1.21.0 - Configuration management
- `github.com/spf13/pflag` v1.0.10 - Flag parsing

### Database Migration
- `github.com/golang-migrate/migrate/v4` v4.19.1 - Migration runner

### Interactive UI
- `github.com/manifoldco/promptui` v0.9.0 - Interactive prompts

### Validation & Utilities
- `github.com/go-playground/validator/v10` v10.30.1 - Input validation
- `go.yaml.in/yaml/v3` v3.0.4 - YAML parsing

### Supporting Libraries
- `golang.org/x/crypto` v0.46.0 - Cryptography utilities
- `golang.org/x/text` v0.32.0 - Text handling
- `golang.org/x/sys` v0.39.0 - System utilities

---

## Configuration

### File: migrate-tool.example.yaml
Multi-environment configuration with database URLs and migration paths.

```yaml
environments:
  dev:
    database_url: "postgres://user:pass@localhost:5432/myapp_dev?sslmode=disable"
    migrations_path: "./migrations"
  staging:
    database_url: "${DATABASE_URL}"        # Env var support
    migrations_path: "./migrations"
    require_confirmation: true             # Safety flag
  prod:
    database_url: "${DATABASE_URL}"
    migrations_path: "./migrations"
    require_confirmation: true
```

**Config Loading:**
- Default: `./migrate-tool.yaml` in current directory
- Override: `--config <path>` flag
- Environment selection: `--env <name>` (default: dev)
- Automatic env var substitution via Viper

---

## Build & Development

### Makefile Commands

```bash
make build          # Build binary to bin/migrate-tool (cross-platform)
make run ARGS="..." # Run with arguments
make test           # Run all tests with verbose output
make lint           # Run golangci-lint
make clean          # Remove bin/ directory
```

### Build Features
- CGO disabled for cross-platform compilation
- Version information injected via ldflags:
  - `-X main.version` (VERSION, default: dev)
  - `-X main.commit` (git short hash or "none")
  - `-X main.date` (UTC timestamp)

---

## Development Patterns

### Command Structure
- Single root command: `migrate-tool`
- Persistent flags shared across subcommands
- Viper handles config file + env var merging
- Config initialized before command execution
- Subcommands use type-safe config.Load() instead of Viper directly

### Error Handling
- CLI exits with code 1 on errors
- Propagates errors up to main for exit handling
- Config validation errors include helpful, user-friendly messages

### Configuration Pattern
1. Viper loads YAML + env vars (in initConfig)
2. Command calls config.Load() for type-safe access
3. Load() unmarshals Viper data into Config struct
4. Applies defaults (migrations_path fallback chain)
5. Expands ${VAR} environment variables
6. Validates structure with go-playground/validator
7. Detects unexpanded env vars and reports them
8. Thread-safe via sync.Once (single initialization)

### Environment Variable Expansion
- Pattern: `${VAR}` in YAML values
- Expanded during Load() via ExpandEnvVars()
- Unexpanded patterns preserved if env var not found
- Custom validation detects incomplete expansions and fails early

---

## Completed in Phase 1
- Project structure setup (cmd/, internal/)
- Root command with config initialization
- Makefile for build automation
- Configuration file template with environment support
- Unit tests for core functionality
- Proper gitignore setup for Go projects

## Completed in Phase 2
- Configuration package (internal/config/)
  - Type-safe Config & Environment structs with validation tags
  - Thread-safe Load() with sync.Once pattern
  - GetEnv() for environment-specific access
  - Type-safe defaults application
  - Full validation with helpful error messages
- Environment variable expansion (${VAR} pattern)
  - Regex-based pattern matching
  - Graceful fallback for unset variables
  - Validation to catch incomplete expansions
- Configuration command (config show)
  - Display all environments and defaults
  - Secure password masking in database URLs
- Helper functions in root.go
  - GetEnvName() - current environment name
  - IsConfigLoaded() - config file status check

## Completed in Phase 3
- Single-file source driver (internal/source/singlefile/)
  - **parser.go:** Migration file parser with regex-based filename validation
    - `parseMigrationFile()` - Reads & parses .sql files to extract version/name/up/down sections
    - `parseContent()` - Splits migration into UP & DOWN sections (markers: `-- +migrate UP`, `-- +migrate DOWN`)
    - `validateFilename()` - Validates filename format: `{version}_{name}.sql` (e.g., `000001_create_users.sql`)
  - **driver.go:** source.Driver implementation
    - `Driver` struct - Implements golang-migrate/migrate v4 source.Driver interface
    - `Open(url)` - Initializes driver from URL: `singlefile://path/to/migrations`
    - `NewWithPath(path)` - Direct filesystem initialization (bypass URL parsing)
    - `First()`, `Next()`, `Prev()` - Version navigation with sorted version list
    - `ReadUp()`, `ReadDown()` - Returns io.ReadCloser for migration content
    - `scanMigrations()` - Scans directory, validates files, detects duplicates, path traversal checks
    - `GetMigrations()`, `GetVersions()` - Debugging/status accessors
  - **Tests (98 test cases across parser_test.go & driver_test.go)**
    - Content parsing: basic up/down, multiline, partial (up-only/down-only), comments, empty
    - Filename validation: valid/invalid formats, edge cases
    - Driver interface: Open, NewWithPath, version navigation, read operations
    - Error handling: nonexistent paths, non-directories, missing migrations, duplicates
    - File filtering: skips non-.sql files, invalid filenames, subdirectories
    - Security: path traversal prevention via absolute path resolution
- Registered with golang-migrate driver registry as "singlefile"
- Sample migration file: migrations/000001_create_users.sql (users table DDL)

## Completed in Phase 4
- Core migrator package (internal/migrator/)
  - `Migrator` struct wrapping golang-migrate instance
  - `New(envName)` - creates Migrator with config loading & validation
  - `Up()` method with step control (0=all, N>0=N steps)
  - `Down()` method with safety default (1 step)
  - `Force()`, `Goto()` for advanced migration control
  - Status tracking with `Status()` returning version/dirty/pending/applied/total
  - `GetMigrationList()` for history display with applied status markers
  - Database driver registration: MySQL, PostgreSQL, SQLite3
- Four migration CLI commands:
  - `up [--steps=N]` - Apply migrations with optional step limit
  - `down [--steps=N]` - Rollback migrations (default: 1 step)
  - `status` - Display current migration status & warnings
  - `history [--limit=N]` - Show migration list with applied markers
- Command tests (16 test cases)
  - Command registration verification
  - Flag defaults & types
  - Error handling for missing config
- Status struct with counts: Applied, Pending, Total, Version, Dirty
- Safety features: dirty state detection & warning messages

## Completed in Phase 5 - Utility Commands
- Create migration command (internal/cmd/create.go)
  - `create <name> [--seq]` - Generate new migration files
  - Name sanitization: converts spaces/special chars to underscores
  - Sequential versioning: auto-increments numeric version (000001, 000002, etc.)
  - Timestamp versioning: Unix timestamp when --seq=false
  - Template generation: includes UP/DOWN markers with TODO placeholders
  - Migrations directory auto-creation
  - Security: path traversal prevention, restrictive file permissions (0600)
  - Name validation: max 100 characters, alphanumeric + underscores
  - Configuration integration: reads migrations_path from config with ./migrations fallback
  - Test coverage (create_test.go):
    - Name sanitization validation (spaces, special chars, case conversion)
    - Sequential version number generation
    - File creation with proper template
    - Integration testing

- Validate command (internal/cmd/validate.go)
  - `validate [--env=ENV]` - Validate config & migration files
  - Config validation: checks file syntax, environment count, required fields
  - Multi-environment support: validate all envs or specific env via --env flag
  - Migration inspection: scans migration files, counts migrations
  - Error/warning detection: identifies empty UP/DOWN sections
  - Output formatting: errors (✗), warnings (!), success (✓) with separator
  - Environment-specific validation: each env checked for migrations_path existence
  - Test coverage (validate_test.go):
    - Config loading & parsing
    - Missing/invalid paths detection
    - Empty migration sections handling
    - Multi-environment validation

- Version command (internal/cmd/version.go)
  - `version` - Display version & build information
  - Version display: shows compiled version or "dev" as fallback
  - Commit info: displays short git hash or "unknown"
  - Build date: shows UTC timestamp or "unknown"
  - Runtime info: Go version (via runtime.Version())
  - Platform info: OS and architecture (via runtime.GOOS, runtime.GOARCH)
  - Build-time injection: version, commit, date via ldflags
  - SetVersionInfo() function: sets version variables from main
  - Test coverage (version_test.go):
    - Version output verification
    - Default values for dev builds
    - Output format validation

---

## Next Phases
- Phase 6: Interactive confirmation & dry-run mode
- Phase 7: Advanced features (undo, seed, hooks)
- Phase 8: UI enhancements & release

