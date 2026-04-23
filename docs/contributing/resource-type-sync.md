# Resource Type Synchronization from resource-types-contrib

## Overview

Radius supports automatic registration of resource types from manifests placed in the `deploy/manifest/built-in-providers` directory. This document describes the mechanism for automatically synchronizing resource type manifests from the [resource-types-contrib](https://github.com/kachawla/resource-types-contrib) repository to ensure schemas stay in sync.

## Problem Statement

Previously, resource types defined in resource-types-contrib had to be manually copied to the Radius repository. This created several challenges:
- **Manual Process**: Developers had to remember to copy manifests manually
- **Schema Drift**: Changes in resource-types-contrib weren't automatically reflected in Radius
- **No Clear Mechanism**: No standard way to indicate which types should be registered by default

## Solution: Automated Sync Mechanism

The solution uses a metadata flag approach combined with an automated GitHub Actions workflow to keep resource types in sync.

### Key Components

1. **Metadata Flag**: Manifests in resource-types-contrib can include `defaultRegistration: true` to indicate they should be synced
2. **Sync Tool**: A Go tool (`hack/sync-resource-types/`) that parses manifests and performs the sync
3. **GitHub Actions Workflow**: Automated workflow that runs weekly and can be manually triggered
4. **Manifest Parser Enhancement**: Updated parser supports the `defaultRegistration` field

## How It Works

### 1. Marking Manifests for Default Registration

In the resource-types-contrib repository, add `defaultRegistration: true` to any manifest that should be automatically registered in Radius:

```yaml
defaultRegistration: true  # This manifest will be synced to Radius
namespace: Applications.Databases
types:
  postgresql:
    apiVersions:
      "2023-10-01-preview":
        schema:
          type: object
          properties:
            host:
              type: string
            port:
              type: integer
```

### 2. Automatic Synchronization

The sync workflow (`.github/workflows/sync-resource-types.yaml`):
- Runs weekly on Sunday at midnight UTC
- Can be manually triggered via GitHub Actions UI
- Checks out both repositories
- Runs the sync tool to identify and copy marked manifests
- Creates a pull request if changes are detected

### 3. Sync Tool Behavior

The sync tool (`hack/sync-resource-types/`):
1. Scans the source directory for YAML manifest files
2. Parses each manifest using the standard Radius manifest parser
3. Identifies manifests with `defaultRegistration: true`
4. Removes the `defaultRegistration` field (not needed in Radius)
5. Compares with existing files in the target directory
6. Adds new files or updates existing ones if content differs
7. Provides a detailed summary of changes

### 4. Pull Request Review

When changes are detected, the workflow creates a PR with:
- Summary of added/updated files
- Source repository and commit information
- Checklist for reviewers
- Automatic labels

Reviewers should:
- Verify the changes are expected
- Check that manifests are valid
- Ensure no breaking changes
- Merge when satisfied

## Using the Sync Tool

### Prerequisites

- Go 1.22 or later
- Access to both repositories

### Command Line Usage

```bash
# Dry run to see what would be synced
go run ./hack/sync-resource-types/main.go \
  --source ../resource-types-contrib/manifests \
  --target ./deploy/manifest/built-in-providers/dev \
  --dry-run --verbose

# Actual sync
go run ./hack/sync-resource-types/main.go \
  --source ../resource-types-contrib/manifests \
  --target ./deploy/manifest/built-in-providers/dev
```

### Options

- `--source`: Source directory containing resource type manifests (required)
- `--target`: Target directory for synced manifests (required)
- `--dry-run`: Print actions without making changes
- `--verbose`: Enable verbose output

## Manual Workflow Trigger

To manually trigger the sync workflow:

1. Go to the Radius repository on GitHub
2. Navigate to **Actions** → **sync-resource-types**
3. Click **Run workflow**
4. (Optional) Customize parameters:
   - Source repository
   - Source branch
   - Target directory
   - Dry run mode
5. Click **Run workflow**

## Development Workflow

### Adding a New Resource Type

**In resource-types-contrib:**
1. Create or update the manifest YAML file
2. Add `defaultRegistration: true` if it should be registered by default
3. Commit and push changes
4. Wait for weekly sync OR manually trigger the workflow

**In Radius:**
1. Review the automatically created PR
2. Verify the manifest is valid
3. Run tests
4. Merge the PR

### Updating an Existing Resource Type

**In resource-types-contrib:**
1. Update the manifest YAML file
2. Commit and push changes
3. Wait for weekly sync OR manually trigger the workflow

**In Radius:**
1. Review the automatically created PR showing the diff
2. Verify changes don't break existing functionality
3. Run tests
4. Merge the PR

### Removing a Resource Type from Default Registration

**In resource-types-contrib:**
1. Remove `defaultRegistration: true` or set it to `false`
2. Commit and push changes

**In Radius:**
1. The sync workflow will NOT remove the file automatically
2. If removal is desired, manually delete the file from Radius
3. This is by design to avoid accidental data loss

## Architecture

### Manifest Parser Enhancement

The `ResourceProvider` struct in `pkg/cli/manifest/manifest.go` now includes:

```go
type ResourceProvider struct {
    Namespace string `yaml:"namespace"`
    Location map[string]string `yaml:"location,omitempty"`
    Types map[string]*ResourceType `yaml:"types"`
    DefaultRegistration bool `yaml:"defaultRegistration,omitempty"`
}
```

The `defaultRegistration` field is:
- Optional (defaults to `false`)
- Used only during sync
- Removed before writing to Radius built-in-providers directory

### Sync Tool Architecture

```
hack/sync-resource-types/
├── main.go          # Main sync logic
├── main_test.go     # Unit and integration tests
├── README.md        # Tool documentation
└── testdata/        # Test fixtures
    ├── source/      # Sample source manifests
    └── target/      # Sample target state
```

Key functions:
- `syncResourceTypes()`: Main sync orchestration
- `removeDefaultRegistrationField()`: Strips sync metadata
- `contentEqual()`: Normalizes and compares YAML content

## Security Considerations

1. **Repository Access**: The workflow only runs on the main repository, not forks
2. **Review Required**: All changes require PR review before merging
3. **No Auto-Merge**: The workflow does not automatically merge changes
4. **Manifest Validation**: Manifests are validated before syncing
5. **Token Permissions**: Workflow uses minimal permissions

## Testing

### Unit Tests

Run sync tool tests:
```bash
cd hack/sync-resource-types
go test -v
```

### Integration Tests

Test with actual repositories:
```bash
# Clone both repos
git clone https://github.com/kachawla/resource-types-contrib.git
git clone https://github.com/kachawla/radius.git

# Run sync tool
cd radius
go run ./hack/sync-resource-types/main.go \
  --source ../resource-types-contrib/manifests \
  --target ./deploy/manifest/built-in-providers/dev \
  --dry-run --verbose
```

## Troubleshooting

### Sync Workflow Fails

1. Check the workflow logs in GitHub Actions
2. Verify source repository is accessible
3. Ensure manifests are valid YAML
4. Check for parser errors

### Manifests Not Syncing

1. Verify `defaultRegistration: true` is present
2. Check the manifest is valid (run parser locally)
3. Ensure the file has `.yaml` or `.yml` extension
4. Check workflow logs for errors

### PR Not Created

1. Verify changes were detected (check workflow output)
2. Ensure workflow has `contents: write` and `pull-requests: write` permissions
3. Check if a PR already exists for the branch

### Content Differences Not Detected

The sync tool normalizes YAML before comparing, but:
- Check for comments that might be stripped
- Verify both files use valid YAML
- Try running with `--verbose` for more details

## Future Enhancements

Potential improvements for future iterations:

1. **Version Tracking**: Record which commit from resource-types-contrib each file came from
2. **Partial Syncs**: Support syncing specific resource types within a manifest
3. **Conflict Detection**: Better handling of local modifications
4. **Notification System**: Alerts for sync failures
5. **Bi-directional Sync**: Support syncing changes back to resource-types-contrib
6. **Schema Validation**: Enhanced validation before syncing

## Related Resources

- [resource-types-contrib repository](https://github.com/kachawla/resource-types-contrib)
- [Sync tool documentation](../hack/sync-resource-types/README.md)
- [Manifest format documentation](../pkg/cli/manifest/doc.go)
- [UCP resources documentation](./ucp/resources.md)

## Support

For questions or issues:
1. Check this documentation
2. Review sync tool logs
3. Open an issue in the Radius repository
4. Ask in the Radius community Slack
