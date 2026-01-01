# CLI Reference - migrate-tool

## Overview

migrate-tool is a cross-platform database migration CLI with support for PostgreSQL, MySQL, and SQLite3. All commands support multi-environment configuration and require a `migrate-tool.yaml` config file.

**Global Flags:**
- `--config` - Path to config file (default: ./migrate-tool.yaml)
- `--env` - Environment name (default: dev)

---

## Commands

### Config Commands

#### config show
Display current configuration with password masking.

```bash
migrate-tool config show [--config=PATH]
```

**Output:**
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

---

### Migration Commands

#### up
Apply pending migrations to a specific environment.

```bash
migrate-tool up [--steps=N] [--env=ENV] [--config=PATH]
```

**Flags:**
- `--steps` - Number of migrations to apply (default: 0 = all pending)
- `--env` - Target environment name (default: dev)

**Behavior:**
1. Validates environment configuration
2. Checks pending migrations count
3. Applies N or all pending migrations
4. Displays count applied and new version
5. Returns error if migration fails

**Examples:**
```bash
# Apply all pending migrations to dev environment
migrate-tool up --env=dev

# Apply next 2 migrations
migrate-tool up --steps=2 --env=staging

# Apply to production (config must exist)
migrate-tool up --env=prod

# Use custom config file
migrate-tool up --config=/path/to/config.yaml --env=prod
```

**Output:**
```
Applied 3 migration(s) successfully
Current version: 3
```

---

#### down
Rollback the last applied migration(s) from a specific environment.

```bash
migrate-tool down [--steps=N] [--env=ENV] [--config=PATH]
```

**Flags:**
- `--steps` - Number of migrations to rollback (default: 1 = safety default)
- `--env` - Target environment name (default: dev)

**Behavior:**
1. Validates environment configuration
2. Checks applied migrations count
3. Rolls back N migrations (default: 1 for safety)
4. Displays count rolled back and new version
5. Returns error if rollback fails

**Safety Feature:**
The default is 1 step (not all) to prevent accidental data loss. Explicit `--steps=N` required for larger rollbacks.

**Examples:**
```bash
# Rollback 1 migration (default, safe)
migrate-tool down --env=dev

# Rollback 3 migrations
migrate-tool down --steps=3 --env=staging

# Rollback all migrations (explicit)
migrate-tool down --steps=99 --env=dev
```

**Output:**
```
Rolled back 1 migration(s)
Current version: 2
```

**Output (at base):**
```
Rolled back 1 migration(s)
Current version: none (clean slate)
```

---

#### status
Display current migration status for a specific environment.

```bash
migrate-tool status [--env=ENV] [--config=PATH]
```

**Flags:**
- `--env` - Target environment name (default: dev)

**Behavior:**
1. Validates environment configuration
2. Gets current version and dirty state from database
3. Counts pending/applied/total migrations
4. Displays status summary
5. Shows warning if database in dirty state

**Dirty State:**
Indicates migration partially executed and failed. Database left in inconsistent state. Use `migrate-tool force` to fix (Phase 5).

**Examples:**
```bash
# Check status of dev environment
migrate-tool status --env=dev

# Check prod status
migrate-tool status --env=prod
```

**Output (normal):**
```
Environment: dev
Current Version: 3
Dirty: false
Applied: 3 / 5
Pending: 2
```

**Output (at base):**
```
Environment: dev
Current Version: none (no migrations applied)
Dirty: false
Applied: 0 / 5
Pending: 5
```

**Output (dirty state warning):**
```
Environment: prod
Current Version: 5
Dirty: true
Applied: 5 / 7
Pending: 2

WARNING: Database is in dirty state.
This usually means a migration failed mid-execution.
Fix with: migrate-tool force 5 --env=prod
```

---

#### history
Display list of available migrations with applied status for a specific environment.

```bash
migrate-tool history [--limit=N] [--env=ENV] [--config=PATH]
```

**Flags:**
- `--limit` - Number of migrations to show (default: 10)
- `--env` - Target environment name (default: dev)

**Behavior:**
1. Validates environment configuration
2. Gets current version from database
3. Loads all migrations from source
4. Marks each migration as applied [x] or pending [ ]
5. Shows up to limit migrations
6. Displays pagination message if more exist

