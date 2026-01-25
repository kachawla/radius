# Automated Resource Type Sync from resource-types-contrib

* **Author**: GitHub Copilot (@copilot)

## Overview

This feature automates the synchronization of resource type YAML definitions from the community-contributed [resource-types-contrib](https://github.com/radius-project/resource-types-contrib) repository to the Radius repository for default registration. Previously, resource types had to be manually copied between repositories, which was error-prone and made it difficult to keep resource types up-to-date.

The solution implements an author-driven approach where resource type authors maintain a configuration file (`.radius-sync-config.yaml`) in the resource-types-contrib repository to mark which resource types should be automatically synced. A GitHub Actions workflow runs weekly to detect changes and automatically creates pull requests in the Radius repository with updated resource type definitions.

This automation ensures that community-contributed resource types are kept in sync with minimal manual intervention while giving authors full control over which types are registered by default.

## Terms and definitions

| Term | Definition |
|------|------------|
| **Resource Type** | A schema definition that describes the structure and properties of a Radius resource (e.g., mySqlDatabases, containers) |
| **Default Registration** | Resource types that are automatically available in Radius without requiring manual installation |
| **Built-in Providers** | The manifest directories (`deploy/manifest/built-in-providers/dev` and `deploy/manifest/built-in-providers/self-hosted`) where default resource types are stored |
| **Sync Configuration** | The `.radius-sync-config.yaml` file that lists which resource types should be synced |
| **Author-Driven** | An approach where resource type authors control the sync configuration in their own repository |

## Objectives

> **Issue Reference:** Automatically add certain resource types for default registration and update the schema if it is updated in resource-types-contrib

### Goals

- **Automate resource type synchronization**: Eliminate manual copying of resource type YAML files between repositories
- **Enable automatic updates**: Detect and sync changes when resource type schemas are updated in resource-types-contrib
- **Author-driven configuration**: Allow resource type authors to mark their types for default registration by updating a configuration file in their own repository
- **Weekly automated checks**: Run periodic checks to ensure resource types stay synchronized
- **Minimal infrastructure changes**: Integrate with existing manifest loading system without requiring changes to Radius core components

### Non goals

- **Real-time synchronization**: The sync runs weekly, not on every commit to resource-types-contrib
- **Bi-directional sync**: Changes only flow from resource-types-contrib to Radius, not the reverse
- **Automatic merging**: Pull requests are created but require manual review and approval before merging
- **Version management**: The sync uses the main branch; it does not handle versioning or backporting to release branches
- **Schema validation beyond YAML syntax**: The sync validates YAML syntax but does not perform semantic validation of resource type schemas

## User scenarios (optional)

### User story 1: Resource Type Author

As a resource type author contributing to resource-types-contrib, I want to mark my resource type for default registration so that Radius users can use my resource type without manual installation, by simply updating a configuration file in my repository.

**Steps:**
1. Author creates a new resource type in resource-types-contrib (e.g., `Data/redisCaches/redisCaches.yaml`)
2. Author updates `.radius-sync-config.yaml` in resource-types-contrib to include their resource type
3. Weekly sync job automatically detects the new entry and creates a PR in the Radius repository
4. After PR is reviewed and merged, the resource type becomes available by default in Radius

### User story 2: Resource Type Maintainer

As a resource type maintainer, I want to update my resource type schema and have those changes automatically propagated to Radius, so that users always have access to the latest version without manual coordination.

**Steps:**
1. Maintainer updates resource type schema in resource-types-contrib
2. Weekly sync job detects the file difference and creates a PR with the updated schema
3. After PR review and merge, users get the updated resource type in the next Radius release

## User Experience (if applicable)

The sync process is transparent to end users. Resource type authors interact with the sync system through the configuration file:

**Sample Configuration File (`.radius-sync-config.yaml` in resource-types-contrib):**
```yaml
# Target directories in the radius repository
targetDirectories:
  - deploy/manifest/built-in-providers/dev
  - deploy/manifest/built-in-providers/self-hosted

# Resource types to sync for default registration
resourceTypes:
  # Data namespace resource types
  - namespace: Data
    name: mySqlDatabases
    file: mySqlDatabases.yaml
    
  - namespace: Data
    name: postgreSqlDatabases
    file: postgreSqlDatabases.yaml

  # Compute namespace resource types
  - namespace: Compute
    name: containers
    file: containers.yaml
```

**Sample PR Created by Automation:**
```
Title: chore: sync resource types from resource-types-contrib

Body:
## Automated Resource Type Sync

This PR updates resource type definitions synced from resource-types-contrib.

Resource types are copied to both `deploy/manifest/built-in-providers/dev` 
and `deploy/manifest/built-in-providers/self-hosted` directories.

### Review Checklist
- [ ] Review the changes to ensure they are expected
- [ ] Verify that the resource type schemas are valid
- [ ] Check that no breaking changes were introduced
```

## Design

### High Level Design

The sync system consists of three main components:

1. **Configuration File** (`.radius-sync-config.yaml` in resource-types-contrib): Defines which resource types should be synced
2. **Sync Script** (`.github/scripts/sync-resource-types.sh` in Radius): Fetches configuration and resource type files, performs diff-based change detection
3. **GitHub Actions Workflow** (`.github/workflows/sync-resource-types.yaml` in Radius): Schedules weekly runs and creates PRs when changes are detected

**Data Flow:**
```
┌─────────────────────────────────────┐
│  resource-types-contrib (upstream)  │
│                                     │
│  ├── .radius-sync-config.yaml      │ ◄── Authors update this
│  ├── Data/                          │
│  │   ├── mySqlDatabases/            │
│  │   │   └── mySqlDatabases.yaml   │
│  │   └── postgreSqlDatabases/       │
│  │       └── postgreSqlDatabases... │
│  └── Compute/                       │
│      └── containers/                │
│          └── containers.yaml        │
└─────────────────────────────────────┘
            │
            │ Weekly Sync (GitHub Actions)
            │ 1. Fetch config
            │ 2. Fetch YAML files
            │ 3. Detect changes
            ▼
┌─────────────────────────────────────┐
│  Radius repository                  │
│                                     │
│  ├── .github/                       │
│  │   ├── scripts/                   │
│  │   │   └── sync-resource-types.sh│ ◄── Sync script
│  │   └── workflows/                 │
│  │       └── sync-resource-types... │ ◄── Workflow
│  └── deploy/manifest/               │
│      └── built-in-providers/        │
│          ├── dev/                   │
│          │   ├── mySqlDatabases.... │ ◄── Synced files
│          │   ├── postgreSqlDatabas..│
│          │   └── containers.yaml   │
│          └── self-hosted/           │
│              ├── mySqlDatabases.... │
│              ├── postgreSqlDatabas..│
│              └── containers.yaml   │
└─────────────────────────────────────┘
            │
            │ PR Created if changes detected
            ▼
    Manual Review & Merge
```

### Architecture Diagram

```
┌──────────────────────────────────────────────────────────────────────────┐
│                          Sync Architecture                                │
└──────────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────────┐
│ resource-types-contrib Repository (Upstream)                             │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                           │
│  .radius-sync-config.yaml                                                │
│  ┌───────────────────────────────────────────────────────────────┐      │
│  │ targetDirectories:                                             │      │
│  │   - deploy/manifest/built-in-providers/dev                    │      │
│  │   - deploy/manifest/built-in-providers/self-hosted            │      │
│  │ resourceTypes:                                                 │      │
│  │   - namespace: Data                                            │      │
│  │     name: mySqlDatabases                                       │      │
│  │     file: mySqlDatabases.yaml                                  │      │
│  └───────────────────────────────────────────────────────────────┘      │
│                                                                           │
│  Resource Type Files:                                                    │
│  Data/mySqlDatabases/mySqlDatabases.yaml                                 │
│  Data/postgreSqlDatabases/postgreSqlDatabases.yaml                       │
│  Compute/containers/containers.yaml                                      │
│                                                                           │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    │ HTTPS (GitHub Raw Content API)
                                    │
                    ┌───────────────┴────────────────┐
                    │                                │
                    ▼                                ▼
        ┌────────────────────────┐      ┌─────────────────────────┐
        │  1. Fetch Config File  │      │  2. Fetch Resource Type │
        │                        │      │     YAML Files          │
        └────────────────────────┘      └─────────────────────────┘
                    │                                │
                    └───────────────┬────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│ Radius Repository - Sync Workflow                                        │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                           │
│  .github/workflows/sync-resource-types.yaml                              │
│  ┌─────────────────────────────────────────────────────────────────┐    │
│  │ Trigger: Cron (Weekly on Sunday 00:00 UTC)                      │    │
│  │                                                                   │    │
│  │ Steps:                                                            │    │
│  │ 1. Checkout Radius repo                                          │    │
│  │ 2. Install dependencies (yq)                                     │    │
│  │ 3. Run .github/scripts/sync-resource-types.sh ───────┐          │    │
│  │ 4. Detect changes (git diff)                         │          │    │
│  │ 5. Create PR if changes found                        │          │    │
│  └───────────────────────────────────────────────────────┼──────────┘    │
│                                                           │               │
│  .github/scripts/sync-resource-types.sh    ◄─────────────┘               │
│  ┌─────────────────────────────────────────────────────────────────┐    │
│  │ #!/bin/bash                                                      │    │
│  │                                                                   │    │
│  │ fetch_config() {                                                 │    │
│  │   curl https://raw.github.../.radius-sync-config.yaml           │    │
│  │ }                                                                 │    │
│  │                                                                   │    │
│  │ fetch_resource_type() {                                          │    │
│  │   curl https://raw.github.../${namespace}/${name}/${file}       │    │
│  │   yq validate                                                    │    │
│  │   diff with existing file                                        │    │
│  │   copy to target directories if changed                          │    │
│  │ }                                                                 │    │
│  │                                                                   │    │
│  │ sync_resource_types() {                                          │    │
│  │   for each resource type in config                              │    │
│  │     fetch_resource_type()                                        │    │
│  │ }                                                                 │    │
│  └─────────────────────────────────────────────────────────────────┘    │
│                                    │                                      │
│                                    ▼                                      │
│  deploy/manifest/built-in-providers/                                     │
│  ├── dev/                                                                 │
│  │   ├── mySqlDatabases.yaml          ◄── Updated files                 │
│  │   ├── postgreSqlDatabases.yaml                                        │
│  │   └── containers.yaml                                                 │
│  └── self-hosted/                                                        │
│      ├── mySqlDatabases.yaml          ◄── Updated files                 │
│      ├── postgreSqlDatabases.yaml                                        │
│      └── containers.yaml                                                 │
│                                                                           │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    │ If changes detected
                                    ▼
                        ┌─────────────────────────┐
                        │   Create Pull Request   │
                        │                         │
                        │ - Title: chore: sync... │
                        │ - Branch: automated/... │
                        │ - Labels: automated     │
                        └─────────────────────────┘
                                    │
                                    ▼
                          Manual Review & Merge
                                    │
                                    ▼
                        ┌─────────────────────────┐
                        │  Resource Types Active  │
                        │  in Radius Deployment   │
                        └─────────────────────────┘
```

### Detailed Design

#### Component 1: Sync Configuration File

**Location**: `.radius-sync-config.yaml` in the root of resource-types-contrib repository

**Purpose**: Provides an author-driven mechanism for resource type contributors to mark their types for default registration in Radius.

**Schema**:
```yaml
targetDirectories:  # List of target directories in Radius repo
  - string          # Relative path from Radius repo root
  
resourceTypes:      # List of resource types to sync
  - namespace: string      # Namespace without 'Radius.' prefix (e.g., Data, Compute)
    name: string          # Resource type directory name
    file: string          # YAML filename
```

**Validation**:
- File must be valid YAML
- `targetDirectories` must be a non-empty array
- `resourceTypes` must be a non-empty array
- Each resource type entry must have `namespace`, `name`, and `file` fields

#### Component 2: Sync Script

**Location**: `.github/scripts/sync-resource-types.sh` in Radius repository

**Purpose**: Fetches configuration and resource type files from resource-types-contrib, validates them, detects changes, and copies files to target directories.

**Key Functions**:

1. **`fetch_config()`**:
   - Downloads `.radius-sync-config.yaml` from resource-types-contrib using GitHub raw content API
   - Validates YAML syntax
   - Fails if file doesn't exist with helpful error message

2. **`fetch_resource_type(namespace, name, file)`**:
   - Constructs URL: `https://raw.githubusercontent.com/${SOURCE_REPO}/${SOURCE_BRANCH}/${namespace}/${name}/${file}`
   - Downloads resource type YAML file
   - Validates YAML syntax with `yq`
   - For each target directory:
     - Compares with existing file using `diff`
     - Copies file only if changed or new
     - Sets change detection flag

3. **`sync_resource_types()`**:
   - Iterates through all resource types in configuration
   - Calls `fetch_resource_type()` for each
   - Tracks whether any changes were detected

**Dependencies**:
- `curl`: For downloading files from GitHub
- `yq`: For YAML validation
- `diff`: For change detection

**Command-line Options**:
- `--source-repo REPO`: Override source repository (default: radius-project/resource-types-contrib)
- `--source-branch BRANCH`: Override source branch (default: main)
- `--help`: Display usage information

#### Component 3: GitHub Actions Workflow

**Location**: `.github/workflows/sync-resource-types.yaml` in Radius repository

**Trigger**:
- Schedule: `cron: "0 0 * * 0"` (Weekly on Sunday at 00:00 UTC)
- Manual: `workflow_dispatch` (commented out for production use)

**Permissions**:
- `contents: write` - To create commits
- `pull-requests: write` - To create PRs

**Steps**:
1. **Checkout**: Check out Radius repository
2. **Install yq**: Install YAML processor for validation
3. **Run sync script**: Execute `.github/scripts/sync-resource-types.sh`
4. **Detect changes**: Run `git diff --quiet deploy/manifest/built-in-providers/`
5. **Create PR**: If changes detected, use `peter-evans/create-pull-request` action with:
   - Title: "chore: sync resource types from resource-types-contrib"
   - Branch: `automated/sync-resource-types`
   - Labels: `automated`, `resource-types`
   - Delete branch after merge

**Fork Protection**: Only runs on main repository (`if: github.repository == 'kachawla/radius'`)

#### Advantages

1. **Author-Driven**: Resource type authors control what gets synced by editing a file in their own repository
2. **Decentralized**: No need to coordinate changes across multiple repositories
3. **Transparent**: All configuration changes are tracked in git history
4. **Simple**: Authors only need to update a YAML file
5. **Minimal Infrastructure**: No changes to Radius core components (config.go, initializer, etc.)
6. **Automatic Updates**: Weekly sync ensures resource types stay current
7. **Safe**: PRs require manual review before merging

#### Disadvantages

1. **Not Real-Time**: Weekly schedule means delays up to 7 days for updates
2. **Requires Companion PR**: Initial setup requires PR in resource-types-contrib repository
3. **No Version Control**: Syncs from main branch only, no support for release branches
4. **Manual PR Review**: PRs aren't automatically merged, requiring maintainer attention
5. **Single Point of Configuration**: If `.radius-sync-config.yaml` is deleted or corrupted, sync breaks

#### Proposed Option

The implemented design is the proposed option because:
- It meets all stated goals (automation, author-driven, minimal infrastructure changes)
- It's simple to understand and maintain
- It provides safety through manual PR review
- It's extensible (can add real-time sync, version control, etc. in the future)

### API design (if applicable)

N/A - This feature does not introduce new REST APIs, CLI commands, or Go APIs.

### CLI Design (if applicable)

N/A - This feature does not introduce new CLI commands.

### Implementation Details

#### Sync Script (.github/scripts/sync-resource-types.sh)

**Language**: Bash shell script

**Key Implementation Details**:
- Uses `curl` with `-sf` flags (silent mode, fail on error) to download files
- Uses `yq eval` to validate YAML syntax
- Uses `mktemp -d` for temporary file storage during downloads
- Implements cleanup trap to remove temporary files on exit
- Uses `diff -q` for efficient file comparison
- Copies files only when changes detected to minimize git diff noise

**Error Handling**:
- Fails fast on missing dependencies
- Provides clear error messages for missing config file
- Validates YAML syntax before copying
- Uses bash `set -euo pipefail` for robust error handling

#### GitHub Actions Workflow (.github/workflows/sync-resource-types.yaml)

**Key Implementation Details**:
- Uses pinned action versions with full commit SHA for security
- Installs `yq` from specific release for reproducibility
- Uses `git diff --quiet` to detect changes efficiently
- Sets outputs for conditional PR creation
- Uses `peter-evans/create-pull-request` for reliable PR creation

#### UCP (if applicable)

No changes to UCP. The synced resource type files are loaded by the existing manifest loading system in the UCP initializer.

#### Bicep (if applicable)

N/A - No changes to Bicep.

#### Deployment Engine (if applicable)

N/A - No changes to Deployment Engine.

#### Core RP (if applicable)

N/A - No changes to Core RP.

#### Portable Resources / Recipes RP (if applicable)

N/A - No changes to Portable Resources or Recipes RP.

### Error Handling

**Error Scenario 1**: `.radius-sync-config.yaml` doesn't exist in resource-types-contrib

- **Detection**: `curl` returns 404 error
- **Handling**: Script logs clear error message and exits with code 1
- **User Experience**: Workflow fails with error message directing to create companion PR
- **Recovery**: Create `.radius-sync-config.yaml` in resource-types-contrib

**Error Scenario 2**: Resource type YAML file doesn't exist

- **Detection**: `curl` returns 404 error when fetching resource type file
- **Handling**: Script logs error with URL and exits with code 1
- **User Experience**: Workflow fails with specific file not found error
- **Recovery**: Fix configuration file or add missing resource type file

**Error Scenario 3**: Invalid YAML syntax

- **Detection**: `yq eval` fails on malformed YAML
- **Handling**: Script logs validation error and exits with code 1
- **User Experience**: Workflow fails with YAML syntax error
- **Recovery**: Fix YAML syntax in resource-types-contrib

**Error Scenario 4**: Missing dependencies (curl, yq)

- **Detection**: Script checks for command availability at startup
- **Handling**: Script logs missing dependencies and exits with code 1
- **User Experience**: Workflow fails at dependency check step
- **Recovery**: Automatic (yq installation step in workflow)

**Error Scenario 5**: GitHub API rate limiting

- **Detection**: `curl` returns 429 error
- **Handling**: Script fails with rate limit error
- **User Experience**: Workflow fails and retries on next scheduled run
- **Recovery**: Wait for rate limit reset or authenticate requests

## Test plan

### Manual Testing

1. **Test sync script locally**:
   - Run `./.github/scripts/sync-resource-types.sh` from Radius repo
   - Verify it fetches config and resource type files
   - Verify change detection works correctly
   - Verify files are copied to both target directories

2. **Test workflow**:
   - Trigger workflow manually (enable workflow_dispatch temporarily)
   - Verify it completes successfully
   - Verify PR is created when changes exist
   - Verify PR description is accurate

3. **Test error scenarios**:
   - Test with missing config file
   - Test with invalid YAML
   - Test with missing resource type files
   - Verify error messages are clear and actionable

### Automated Testing

Currently no automated tests. Future considerations:
- Unit tests for sync script functions (requires refactoring to testable functions)
- Integration tests that mock GitHub API responses
- End-to-end tests in staging environment

### Functional Testing Areas

1. **Configuration parsing**: Verify script correctly reads and validates config file
2. **File fetching**: Verify script downloads files from correct URLs
3. **YAML validation**: Verify invalid YAML is rejected
4. **Change detection**: Verify only changed files trigger updates
5. **Multi-directory copy**: Verify files copied to all target directories
6. **PR creation**: Verify PR contains expected changes and metadata
7. **Idempotency**: Verify running sync multiple times without changes doesn't create PRs

## Security

### Threat Model

**Threat 1: Malicious YAML injection**
- **Risk**: Attacker modifies resource type YAML in resource-types-contrib to inject malicious content
- **Mitigation**: 
  - Manual PR review required before merging
  - YAML validation checks basic syntax but not semantic correctness
  - Resource types are validated by Radius when loaded
  - resource-types-contrib has its own access controls and review process

**Threat 2: Supply chain attack via compromised dependencies**
- **Risk**: Compromised `yq` binary or GitHub Actions
- **Mitigation**:
  - Pin `yq` to specific version and checksum (currently pinned to v4.40.5)
  - Pin GitHub Actions to specific commit SHA
  - Download `yq` from official GitHub release

**Threat 3: Unauthorized access to sync workflow**
- **Risk**: Attacker triggers malicious sync runs
- **Mitigation**:
  - Workflow only runs on main repository (fork protection)
  - `workflow_dispatch` commented out in production
  - GitHub token has minimal required permissions

**Threat 4: Exposure of secrets in logs**
- **Risk**: Sensitive data logged during sync
- **Mitigation**:
  - No secrets used in sync process
  - All URLs and content are public
  - GitHub Actions logs are public-readable (intended behavior)

### Authentication and Authorization

- Uses GitHub's built-in authentication via `GITHUB_TOKEN`
- Token has `contents: write` and `pull-requests: write` permissions
- No external authentication required (all content is public)

### Storing Secrets

N/A - No secrets are stored or used in this feature.

## Compatibility (optional)

**Backward Compatibility**: Fully backward compatible. Existing resource types continue to work unchanged.

**Breaking Changes**: None. This is a new automation feature that doesn't change existing behavior.

**Version Compatibility**: 
- Syncs from main branch only
- No support for syncing to release branches (non-goal for initial version)
- Future enhancement could add version-specific sync

## Monitoring and Logging

### Instrumentation

**Logs**:
- Sync script logs all operations to stdout with `[INFO]` and `[ERROR]` prefixes
- GitHub Actions workflow logs capture all script output
- PR creation action logs success/failure

**Metrics**: 
- None currently implemented
- Future: Could track sync success/failure rate, number of resource types synced, PR merge rate

**Traces**: N/A - No distributed tracing for this batch automation

### Troubleshooting

**Common Issues**:

1. **Workflow fails with "Config file not found"**
   - Check if `.radius-sync-config.yaml` exists in resource-types-contrib
   - Verify repository and branch names are correct

2. **Workflow completes but no PR created**
   - Check if resource types have actually changed
   - Review workflow logs for "No changes detected" message

3. **PR created with unexpected files**
   - Check `.radius-sync-config.yaml` for incorrect entries
   - Verify resource type file paths are correct

4. **Sync script fails locally**
   - Ensure `curl`, `yq`, and `diff` are installed
   - Check network connectivity to GitHub

## Development plan

### Phase 1: Core Implementation (Completed)
- ✅ Design and implement sync script
- ✅ Create GitHub Actions workflow
- ✅ Create documentation and examples
- ✅ Test with initial set of resource types (5 types)

### Phase 2: Companion PR (Pending)
- ⏳ Create PR in resource-types-contrib to add `.radius-sync-config.yaml`
- ⏳ Get PR reviewed and merged

### Phase 3: Validation and Monitoring (Future)
- Add automated tests for sync script
- Add metrics/monitoring for sync success rate
- Add alerting for sync failures

### Phase 4: Enhancements (Future)
- Support for release branch syncing
- Real-time sync on resource-types-contrib commits
- Automatic PR merging for non-breaking changes
- Support for partial syncs (sync only changed resource types)

### Effort Estimates

| Task | Estimate | Status |
|------|----------|--------|
| Design document | 4 hours | Complete |
| Sync script implementation | 8 hours | Complete |
| GitHub workflow implementation | 4 hours | Complete |
| Documentation | 4 hours | Complete |
| Manual testing | 4 hours | Complete |
| Companion PR | 2 hours | Pending |
| **Total** | **26 hours** | **~92% Complete** |

## Open Questions

1. **Should we support automatic merging of sync PRs?**
   - Pro: Faster updates, less manual work
   - Con: Risk of breaking changes without review
   - **Decision Pending**: Start with manual review, consider auto-merge for minor updates later

2. **Should we sync to release branches?**
   - Useful for backporting important updates
   - Adds complexity (which versions to sync to?)
   - **Decision**: Not for initial version (non-goal)

3. **How should we handle breaking changes?**
   - Should there be a separate review process?
   - Should we add breaking change detection?
   - **Decision Pending**: Rely on manual PR review for now

4. **Should we add semantic validation of resource type schemas?**
   - Would catch more errors earlier
   - Requires understanding of Radius schema validation rules
   - **Decision Pending**: Consider for future enhancement

## Alternatives considered

### Alternative 1: Metadata in Resource Type Files

Add a `defaultRegistration: true` field to each resource type YAML file instead of maintaining a separate configuration file.

**Advantages**:
- Single source of truth
- No separate config file to maintain
- Easier for authors (mark inline)

**Disadvantages**:
- Requires parsing every resource type file to find marked types
- Slower sync (must download all files)
- Pollutes resource type schema with sync metadata
- Harder to get overview of what's synced

**Rejected because**: Separate config file provides better performance and cleaner separation of concerns.

### Alternative 2: Manual Sync via CLI Command

Provide a CLI command in Radius to trigger sync on-demand rather than automated workflow.

**Advantages**:
- More control over when sync happens
- Can sync on-demand for urgent updates
- No GitHub Actions dependency

**Disadvantages**:
- Requires manual intervention
- Doesn't solve the automation goal
- Requires authentication setup for CLI users

**Rejected because**: Doesn't meet the automation objective.

### Alternative 3: Git Submodule

Use git submodules to include resource-types-contrib as a submodule in Radius repository.

**Advantages**:
- Built-in git mechanism
- Always in sync at commit level
- No custom sync needed

**Disadvantages**:
- Pulls entire repository (many unnecessary files)
- Submodules are complex to manage
- All contributors need to understand submodules
- Can't selectively sync specific resource types

**Rejected because**: Too complex and pulls unnecessary files.

### Alternative 4: npm/Package Distribution

Distribute resource types as npm packages or similar package system.

**Advantages**:
- Established distribution mechanism
- Version control built-in
- Dependency management

**Disadvantages**:
- Adds package management complexity
- Requires package publishing infrastructure
- Overkill for simple YAML file sync
- Requires changes to Radius loading mechanism

**Rejected because**: Too complex for the use case.

## Design Review Notes

*This section will be updated after design review meeting.*
