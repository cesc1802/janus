# Multi-Environment Workflows

This guide covers managing migrations across development, staging, and production environments.

## Environment Configuration

### Typical Setup

```yaml
environments:
  # Local development - fast iteration
  dev:
    database_url: "postgres://localhost:5432/myapp_dev?sslmode=disable"
    migrations_path: "./migrations"
    require_confirmation: false

  # Testing environment
  test:
    database_url: "postgres://localhost:5432/myapp_test?sslmode=disable"
    migrations_path: "./migrations"
    require_confirmation: false

  # Staging - mirrors production
  staging:
    database_url: "${STAGING_DATABASE_URL}"
    migrations_path: "./migrations"
    require_confirmation: true

  # Production - highest caution
  prod:
    database_url: "${PROD_DATABASE_URL}"
    migrations_path: "./migrations"
    require_confirmation: true

defaults:
  migrations_path: "./migrations"
```

### Environment Variables

Set database URLs per environment:

**Linux/macOS:**
```bash
export STAGING_DATABASE_URL="postgres://user:pass@staging.example.com:5432/myapp"
export PROD_DATABASE_URL="postgres://user:pass@prod.example.com:5432/myapp"
```

**Windows (PowerShell):**
```powershell
$env:STAGING_DATABASE_URL = "postgres://user:pass@staging.example.com:5432/myapp"
$env:PROD_DATABASE_URL = "postgres://user:pass@prod.example.com:5432/myapp"
```

## Promotion Workflow

### Development → Staging → Production

```
                 ┌─────────────┐
                 │   Create    │
                 │  Migration  │
                 └──────┬──────┘
                        │
                        ▼
                 ┌─────────────┐
                 │   dev up    │  Test locally
                 └──────┬──────┘
                        │
                        ▼
                 ┌─────────────┐
                 │ staging up  │  Verify in staging
                 └──────┬──────┘
                        │
                        ▼
                 ┌─────────────┐
                 │  prod up    │  Deploy to production
                 └─────────────┘
```

### Step-by-Step

**1. Develop Locally**

```bash
# Create new migration
migrate-tool create add_user_phone

# Edit the migration file
# vim migrations/000005_add_user_phone.sql

# Apply to dev
migrate-tool up --env=dev

# Test your application

# Need changes? Rollback and edit
migrate-tool down --env=dev
# Edit migration
migrate-tool up --env=dev
```

**2. Test in Staging**

```bash
# Check staging status
migrate-tool status --env=staging

# Apply to staging
migrate-tool up --env=staging

# Verify application works in staging
```

**3. Deploy to Production**

```bash
# Check production status
migrate-tool status --env=prod

# Preview pending migrations
migrate-tool history --env=prod

# Apply (with confirmation prompt)
migrate-tool up --env=prod
```

## Quick Status Check

Check all environments at once:

**Linux/macOS:**
```bash
for env in dev staging prod; do
  echo "=== $env ==="
  migrate-tool status --env=$env
  echo
done
```

**Windows (PowerShell):**
```powershell
foreach ($env in @("dev", "staging", "prod")) {
    Write-Host "=== $env ==="
    migrate-tool status --env=$env
    Write-Host ""
}
```

Output:
```
=== dev ===
Environment: dev
Current Version: 5
Dirty: false
Applied: 5 / 5
Pending: 0

=== staging ===
Environment: staging
Current Version: 4
Dirty: false
Applied: 4 / 5
Pending: 1

=== prod ===
Environment: prod
Current Version: 4
Dirty: false
Applied: 4 / 5
Pending: 1
```

## Staged Rollouts

### Apply One Migration at a Time

Safer for production deployments:

```bash
# Check pending count
migrate-tool status --env=prod

# Apply one migration
migrate-tool up --steps=1 --env=prod

# Verify application
# ... test critical paths ...

# Continue with next
migrate-tool up --steps=1 --env=prod
```

### Quick Rollback Plan

If issues occur:

```bash
# Rollback the last migration
migrate-tool down --env=prod

# Verify rollback
migrate-tool status --env=prod
```

## Environment Parity

### Same Migrations, Different Data

All environments share the same `migrations_path`:

```yaml
environments:
  dev:
    migrations_path: "./migrations"
  staging:
    migrations_path: "./migrations"
  prod:
    migrations_path: "./migrations"
```

This ensures schema parity across environments.

### Sync Check

Verify environments are at same version:

```bash
# Get versions
DEV_VER=$(migrate-tool status --env=dev | grep "Current Version" | awk '{print $3}')
STAGING_VER=$(migrate-tool status --env=staging | grep "Current Version" | awk '{print $3}')
PROD_VER=$(migrate-tool status --env=prod | grep "Current Version" | awk '{print $3}')

echo "dev: $DEV_VER, staging: $STAGING_VER, prod: $PROD_VER"
```

## Confirmation Prompts

### require_confirmation

When enabled, prompts before migrations:

```yaml
environments:
  prod:
    require_confirmation: true
```

Behavior:
```
About to apply 2 migrations to production.
Continue? [y/N]: y
Applied 2 migration(s) successfully
```

### Auto-Approve (CI/CD)

Skip prompts in automated pipelines:

```bash
migrate-tool up --env=prod --auto-approve
```

Use with caution - typically in controlled CI/CD environments only.

## Best Practices

### 1. Test Migrations Before Production

Always test in dev and staging first:

```bash
# 1. Dev
migrate-tool up --env=dev
# Test locally

# 2. Staging
migrate-tool up --env=staging
# Integration tests

# 3. Production
migrate-tool up --env=prod
```

### 2. Use require_confirmation for Protected Environments

```yaml
environments:
  staging:
    require_confirmation: true
  prod:
    require_confirmation: true
```

### 3. Monitor After Deployment

After production migrations:
- Check application logs
- Monitor error rates
- Verify key functionality

### 4. Have a Rollback Plan

Before deploying:
- Know which version to rollback to
- Test rollback in staging
- Have `down` command ready

```bash
# Rollback command ready
migrate-tool down --env=prod
```

### 5. Document Breaking Changes

For migrations that require application changes:

```sql
-- Migration: remove_legacy_column
-- BREAKING: Requires app version 2.0+ deployed first
-- +migrate UP
ALTER TABLE users DROP COLUMN legacy_field;

-- +migrate DOWN
ALTER TABLE users ADD COLUMN legacy_field VARCHAR(100);
```

## Troubleshooting

### Environment Not Found

```
Error: environment 'stage' not found
```

Fix: Check environment name spelling
```bash
migrate-tool config show  # List available environments
```

### Different Versions Across Environments

Check status and sync:

```bash
# Check all environments
migrate-tool status --env=dev
migrate-tool status --env=staging
migrate-tool status --env=prod

# Apply missing migrations
migrate-tool up --env=staging
migrate-tool up --env=prod
```

## Next Steps

- [Troubleshooting](./06-troubleshooting.md) - Handle errors and dirty state
- [CI/CD Integration](./07-ci-cd-integration.md) - Automated deployments
- [CLI Reference](../cli-reference.md) - Complete command documentation
