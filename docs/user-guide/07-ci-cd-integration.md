# CI/CD Integration

This guide covers integrating migrate-tool into automated deployment pipelines.

## GitHub Actions

### Basic Workflow

`.github/workflows/migrate.yml`:

```yaml
name: Database Migration

on:
  push:
    branches: [main]
  workflow_dispatch:

jobs:
  migrate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Install migrate-tool
        run: |
          curl -sSL https://raw.githubusercontent.com/cesc1802/migration-tool/master/scripts/install.sh | sh

      - name: Run migrations
        env:
          DATABASE_URL: ${{ secrets.DATABASE_URL }}
        run: |
          migrate-tool up --env=prod --auto-approve
```

### With Validation

```yaml
name: Database Migration

on:
  push:
    branches: [main]

jobs:
  migrate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Install migrate-tool
        run: |
          curl -sSL https://raw.githubusercontent.com/cesc1802/migration-tool/master/scripts/install.sh | sh

      - name: Validate migrations
        run: |
          migrate-tool validate --env=prod

      - name: Check status
        env:
          DATABASE_URL: ${{ secrets.DATABASE_URL }}
        run: |
          migrate-tool status --env=prod

      - name: Run migrations
        env:
          DATABASE_URL: ${{ secrets.DATABASE_URL }}
        run: |
          migrate-tool up --env=prod --auto-approve
```

### PR Validation Only

Run validation on PRs without applying:

```yaml
name: Validate Migrations

on:
  pull_request:
    paths:
      - 'migrations/**'
      - 'migrate-tool.yaml'

jobs:
  validate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Install migrate-tool
        run: |
          curl -sSL https://raw.githubusercontent.com/cesc1802/migration-tool/master/scripts/install.sh | sh

      - name: Validate migrations
        run: |
          migrate-tool validate
```

### Staged Deployment

Deploy to staging first, then production:

```yaml
name: Staged Migration

on:
  push:
    branches: [main]

jobs:
  staging:
    runs-on: ubuntu-latest
    environment: staging
    steps:
      - uses: actions/checkout@v4

      - name: Install migrate-tool
        run: |
          curl -sSL https://raw.githubusercontent.com/cesc1802/migration-tool/master/scripts/install.sh | sh

      - name: Migrate staging
        env:
          DATABASE_URL: ${{ secrets.STAGING_DATABASE_URL }}
        run: |
          migrate-tool up --env=staging --auto-approve

  production:
    needs: staging
    runs-on: ubuntu-latest
    environment: production
    steps:
      - uses: actions/checkout@v4

      - name: Install migrate-tool
        run: |
          curl -sSL https://raw.githubusercontent.com/cesc1802/migration-tool/master/scripts/install.sh | sh

      - name: Migrate production
        env:
          DATABASE_URL: ${{ secrets.PROD_DATABASE_URL }}
        run: |
          migrate-tool up --env=prod --auto-approve
```

## GitLab CI

### Basic Pipeline

`.gitlab-ci.yml`:

```yaml
stages:
  - validate
  - migrate

variables:
  MIGRATE_TOOL_VERSION: "latest"

.install_migrate_tool: &install_migrate_tool
  before_script:
    - curl -sSL https://raw.githubusercontent.com/cesc1802/migration-tool/master/scripts/install.sh | sh

validate:
  stage: validate
  <<: *install_migrate_tool
  script:
    - migrate-tool validate
  rules:
    - if: $CI_MERGE_REQUEST_ID
      changes:
        - migrations/**
        - migrate-tool.yaml

migrate:staging:
  stage: migrate
  <<: *install_migrate_tool
  script:
    - migrate-tool up --env=staging --auto-approve
  environment:
    name: staging
  rules:
    - if: $CI_COMMIT_BRANCH == "main"

migrate:production:
  stage: migrate
  <<: *install_migrate_tool
  script:
    - migrate-tool up --env=prod --auto-approve
  environment:
    name: production
  needs:
    - migrate:staging
  rules:
    - if: $CI_COMMIT_BRANCH == "main"
      when: manual
```

### With Rollback

```yaml
migrate:production:
  stage: migrate
  <<: *install_migrate_tool
  script:
    - migrate-tool status --env=prod
    - migrate-tool up --env=prod --auto-approve
  environment:
    name: production
  rules:
    - if: $CI_COMMIT_BRANCH == "main"
      when: manual

rollback:production:
  stage: migrate
  <<: *install_migrate_tool
  script:
    - migrate-tool down --env=prod --auto-approve
  environment:
    name: production
  rules:
    - if: $CI_COMMIT_BRANCH == "main"
      when: manual
```

## Best Practices

### 1. Use Environment-Specific Secrets

GitHub:
```yaml
env:
  DATABASE_URL: ${{ secrets.PROD_DATABASE_URL }}
```

GitLab:
```yaml
variables:
  DATABASE_URL: $PROD_DATABASE_URL
```

### 2. Validate Before Applying

Always validate first:

