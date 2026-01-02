# Configuration

This guide covers all configuration options for migrate-tool.

## Config File

migrate-tool uses a YAML configuration file. Default location: `./migrate-tool.yaml`

### Basic Structure

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

defaults:
  migrations_path: "./migrations"
  require_confirmation: false
```

### Options Reference

| Option | Description | Default |
|--------|-------------|---------|
| `database_url` | Database connection string | Required |
| `migrations_path` | Path to migrations directory | `./migrations` |
| `require_confirmation` | Prompt before running migrations | `false` |

## Environment Variables

### Variable Substitution

Use `${VAR}` syntax to reference environment variables:

```yaml
environments:
  prod:
    database_url: "${DATABASE_URL}"
```

At runtime, `${DATABASE_URL}` expands to the value of the `DATABASE_URL` environment variable.

### Setting Environment Variables

**Linux/macOS:**
```bash
export DATABASE_URL="postgres://user:pass@prod-host:5432/myapp"
migrate-tool up --env=prod
```

**Windows (PowerShell):**
```powershell
$env:DATABASE_URL = "postgres://user:pass@prod-host:5432/myapp"
migrate-tool up --env=prod
```

**Windows (CMD):**
```cmd
set DATABASE_URL=postgres://user:pass@prod-host:5432/myapp
migrate-tool up --env=prod
```

## Database URLs

### PostgreSQL

```
postgres://[user[:password]@][host][:port][/database][?param=value]
```

Examples:
```yaml
# Local development
database_url: "postgres://localhost:5432/myapp_dev?sslmode=disable"

# With credentials
database_url: "postgres://user:pass@localhost:5432/myapp_dev?sslmode=disable"

# Remote with SSL
database_url: "postgres://user:pass@db.example.com:5432/myapp?sslmode=require"
```

**SSL Mode Options:**
- `disable` - No SSL (dev only)
- `require` - SSL required, no verification
- `verify-ca` - Verify server certificate
- `verify-full` - Verify server certificate and hostname

### MySQL

```
mysql://[user[:password]@][host][:port]/database[?param=value]
```

Examples:
```yaml
# Local development
database_url: "mysql://root@localhost:3306/myapp_dev"

# With credentials
database_url: "mysql://user:pass@localhost:3306/myapp_dev"

# With charset
database_url: "mysql://user:pass@localhost:3306/myapp?charset=utf8mb4"
```

### SQLite3

```
sqlite3:///path/to/database.db
```

Examples:
```yaml
# Relative path
database_url: "sqlite3:///./data/app.db"

# Absolute path
database_url: "sqlite3:////var/lib/myapp/data.db"

# In-memory (for testing)
database_url: "sqlite3:///:memory:"
```

## Custom Config Path

Use `--config` flag to specify an alternate config file:

```bash
migrate-tool status --config=/path/to/custom-config.yaml --env=prod
```

## View Configuration

Use `config show` to display current configuration (passwords masked):

```bash
migrate-tool config show
```

Output:
```
Config file: migrate-tool.yaml

Environments:
  dev:
    database_url: postgres://user:***@localhost:5432/dev
    migrations_path: ./migrations
    require_confirmation: false
  prod:
    database_url: postgres://user:***@prod:5432/prod
    migrations_path: ./migrations
    require_confirmation: true

Defaults:
  migrations_path: ./migrations
```

## Validate Configuration

Check for configuration errors before running migrations:

```bash
migrate-tool validate
```

Output:
```
Validating configuration...
  Found 3 environment(s)

Validating migrations for 'dev'...
  Found 5 migration(s)

─────────────────────────────
✓ All validations passed
```

## Best Practices

### 1. Never Commit Secrets

Use environment variables for sensitive data:

```yaml
# Good - uses environment variable
database_url: "${DATABASE_URL}"

# Bad - hardcoded credentials
database_url: "postgres://user:real-password@host/db"
```

### 2. Use require_confirmation for Production

Prevent accidental migrations:

```yaml
environments:
  prod:
    require_confirmation: true
```

### 3. Keep Consistent migrations_path

Use same path across environments to avoid confusion:

```yaml
defaults:
  migrations_path: "./migrations"
```

### 4. Organize by Environment

Example for a typical setup:

```yaml
environments:
  # Local development - no confirmation needed
  dev:
    database_url: "postgres://localhost/myapp_dev?sslmode=disable"
    migrations_path: "./migrations"

  # Testing - local database, no confirmation
  test:
    database_url: "postgres://localhost/myapp_test?sslmode=disable"
    migrations_path: "./migrations"

  # Staging - remote, confirmation required
  staging:
    database_url: "${STAGING_DATABASE_URL}"
    migrations_path: "./migrations"
    require_confirmation: true

  # Production - remote, confirmation required
  prod:
    database_url: "${PROD_DATABASE_URL}"
    migrations_path: "./migrations"
    require_confirmation: true
```

## Next Steps

- [Creating Migrations](./03-creating-migrations.md) - Learn to create migration files
- [Running Migrations](./04-running-migrations.md) - Apply and rollback migrations
- [CLI Reference](../cli-reference.md) - Complete command documentation
