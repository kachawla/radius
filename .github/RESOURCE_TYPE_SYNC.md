# Resource Type Sync - Quick Reference

## Overview

Automatic syncing of resource types from resource-types-contrib to Radius.

## Quick Start

### For Contributors (resource-types-contrib)

Add to your resource type YAML:
```yaml
defaultRegistration: true
namespace: YourCompany.Resources
types:
  yourType:
    # ... rest of definition
```

### For Maintainers (Radius)

1. Review automated PRs when created
2. Verify changes match expectations
3. Merge to activate resource types

## Architecture

```
resource-types-contrib → GitHub API → Sync Script → PR → Merge → UCP Registration
```

## Key Files

| File | Purpose | Size |
|------|---------|------|
| `.github/resource-type-sync-config.yaml` | Configuration | 3.3KB |
| `.github/workflows/sync-resource-types.yaml` | Workflow | 4.2KB |
| `.github/scripts/sync-resource-types.py` | Sync logic | 9.7KB |
| `.github/scripts/test_sync.py` | Tests | 4.0KB |

## Configuration

```yaml
# Source
repository: radius-project/resource-types-contrib
branch: main
basePath: resource-types

# Target
directory: deploy/manifest/built-in-providers/self-hosted
filePrefix: synced_

# Strategy
strategy: metadata
metadataField: defaultRegistration
```

## Workflow

- **Schedule**: Daily at 2 AM UTC
- **Trigger**: Can be manually triggered via workflow_dispatch (when uncommented)
- **Fork Protection**: Only runs on main repository
- **Output**: Pull request with changes

## Testing

```bash
# Run unit tests
python .github/scripts/test_sync.py

# Dry run
export DRY_RUN=true
export CONFIG_FILE=.github/resource-type-sync-config.yaml
export SOURCE_REPO=radius-project/resource-types-contrib
python .github/scripts/sync-resource-types.py
```

## File Naming

- Manual files: `applications_core.yaml`, `radius_compute.yaml`, etc.
- Synced files: `synced_*.yaml` (auto-generated, gitignored)

## Documentation

1. **Technical**: [docs/contributing/resource-type-sync.md](../../docs/contributing/resource-type-sync.md) (365 lines)
2. **User Guide**: [docs/contributing/marking-resource-types-for-default-registration.md](../../docs/contributing/marking-resource-types-for-default-registration.md) (239 lines)
3. **Directory**: [deploy/manifest/built-in-providers/README.md](../../deploy/manifest/built-in-providers/README.md)

## Validation Rules

Required fields:
- `namespace`
- `types`

Additional checks:
- Valid YAML syntax
- `types` is a dictionary
- Proper schema structure

## Security

- ✅ Fork protection enabled
- ✅ Minimal token permissions
- ✅ Content validation
- ✅ PR review required
- ✅ Pinned action versions

## Troubleshooting

| Issue | Solution |
|-------|----------|
| File not syncing | Check `defaultRegistration: true` is set |
| Validation error | Review manifest format |
| Workflow not running | Verify repository is main, not fork |
| No PR created | Check if changes detected |

## Monitoring

- Workflow runs: GitHub Actions → sync-resource-types
- Sync PRs: Labels: `resource-types`, `sync`, `automated`
- Logs: Workflow run details

## Support

- Discord: [Radius Discord](https://discord.gg/SRG3ePMKNy) #dev channel
- Issues: [GitHub Issues](https://github.com/radius-project/radius/issues)
- Docs: [Resource Type Sync Documentation](../../docs/contributing/resource-type-sync.md)

---

**Status**: ✅ Implemented and ready for production use

**Version**: 1.0

**Last Updated**: 2026-02-05