```yaml
- name: Validate migrations
  run: migrate-tool validate --env=prod

- name: Run migrations
  run: migrate-tool up --env=prod --auto-approve
```

### 3. Show Status for Visibility

Log migration status:

```yaml
- name: Pre-migration status
  run: migrate-tool status --env=prod

- name: Run migrations
  run: migrate-tool up --env=prod --auto-approve

- name: Post-migration status
  run: migrate-tool status --env=prod
```

### 4. Use --auto-approve in CI

Skip interactive prompts:

```bash
migrate-tool up --env=prod --auto-approve
```

### 5. Pin Versions in Production

Use specific version instead of latest:

```yaml
- name: Install migrate-tool
  run: |
    curl -sSL https://raw.githubusercontent.com/cesc1802/migration-tool/master/scripts/install.sh | sh -s -- --version v1.0.0
```

### 6. Use Environment Protection

GitHub Environments:
```yaml
jobs:
  production:
    environment: production  # Requires approval
```

GitLab Protected Environments:
```yaml
migrate:production:
  environment:
    name: production
  rules:
    - if: $CI_COMMIT_BRANCH == "main"
      when: manual  # Manual approval
```

### 7. Cache Installation

GitHub:
```yaml
- name: Cache migrate-tool
  uses: actions/cache@v4
  with:
    path: /usr/local/bin/migrate-tool
    key: migrate-tool-${{ runner.os }}-v1.0.0

- name: Install migrate-tool
  run: |
    if ! command -v migrate-tool &> /dev/null; then
      curl -sSL https://raw.githubusercontent.com/cesc1802/migration-tool/master/scripts/install.sh | sh -s -- --version v1.0.0
    fi
```

## Complete Example

Full production-ready workflow:

```yaml
name: Database Migration

on:
  push:
    branches: [main]
  pull_request:
    paths:
      - 'migrations/**'
      - 'migrate-tool.yaml'

env:
  MIGRATE_TOOL_VERSION: "v1.0.0"

jobs:
  validate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Install migrate-tool
        run: |
          curl -sSL https://raw.githubusercontent.com/cesc1802/migration-tool/master/scripts/install.sh | sh -s -- --version ${{ env.MIGRATE_TOOL_VERSION }}

      - name: Validate migrations
        run: migrate-tool validate

  staging:
    needs: validate
    if: github.event_name == 'push' && github.ref == 'refs/heads/main'
    runs-on: ubuntu-latest
    environment: staging
    steps:
      - uses: actions/checkout@v4

      - name: Install migrate-tool
        run: |
          curl -sSL https://raw.githubusercontent.com/cesc1802/migration-tool/master/scripts/install.sh | sh -s -- --version ${{ env.MIGRATE_TOOL_VERSION }}

      - name: Pre-migration status
        env:
          DATABASE_URL: ${{ secrets.STAGING_DATABASE_URL }}
        run: migrate-tool status --env=staging

      - name: Run migrations
        env:
          DATABASE_URL: ${{ secrets.STAGING_DATABASE_URL }}
        run: migrate-tool up --env=staging --auto-approve

      - name: Post-migration status
        env:
          DATABASE_URL: ${{ secrets.STAGING_DATABASE_URL }}
        run: migrate-tool status --env=staging

  production:
    needs: staging
    runs-on: ubuntu-latest
    environment: production
    steps:
      - uses: actions/checkout@v4

      - name: Install migrate-tool
        run: |
          curl -sSL https://raw.githubusercontent.com/cesc1802/migration-tool/master/scripts/install.sh | sh -s -- --version ${{ env.MIGRATE_TOOL_VERSION }}

      - name: Pre-migration status
        env:
          DATABASE_URL: ${{ secrets.PROD_DATABASE_URL }}
        run: migrate-tool status --env=prod

      - name: Run migrations
        env:
          DATABASE_URL: ${{ secrets.PROD_DATABASE_URL }}
        run: migrate-tool up --env=prod --auto-approve

      - name: Post-migration status
        env:
          DATABASE_URL: ${{ secrets.PROD_DATABASE_URL }}
        run: migrate-tool status --env=prod
```

## Rollback Automation

### Manual Rollback Job

```yaml
rollback:
  if: github.event_name == 'workflow_dispatch'
  runs-on: ubuntu-latest
  environment: production
  steps:
    - uses: actions/checkout@v4

    - name: Install migrate-tool
      run: |
        curl -sSL https://raw.githubusercontent.com/cesc1802/migration-tool/master/scripts/install.sh | sh

    - name: Rollback
      env:
        DATABASE_URL: ${{ secrets.PROD_DATABASE_URL }}
      run: migrate-tool down --env=prod --auto-approve
```

### Trigger via CLI

```bash
gh workflow run rollback.yml
```

## Next Steps

- [CLI Reference](../cli-reference.md) - Complete command documentation
- [Troubleshooting](./06-troubleshooting.md) - Handle errors and dirty state
- [Deployment Guide](../deployment-guide.md) - Installation details
