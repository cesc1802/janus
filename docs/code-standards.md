# Code Standards & Development Guidelines

## Project: Golang Migration CLI Tool

---

## Go Code Standards

### Package Organization

#### Directory Structure
```
project/
├── cmd/
│   └── migrate-tool/
│       └── main.go                   # Executable entry point
├── internal/
│   ├── cmd/                          # CLI command handlers
│   │   ├── root.go                   # Root command (Phase 1)
│   │   ├── config_show.go            # Config command (Phase 2)
│   │   ├── up.go, down.go            # Migration commands (Phase 4)
│   │   ├── status.go, history.go     # Status commands (Phase 4)
│   │   ├── force.go, goto.go         # Advanced commands (Phase 6)
│   │   ├── create.go                 # Create command (Phase 5)
│   │   ├── validate.go               # Validate command (Phase 5)
│   │   └── version.go                # Version command (Phase 5)
│   ├── config/                       # Configuration (Phase 2)
│   ├── ui/                           # UI & confirmations (Phase 7)
│   │   ├── prompt.go                 # TTY detection, confirmation prompts
│   │   └── output.go                 # Colored output helpers
│   ├── source/                       # Migration source drivers (Phase 3)
│   │   └── singlefile/               # Single-file up/down migrations
│   │       ├── parser.go             # Migration file parser
│   │       ├── parser_test.go        # Parser tests
│   │       ├── driver.go             # source.Driver implementation
│   │       └── driver_test.go        # Driver tests
│   ├── migrator/                     # Migration logic (Phase 4)
│   └── util/                         # Utilities & helpers
├── migrations/                       # Migration files (Phase 3+)
├── go.mod                            # Module definition
├── Makefile                          # Build automation
└── README.md                         # Project documentation
```

#### Package Naming
- Use lowercase, single-word package names when possible
- Use `internal/` for non-exported packages
- Package names should describe their purpose (e.g., `cmd`, `config`, `migrator`)
- Package organization by feature: migration logic in `migrator/`, CLI commands in `cmd/`

### Import Organization
```go
import (
	"standard/library"

	"github.com/external-package"
	"github.com/cesc1802/migrate-tool/internal/package"
)
```

Order: Standard library → External → Internal (separated by blank lines)

---

## Command Structure (Cobra)

### Root Command Pattern
```go
var rootCmd = &cobra.Command{
	Use:   "migrate-tool",
	Short: "Brief description",
	Long:  `Detailed description with examples.`,
	// PersistentPreRunE for setup
	// RunE for execution
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file path")
}
```

### Subcommand Pattern (Future)
```go
func init() {
	rootCmd.AddCommand(migrateUpCmd)
	migrateUpCmd.Flags().StringVar(&target, "target", "", "target version")
}

var migrateUpCmd = &cobra.Command{
	Use:   "up [version]",
	Short: "Migrate up to version",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Implementation
		return nil
	},
}
```

### Flag Naming Conventions
- Use hyphens for multi-word flags: `--config-path` (not `--configPath`)
- Single-letter short flags for common options: `-c` for `--config`
- Persistent flags for global options (--config, --env)
- Local flags for command-specific options

---

## Configuration Management (Viper)

### Loading Pattern
```go
func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.SetConfigName("migrate-tool")
		viper.SetConfigType("yaml")
		viper.AddConfigPath(".")
	}
	viper.AutomaticEnv()
	_ = viper.ReadInConfig()
}
```

### Accessing Configuration
```go
// Get string value
dbURL := viper.GetString("environments.dev.database_url")

// Get with nested keys
path := viper.GetString("environments." + envName + ".migrations_path")

// Get with defaults
confirmation := viper.GetBool("environments." + envName + ".require_confirmation")
```

### Environment Variable Mapping
- Pattern: `MIGRATE_TOOL_ENVIRONMENTS_DEV_DATABASE_URL`
- Viper auto-converts underscores to nested keys
- Case-insensitive

---

## Migrator Package Standards (Phase 4)

### Migrator Design Pattern
```go
// Migrator wraps golang-migrate with our config system
type Migrator struct {
	m            *migrate.Migrate
	env          config.Environment
	envName      string
	sourceDriver source.Driver
}

// Factory function loads config, creates source driver, initializes migrator
func New(envName string) (*Migrator, error) {
	// 1. Load configuration
	// 2. Validate environment exists
	// 3. Create source driver with migrations path
	// 4. Initialize golang-migrate instance with database URL
	// 5. Return wrapped Migrator
}

// Always cleanup with defer
defer mg.Close()
```

