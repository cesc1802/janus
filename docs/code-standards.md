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
│   ├── config/                       # Configuration (Phase 2)
│   ├── source/                       # Migration source drivers (Phase 3)
│   │   └── singlefile/               # Single-file up/down migrations
│   │       ├── parser.go             # Migration file parser
│   │       ├── parser_test.go        # Parser tests
│   │       ├── driver.go             # source.Driver implementation
│   │       └── driver_test.go        # Driver tests
│   ├── migration/                    # Migration logic (Phase 4)
│   └── util/                         # Utilities & helpers
├── migrations/                       # Migration files (Phase 3+)
├── go.mod                            # Module definition
├── Makefile                          # Build automation
└── README.md                         # Project documentation
```

#### Package Naming
- Use lowercase, single-word package names when possible
- Use `internal/` for non-exported packages
- Package names should describe their purpose (e.g., `cmd`, `config`, `driver`)

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
- Never ignore errors with `_ =` unless intentional (document why)

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
- Close file handles properly (use defer)

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

