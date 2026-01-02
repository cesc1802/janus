# Project Overview & Product Development Requirements

## Project: Golang Migration CLI Tool

### Vision
A robust, cross-platform database migration CLI tool that provides single-file up/down support with a focus on simplicity, safety, and multi-environment management.

---

## Phase 1: Project Setup & Unix Installation (COMPLETED)

### Functional Requirements
- [x] Go project structure with proper module organization
- [x] Root CLI command framework using Cobra
- [x] Configuration management system (YAML + environment variables)
- [x] Build automation with version injection
- [x] Unit test foundation
- [x] Automated Unix installation script with checksum verification

### Non-Functional Requirements
- [x] Cross-platform compilation (CGO disabled)
- [x] Proper error handling and exit codes
- [x] Version tracking (version, commit hash, date)
- [x] Configuration file templating
- [x] Secure binary distribution with SHA256 verification
- [x] Curl/wget fallback for maximum compatibility
- [x] User-friendly colored output

### Acceptance Criteria
- [x] CLI runs without errors
- [x] Root command initializes properly
- [x] Configuration loads from YAML and environment variables
- [x] Unit tests pass (TestSetVersionInfo, TestExecute, TestRootCmdExists)
- [x] Binary builds for multiple platforms
- [x] Install script detects OS/architecture correctly
- [x] Install script verifies archive checksum before installation
- [x] Install script supports custom directories and version pinning
- [x] Install script handles missing dependencies gracefully

### Deliverables
- cmd/migrate-tool/main.go - Entry point with version injection
- internal/cmd/root.go - Root command & config initialization
- internal/cmd/root_test.go - Core unit tests
- Makefile - Build automation
- migrate-tool.example.yaml - Configuration template
- .gitignore - Properly configured for Go project
- scripts/install.sh - Automated Unix installation script (279 lines)
- docs/deployment-guide.md - Installation and distribution documentation

---

## Tech Stack

### Core
- **Language:** Go 1.25.1
- **CLI Framework:** Cobra v1.10.2
- **Config Management:** Viper v1.21.0
- **Flag Parsing:** Pflag v1.0.10

### Database
- **Migration Engine:** golang-migrate/migrate v4.19.1
- **Database Drivers:** PostgreSQL (primary), extensible to MySQL, SQLite

### User Experience
- **Interactive Prompts:** manifoldco/promptui v0.9.0
- **Input Validation:** go-playground/validator v10.30.1

### Infrastructure
- **YAML Processing:** go.yaml.in/yaml v3.0.4
- **Cryptography:** golang.org/x/crypto v0.46.0
- **System Utilities:** golang.org/x/sys v0.39.0
- **Text Processing:** golang.org/x/text v0.32.0

---

## Architecture Overview

### Layer 1: Entry Point
- `cmd/migrate-tool/main.go` - Bootstrap application
- Injects version information (version, commit, date)
- Delegates to internal command handler

### Layer 2: Command Handler
- `internal/cmd/` - Command definitions and handlers
- Root command with persistent flags
- Configuration initialization pipeline

### Layer 3: Configuration
- YAML file-based configuration
- Environment variable override support
- Multi-environment support (dev, staging, prod)
- Safety flags (e.g., require_confirmation)

### Layer 4: Core Functionality
- Database drivers (Phase 3)
- Migration execution engine (Phase 4)
- Advanced commands (Phase 6)

---

## Configuration Structure

### Default Location
- `./migrate-tool.yaml` in current working directory
- Override with `--config <path>`

### Environment Selection
- Flag: `--env <name>` (default: "dev")
- Loads configuration from `environments.<name>` section

### Field Definitions
- `database_url` - Database connection string (supports env var substitution)
- `migrations_path` - Directory containing migration files
- `require_confirmation` - Safety flag for production databases

### Environment Variables
- Viper auto-loads environment variables
- Supports pattern: `MIGRATE_TOOL_*` for CLI-specific vars
- Can override any config field

---

## Development Rules

### Code Organization
- **cmd/** - Executable entry points
- **internal/** - Internal packages (not exported)
- **test files** - Inline with source (e.g., `root_test.go`)

### Build Process
1. Cross-platform compilation with CGO disabled
2. Version info injection via ldflags
3. Binary output to `bin/` directory
4. No dependencies on system-specific libraries

### Testing Standards
- Unit tests required for public functions
- Test naming: `Test<FunctionName>`
- Run tests with `make test`

### Error Handling
- All errors propagate to main
- Exit code 1 for any error
- Graceful shutdown on errors

---

## Phase Roadmap

| Phase | Focus | Status |
|-------|-------|--------|
| 1 | Project Setup, Root Command | COMPLETED |
| 2 | Configuration & Validation | Planned |
| 3 | Source Driver Implementation | Planned |
| 4 | Core Commands (up, down, status) | Planned |
| 5 | Utility Commands (info, rollback) | Planned |
| 6 | Advanced Commands (force, repair) | Planned |
| 7 | Interactive UI & Prompts | Planned |
| 8 | Release & Distribution | Planned |

---

## Success Metrics (Phase 1)

- [x] All unit tests passing
- [x] Binary builds successfully for Linux, macOS, Windows
- [x] Configuration loads correctly
- [x] CLI responds to version flag
- [x] Error handling returns proper exit codes

---

## Dependencies & Constraints

### Required
- Go 1.25.1 or higher
- PostgreSQL driver support (golang-migrate)

### Optional
- golangci-lint for code quality checks
- Git (for commit hash injection)

### Constraints
- CGO disabled for cross-platform compatibility
- No C dependencies allowed
- Configuration must be file-based with env var override
- Single binary distribution model

---

## Known Limitations & Future Work

### Phase 1 Limitations
- No subcommands implemented yet
- Configuration validation not implemented
- No database driver integration yet
- No migration file parsing

### Future Enhancements
- Multiple database driver support
- Interactive migration UI
- Migration history tracking
- Rollback capabilities
- Advanced conflict resolution
- CI/CD integration templates

---

## Contact & Maintenance
- **Module:** github.com/cesc1802/janus
- **Maintained By:** Development Team
- **Last Updated:** 2026-01-01

