# Radius Bicep Types

This directory contains the Bicep type definitions for Radius, generated from the OpenAPI specifications in `/swagger/specification/`.

## Directory Structure

```
hack/bicep-types-radius/
├── generated/                          # Generated Bicep type definitions
│   ├── index.json                      # All types (backward compatibility)
│   ├── applications-index.json         # Deprecated Applications.* types
│   ├── radius-index.json              # Current Radius.* types
│   ├── index.md                       # Documentation for all types
│   ├── applications-index.md          # Documentation for deprecated types
│   ├── radius-index.md                # Documentation for current types
│   └── applications/                  # Applications.* type files
│   └── radius/                        # Radius.* type files
├── src/                                # Source code for type generation
│   ├── autorest.bicep/                # AutoRest extension for Bicep
│   └── generator/                     # Generator CLI tool
├── DEPRECATION.md                      # Deprecation guide
└── README.md                           # This file
```

## Type Indices

### Main Index (`index.json`)

Contains all resource types from both Applications.* and Radius.* namespaces. Used for backward compatibility.

Published to: `br:biceptypes.azurecr.io/radius:latest`

### Applications Index (`applications-index.json`) - **DEPRECATED**

Contains only Applications.* resource types, marked as deprecated with `isDeprecated: true` in TypeSettings.

Published to: `br:biceptypes.azurecr.io/applications:latest`

When configured, Bicep IntelliSense in VS Code will show deprecation warnings for all Applications.* types.

### Radius Index (`radius-index.json`)

Contains only the current Radius.* resource types.

Currently only used internally. The main index.json is used for the Radius extension.

## Building and Generating Types

### Prerequisites

- Node.js 20+
- npm
- AutoRest CLI

### Build Steps

1. **Build bicep-types library:**
   ```bash
   cd ../../bicep-types/src/bicep-types
   npm install
   npm run build
   ```

2. **Build autorest.bicep extension:**
   ```bash
   cd src/autorest.bicep
   npm install
   npm run build
   ```

3. **Build generator:**
   ```bash
   cd src/generator
   npm install
   npm run build
   ```

4. **Generate types:**
   ```bash
   cd src/generator
   npm run generate -- \
     --specs-dir /path/to/radius/swagger \
     --out-dir ../../generated
   ```

This will generate:
- Type JSON files for each API version
- Three index files (index.json, applications-index.json, radius-index.json)
- Markdown documentation files

## Publishing Extensions

Extensions are published automatically by the CI/CD pipeline in `.github/workflows/build.yaml` when:
- Pushing to `main` branch (published as `latest`)
- Creating version tags (published with the tag version)

Manual publishing (for testing):
```bash
bicep publish-extension ./generated/index.json \
  --target br:biceptypes.azurecr.io/radius:latest \
  --force

bicep publish-extension ./generated/applications-index.json \
  --target br:biceptypes.azurecr.io/applications:latest \
  --force
```

## Using the Extensions

### Current Radius Types (Recommended)

```json
{
  "experimentalFeaturesEnabled": {
    "extensibility": true
  },
  "extensions": {
    "radius": "br:biceptypes.azurecr.io/radius:latest"
  }
}
```

### Deprecated Applications Types (for backward compatibility)

To see deprecation warnings for Applications.* types:

```json
{
  "experimentalFeaturesEnabled": {
    "extensibility": true
  },
  "extensions": {
    "applications": "br:biceptypes.azurecr.io/applications:latest",
    "radius": "br:biceptypes.azurecr.io/radius:latest"
  }
}
```

## Deprecation

The Applications.* namespace is deprecated. See [DEPRECATION.md](./DEPRECATION.md) for:
- Migration guide
- Resource type mapping
- Timeline
- Support information

## Source Code

The type generation is based on:
- **Bicep Types Library**: `../../bicep-types` (git submodule from Azure/bicep-types)
- **AutoRest Extension**: `src/autorest.bicep` (forked from Azure/bicep-types-az)
- **Generator CLI**: `src/generator` (custom tool for Radius)

## More Information

- [Bicep Documentation](https://learn.microsoft.com/en-us/azure/azure-resource-manager/bicep/)
- [Radius Documentation](https://docs.radapp.io/)
- [Azure Bicep Types Repository](https://github.com/Azure/bicep-types)
