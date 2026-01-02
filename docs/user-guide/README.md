<div align="center">
  <img src="{{ '/assets/logo/janus-roman-pillar.svg' | relative_url }}" alt="Janus Logo" width="96" height="96">
</div>

# Janus User Guide

Welcome to the Janus user guide. This tutorial-style documentation covers everything from installation to CI/CD integration.

## Prerequisites

- PostgreSQL, MySQL, or SQLite3 database
- Terminal access (Linux/macOS) or PowerShell (Windows)
- Basic SQL knowledge

## Guide Structure

| Chapter | Topic | Description |
|---------|-------|-------------|
| [01](./01-getting-started.md) | Getting Started | Install, verify, run first migration |
| [02](./02-configuration.md) | Configuration | Config file, env vars, database URLs |
| [03](./03-creating-migrations.md) | Creating Migrations | Create command, SQL syntax, best practices |
| [04](./04-running-migrations.md) | Running Migrations | up/down/status/history/goto commands |
| [05](./05-multi-environment.md) | Multi-Environment | Dev/staging/prod workflows |
| [06](./06-troubleshooting.md) | Troubleshooting | Dirty state, force, common errors |
| [07](./07-ci-cd-integration.md) | CI/CD Integration | GitHub Actions, GitLab CI examples |

## Quick Links

- [CLI Reference](../cli-reference.md) - Complete command documentation
- [Deployment Guide](../deployment-guide.md) - Installation scripts details

## Sample Project

All examples use a consistent sample project with a `users` table:

```sql
-- Sample migration used throughout this guide
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) NOT NULL UNIQUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

## Next Steps

Start with [Getting Started](./01-getting-started.md) to install Janus and run your first migration.
