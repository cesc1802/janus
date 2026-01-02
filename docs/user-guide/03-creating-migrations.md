# Creating Migrations

This guide covers how to create and structure migration files.

## The create Command

Generate a new migration file:

```bash
janus create <name>
```

### Examples

```bash
# Create a users table migration
janus create create_users

# Add a column
janus create add_email_to_users

# Create an index
janus create add_email_index
```

### Output

```
Created: migrations/000001_create_users.sql
```

### Naming Conventions

Input names are sanitized automatically:
- `create users` → `create_users`
- `add-email` → `add_email`
- `Create_Users_Table` → `create_users_table`

## Migration File Format

### Basic Structure

```sql
-- Migration: create_users
-- Created: 2026-01-02

-- +migrate UP
-- SQL to apply the migration

-- +migrate DOWN
-- SQL to reverse the migration
```

### Section Markers

- `-- +migrate UP` - SQL executed when running `up`
- `-- +migrate DOWN` - SQL executed when running `down`

Both sections are optional but recommended.

## Writing UP Migrations

The UP section contains SQL to apply your changes.

### Create Table

```sql
-- +migrate UP
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) NOT NULL UNIQUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### Add Column

```sql
-- +migrate UP
ALTER TABLE users ADD COLUMN name VARCHAR(100);
```

### Create Index

```sql
-- +migrate UP
CREATE INDEX idx_users_email ON users(email);
```

### Multiple Statements

```sql
-- +migrate UP
CREATE TABLE posts (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id),
    title VARCHAR(255) NOT NULL,
    content TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_posts_user_id ON posts(user_id);
CREATE INDEX idx_posts_created_at ON posts(created_at);
```

## Writing DOWN Migrations

The DOWN section reverses the UP changes. Order should be reversed.

### Drop Table

```sql
-- +migrate DOWN
DROP TABLE users;
```

### Remove Column

```sql
-- +migrate DOWN
ALTER TABLE users DROP COLUMN name;
```

### Drop Index

```sql
-- +migrate DOWN
DROP INDEX idx_users_email;
```

### Multiple Statements (Reversed Order)

```sql
-- +migrate DOWN
DROP INDEX idx_posts_created_at;
DROP INDEX idx_posts_user_id;
DROP TABLE posts;
```

## Complete Example

Full migration with UP and DOWN:

```sql
-- Migration: create_users
-- Created: 2026-01-02

-- +migrate UP
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) NOT NULL UNIQUE,
    name VARCHAR(100),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_users_email ON users(email);

-- +migrate DOWN
DROP INDEX idx_users_email;
DROP TABLE users;
```

## Versioning

### Sequential Versioning (Default)

Files are numbered sequentially: `000001`, `000002`, `000003`

```bash
janus create create_users
# Creates: 000001_create_users.sql

janus create create_posts
# Creates: 000002_create_posts.sql
```

### Timestamp Versioning

Use `--seq=false` for timestamp-based versions:

```bash
janus create add_email --seq=false
# Creates: 20260102150405_add_email.sql
```

## Best Practices

### 1. Make Migrations Reversible

Always write DOWN migrations:

```sql
-- +migrate UP
ALTER TABLE users ADD COLUMN phone VARCHAR(20);

-- +migrate DOWN
ALTER TABLE users DROP COLUMN phone;
```

### 2. One Change Per Migration

Keep migrations focused:

```bash
# Good - separate migrations
janus create create_users
janus create create_posts
janus create add_user_phone

# Bad - too many changes in one
janus create create_all_tables_and_add_phone
```

### 3. Use Descriptive Names

Names should describe the change:

```bash
# Good
janus create add_email_unique_constraint
janus create remove_legacy_columns

# Bad
janus create update_users
janus create fix
```

### 4. Test Rollbacks

Verify DOWN migrations work:

```bash
janus up --env=dev    # Apply
janus down --env=dev  # Rollback
janus up --env=dev    # Apply again
```

### 5. Handle Data Migrations Carefully

For data changes, consider:

```sql
-- +migrate UP
-- Add column with default
ALTER TABLE users ADD COLUMN status VARCHAR(20) DEFAULT 'active';

-- Update existing rows
UPDATE users SET status = 'active' WHERE status IS NULL;

-- Make NOT NULL after update
ALTER TABLE users ALTER COLUMN status SET NOT NULL;

-- +migrate DOWN
ALTER TABLE users DROP COLUMN status;
```

## Database-Specific Syntax

### PostgreSQL

```sql
-- +migrate UP
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) NOT NULL,
    data JSONB,
    created_at TIMESTAMPTZ DEFAULT NOW()
);
```

### MySQL

```sql
-- +migrate UP
CREATE TABLE users (
    id INT AUTO_INCREMENT PRIMARY KEY,
    email VARCHAR(255) NOT NULL,
    data JSON,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

### SQLite3

```sql
-- +migrate UP
CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    email TEXT NOT NULL,
    created_at TEXT DEFAULT CURRENT_TIMESTAMP
);
```

## Validating Migrations

Check for syntax errors before running:

```bash
janus validate --env=dev
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

Warnings for empty sections:
```
WARNINGS:
  ! Env dev: 1 migration(s) with empty UP section
  ! Env dev: 2 migration(s) with empty DOWN section
```

## Next Steps

- [Running Migrations](./04-running-migrations.md) - Apply and rollback migrations
- [Multi-Environment](./05-multi-environment.md) - Manage multiple environments
- [CLI Reference](../cli-reference.md) - Complete command documentation