### Command Handler Pattern
```go
// Command flag variables scoped at package level
var (
	upSteps int
	downSteps int
	historyLimit int
)

// Command defined as package-level variable
var upCmd = &cobra.Command{
	Use:   "up",
	Short: "Brief description",
	Long:  "Longer description",
	RunE:  runUp,  // Function pointer to handler
}

// Handler function
func runUp(cmd *cobra.Command, args []string) error {
	// 1. Create Migrator instance
	// 2. Get current status
	// 3. Perform operation
	// 4. Display results
	// 5. Return error if any
}

// Register command in init()
func init() {
	upCmd.Flags().IntVar(&upSteps, "steps", 0, "Help text")
	rootCmd.AddCommand(upCmd)
}
```

### Status Struct Pattern
```go
// Status represents migration state at a point in time
type Status struct {
	Version uint  // Current applied version (0 = no migrations)
	Dirty   bool  // Database in dirty state (migration failed)
	Pending int   // Migrations not yet applied
	Applied int   // Migrations already applied
	Total   int   // Total migrations in source
}

// Status() method counts migrations by comparing against current version
func (mg *Migrator) Status() (*Status, error) {
	// Get current version from database
	// Count pending/applied/total from source driver
	// Return Status struct
}
```

### Safety Defaults
- **Down command:** Default to 1 step (not all) to prevent accidental data loss
- **Force command:** Require explicit confirmation (Phase 7)
- **Status output:** Show dirty state warning with recovery instructions
- **History display:** Paginate results with "... and N more" indicator

---

## Source Driver Standards (Phase 3)

### Migration File Format
```
-- +migrate UP
[SQL statements for upgrade]

-- +migrate DOWN
[SQL statements for downgrade]
```

- Filename format: `{version}_{name}.sql` (e.g., `000001_create_users.sql`)
- Version: numeric, up to 64-bit unsigned integer
- Name: alphanumeric with underscores, extracted from filename
- Markers: literal strings `-- +migrate UP` and `-- +migrate DOWN` on separate lines
- Sections optional: file can have UP-only or DOWN-only content

### Driver Implementation Pattern
```go
// Implement source.Driver interface from github.com/golang-migrate/migrate/v4/source
type CustomDriver struct {
	path       string
	migrations map[uint]Migration  // version -> migration data
	versions   []uint              // sorted ascending
}

// Core interface methods (all required)
func (d *CustomDriver) Open(url string) (source.Driver, error)
func (d *CustomDriver) Close() error
func (d *CustomDriver) First() (uint, error)
func (d *CustomDriver) Prev(version uint) (uint, error)
func (d *CustomDriver) Next(version uint) (uint, error)
func (d *CustomDriver) ReadUp(version uint) (io.ReadCloser, string, error)
func (d *CustomDriver) ReadDown(version uint) (io.ReadCloser, string, error)
```

- Register with: `source.Register("drivername", &CustomDriver{})`
- Return `os.ErrNotExist` for missing versions/migrations
- Version list must stay sorted in ascending order
- ReadUp/ReadDown return (reader, name, error); name is migration name

### Security Considerations
- Validate path exists and is directory before processing
- Prevent path traversal: resolve absolute paths, verify resolved path stays within migrations dir
- Filename validation: reject files not matching expected pattern
- Handle non-migration files: skip `.md`, invalid names, subdirectories
- Detect duplicate versions: error if multiple files map to same version

### Testing Pattern
```go
func TestDriver_InterfaceCompliance(t *testing.T) {
	// Verify implements source.Driver
	var _ source.Driver = (*Driver)(nil)
}

func TestDriver_Functionality(t *testing.T) {
	dir := t.TempDir()
	// Create test migrations with os.WriteFile
	// Test Open/NewWithPath, First/Next/Prev, ReadUp/ReadDown
	// Test error cases: missing files, invalid formats, duplicates
}
```

- Use `t.TempDir()` for test migrations
- Test both success & error paths
- Verify interface compliance with var _ pattern
- Test version ordering with multiple files
- Test file skipping (non-.sql, invalid names)

---

## Phase 5 Utility Command Standards

