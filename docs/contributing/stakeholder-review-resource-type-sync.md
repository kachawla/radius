# Automated Resource Type Sync: Stakeholder Review Document

## Executive Summary

We've implemented an automated mechanism to synchronize resource type manifests from the **resource-types-contrib** repository to the **Radius** repository. This eliminates manual copying, prevents schema drift, and ensures default resource types stay current.

## The Challenge

**Before Implementation:**
- Resource type manifests had to be manually copied from resource-types-contrib to Radius
- No clear way to indicate which types should be registered by default
- Schema drift occurred when resource-types-contrib was updated but Radius wasn't
- Manual process was error-prone and time-consuming

## The Solution

### Approach: Metadata Flag

We selected the **Metadata Flag approach** (one of five evaluated alternatives) as it provides the best balance of simplicity, clarity, and maintainability.

**How It Works:**

1. **Marking for Sync**: In resource-types-contrib, add a single line to manifests:
   ```yaml
   defaultRegistration: true  # Marks this for automatic sync
   namespace: Applications.Databases
   types:
     postgresql:
       apiVersions:
         "2023-10-01-preview":
           schema: {...}
   ```

2. **Automatic Detection**: A GitHub Actions workflow (runs weekly or on-demand) identifies manifests marked with `defaultRegistration: true`

3. **Smart Sync**: The sync tool:
   - Parses manifests from resource-types-contrib
   - Removes the `defaultRegistration` field (not needed in Radius)
   - Compares content intelligently (ignores formatting differences)
   - Copies new files or updates existing ones

4. **Review & Merge**: Creates a pull request in Radius for team review before changes are applied

## Implementation Components

### 1. Manifest Parser Enhancement
**File**: `pkg/cli/manifest/manifest.go`
- Added optional `DefaultRegistration` field to existing `ResourceProvider` struct
- Minimal change: 5 lines
- Fully backwards compatible

### 2. Sync Tool
**Location**: `hack/sync-resource-types/`
- Command-line Go tool with dry-run and verbose modes
- Intelligent YAML comparison (handles formatting variations)
- Comprehensive error handling and reporting
- ~260 lines of production code + 260 lines of tests

### 3. GitHub Actions Workflow
**File**: `.github/workflows/sync-resource-types.yaml`
- **Automatic**: Runs weekly on Sundays at midnight UTC
- **Manual**: Can be triggered on-demand via GitHub Actions UI
- **Secure**: Only runs on main repository, requires PR review
- **Configurable**: Parameters for source repo, branch, and target directory

### 4. Documentation
- Design document analyzing 5 alternative approaches
- Comprehensive user guide (350+ lines)
- Tool documentation with examples
- Implementation summaries

## Workflow Example

```
┌─────────────────────────────────────────────────────────────┐
│ 1. Developer in resource-types-contrib                      │
│    Adds: defaultRegistration: true                          │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│ 2. Weekly GitHub Actions Workflow                           │
│    Detects marked manifests                                 │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│ 3. Sync Tool Processes                                      │
│    • Parses manifests                                       │
│    • Removes sync metadata                                  │
│    • Compares with Radius files                             │
│    • Identifies additions/updates                           │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│ 4. Pull Request Created in Radius                           │
│    Team reviews changes before merge                        │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│ 5. After Merge: Resource Types Auto-Registered              │
│    Applications can use new types immediately               │
└─────────────────────────────────────────────────────────────┘
```

## Key Benefits

### For Contributors
- ✅ **Simple**: Add one line (`defaultRegistration: true`) to mark manifests
- ✅ **Clear**: Obvious which types will be synced
- ✅ **No Setup**: No additional tools or permissions required

### For Maintainers
- ✅ **Automated**: Weekly sync with no manual intervention
- ✅ **Controlled**: All changes reviewed via PR before merge
- ✅ **Flexible**: Can trigger manual sync for urgent updates
- ✅ **Auditable**: Full Git history of all changes

### For Users
- ✅ **Current**: Resource types stay up-to-date automatically
- ✅ **Reliable**: Reduces human error in manual copying
- ✅ **Predictable**: Consistent weekly update schedule

## Security & Safety

- **Fork Protection**: Workflow only runs on main repository
- **Review Required**: No auto-merge; all changes need approval
- **Validation**: Manifests validated before syncing
- **Minimal Permissions**: Workflow uses least-privilege tokens
- **Audit Trail**: Complete Git history maintained

## Testing & Quality Assurance

✅ **Unit Tests**: 4 comprehensive test cases, all passing
✅ **Integration Tests**: Add/update scenarios verified
✅ **Manual Testing**: All functionality validated with demo scenarios
✅ **Code Quality**: Passes go vet, gofmt, and YAML validation
✅ **Zero Breaking Changes**: Fully backwards compatible

## Usage

### For Resource Type Authors (resource-types-contrib)
Add to your manifest:
```yaml
defaultRegistration: true
```

