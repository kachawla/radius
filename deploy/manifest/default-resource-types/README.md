# Default Resource Types

This directory contains resource type definitions that are automatically synced from the [resource-types-contrib](https://github.com/radius-project/resource-types-contrib) repository.

## Overview

Resource types in this directory are automatically registered by default when Radius is deployed. These are community-contributed resource types that have been marked for default registration.

## Auto-Sync Process

The resource type YAML files in this directory are automatically synchronized from the resource-types-contrib repository through a GitHub Actions workflow.

### Configuration

The sync configuration is defined in `.github/resource-types-sync-config.yaml`, which specifies:
- Which resource types to sync from resource-types-contrib
- Where to place them in this directory

### Workflow

The sync workflow (`.github/workflows/sync-resource-types.yaml`):
- Runs on a schedule (weekly)
- Can be triggered manually via workflow_dispatch
- Detects changes in the configured resource types
- Creates a PR automatically when changes are detected

## Manual Sync

To manually trigger a sync:
1. Go to the Actions tab in the GitHub repository
2. Select the "Sync Resource Types from resource-types-contrib" workflow
3. Click "Run workflow"

## Adding New Resource Types

To add a new resource type for automatic sync:
1. Edit `.github/resource-types-sync-config.yaml`
2. Add the resource type path under the `resourceTypes` section
3. Commit and push the changes
4. The next sync run will include the new resource type

## File Structure

Each resource type YAML file follows the Radius resource type manifest format:
```yaml
namespace: <Namespace>
types:
  <resourceTypeName>:
    apiVersions:
      <version>:
        schema: <JSON schema>
```

## Important Notes

- **Do not manually edit files in this directory** - they will be overwritten by the sync process
- Any manual changes should be made in the upstream [resource-types-contrib](https://github.com/radius-project/resource-types-contrib) repository
- If you need to customize a resource type, fork it or propose changes to the upstream repository