### Create Command Pattern
```go
// Command flag for versioning mode
var createSeq bool

var createCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new migration file",
	Args:  cobra.ExactArgs(1),
	RunE:  runCreate,
}

func init() {
	createCmd.Flags().BoolVar(&createSeq, "seq", true, "Use sequential versioning")
	rootCmd.AddCommand(createCmd)
}

func runCreate(cmd *cobra.Command, args []string) error {
	name := args[0]

	// 1. Sanitize input
	name = sanitizeName(name)
	if name == "" {
		return fmt.Errorf("invalid migration name")
	}

	// 2. Validate constraints
	if len(name) > 100 {
		return fmt.Errorf("migration name too long (max 100 chars)")
	}

	// 3. Resolve paths securely
	absPath, err := filepath.Abs(migrationsPath)
	if err != nil {
		return fmt.Errorf("resolve path: %w", err)
	}

	// 4. Create directory & file
	// 5. Verify path security (no traversal)
	// 6. Write with template
	// 7. Report success
}

// Helper: Name sanitization via regex
func sanitizeName(name string) string {
	re := regexp.MustCompile(`[^a-zA-Z0-9_]`)
	name = re.ReplaceAllString(name, "_")
	name = strings.Trim(name, "_")
	return strings.ToLower(name)
}

// Helper: Sequential version generation
func getNextSequentialVersion(dir string) string {
	// Read directory, find highest numeric prefix
	// Return next sequential number (e.g., "000006")
}

// Helper: Migration template generation
func migrationTemplate(name string) string {
	// Return SQL template with UP/DOWN markers and TODO comments
}
```

### Validate Command Pattern
```go
var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate configuration and migration files",
	RunE:  runValidate,
}

func runValidate(cmd *cobra.Command, args []string) error {
	var errors []string
	var warnings []string

	// 1. Validate config
	cfg, err := config.Load()
	if err != nil {
		errors = append(errors, fmt.Sprintf("Config: %v", err))
	}

	// 2. Determine environments to validate
	var envs []string
	if envName != "" {
		envs = append(envs, envName)
	} else if cfg != nil {
		for name := range cfg.Environments {
			envs = append(envs, name)
		}
	}

	// 3. Validate each environment
	for _, env := range envs {
		// Check path exists
		// Load migrations
		// Count & inspect migrations
		// Detect empty UP/DOWN sections
	}

	// 4. Format & display results
	// 5. Return error if any errors found
}
```

### Version Command Pattern
```go
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Run:   runVersion,
}

func runVersion(cmd *cobra.Command, args []string) {
	// Use fallback defaults for empty values
	v := version
	if v == "" {
		v = "dev"
	}

	// Display version, commit, date, Go version, OS/arch
	fmt.Printf("migrate-tool %s\n", v)
	fmt.Printf("  commit: %s\n", c)
	fmt.Printf("  built:  %s\n", d)
	fmt.Printf("  go:     %s\n", runtime.Version())
	fmt.Printf("  os:     %s/%s\n", runtime.GOOS, runtime.GOARCH)
}

// SetVersionInfo is called from main.go to inject build info
func SetVersionInfo(v, c, d string) {
	version = v
	commit = c
	date = d
}
```

### UI & Confirmation Patterns (Phase 7)

**TTY Detection & Non-Interactive Mode:**
```go
// Check if stdout is a terminal
if !ui.IsTTY() && !AutoApprove() {
	return fmt.Errorf("not a TTY: use --auto-approve for non-interactive mode")
}
```

**Confirmation Pattern - Standard Operation:**
```go
if !AutoApprove() {
	confirmed, err := ui.Confirm("Proceed with operation?", false)
	if err != nil {
		return err  // Non-TTY error with clear guidance
	}
	if !confirmed {
		ui.Warning("Cancelled")
		return nil
	}
}
```

**Confirmation Pattern - Production Environment:**
```go
if !AutoApprove() {
	if mg.RequiresConfirmation() {
		// Double confirmation for production safety
		confirmed, err := ui.ConfirmProduction(envName)
		if err != nil {
			return err
		}
		if !confirmed {
			ui.Warning("Cancelled")
			return nil
		}
	} else {
		// Single confirmation for non-production
		confirmed, err := ui.Confirm("Apply migrations?", false)
		if err != nil {
			return err
		}
		if !confirmed {
			ui.Warning("Cancelled")
			return nil
		}
	}
}
```

**Dangerous Operation Confirmation:**
```go
details := fmt.Sprintf("This will %s in %s environment", operation, envName)
confirmed, err := ui.ConfirmDangerous("force version change", details)
if err != nil {
	return err
}
if !confirmed {
	ui.Warning("Cancelled")
	return nil
}
```

