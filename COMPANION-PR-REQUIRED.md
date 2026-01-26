# Companion PR Required for resource-types-contrib

This PR implements fetching sync configuration from the resource-types-contrib repository.
A companion PR is required in the resource-types-contrib repository to add the configuration file.

## Required Action

Create a PR in [radius-project/resource-types-contrib](https://github.com/radius-project/resource-types-contrib) that adds `.radius-sync-config.yaml` at the repository root.

## File Content

Use the content from `.radius-sync-config.yaml.example` in this repository as a template:

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
    
  - namespace: Data
    name: postgreSqlDatabases
    file: postgreSqlDatabases.yaml

  - namespace: Compute
    name: containers
    file: containers.yaml
    
  - namespace: Compute
    name: persistentVolumes
    file: persistentVolumes.yaml
    
  - namespace: Compute
    name: routes
    file: routes.yaml
```

## Benefits

Once the companion PR is merged, resource type authors can update the configuration directly in their repository to mark types for default registration, making the process more author-driven and decentralized.

## Testing

Until the companion PR is merged in resource-types-contrib, the sync script will fail with an error message:
```
Failed to fetch config file from https://raw.githubusercontent.com/radius-project/resource-types-contrib/main/.radius-sync-config.yaml
Make sure .radius-sync-config.yaml exists in the radius-project/resource-types-contrib repository
```

This is expected and the workflow should be disabled or the companion PR should be merged before this PR is merged.
