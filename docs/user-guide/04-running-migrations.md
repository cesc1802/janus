# Running Migrations

This guide covers all commands for applying and managing migrations.

## Check Status

View current migration state:

```bash
migrate-tool status --env=dev
```

Output:
```
Environment: dev
Current Version: 3
Dirty: false
Applied: 3 / 5
Pending: 2
```

### Status Fields

| Field | Description |
|-------|-------------|
| Environment | Active environment name |
| Current Version | Last applied migration version |
| Dirty | `true` if migration failed mid-execution |
| Applied | Count of applied / total migrations |
| Pending | Migrations waiting to be applied |

## Apply Migrations (up)

### Apply All Pending

```bash
migrate-tool up --env=dev
```

Output:
```
Applied 3 migration(s) successfully
Current version: 3
```

### Apply Specific Count

Use `--steps` to limit migrations applied:

```bash
# Apply next 2 migrations only
migrate-tool up --steps=2 --env=dev
```

Output:
```
Applied 2 migration(s) successfully
Current version: 2
```

### No Pending Migrations

```bash
migrate-tool up --env=dev
```

Output:
```
No pending migrations
Current version: 3
```

## Rollback Migrations (down)

### Rollback One (Default)

Safety default - only rolls back one migration:

```bash
migrate-tool down --env=dev
```

Output:
```
Rolled back 1 migration(s)
Current version: 2
```

### Rollback Multiple

Explicit `--steps` required for larger rollbacks:

```bash
# Rollback 3 migrations
migrate-tool down --steps=3 --env=dev
```

Output:
```
Rolled back 3 migration(s)
Current version: 0
```

### Rollback to Base

```bash
migrate-tool down --steps=99 --env=dev
```

Output:
```
Rolled back 5 migration(s)
Current version: none (clean slate)
```

## View History

List migrations with applied status:

```bash
migrate-tool history --env=dev
```

Output:
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

### Show More

```bash
# Show last 20 migrations
migrate-tool history --limit=20 --env=dev

# Show all migrations
migrate-tool history --limit=999 --env=dev
```

## Go to Specific Version (goto)

Migrate to a target version (up or down):

### Migrate Up to Version

```bash
migrate-tool goto 10 --env=dev
```

Output:
```
┌─────────────────────────────────────────┐
│  Migration Target                       │
└─────────────────────────────────────────┘
Environment: dev
Current version: 3
Target version: 10
Direction: UP (7 migration(s))

Migrated to version 10
```

### Rollback to Version

```bash
migrate-tool goto 5 --env=dev
```

Output:
```
┌─────────────────────────────────────────┐
│  Migration Target                       │
└─────────────────────────────────────────┘
Environment: dev
Current version: 10
Target version: 5
Direction: DOWN (5 migration(s))

Migrated to version 5
```

### Reset to Base

```bash
migrate-tool goto 0 --env=dev
```

Rolls back all migrations.

### Already at Version

```bash
migrate-tool goto 5 --env=dev
```

Output:
```
Already at version 5
```

## Common Workflows

### Fresh Database Setup

```bash
# Check configuration
migrate-tool config show --env=dev

# Verify migrations
migrate-tool validate --env=dev

# Apply all migrations
migrate-tool up --env=dev

# Confirm status
migrate-tool status --env=dev
```

### Staged Deployment

Apply migrations one at a time for safer rollout:

```bash
# Check pending
migrate-tool status --env=prod

# Apply one migration
migrate-tool up --steps=1 --env=prod

# Verify application works
# ... test your app ...

# Apply next migration
migrate-tool up --steps=1 --env=prod
```

### Quick Rollback

```bash
# Something went wrong, rollback last migration
migrate-tool down --env=prod

# Verify
migrate-tool status --env=prod
```

### Development Reset

Reset to specific state during development:

```bash
# Rollback to version 3
migrate-tool goto 3 --env=dev

# Make changes to migration 4
# Edit migrations/000004_xxx.sql

# Re-apply
migrate-tool up --env=dev
```

## Environment Examples

### Development

```bash
# Fast iteration - apply all
migrate-tool up --env=dev

# Quick rollback
migrate-tool down --env=dev
```

### Staging

```bash
# Verify before production
migrate-tool status --env=staging
migrate-tool up --env=staging
```

### Production

```bash
# Check status
migrate-tool status --env=prod

# Preview changes
migrate-tool history --env=prod

# Apply step by step
migrate-tool up --steps=1 --env=prod
```

## Command Summary

| Command | Purpose | Default |
|---------|---------|---------|
| `status` | View current state | - |
| `up` | Apply migrations | All pending |
| `down` | Rollback migrations | 1 migration |
| `history` | List migrations | Last 10 |
| `goto` | Go to version | - |

See [CLI Reference](../cli-reference.md) for complete flag documentation.

## Next Steps

- [Multi-Environment](./05-multi-environment.md) - Manage dev/staging/prod
- [Troubleshooting](./06-troubleshooting.md) - Handle dirty state and errors
- [CI/CD Integration](./07-ci-cd-integration.md) - Automated deployments