**Colored Output:**
```go
ui.Success("Migration applied successfully")
ui.Warning("Database in dirty state - run force to recover")
ui.Error("Connection failed")
ui.Info("Checking migration files...")
```

**NO_COLOR Support:**
- Respects environment variable for accessibility
- Automatically detects TTY for color output
- Plain text fallback for CI/CD systems
- Set via: `NO_COLOR=1 migrate-tool up`

### File Security Patterns (Phase 5)

**Path Traversal Prevention:**
```go
// 1. Resolve to absolute path
absPath, err := filepath.Abs(userPath)
if err != nil {
	return err
}

// 2. Verify resolved path stays within expected dir
if !strings.HasPrefix(absPath, expectedDir+string(filepath.Separator)) {
	return fmt.Errorf("path traversal detected")
}
```

**Restrictive File Permissions:**
```go
// Create migration files with owner-only read/write (0600)
// Prevents other users/processes from reading sensitive SQL
os.WriteFile(fpath, []byte(content), 0600)
```

**Name Sanitization Patterns:**
```go
// Use regex to allow only safe characters
// Convert to lowercase for consistency
// Trim leading/trailing underscores
re := regexp.MustCompile(`[^a-zA-Z0-9_]`)
name = re.ReplaceAllString(name, "_")
name = strings.Trim(name, "_")
name = strings.ToLower(name)
```

---

## Testing Standards

### Test File Naming
- Source: `package.go`
- Tests: `package_test.go`
- Located in same package

### Test Function Naming
```go
func TestFunctionName(t *testing.T) {
	// Table-driven tests preferred for multiple scenarios
}

func BenchmarkFunctionName(b *testing.B) {
	// Performance benchmarks
}
```

### Test Structure - Table-Driven Pattern
```go
func TestMigrate(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expected  string
		wantError bool
	}{
		{"case1", "input1", "expected1", false},
		{"case2", "input2", "expected2", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Migrate(tt.input)
			if (err != nil) != tt.wantError {
				t.Errorf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("got %v, want %v", result, tt.expected)
			}
		})
	}
}
```

### Error Testing
```go
err := functionThatShouldFail()
if err == nil {
	t.Error("expected error, got nil")
}

if !strings.Contains(err.Error(), "expected message") {
	t.Errorf("unexpected error message: %v", err)
}
```

### Coverage Requirements
- Minimum 80% coverage for core packages
- All public functions must have tests
- Run: `go test -cover ./...`

---

## Error Handling

### Error Pattern
```go
if err != nil {
	return fmt.Errorf("operation failed: %w", err)
}
```

### Error Wrapping
- Use `%w` in fmt.Errorf for error chain preservation
- Include context about what operation failed
- Include the failed operation in the message (e.g., "get status:", "migration failed:")

### Resource Cleanup Pattern
Always use defer for cleanup, with anonymous function wrapper:
```go
mg, err := migrator.New(envName)
if err != nil {
	return err
}
defer func() { _ = mg.Close() }()
```

**Why anonymous function:** Avoids propagating deferred call evaluation issues and allows explicit error ignore with documentation.

### Handling Non-Critical Errors
Some errors are safely ignored (e.g., reader close in loops, non-TTY checks):
```go
// Reader close failures don't affect loop continuation
upReader, _, upErr := driver.ReadUp(v)
if upErr != nil {
	emptyUp++
} else {
	_ = upReader.Close()
}
```

**Never ignore** errors that represent actual operation failures.

### CLI Error Handling
```go
func main() {
	cmd.SetVersionInfo(version, commit, date)
	if err := cmd.Execute(); err != nil {
		// Error already printed by Cobra
		os.Exit(1)
	}
}
```

---

## Code Comments & Documentation

### Package Documentation
```go
// Package cmd provides CLI command definitions and handlers.
package cmd
```

### Function Documentation
```go
// Execute runs the root command and returns any error encountered.
func Execute() error {
	return rootCmd.Execute()
}
```

### Inline Comments
- Use for non-obvious logic
- Explain "why" not "what"
- Keep brief and clear

```go
// Viper auto-loads environment variables; we ignore the read error
// as we fall back to defaults if config file doesn't exist
_ = viper.ReadInConfig()
```

