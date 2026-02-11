# Automated Resource Type Sync

This directory contains the automation for syncing resource types from the [resource-types-contrib](https://github.com/radius-project/resource-types-contrib) repository.

## How It Works

### Configuration (Author-Driven)

The sync configuration is maintained in the **resource-types-contrib** repository (not here) as `.radius-sync-config.yaml` at the repository root. This makes it author-driven - resource type authors can mark their types for default registration when they create or update them.

See `.radius-sync-config.yaml.example` in this repository for the expected format.

### Sync Script

The `.github/scripts/sync-resource-types.sh` script:
1. Fetches the configuration file from resource-types-contrib repository
2. Reads the list of resource types marked for default registration
3. Fetches each resource type YAML from resource-types-contrib
4. Validates YAML syntax
5. Copies files to `deploy/manifest/built-in-providers/dev` and `deploy/manifest/built-in-providers/self-hosted`
6. Detects changes via file diff

### GitHub Workflow

The `.github/workflows/sync-resource-types.yaml` workflow:
- Runs weekly on Sundays at 00:00 UTC
- Can be triggered manually (commented out for production)
- Automatically creates a pull request when changes are detected

## Usage

### Manual Sync

To manually sync resource types:

```bash
# From the repository root
./.github/scripts/sync-resource-types.sh

# Override source repository or branch
./.github/scripts/sync-resource-types.sh --source-repo <owner/repo> --source-branch <branch>
```

### Adding New Resource Types for Default Registration

1. In the **resource-types-contrib** repository, update `.radius-sync-config.yaml`
2. Add your resource type to the `resourceTypes` list:
   ```yaml
   resourceTypes:
     - namespace: <namespace>  # e.g., Data, Compute, Security
       name: <resourceTypeName>  # e.g., redisCaches
       file: <filename>.yaml     # e.g., redisCaches.yaml
   ```
3. Commit and merge the change
4. The next sync run will automatically include your resource type

## Configuration File Format

The `.radius-sync-config.yaml` file in resource-types-contrib should have this structure:

```yaml
# Target directories in the radius repository
targetDirectories:
  - deploy/manifest/built-in-providers/dev
  - deploy/manifest/built-in-providers/self-hosted

# Resource types to sync
resourceTypes:
  - namespace: Data
    name: mySqlDatabases
    file: mySqlDatabases.yaml
```

## Benefits of This Approach

- **Author-Driven**: Resource type authors control which types are synced
- **Decentralized**: Configuration lives with the resource types
- **Simple**: Just update the config file in resource-types-contrib
- **Transparent**: All changes tracked in resource-types-contrib commits
- **Automatic**: Weekly sync ensures types stay up-to-date
