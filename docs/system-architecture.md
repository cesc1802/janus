# System Architecture

## Golang Migration CLI Tool - Phase 1-2

---

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│                    User / CI/CD Pipeline                    │
└────────────────────────┬──────────────────────────────────────┘
                         │
                         ▼
                ┌────────────────────┐
                │  migrate-tool CLI  │
                │   (Binary Entry)   │
                └────────────┬───────┘
                             │
                             ▼
                ┌────────────────────────────────────┐
                │  Command Handler (Cobra)           │
                │  - Root command initialization     │
                │  - Flag parsing                    │
                │  - Subcommand routing              │
                └────────────┬───────────────────────┘
                             │
                ┌────────────┴────────────┐
                ▼                         ▼
        ┌───────────────────┐     ┌──────────────────┐
        │ Viper Config Mgr  │     │ Root Command     │
        │                   │     │ - Flags          │
        │ - YAML file       │     │ - initConfig()   │
        │ - Env vars        │     │ - GetEnvName()   │
        │ - AutomaticEnv()  │     │ - IsConfigLoaded │
        └────────┬──────────┘     └──────┬───────────┘
                 │                        │
                 └────────────┬───────────┘
                              ▼
                 ┌────────────────────────────┐
                 │ Config Package (Load)      │
                 │ - Unmarshal to struct      │
                 │ - Apply defaults           │
                 │ - ExpandEnvVars(${VAR})    │
                 │ - Validate() [go-validator]│
                 │ - Thread-safe (sync.Once)  │
                 └────────────┬───────────────┘
                              ▼
                 ┌────────────────────────────┐
                 │ Type-Safe Config           │
                 │ - environments (map)       │
                 │   - dev/staging/prod       │
                 │   - DatabaseURL            │
                 │   - MigrationsPath         │
                 │   - RequireConfirmation    │
                 │ - defaults                 │
                 └────────────┬───────────────┘
                              │
                              ▼
                ┌────────────────────────────┐
                │ Core Functionality         │  (Phase 3+)
                │ - Database Drivers         │
                │ - Migration Engine         │
                │ - Execution Handlers       │
                └────────────────────────────┘
```

---

## Layered Architecture

### Layer 1: Presentation (CLI)
**Files:** `cmd/janus/main.go`

**Responsibilities:**
- Application bootstrap
- Version information injection
- Command execution entry point
- Error handling and exit codes

**Key Components:**
```
main()
├── SetVersionInfo(version, commit, date)
└── Execute() -> rootCmd.Execute()
```

---

### Layer 2: Command Handler
**Files:** `internal/cmd/root.go`, `internal/cmd/root_test.go`

**Responsibilities:**
- Root command definition
- Persistent flag management
- Configuration initialization
- Command routing

**Key Components:**
```
rootCmd (Cobra Command)
├── Use: "migrate-tool"
├── Short: "Database migration CLI tool"
├── Long: "Cross-platform database migration..."
├── PersistentFlags:
│   ├── --config (string): config file path
│   └── --env (string): environment name [default: dev]
└── Subcommands: (Phase 2+)
    ├── migrate up
    ├── migrate down
    ├── migrate status
    └── ...
```

**Flag Processing:**
```
Flag Input
    ↓
rootCmd.PersistentFlags()
    ↓
cobra.OnInitialize(initConfig)
    ↓
initConfig() runs before command
    ↓
Viper loads config & env vars
```

---

### Layer 3: Configuration Management
**Framework:** Viper (for loading) + Type-Safe Config Package

**Configuration Flow:**
```
1. Viper loads YAML + env vars (in initConfig)
   ├── Check --config flag
   ├── If set: viper.SetConfigFile(cfgFile)
   └── If not: Search for janus.yaml

2. Viper sets config parameters
   ├── ConfigName: "janus"
   ├── ConfigType: "yaml"
   ├── ConfigPath: "." (current directory)
   └── AutomaticEnv() for env var overrides

3. Command calls config.Load() for type-safe access
   ├── Unmarshal Viper data into Config struct
   ├── Apply defaults (migrations_path fallback)
   ├── ExpandEnvVars() - replace ${VAR} with env values
   ├── Validate() - struct + custom env var checks
   └── Thread-safe with sync.Once (single init)

4. Access typed config during command execution
   └── env, _ := config.GetEnv("dev")
       dbURL := env.DatabaseURL