### For Radius Maintainers
**Automatic Mode**: Wait for Sunday's scheduled sync

**Manual Mode**: 
1. Go to GitHub Actions → sync-resource-types
2. Click "Run workflow"
3. Review and merge the generated PR

### For Local Testing
```bash
go run ./hack/sync-resource-types/main.go \
  --source <path-to-resource-types-contrib> \
  --target deploy/manifest/built-in-providers/dev \
  --dry-run --verbose
```

## Alternative Approaches Considered

We evaluated 5 different approaches before selecting the Metadata Flag:

1. **Metadata Flag** ⭐ **(Selected)** - Explicit marking in manifests
2. **Configuration File** - Centralized list in Radius repo
3. **Directory Convention** - Special folder for default types
4. **Hybrid Approach** - Combination of metadata and config
5. **Git Submodule** - Include resource-types-contrib as submodule

**Why Metadata Flag Won:**
- Clear and discoverable intent
- Minimal infrastructure changes
- Easy for contributors to understand
- Non-invasive to existing workflows
- Flexible for future enhancements

## Impact Analysis

### Code Changes
- **Files Modified**: 1 (added optional field)
- **Files Created**: 8 (tool, tests, docs, workflow)
- **Total Lines**: +1,106 (including comprehensive tests and documentation)
- **Breaking Changes**: None

### Deployment Requirements
- No infrastructure changes needed
- No database migrations required
- No API changes
- No dependency updates needed

## Future Enhancements

Potential improvements identified for future iterations:

1. **Version Tracking**: Record source commit SHA in synced files
2. **Partial Syncs**: Sync specific types within a manifest
3. **Conflict Detection**: Better handling of local modifications
4. **Bi-directional Sync**: Support syncing changes back to resource-types-contrib
5. **Enhanced Validation**: Deeper schema validation before sync
6. **Notifications**: Alerts for sync failures or important updates

## Deployment Plan

### Phase 1: Merge to Main (Ready Now)
- Merge this PR to Radius main branch
- Workflow is immediately available but inactive (no manifests marked yet)

### Phase 2: Mark Initial Types (Post-Merge)
- Identify resource types for default registration in resource-types-contrib
- Add `defaultRegistration: true` to selected manifests
- Commit and push changes

### Phase 3: First Sync (Within 1 Week)
- Wait for next Sunday's scheduled run OR trigger manually
- Review generated PR in Radius
- Merge after verification

### Phase 4: Monitor & Iterate
- Monitor weekly syncs for issues
- Gather feedback from team
- Implement enhancements as needed

## Success Metrics

### Immediate (Week 1-4)
- ✅ Workflow successfully runs weekly
- ✅ PRs created and merged without issues
- ✅ Zero manual manifest copying needed

### Short-term (Month 1-3)
- ✅ All identified default types marked and synced
- ✅ No schema drift incidents
- ✅ Contributors using the mechanism successfully

### Long-term (Month 3+)
- ✅ 100% of default types managed via sync
- ✅ Reduced maintenance burden
- ✅ Faster time-to-availability for new types

## Risks & Mitigations

| Risk | Mitigation |
|------|-----------|
| **Sync failures** | Comprehensive error handling, alerts in PR |
| **Schema conflicts** | PR review process catches issues before merge |
| **Workflow bugs** | Extensive testing, dry-run mode available |
| **Breaking changes** | Manifest validation, review required |
| **Resource-types-contrib unavailable** | Workflow fails gracefully, can retry manually |

## Questions & Support

### Common Questions

**Q: What if I don't want weekly syncs?**
A: You can adjust the cron schedule in the workflow file.

**Q: Can I sync from a different branch?**
A: Yes, use workflow_dispatch with custom branch parameter.

**Q: What if a manifest should no longer be default?**
A: Remove `defaultRegistration: true`. Manual deletion in Radius needed (by design to prevent accidental removal).

**Q: How do I test my manifest before marking it?**
A: Run the sync tool locally with --dry-run flag.

### Getting Help

- **Documentation**: `docs/contributing/resource-type-sync.md`
- **Tool Help**: `hack/sync-resource-types/README.md`
- **Issues**: Open issue in Radius repository
- **Questions**: Ask in Radius community Slack

## Approval Checklist

- [ ] Architecture review completed
- [ ] Security review completed
- [ ] Documentation reviewed
- [ ] Deployment plan approved
- [ ] Success metrics agreed upon
- [ ] Stakeholder sign-off obtained

## Conclusion

This implementation provides a **production-ready**, **low-risk**, **high-value** solution to automate resource type synchronization. It requires minimal changes to existing infrastructure while delivering significant benefits in reducing manual work and preventing schema drift.

The metadata flag approach is clear, simple, and maintainable—making it easy for contributors to use and for maintainers to operate. With comprehensive testing, documentation, and security controls in place, this solution is ready for immediate deployment.

---

**Prepared by**: GitHub Copilot  
**Date**: 2026-02-19  
**Implementation Status**: Complete and Ready for Review
