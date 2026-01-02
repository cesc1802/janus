# migrate-tool

Cross-platform database migration CLI tool with single-file up/down support.

## Features

- Multi-environment configuration (dev, staging, prod)
- Single file migrations with up/down sections
- PostgreSQL support with sslmode options
- Confirmation prompts for protected environments
- Auto-approve mode for CI/CD pipelines

## Installation

### Quick Install (Recommended)

**Linux/macOS:**
```bash
curl -sSL https://raw.githubusercontent.com/cesc1802/migration-tool/master/scripts/install.sh | sh
```

**Windows (PowerShell):**
```powershell
irm https://raw.githubusercontent.com/cesc1802/migration-tool/master/scripts/install.ps1 | iex
```

### Install Specific Version

```bash
# Linux/macOS
curl -sSL https://raw.githubusercontent.com/cesc1802/migration-tool/master/scripts/install.sh | sh -s -- --version v1.0.0

# Windows PowerShell
irm https://raw.githubusercontent.com/cesc1802/migration-tool/master/scripts/install.ps1 -OutFile install.ps1; .\install.ps1 -Version v1.0.0
```

### Platform Support

| OS | Architecture | Status |
|----|--------------|--------|
| Linux | amd64, arm64 | ✅ |
| macOS | amd64, arm64 | ✅ |
| Windows | amd64, arm64 | ✅ |

### Manual Download

Download binaries from [Releases](https://github.com/cesc1802/migration-tool/releases).

```bash
# Verify checksum
sha256sum -c checksums.txt --ignore-missing
```

### Build from Source

```bash
go install github.com/cesc1802/migrate-tool/cmd/migrate-tool@latest
```

For detailed options and troubleshooting, see [Deployment Guide](./docs/deployment-guide.md).

## Quick Start

1. Copy the example config:
```bash
cp migrate-tool.example.yaml migrate-tool.yaml
```

2. Configure your database connection in `migrate-tool.yaml`

3. Run migrations:
```bash
migrate-tool up --env dev
```

## Usage

```bash
# Show help
migrate-tool --help

# Run migrations
migrate-tool up --env <environment>

# Rollback migrations
migrate-tool down --env <environment>

# Show version
migrate-tool version
```

### Flags

| Flag | Description |
|------|-------------|
| `--config` | Config file path (default: ./migrate-tool.yaml) |
| `--env` | Environment name (default: dev) |
| `--auto-approve` | Skip confirmation prompts (for CI/CD) |

## Configuration

Create a `migrate-tool.yaml` file:

```yaml
environments:
  dev:
    database_url: "postgres://user:pass@localhost:5432/myapp_dev?sslmode=disable"
    migrations_path: "./migrations"
  staging:
    database_url: "${DATABASE_URL}"
    migrations_path: "./migrations"
    require_confirmation: true
  prod:
    database_url: "${DATABASE_URL}"
    migrations_path: "./migrations"
    require_confirmation: true
```

## License

See [LICENSE](LICENSE) file.
