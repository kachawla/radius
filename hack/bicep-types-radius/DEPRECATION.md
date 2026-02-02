# Deprecation of Applications.* Namespace

## Overview

The `Applications.*` namespace (including `Applications.Core`, `Applications.Dapr`, `Applications.Datastores`, and `Applications.Messaging`) has been deprecated in favor of the new `Radius.*` namespace.

## Migration Guide

### Old (Deprecated) Types

```bicep
extension radius

resource env 'Applications.Core/environments@2023-10-01-preview' = {
  name: 'myenv'
  properties: {
    compute: {
      kind: 'kubernetes'
      namespace: 'default'
    }
  }
}

resource app 'Applications.Core/applications@2023-10-01-preview' = {
  name: 'myapp'
  properties: {
    environment: env.id
  }
}
```

### New (Current) Types

```bicep
extension radius

resource env 'Radius.Core/environments@2025-08-01-preview' = {
  name: 'myenv'
  properties: {
    compute: {
      kind: 'kubernetes'
      namespace: 'default'
    }
  }
}

resource app 'Radius.Core/applications@2025-08-01-preview' = {
  name: 'myapp'
  properties: {
    environment: env.id
  }
}
```

## Using the Deprecated Extension

If you need to continue using the deprecated `Applications.*` types, you can configure your `bicepconfig.json` to use the `applications-index.json` extension:

```json
{
  "experimentalFeaturesEnabled": {
    "extensibility": true
  },
  "extensions": {
    "applications": "br:biceptypes.azurecr.io/applications:latest"
  }
}
```

**Note:** When using the deprecated extension, Bicep IntelliSense in VS Code will show deprecation warnings for all `Applications.*` resource types, indicating that you should migrate to the `Radius.*` namespace.

## Extension Files

The Bicep types are now published in three variants:

1. **`index.json`** - Contains all resource types (both `Applications.*` and `Radius.*`) for backward compatibility
2. **`applications-index.json`** - Contains only `Applications.*` resource types, marked as deprecated with `isDeprecated: true`
3. **`radius-index.json`** - Contains only `Radius.*` resource types (current)

## Resource Type Mapping

| Deprecated Type | New Type | API Version |
|----------------|----------|-------------|
| `Applications.Core/applications` | `Radius.Core/applications` | `2025-08-01-preview` |
| `Applications.Core/environments` | `Radius.Core/environments` | `2025-08-01-preview` |
| `Applications.Core/containers` | _(removed)_ | N/A |
| `Applications.Core/gateways` | _(removed)_ | N/A |
| `Applications.Core/volumes` | _(removed)_ | N/A |
| `Applications.Core/secretStores` | _(removed)_ | N/A |
| `Applications.Core/extenders` | _(removed)_ | N/A |
| `Applications.Dapr/*` | _(removed)_ | N/A |
| `Applications.Datastores/*` | _(removed)_ | N/A |
| `Applications.Messaging/*` | _(removed)_ | N/A |

**Note:** Many resource types have been removed in the new namespace. Please refer to the Radius documentation for the current resource model and migration strategies.

## How Deprecation Warnings Work

When you use the `applications-index.json` extension (which has `isDeprecated: true` in its settings), Bicep's IntelliSense in Visual Studio Code will display deprecation warnings for all resource types in the `Applications.*` namespace. This helps developers identify deprecated types and migrate to the current `Radius.*` namespace.

The deprecation flag is set at the namespace level in the TypeSettings, which means all resource types under the `Applications.*` namespace will show deprecation warnings.

## Timeline

- **Current:** Deprecated types are still functional but marked as deprecated
- **Future:** The deprecated `Applications.*` namespace may be removed in a future version of Radius

## Support

For questions or migration assistance, please:
- Consult the [Radius Documentation](https://docs.radapp.io/)
- Open an issue in the [Radius GitHub Repository](https://github.com/radius-project/radius)
