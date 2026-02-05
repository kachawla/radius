# GitHub Configuration and Automation

This directory contains GitHub-specific configuration files, workflows, and automation scripts for the Radius project.

## Directory Structure

```
.github/
├── workflows/              # GitHub Actions workflows
│   ├── sync-resource-types.yaml  # Syncs resource types from resource-types-contrib
│   └── ...                # Other CI/CD workflows
├── scripts/               # Automation scripts
│   ├── sync-resource-types.py   # Resource type sync script
│   └── test_sync.py            # Tests for sync script
├── actions/               # Custom GitHub Actions
├── resource-type-sync-config.yaml  # Configuration for resource type sync
└── ...                    # Other GitHub config files
```

## Resource Type Sync Automation

### Overview

The resource type sync mechanism automatically manages default resource type registration in Radius by syncing definitions from the [resource-types-contrib](https://github.com/radius-project/resource-types-contrib) repository.

### Components

1. **Configuration** (`resource-type-sync-config.yaml`)
   - Defines source and target repositories
   - Specifies sync strategy and file patterns
   - Configures validation rules and PR settings

2. **Workflow** (`workflows/sync-resource-types.yaml`)
   - Runs daily at 2 AM UTC
   - Can be manually triggered for testing
   - Creates pull requests with detected changes

3. **Sync Script** (`scripts/sync-resource-types.py`)
   - Fetches resource type files via GitHub API
   - Validates manifest format
   - Syncs files with configured prefix

### How It Works

1. Resource types in resource-types-contrib are marked with `defaultRegistration: true`
2. Daily workflow checks for updates
3. Marked files are validated and synced to `deploy/manifest/built-in-providers/self-hosted/` with `synced_` prefix
4. Pull request is created if changes are detected
5. After review and merge, synced files are automatically registered by UCP

### Testing

Run the sync script tests:
```bash
python .github/scripts/test_sync.py
```

Test the sync script in dry-run mode:
```bash
export DRY_RUN=true
export CONFIG_FILE=.github/resource-type-sync-config.yaml
export SOURCE_REPO=radius-project/resource-types-contrib
python .github/scripts/sync-resource-types.py
```

### Documentation

- [Resource Type Sync Mechanism](../docs/contributing/resource-type-sync.md) - Technical documentation
- [Marking Resource Types for Default Registration](../docs/contributing/marking-resource-types-for-default-registration.md) - User guide

## Other Workflows

See individual workflow files in the `workflows/` directory for documentation on other CI/CD processes.

## Contributing

When adding new automation:

1. Add configuration files to this directory
2. Add scripts to the `scripts/` subdirectory
3. Add workflows to the `workflows/` subdirectory
4. Document in this README
5. Add tests for any scripts
6. Update relevant documentation

## Security

- All workflows use pinned action versions (commit SHA)
- Secrets are managed through GitHub Secrets
- External API calls use authentication tokens
- Validation is performed on all synced content
