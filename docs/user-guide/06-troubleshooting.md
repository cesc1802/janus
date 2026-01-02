# Troubleshooting

This guide covers common issues and how to resolve them.

## Dirty State

### What is Dirty State?

A "dirty" database indicates a migration failed mid-execution. The database is in an inconsistent state where a migration partially applied.

### Detecting Dirty State

```bash
janus status --env=dev
```

Output:
```
Environment: dev
Current Version: 5
Dirty: true
Applied: 5 / 7
Pending: 2

WARNING: Database is in dirty state.
This usually means a migration failed mid-execution.
Fix with: migrate-tool force 5 --env=dev
```

### Fixing Dirty State

**1. Assess the Situation**

Check what migration failed:
```bash
janus history --env=dev
```

Look at the migration file to understand what was attempted.

**2. Manually Fix Database (if needed)**

If partial changes exist, you may need to:
- Complete the migration manually via SQL
- Or rollback partial changes manually

**3. Force Version**

Use `force` to mark the database as clean:

```bash
# Set to last successful version
janus force 4 --env=dev
```

Output:
```
┌─────────────────────────────────────────┐
│  WARNING: Force Version Change          │
└─────────────────────────────────────────┘
Environment: dev
Current version: 5 (dirty: true)
New version: 4

This will NOT run any migrations.
Use this only to recover from dirty state.

Version forced to 4
```

**4. Fix and Re-apply**

After forcing version:
```bash
# Fix the migration file if needed
# vim migrations/000005_xxx.sql

# Re-apply
janus up --env=dev
```

### Force Command Reference

```bash
# Set to specific version
janus force 5 --env=dev

# Reset to base (version 0)
janus force 0 --env=dev

# Clear all version info (NilVersion)
janus force -1 --env=dev
```

## Common Errors

### Config File Not Found

```
Error: config file not found: migrate-tool.yaml
```

**Solutions:**

1. Create the config file:
   ```bash
   cp janus.example.yaml janus.yaml
   ```

2. Or specify path:
   ```bash
   janus status --config=/path/to/config.yaml
   ```

### Environment Not Found

```
Error: environment 'production' not found
```

**Solutions:**

1. Check available environments:
   ```bash
   janus config show
   ```

2. Check spelling - environments are case-sensitive
   ```bash
   # Wrong
   janus up --env=Production

   # Correct
   janus up --env=prod
   ```

### Database Connection Failed

```
Error: failed to connect to database: dial tcp 127.0.0.1:5432: connect: connection refused
```

**Solutions:**

1. Verify database is running:
   ```bash
   # PostgreSQL
   pg_isready -h localhost -p 5432

   # MySQL
   mysqladmin ping -h localhost
   ```

2. Check connection string:
   ```bash
   janus config show --env=dev
   ```

3. Test connection directly:
   ```bash
   # PostgreSQL
   psql "postgres://user:pass@localhost:5432/mydb"

   # MySQL
   mysql -h localhost -u user -p mydb
   ```

### Authentication Failed

```
Error: password authentication failed for user "postgres"
```

**Solutions:**

1. Verify credentials in config:
   ```bash
   janus config show --env=dev
   ```

2. Check environment variable expansion:
   ```bash
   echo $DATABASE_URL
   ```

3. Test with correct credentials

### No Migrations Found

```
No migrations to apply
```

**Check:**

1. Migrations directory exists:
   ```bash
   ls -la ./migrations/
   ```

2. Config points to correct path:
   ```bash
   janus config show --env=dev
   ```

3. Migration files have correct format:
   ```
   000001_create_users.sql
   000002_add_email.sql
   ```

### Migration File Parse Error

```
Error: failed to parse migration 000003_xxx.sql: missing UP section
```

**Solutions:**

1. Check file has correct section markers:
   ```sql
   -- +migrate UP
   CREATE TABLE ...

   -- +migrate DOWN
   DROP TABLE ...
   ```

2. Validate all migrations:
   ```bash
   janus validate --env=dev
   ```

### SSL Connection Error

```
Error: SSL is not enabled on the server
```

**Solutions:**

For development, disable SSL:
```yaml
database_url: "postgres://...?sslmode=disable"
```

For production, use appropriate SSL mode:
```yaml
database_url: "postgres://...?sslmode=require"
```

## Validation Errors

### Run Validation

```bash
janus validate --env=dev
```

### Common Warnings

```
WARNINGS:
  ! Env dev: 1 migration(s) with empty UP section
  ! Env dev: 2 migration(s) with empty DOWN section
```

**Fix:** Add missing SQL to migration files.

### Common Errors

```
ERRORS:
  ✗ Config: environments required
  ✗ Env prod: migrations_path directory not found
```

**Fix:**
- Add `environments:` section to config
- Create migrations directory or fix path

## Recovery Scenarios

### Scenario 1: Migration Failed, Database Unchanged

The migration failed before any changes applied.

```bash
# Check status
janus status --env=dev

# If dirty, force to previous version
janus force 4 --env=dev

# Fix migration file
# vim migrations/000005_xxx.sql

# Re-apply
janus up --env=dev
```

### Scenario 2: Migration Partially Applied

Some changes applied before failure.

```bash
# 1. Assess what applied
psql -c "\d tablename" postgres://...

# 2. Manually complete OR rollback changes
psql postgres://... <<SQL
-- Complete the migration or rollback
SQL

# 3. Force to appropriate version
janus force 5 --env=dev  # If completed
# OR
janus force 4 --env=dev  # If rolled back

# 4. Continue
janus up --env=dev
```

### Scenario 3: Need to Rollback Production

```bash
# 1. Check current state
janus status --env=prod

# 2. Rollback one migration
janus down --env=prod

# 3. Verify
janus status --env=prod
```

### Scenario 4: Reset Development Database

```bash
# Rollback all migrations
janus goto 0 --env=dev

# Re-apply all
janus up --env=dev
```

## Getting Help

### View Help

```bash
# General help
janus --help

# Command-specific help
janus up --help
janus down --help
```

### Show Version

```bash
janus version
```

Include version info when reporting issues.

### Verbose Output

Check config resolution:
```bash
janus config show --env=dev
```

### Debug Connection

Test database connectivity:
```bash
# PostgreSQL
PGPASSWORD=pass psql -h host -U user -d dbname -c "SELECT 1"

# MySQL
mysql -h host -u user -p dbname -e "SELECT 1"
```

## Next Steps

- [CI/CD Integration](./07-ci-cd-integration.md) - Automated deployments
- [CLI Reference](../cli-reference.md) - Complete command documentation
- [Getting Started](./01-getting-started.md) - Start from scratch