### Intentional Error Ignoring
When deferring resource cleanup or handling non-critical errors, use blank identifier pattern:
```go
// Defer resource cleanup - Close errors are typically non-critical
// in cleanup paths and we proceed regardless of error
defer func() { _ = mg.Close() }()

// In loops - reader close failures are non-critical to loop continuation
_ = reader.Close()
```

**Guidelines:**
- Use `_ = value` only for truly non-critical errors (e.g., cleanup, non-TTY checks)
- Always use anonymous function wrapper in defer to avoid deferred call issues
- Document the reason why the error is intentionally ignored
- Never ignore errors that indicate actual operation failures

---

## Naming Conventions

### Variables
- Lowercase with camelCase: `cfgFile`, `envName`, `version`
- Single-letter for loop indices: `i`, `j`, `k`
- Exported (public): PascalCase: `Execute()`, `SetVersionInfo()`
- Unexported: camelCase: `rootCmd`, `initConfig()`

### Constants
```go
const (
	DefaultEnv        = "dev"
	DefaultConfigName = "migrate-tool"
	ConfigTypeYAML    = "yaml"
)
```

### Boolean Variables
- Prefix with verb: `isValid`, `hasError`, `requireConfirmation`
- Or descriptive: `dryRun`, `verbose`, `interactive`

---

## Build & Versioning

### Version Injection (Makefile)
```makefile
VERSION?=dev
COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS=-ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)"
```

### Version Variables (main.go)
```go
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)
```

### Cross-Platform Build
```bash
# Linux
GOOS=linux GOARCH=amd64 go build -o bin/migrate-tool-linux-amd64

# macOS
GOOS=darwin GOARCH=amd64 go build -o bin/migrate-tool-darwin-amd64

# Windows
GOOS=windows GOARCH=amd64 go build -o bin/migrate-tool-windows-amd64.exe
```

---

## Dependency Management

### Adding Dependencies
```bash
go get github.com/package/name
go mod tidy          # Clean up unused dependencies
go mod verify        # Verify integrity
```

### Updating Dependencies
```bash
go get -u ./...      # Update all dependencies
go mod tidy
```

### Checking for Vulnerabilities
```bash
go list -json -m all | nancy sleuth  # Requires nancy tool
```

---

## File Organization Guidelines

### max-length Guidelines
- Lines: Keep under 100 characters when practical
- Functions: Keep under 40 lines (extract helpers if needed)
- Files: Keep under 500 lines (split into multiple files)

### Formatting
- Run `gofmt` before committing
- Use `make lint` to check code quality
- Configure IDE to format on save

---

## Performance Considerations

### Memory
- Reuse buffers for high-frequency operations
- Use `strings.Builder` instead of string concatenation
- Avoid unnecessary allocations in loops

### I/O
- Batch database operations when possible
- Use buffered readers for file operations
- Always defer cleanup: `defer func() { _ = resource.Close() }()`
- In loops, close readers after processing (errors can be safely ignored)

### Concurrency
- Use `sync.Once` for initialization
- Protect shared state with mutexes
- Prefer channels for goroutine coordination

---

## Security Best Practices

### Configuration Secrets
- Never commit `migrate-tool.yaml` with real credentials
- Use environment variable substitution for secrets
- Document required env vars in `.example` files

### Database Connections
- Validate connection strings before use
- Use prepared statements to prevent SQL injection
- Sanitize user input

### Error Messages
- Don't expose sensitive information in error messages
- Log full errors internally, show sanitized errors to users

---

## Git Workflow

### Commit Messages
```
type: brief description (50 chars max)

Longer explanation if needed (72 char line limit)

- Bullet point 1
- Bullet point 2
```

Types: `feat`, `fix`, `docs`, `test`, `refactor`, `chore`

### Branch Naming
- Feature: `feature/description`
- Fix: `fix/issue-description`
- Release: `release/v1.0.0`

---

## Review Checklist

Before submitting code for review:
- [ ] Tests pass: `make test`
- [ ] Code formatted: `gofmt`
- [ ] Linter passes: `make lint`
- [ ] No hardcoded secrets
- [ ] Error handling added
- [ ] Comments added for complex logic
- [ ] Documentation updated if needed
- [ ] Commit messages clear

---

## Resources & Tools

- **Go Docs:** https://golang.org/doc/
- **Effective Go:** https://golang.org/doc/effective_go
- **Cobra Docs:** https://cobra.dev/
- **Viper Docs:** https://github.com/spf13/viper
- **golangci-lint:** https://golangci-lint.run/