**Examples:**
```bash
# Show last 10 migrations (default)
migrate-tool history --env=dev

# Show last 20 migrations
migrate-tool history --limit=20 --env=staging

# Show all migrations (large limit)
migrate-tool history --limit=999 --env=dev
```

**Output:**
```
Migration History (env: dev)
----------------------------------------
  [x] 000001 - create_users
  [x] 000002 - add_email_index
  [x] 000003 - create_posts
  [ ] 000004 - add_post_tags
  [ ] 000005 - create_comments

  ... and 10 more (use --limit to show more)
```

---

## Environment Configuration

### Configuration File (migrate-tool.yaml)

```yaml
environments:
  dev:
    database_url: "postgres://user:pass@localhost:5432/myapp_dev?sslmode=disable"
    migrations_path: "./migrations"
    require_confirmation: false

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

### Environment Variable Support

Database URL can reference environment variables using `${VAR}` pattern:

```yaml
database_url: "${DATABASE_URL}"
```

Will expand to the value of `DATABASE_URL` environment variable at runtime.

### require_confirmation

Set to `true` for environments requiring user confirmation before migrations. Used in Phase 7 for interactive prompts.

---

## Exit Codes

- `0` - Success
- `1` - Error (invalid config, migration failure, missing environment)

---

## Common Workflows

### Initial Setup
```bash
# Check configuration
migrate-tool config show --env=dev

# Check current status
migrate-tool status --env=dev

# Apply all pending migrations
migrate-tool up --env=dev
```

### Deploy to Production
```bash
# Preview migrations
migrate-tool history --env=prod

# Check current status
migrate-tool status --env=prod

# Apply migrations (with --steps for staged rollout)
migrate-tool up --steps=1 --env=prod
```

### Rollback on Error
```bash
# Check status
migrate-tool status --env=prod

# Rollback 1 migration
migrate-tool down --env=prod

# Verify state
migrate-tool status --env=prod
```

### Multi-Environment Management
```bash
# Check all environments
for env in dev staging prod; do
  echo "=== $env ==="
  migrate-tool status --env=$env
done
```

---

## Troubleshooting

### No migrations to apply
- Use `migrate-tool history --env=ENV` to verify migrations exist
- Check `migrations_path` in config points to correct directory
- Verify migration files use format: `{version}_{name}.sql`

### Database in dirty state
- Run: `migrate-tool status --env=ENV` to see current version
- Use: `migrate-tool force VERSION --env=ENV` (Phase 5)
- This marks database as clean without rerunning migration

### Config file not found
- Ensure `migrate-tool.yaml` exists in current directory
- Or specify path: `migrate-tool --config=/path/to/config.yaml status`
- Use `config show` to verify configuration is loaded

### Environment not found
- Run: `migrate-tool config show` to list available environments
- Check environment name spelling
- Verify `migrate-tool.yaml` has `environments:` section

### Database connection error
- Verify `database_url` in config is correct
- Check environment variables expanded: `migrate-tool config show`
- Test connection: `psql "postgres://user@host:5432/db"` (PostgreSQL)

---

## Database Support

Tested with:
- PostgreSQL 10+
- MySQL 5.7+
- SQLite3

Connection string format varies by database:

**PostgreSQL:**
```
postgres://[user[:password]@][netloc][:port][/dbname][?param1=value1&...]
```

**MySQL:**
```
mysql://[user[:password]@][netloc][:port]/dbname[?param=value&...]
```

**SQLite3:**
```
sqlite3:///path/to/database.db
```

---

## Migration File Format

Migration files use format: `{version}_{name}.sql`

Example: `000001_create_users.sql`

Content structure:
```sql
-- +migrate UP
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) NOT NULL
);