```

**Configuration Schema (janus.yaml):**
```yaml
environments:
  <env-name>:
    database_url: <connection-string> # or ${ENV_VAR}
    migrations_path: <directory-path>  # optional, uses defaults if missing
    require_confirmation: <boolean>     # optional
defaults:
  migrations_path: "./migrations"       # fallback for envs without it
  require_confirmation: false           # fallback for envs
```

**Validation Rules:**
- `environments` is required with at least 1 environment
- Each environment's `database_url` is required
- `migrations_path` defaults to: env > defaults > "./migrations"
- Unexpanded env vars detected and reported with helpful errors

**Environment Variable Support:**
- Pattern in YAML: `${VAR_NAME}`
- Expanded via ExpandEnvVars() during Load()
- Viper also supports: `MIGRATE_TOOL_<KEY_PATH_UPPERCASE>`
- If env var not found, pattern preserved (validation catches it)

---

### Layer 4: Configuration Commands (Phase 2)
**Files:** `internal/cmd/config_show.go`

**Commands:**
- `config show` - Display environments & defaults with password masking
  - Masks passwords in database URLs for security
  - Shows current configuration without exposing secrets

---

### Layer 5: Core Functionality (Future Phases)
**Planned Modules:**
- `internal/driver/` - Database driver interface & implementations (Phase 3)
- `internal/migration/` - Migration parsing & execution (Phase 4)
- `internal/util/` - Utilities & helpers

---

## Data Flow

### Startup Sequence
```
1. main.go main()
   │
   ├─ SetVersionInfo(version, commit, date)
   │  └─ Stores version data globally
   │
   └─ Execute()
      └─ rootCmd.Execute()
         │
         ├─ Trigger: cobra.OnInitialize()
         │  └─ initConfig()
         │     ├─ Check --config flag
         │     ├─ Set Viper config name/type/path
         │     ├─ Enable auto env vars
         │     └─ Load janus.yaml
         │
         └─ Execute root command or route to subcommand
            └─ Subcommand accesses config via viper.Get*()
```

### Configuration Access Pattern
```
Subcommand Execution
    ↓
cfgFile (from --config flag)
envName (from --env flag, default: "dev")
    ↓
Viper Configuration Merged
├─ YAML file values
├─ Environment variable overrides
└─ Compiled into in-memory map
    ↓
Access pattern:
environments.dev.database_url → viper.GetString(...)
environments.dev.migrations_path → viper.GetString(...)
```

---

## Component Relationships

### Dependencies
```
main.go
└─ depends on → internal/cmd/root.go
                ├─ depends on → github.com/spf13/cobra
                └─ depends on → github.com/spf13/viper

root.go
├─ depends on → cobra.Command (Cobra framework)
└─ depends on → viper (Configuration)
```

### External Dependencies (Go Modules)
```
Core CLI:
├─ spf13/cobra v1.10.2 (command framework)
├─ spf13/viper v1.21.0 (config management)
├─ spf13/pflag v1.0.10 (flag parsing)
└─ spf13/afero v1.15.0 (file system abstraction)

Database:
└─ golang-migrate/migrate/v4 v4.19.1

Utilities:
├─ manifoldco/promptui v0.9.0 (interactive UI)
├─ go-playground/validator/v10 v10.30.1 (validation)
└─ go.yaml.in/yaml/v3 v3.0.4 (YAML parsing)

System:
├─ golang.org/x/sys v0.39.0
├─ golang.org/x/text v0.32.0
└─ golang.org/x/crypto v0.46.0
```

---

## Phase 1 Design Decisions

### 1. Cobra for CLI Framework
**Rationale:** Industry standard, excellent for complex CLIs
**Benefit:** Easy subcommand management, automatic help, version flags

### 2. Viper for Configuration
**Rationale:** Supports YAML + env var override elegantly
**Benefit:** Multi-environment support, automatic env binding

### 3. Package Structure (cmd/ + internal/)
**Rationale:** Go conventions - cmd/ for binaries, internal/ for private packages
**Benefit:** Clear public/private separation, encapsulation

### 4. YAML Configuration Format
**Rationale:** Human-readable, hierarchical structure
**Benefit:** Multi-environment definitions, environment variable templating

### 5. Version Injection via LDFLAGS
**Rationale:** Single binary with build metadata
**Benefit:** No external version files, accurate version info from git

### 6. Configuration-First Design
**Rationale:** Support DevOps/CI/CD workflows with environment-specific configs
**Benefit:** Same binary, different configurations per environment

---

## Scalability & Extension Points

### For Phase 2-3: Adding Subcommands
```go
// Phase 2: Configuration validation
var validateCmd = &cobra.Command{
	Use: "validate",
	Run: func(cmd *cobra.Command, args []string) {
		// Access config via viper
	},
}
func init() {
	rootCmd.AddCommand(validateCmd)
}

