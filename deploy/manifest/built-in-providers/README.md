# Built-in Resource Providers

This directory contains resource type manifests for built-in resource providers in Radius.

## Directory Structure

- `self-hosted/` - Resource type manifests for self-hosted Radius installations
- `dev/` - Resource type manifests for development/testing environments

## File Types

### Manually Maintained Files

Files without the `synced_` prefix are manually maintained and represent core Radius resource types:

- `applications_core.yaml` - Core application resource types
- `applications_dapr.yaml` - Dapr integration resource types
- `applications_datastores.yaml` - Datastore resource types
- `applications_messaging.yaml` - Messaging resource types
- `microsoft_resources.yaml` - Microsoft.Resources provider
- `radius_compute.yaml` - Compute resource types (containers, etc.)
- `radius_core.yaml` - Core Radius resource types
- `radius_security.yaml` - Security resource types

### Auto-Synced Files

Files with the `synced_` prefix are automatically synced from the [resource-types-contrib](https://github.com/radius-project/resource-types-contrib) repository:

- `synced_*.yaml` - Resource types marked for default registration

⚠️ **Do not manually edit `synced_*` files.** Changes should be made in resource-types-contrib and will be automatically synced.

## How Auto-Sync Works

1. Resource types in resource-types-contrib are marked with `defaultRegistration: true`
2. A GitHub Actions workflow runs daily to check for updates
3. Marked resource types are synced to this directory with the `synced_` prefix
4. A pull request is created with the changes for review
5. Once merged, the resource types are automatically registered by UCP at startup

See [Resource Type Sync Documentation](../../../docs/contributing/resource-type-sync.md) for more details.

## Adding New Resource Types

### Core Radius Resource Types (Manual)

To add a new core resource type:

1. Create a YAML file in the appropriate subdirectory (`self-hosted/` or `dev/`)
2. Follow the manifest format defined in `pkg/cli/manifest/manifest.go`
3. Ensure the file is validated before committing
4. Submit a pull request

### Default Registration Resource Types (Auto-Sync)

To add a resource type for default registration:

1. Create the resource type in [resource-types-contrib](https://github.com/radius-project/resource-types-contrib)
2. Add `defaultRegistration: true` to the YAML file
3. The sync workflow will automatically copy it here with the `synced_` prefix

## Integration with Radius

These manifests are registered when Radius starts:

1. UCP configuration sets `manifestDirectory: /manifest/built-in-providers`
2. The directory is mounted into the UCP container
3. UCP initializer service reads all YAML files in the directory
4. Resource types are registered and become available

## Related Documentation

- [Resource Type Sync Mechanism](../../../docs/contributing/resource-type-sync.md)
- [Resource Type Manifest Format](../../../pkg/cli/manifest/manifest.go)
- [UCP Initializer Service](../../../pkg/ucp/initializer/service.go)
