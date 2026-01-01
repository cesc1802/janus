# Codebase Summary - migrate-tool

## Project Overview

Golang migration CLI tool - Phase 1-2 complete. A cross-platform database migration tool with configuration system, validation, and type-safe accessors.

**Module:** `github.com/cesc1802/migrate-tool`
**Go Version:** 1.25.1
**Total Files:** 13+ files (expanded with config subsystem)

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
│   │   ├── root_test.go         # Unit tests for root command
│   │   └── config_show.go       # "config show" command with masking
│   └── config/
│       ├── config.go            # Config struct, Load(), Get(), GetEnv()
│       ├── env_expand.go        # ExpandEnvVars() for ${VAR} pattern
│       └── validation.go        # Validate() with go-playground/validator
├── Makefile                      # Build & development tasks
├── go.mod                        # Module dependencies
├── go.sum                        # Go dependencies checksums
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

---

## Next Phases
- Phase 4: Core migration commands (up, down, status)
- Phase 5-6: Utility & advanced commands
- Phase 7-8: Interactive UI & release

