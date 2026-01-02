# Getting Started

This guide walks you through installing Janus and running your first migration.

## Installation

### Quick Install

**Linux/macOS:**
```bash
curl -sSL https://raw.githubusercontent.com/cesc1802/migration-tool/master/scripts/install.sh | sh
```

**Windows (PowerShell):**
```powershell
irm https://raw.githubusercontent.com/cesc1802/migration-tool/master/scripts/install.ps1 | iex
```

### Install Specific Version

**Linux/macOS:**
```bash
curl -sSL https://raw.githubusercontent.com/cesc1802/migration-tool/master/scripts/install.sh | sh -s -- --version v1.0.0
```

**Windows (PowerShell):**
```powershell
irm https://raw.githubusercontent.com/cesc1802/migration-tool/master/scripts/install.ps1 -OutFile install.ps1
.\install.ps1 -Version v1.0.0
```

### Alternative Methods

**Build from source:**
```bash
go install github.com/cesc1802/janus/cmd/janus@latest
```

**Manual download:** Visit [GitHub Releases](https://github.com/cesc1802/migration-tool/releases)

See [Deployment Guide](../deployment-guide.md) for advanced installation options.

## Verify Installation

```bash
janus version
```

Expected output:
```
Janus 1.0.0
  commit: a1b2c3d
  built:  2026-01-01T10:30:00Z
  go:     go1.25.1
  os:     darwin/amd64
```

## First Migration

### 1. Create Config File

Create `janus.yaml` in your project root:

```yaml
environments:
  dev:
    database_url: "postgres://user:pass@localhost:5432/myapp_dev?sslmode=disable"
    migrations_path: "./migrations"

defaults:
  migrations_path: "./migrations"
```

### 2. Create Migrations Directory

**Linux/macOS:**
```bash
mkdir -p migrations
```

**Windows (PowerShell):**
```powershell
New-Item -ItemType Directory -Path migrations -Force
```

### 3. Create Your First Migration

```bash
janus create create_users
```

Output:
```
Created: migrations/000001_create_users.sql
```

### 4. Edit the Migration

Open `migrations/000001_create_users.sql` and add:

```sql
-- Migration: create_users
-- Created: 2026-01-02

-- +migrate UP
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) NOT NULL UNIQUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- +migrate DOWN
DROP TABLE users;
```

### 5. Check Status

```bash
janus status --env=dev
```

Output:
```
Environment: dev
Current Version: none (no migrations applied)
Dirty: false
Applied: 0 / 1
Pending: 1
```

### 6. Apply the Migration

```bash
janus up --env=dev
```

Output:
```
Applied 1 migration(s) successfully
Current version: 1
```

### 7. Verify

```bash
janus status --env=dev
```

Output:
```
Environment: dev
Current Version: 1
Dirty: false
Applied: 1 / 1
Pending: 0
```

## Project Structure

After setup, your project should look like:

```
my-project/
├── janus.yaml
└── migrations/
    └── 000001_create_users.sql
```

## Next Steps

- [Configuration](./02-configuration.md) - Learn about config options and environment variables
- [Creating Migrations](./03-creating-migrations.md) - SQL syntax and best practices
- [CLI Reference](../cli-reference.md) - Complete command documentation
