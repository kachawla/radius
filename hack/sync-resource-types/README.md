# Resource Type Sync Tool

This tool synchronizes resource type manifests from the resource-types-contrib repository to the Radius built-in-providers directory.

## Overview

The tool automatically identifies manifests marked for default registration (with `defaultRegistration: true`) and copies them to the Radius repository, stripping out the sync-specific metadata.

## Usage

```bash
go run ./hack/sync-resource-types/main.go \
  --source <path-to-resource-types-contrib> \
  --target <path-to-built-in-providers>
```

### Options

- `--source`: Source directory containing resource type manifests (required)
- `--target`: Target directory for synced manifests (required)
- `--dry-run`: Print actions without making changes
- `--verbose`: Enable verbose output

### Examples

**Dry run to see what would be synced:**
```bash
go run ./hack/sync-resource-types/main.go \
  --source ../resource-types-contrib/manifests \
  --target ./deploy/manifest/built-in-providers/dev \
  --dry-run --verbose
```

**Actual sync:**
```bash
go run ./hack/sync-resource-types/main.go \
  --source ../resource-types-contrib/manifests \
  --target ./deploy/manifest/built-in-providers/dev
```

## How It Works

1. **Discovery**: Scans the source directory for YAML manifest files
2. **Parsing**: Parses each manifest using the standard Radius manifest parser
3. **Filtering**: Identifies manifests with `defaultRegistration: true`
4. **Cleaning**: Removes the `defaultRegistration` field (not needed in Radius)
5. **Comparison**: Compares with existing files in the target directory
6. **Syncing**: Adds new files or updates existing ones if content differs
7. **Reporting**: Provides a summary of added, updated, and skipped files

## Exit Codes

- `0`: Success, changes detected and applied
- `1`: Errors occurred during sync
- `2`: Success, no changes detected

## Integration with GitHub Actions

This tool is used by the automated sync workflow at `.github/workflows/sync-resource-types.yaml` to keep resource types in sync between repositories.
