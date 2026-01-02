---
layout: default
title: Home
---

# Janus Documentation

Cross-platform database migration CLI with single-file up/down support.

## Quick Links

- [User Guide](./user-guide/) - Complete tutorial from installation to CI/CD
- [CLI Reference](./cli-reference.md) - All commands and flags
- [Deployment Guide](./deployment-guide.md) - Installation scripts

## Getting Started

```bash
# Install
curl -sSL https://raw.githubusercontent.com/cesc1802/migration-tool/master/scripts/install.sh | sh

# Run migrations
janus up --env=dev
```

## Features

- Multi-environment configuration (dev, staging, prod)
- Single file migrations with up/down sections
- PostgreSQL, MySQL, SQLite3 support
- CI/CD ready with auto-approve mode

[Get Started →](./user-guide/01-getting-started.md)