-- +migrate DOWN
DROP TABLE users;
```

- Sections marked by comment lines: `-- +migrate UP`, `-- +migrate DOWN`
- Both sections optional (UP-only or DOWN-only migrations supported)
- Version: numeric only (no leading zeros required but recommended)
- Name: alphanumeric + underscores
- File must be `.sql` with proper name format

---

## Utility Commands

### create
Create a new migration file with standard UP/DOWN template.

```bash
migrate-tool create <name> [--seq]
```

**Arguments:**
- `<name>` - Migration name (sanitized to lowercase alphanumeric + underscores)

**Flags:**
- `--seq` - Use sequential versioning (default: true). When false, uses timestamp versioning

**Behavior:**
1. Sanitizes migration name (converts spaces/special chars to underscores)
2. Validates name is not empty and under 100 characters
3. Gets migrations path from config or uses default `./migrations`
4. Creates migrations directory if not exists
5. Generates next sequential or timestamp version
6. Creates `.sql` file with migration template
7. Sets secure file permissions (0600 - owner read/write only)

**Naming Convention:**
- Input: `create users`, `add-email`, `Create_Users_Table`
- Output: `create_users`, `add_email`, `create_users_table`

**Examples:**
```bash
# Create migration with sequential version
migrate-tool create create_users_table

# Create migration with timestamp version
migrate-tool create add_email_to_users --seq=false

# Migration name with spaces (sanitized to underscores)
migrate-tool create "add post tags"
```

**Template Output:**
```
-- Migration: create_users
-- Created: 2026-01-01 23:09:15

-- +migrate UP
-- TODO: Add your UP migration SQL here


-- +migrate DOWN
-- TODO: Add your DOWN migration SQL here
```

**Output:**
```
Created: /path/to/migrations/000001_create_users.sql
```

**Security:**
- Path traversal protection via absolute path validation
- Filename sanitization prevents directory escape
- File created with restrictive permissions (0600)
- Name length validated (max 100 chars)

---

### validate
Validate configuration file and migration files for syntax errors.

```bash
migrate-tool validate [--env=ENV]
```

**Flags:**
- `--env` - Validate specific environment only (default: validates all)

**Behavior:**
1. Loads and validates config file
2. Displays environment count
3. For each environment (or specified env):
   - Validates environment configuration
   - Checks migrations path exists
   - Loads all migration files
   - Counts migrations
   - Detects empty UP/DOWN sections
4. Displays errors (red) and warnings (yellow)
5. Returns success if no errors, exit code 1 if errors found

**Examples:**
```bash
# Validate all environments
migrate-tool validate

# Validate only production
migrate-tool validate --env=prod

# Validate staging
migrate-tool validate --env=staging
```

**Output (success):**
```
Validating configuration...
  Found 3 environment(s)

Validating migrations for 'dev'...
  Found 5 migration(s)

Validating migrations for 'staging'...
  Found 5 migration(s)

Validating migrations for 'prod'...
  Found 5 migration(s)

─────────────────────────────
✓ All validations passed
```

**Output (with warnings):**
```
Validating configuration...
  Found 3 environment(s)

Validating migrations for 'dev'...
  Found 5 migration(s)

─────────────────────────────
WARNINGS:
  ! Env dev: 1 migration(s) with empty UP section
  ! Env dev: 2 migration(s) with empty DOWN section
```

**Output (with errors):**
```
Validating configuration...
  Found 0 environment(s)

─────────────────────────────
ERRORS:
  ✗ Config: environments required
```

---

### version
Display version information including commit hash, build date, and Go runtime details.

```bash
migrate-tool version
```

**Behavior:**
1. Displays version (or "dev" if not set)
2. Shows git commit hash (or "unknown" if not available)
3. Shows build date in UTC (or "unknown" if not set)
4. Shows Go version used to compile binary
5. Shows OS and architecture information

**Examples:**
```bash
migrate-tool version
```

**Output (Release):**
```
migrate-tool 1.2.0
  commit: a1b2c3d
  built:  2026-01-01T10:30:00Z
  go:     go1.25.1
  os:     darwin/amd64
```

**Output (Development):**
```
migrate-tool dev
  commit: unknown
  built:  unknown
  go:     go1.25.1
  os:     linux/amd64
```

**Build Information:**
- Version: Injected at compile time via `-X main.version`
- Commit: Short git hash (7 chars) or "none"
- Date: UTC timestamp in ISO 8601 format
- Go: Runtime version
- OS: Platform and architecture

---

## Version Information

To see installed version:

```bash
migrate-tool version
```

(Displays version, git commit, build date, Go version, and OS/arch information injected at compile time)
