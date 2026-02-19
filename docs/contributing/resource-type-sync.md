# Resource Type Sync Mechanism

This document explains how the automatic resource type sync mechanism works in Radius.

## Overview

The resource type sync mechanism automatically manages default resource type registration in Radius by syncing resource type definitions from the [resource-types-contrib](https://github.com/radius-project/resource-types-contrib) repository. This eliminates manual copying and ensures consistency between repositories.

## Problem Statement

Previously, resource type YAML files were manually copied from resource-types-contrib to enable default registration in Radius. This manual process:

- Did not scale well
- Could lead to schema drift when resource definitions were updated upstream
- Required manual intervention to keep schemas synchronized

## Solution

The automated sync mechanism addresses these issues through:

1. **Metadata-based marking**: Resource types in resource-types-contrib can be marked with `defaultRegistration: true` to indicate they should be automatically registered in Radius
2. **Automated syncing**: A GitHub Actions workflow automatically copies marked resource types to the Radius repository
3. **Change detection**: The workflow only creates PRs when changes are detected, reducing noise
4. **Review process**: All synced changes go through pull request review before being merged

## Architecture

### Components

1. **Sync Configuration** (`.github/resource-type-sync-config.yaml`)
   - Defines source and target repositories
   - Specifies sync strategy (metadata or convention-based)
   - Configures file patterns, validation rules, and PR settings

2. **Sync Workflow** (`.github/workflows/sync-resource-types.yaml`)
   - Runs on a daily schedule (2 AM UTC)
   - Can be manually triggered for testing
   - Creates pull requests with detected changes

3. **Sync Script** (`.github/scripts/sync-resource-types.py`)
   - Fetches resource type files from resource-types-contrib
   - Identifies files marked for default registration
   - Validates file format and required fields
   - Syncs files to the target directory

4. **Target Directory** (`deploy/manifest/built-in-providers/self-hosted/`)
   - Stores synced resource type definitions (with `synced_` prefix)
   - Also contains manually maintained core resource types
   - Automatically registered by UCP initializer at startup

### Data Flow

```
resource-types-contrib repository
         |
         | (1) GitHub Actions workflow triggers
         v
    Fetch repository tree via API
         |
         | (2) Filter files based on patterns
         v
    Check defaultRegistration field
         |
         | (3) Download marked files
         v
    Validate manifest format
         |
         | (4) Sync to target directory with synced_ prefix
         v
deploy/manifest/built-in-providers/self-hosted/
         |
         | (5) Create PR if changes detected
         v
    Review and merge PR
         |
         | (6) UCP reads manifests at startup
         v
    Resource types registered in Radius
```

## Marking Resource Types for Default Registration

In the resource-types-contrib repository, add the `defaultRegistration: true` field to the top-level of your resource type YAML:

```yaml
# Mark this resource type for default registration in Radius
defaultRegistration: true

namespace: MyCompany.Resources
types:
  myResourceType:
    description: A sample resource type
    apiVersions:
      2023-10-01-preview:
        schema:
          type: object
          properties:
            name:
              type: string
              description: The resource name
          required:
            - name
```

### When to Mark for Default Registration

Mark a resource type for default registration when:

- It provides core functionality needed by most Radius users
- It has stable schemas and is production-ready
- It should be available immediately without manual registration
- It follows all Radius resource type best practices

### When NOT to Mark for Default Registration

Do not mark a resource type for default registration when:

- It's experimental or under active development
- It's specific to a particular organization or use case
- It has dependencies that may not be available in all environments
- It's primarily for testing or examples

## Sync Workflow

### Schedule

The sync workflow runs automatically:

- **Daily at 2 AM UTC** to check for updates
- **On-demand** via workflow_dispatch (for testing)

### Process

1. **Checkout**: Clone both Radius and resource-types-contrib repositories
2. **Fetch**: Retrieve the file tree from resource-types-contrib
3. **Filter**: Identify files matching configured patterns
4. **Check**: Verify `defaultRegistration: true` is set
5. **Validate**: Ensure files follow the proper manifest format
6. **Sync**: Copy changed files to the target directory
7. **PR**: Create a pull request if changes are detected

### Pull Request Workflow

When changes are detected:

1. A new branch is created: `sync/resource-types-{run-number}`
2. Changes are committed to the branch
3. A PR is opened with:
   - Clear title indicating it's an automated sync
   - Description listing changed files
   - Labels: `resource-types`, `sync`, `automated`
   - Checklist for reviewers

4. Reviewers verify:
   - Manifest format is correct
   - Required fields are present
   - No sensitive data is included
   - Changes are expected and valid

5. Once approved, the PR is merged and resource types are updated

## Configuration

### Sync Strategy

Two strategies are supported:

1. **Metadata** (default): Look for `defaultRegistration: true` in YAML files
   ```yaml
   sync:
     strategy: metadata
     metadataField: defaultRegistration
   ```

2. **Convention**: Sync all files in a specific directory
   ```yaml
   sync:
     strategy: convention
     conventionPath: default
   ```

### File Patterns

Control which files are considered for syncing:

```yaml
filePatterns:
  - "**/*.yaml"
  - "**/*.yml"

excludePatterns:
  - "**/test/**"
  - "**/testdata/**"
```

### Validation

Ensure synced files meet quality standards:

```yaml
validation:
  enabled: true
  requiredFields:
    - namespace
    - types
```

## Integration with Radius

### UCP Configuration

The UCP is configured to read manifests from the built-in-providers directory:

```yaml
initialization:
  manifestDirectory: "/manifest/built-in-providers"
```

This directory contains both manually maintained core resource types and auto-synced resource types (prefixed with `synced_`).

This is set in `deploy/Chart/templates/ucp/configmaps.yaml`.

### Startup Registration

When Radius starts:

1. UCP initializer service (`pkg/ucp/initializer/service.go`) reads the manifest directory
2. All YAML files are parsed as resource type manifests
3. Resource types are registered with UCP
4. They become available for use in Bicep templates

### Volume Mounting

In Kubernetes deployments, the manifest directory is mounted from a ConfigMap or included in the container image, making the synced resource types available at runtime.

## Troubleshooting

### Workflow Not Running

- Check if the repository is `kachawla/radius` (workflow only runs on main repo, not forks)
- Verify the schedule cron expression
- Check workflow permissions in repository settings

### Files Not Syncing

- Verify `defaultRegistration: true` is set in the source file
- Check file patterns in configuration (includes and excludes)
- Ensure file follows the proper manifest format
- Review workflow logs for validation errors

### Validation Failures

Common validation errors:

- Missing required fields (`namespace`, `types`)
- Invalid YAML syntax
- `types` field is not a dictionary
- File doesn't match the manifest schema

Fix these in the source repository, then trigger the workflow again.

### PR Not Created

- No changes detected (files are already up to date)
- Dry run mode is enabled
- Check GitHub token permissions

## Security Considerations

### GitHub Token

The workflow uses `GITHUB_TOKEN` with minimal permissions:
- `contents: read` - To fetch files
- `pull-requests: write` - To create PRs

### Source Repository Trust

Only sync from trusted source repositories. The default configuration uses:
- `radius-project/resource-types-contrib` (official repository)

### Validation

All synced files are validated to ensure:
- Proper YAML format
- Required fields are present
- No malicious content (basic checks)

### Review Process

All changes go through pull request review before being merged, providing human oversight of automated changes.

## Testing

### Manual Trigger

To test the sync workflow:

1. Uncomment the `workflow_dispatch` section in the workflow file
2. Go to Actions → sync-resource-types
3. Click "Run workflow"
4. Select dry_run: true for testing without creating a PR

### Dry Run Mode

Set `DRY_RUN=true` to:
- See what files would be synced
- Validate configuration
- Test without making changes

### Local Testing

Run the sync script locally:

```bash
export GITHUB_TOKEN="your-token"
export SOURCE_REPO="radius-project/resource-types-contrib"
export CONFIG_FILE=".github/resource-type-sync-config.yaml"
export DRY_RUN="true"

python .github/scripts/sync-resource-types.py
```

## Maintenance

### Updating Configuration

To change sync behavior:

1. Edit `.github/resource-type-sync-config.yaml`
2. Test with dry run mode
3. Commit and push changes
4. Manually trigger workflow to verify

### Monitoring

Monitor the workflow:
- Check GitHub Actions for failed runs
- Review sync PRs regularly
- Watch for validation errors in logs

### Keeping in Sync

The sync is idempotent and safe to run repeatedly. If upstream changes are missed:

1. Manually trigger the workflow
2. The next scheduled run will catch up
3. Review and merge the resulting PR

## Future Enhancements

Potential improvements:

- Support for multiple source repositories
- Webhook-triggered sync on changes in resource-types-contrib
- Auto-merge for trusted changes
- Slack/email notifications for sync failures
- More sophisticated validation rules
- Support for removing resource types when unmarked

## References

- [Resource Type Manifest Format](../../pkg/cli/manifest/manifest.go)
- [UCP Initializer](../../pkg/ucp/initializer/service.go)
- [Resource Types Contrib Repository](https://github.com/radius-project/resource-types-contrib)
- [GitHub Actions - create-pull-request](https://github.com/peter-evans/create-pull-request)