// Phase 4: Migration commands
rootCmd.AddCommand(upCmd, downCmd, statusCmd)
```

### For Phase 3: Database Driver Interface
```go
// Driver interface for multiple database support
type Driver interface {
	Connect(dsn string) error
	RunMigration(migration *Migration) error
	Close() error
}

// Implementations: PostgreSQL, MySQL, SQLite
```

### For Phase 4: Migration Engine
```go
// Migration execution with rollback support
type MigrationRunner struct {
	driver Driver
	migrations []*Migration
}

func (m *MigrationRunner) Up(target string) error
func (m *MigrationRunner) Down(steps int) error
func (m *MigrationRunner) Status() (*MigrationStatus, error)
```

---

## Error Handling Architecture

### Error Propagation Path
```
Driver Error
    ↓
Runner Error
    ↓
Command Handler (returns error)
    ↓
Cobra (prints error)
    ↓
main.go (exits with code 1)
```

### Error Types
- Configuration errors: Invalid YAML, missing env vars
- Driver errors: Connection failed, invalid DSN
- Migration errors: Invalid SQL, constraint violations
- CLI errors: Invalid flags, missing arguments

---

## Security Architecture

### Secret Management
```
Sensitive Data (passwords, tokens)
    ↓
Environment Variables
    ↓
Viper (never logged)
    ↓
Driver (connection string passed securely)
    ↓
Database (over secure connection)
```

### Protection Mechanisms
- Environment variable support for secrets (not in YAML)
- No hardcoded credentials
- Connection string validation before execution
- Confirmation required for destructive operations

---

## Performance Considerations

### Phase 1 Performance Profile
- **Startup:** ~50-100ms (config loading + cobra initialization)
- **Memory:** ~10-20MB (CLI tool, minimal)
- **Binary Size:** ~5-10MB (depends on dependencies)

### Optimization Points for Future Phases
- Cache compiled migration scripts
- Connection pooling for database operations
- Parallel migration execution (Phase 5+)

---

## Testing Architecture

### Unit Test Organization
```
internal/cmd/root_test.go
├── TestSetVersionInfo
│   └── Tests version info injection
├── TestExecute
│   └── Tests command execution
└── TestRootCmdExists
    └── Tests root command initialization
```

### Future Test Layers
- **Unit Tests:** Individual functions & commands
- **Integration Tests:** Config + Command + Database
- **E2E Tests:** Full workflow from CLI to migration

---

## Deployment Model

### Binary Distribution
- Single stateless binary
- Platform-specific builds (Linux, macOS, Windows)
- No external dependencies (CGO disabled)

### Configuration Management
- Local YAML file per environment
- Environment variables for overrides
- No configuration server needed

### CI/CD Integration
```
Build Stage:
├─ make build (builds with version info)
└─ Output: bin/janus-<os>-<arch>

Deploy Stage:
├─ Copy binary to server
├─ Copy janus.yaml to server
└─ Run: ./janus up --env prod
```

---

## Phase 1 Summary

**Completed:**
- CLI framework foundation (Cobra)
- Configuration system (Viper + YAML)
- Version injection pipeline
- Error handling scaffolding
- Build automation (Makefile)
- Unit test foundation

## Phase 2 Summary

**Completed:**
- Type-safe configuration system (internal/config/)
  - Struct-based Config & Environment with validation tags
  - Thread-safe Load() with sync.Once pattern
  - GetEnv() accessor for environment-specific configs
  - Type-safe defaults application chain
- Environment variable expansion & validation
  - ${VAR} pattern replacement with graceful fallback
  - Detection of unset/unexpanded variables
  - User-friendly error reporting
- Configuration inspection command
  - `config show` command with password masking
  - Displays all environments and defaults safely
- Helper functions in root command
  - GetEnvName() - retrieve current environment
  - IsConfigLoaded() - check config file status

**Ready for Phase 3:**
- Database driver interface design
- Connection string validation
- Multi-database support (PostgreSQL, MySQL, SQLite)

