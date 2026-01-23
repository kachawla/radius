# Automated Resource Type Sync

This document describes the automated synchronization of resource types from the [resource-types-contrib](https://github.com/radius-project/resource-types-contrib) repository.

## Overview

Radius automatically syncs certain community-contributed resource types from the resource-types-contrib repository to enable default registration. This means these resource types are available out-of-the-box when Radius is deployed, without requiring manual installation.

## How It Works

### Configuration

The resource types to sync are defined in `.github/resource-types-sync-config.yaml`:

```yaml
sourceRepo: radius-project/resource-types-contrib
sourceBranch: main
targetDirectory: deploy/manifest/default-resource-types

resourceTypes:
  - namespace: Data
    name: mySqlDatabases
    file: mySqlDatabases.yaml
    
  - namespace: Compute
    name: containers
    file: containers.yaml
```

### Sync Script

The `hack/sync-resource-types.sh` script:
1. Reads the configuration file
2. Fetches each resource type YAML from the resource-types-contrib repository
3. Validates the YAML syntax
4. Compares with existing files to detect changes
5. Updates files only if changes are detected

### GitHub Workflow

The `.github/workflows/sync-resource-types.yaml` workflow:
- Runs weekly on Sundays at 00:00 UTC
- Can be triggered manually via workflow_dispatch (during development)
- Detects changes and creates a pull request automatically
- Includes changed files in the PR for review

## Adding New Resource Types

To add a new resource type for automatic synchronization:

1. **Update the configuration file** (`.github/resource-types-sync-config.yaml`):
   ```yaml
   resourceTypes:
     - namespace: <namespace>  # e.g., Data, Compute, Security
       name: <resourceTypeName>  # e.g., redisCaches
       file: <filename>.yaml  # e.g., redisCaches.yaml
   ```

2. **Commit and push the changes**:
   ```bash
   git add .github/resource-types-sync-config.yaml
   git commit -m "Add <resourceType> to sync configuration"
   git push
   ```

3. **Trigger the sync** (optional):
   - Wait for the weekly scheduled run, or
   - Manually trigger the workflow from the GitHub Actions UI

4. **Review the PR**: The workflow will create a PR with the new resource type files

## Resource Type Loading

Resource types from the `deploy/manifest/default-resource-types/` directory are automatically loaded when Radius starts:

1. **UCP Configuration**: The `manifestDirectories` field in UCP configuration includes the default-resource-types directory
2. **Initialization**: The UCP initializer service loads and registers all YAML files from configured directories
3. **Docker Image**: The default-resource-types directory is included in the UCP Docker image

### Configuration Files

- **Development**: `cmd/ucpd/ucp-dev.yaml`
- **Build**: `build/configs/ucp.yaml`
- **Helm Chart**: `deploy/Chart/templates/ucp/configmaps.yaml`

Example configuration:
```yaml
initialization:
  manifestDirectory: "deploy/manifest/built-in-providers/dev"
  manifestDirectories:
    - "deploy/manifest/default-resource-types"
```

## Manual Sync

To manually sync resource types:

```bash
# From the repository root
./hack/sync-resource-types.sh

# With a custom config file
./hack/sync-resource-types.sh --config path/to/config.yaml

# Dry run (show what would be synced without making changes)
./hack/sync-resource-types.sh --dry-run
```

## Troubleshooting

### Sync script fails to fetch a resource type

**Symptom**: Error message "Failed to fetch <URL>"

**Solution**:
1. Verify the resource type exists in the resource-types-contrib repository
2. Check the namespace, name, and file name in the configuration
3. Ensure the resource-types-contrib repository is accessible

### Resource type not loaded in Radius

**Symptom**: Resource type not available after deployment

**Solution**:
1. Verify the YAML file exists in `deploy/manifest/default-resource-types/`
2. Check that the file is included in the Docker image build
3. Verify the UCP configuration includes the `manifestDirectories` setting
4. Check UCP logs for errors during initialization

### Workflow fails to create PR

**Symptom**: Sync workflow completes but no PR is created

**Solution**:
1. Check if there were actually any changes to sync
2. Verify GitHub Actions has permission to create pull requests
3. Check workflow logs for errors

## Development

When developing or testing changes to the sync mechanism:

1. **Uncomment `workflow_dispatch`** in `.github/workflows/sync-resource-types.yaml`:
   ```yaml
   on:
     schedule:
       - cron: "0 0 * * 0"
     workflow_dispatch:  # Uncomment this line
   ```

2. **Test the sync script locally**:
   ```bash
   ./hack/sync-resource-types.sh
   ```

3. **Comment out `workflow_dispatch` before merging** to prevent manual triggers in production

## Future Enhancements

Potential improvements to consider:

- Support for detecting new resource types automatically (without manual configuration)
- Integration with resource-types-contrib repository webhooks for real-time sync
- Automated testing of synced resource types
- Support for versioning and rollback of resource types
- Ability to exclude specific versions or branches
