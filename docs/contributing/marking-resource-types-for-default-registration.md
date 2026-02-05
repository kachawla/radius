# Marking Resource Types for Default Registration

This guide explains how to mark a resource type in the resource-types-contrib repository for automatic registration in Radius.

## Overview

Resource types in the resource-types-contrib repository can be marked for "default registration," which means they will be automatically synced to the Radius repository and registered when Radius starts. This eliminates the need for users to manually register these resource types.

## When to Use Default Registration

Mark a resource type for default registration when:

- ✅ It provides core functionality needed by most Radius users
- ✅ It has stable schemas and is production-ready
- ✅ It should be available immediately without manual registration
- ✅ It follows all Radius resource type best practices and conventions

Do NOT mark a resource type for default registration when:

- ❌ It's experimental or under active development
- ❌ It's specific to a particular organization or narrow use case
- ❌ It has dependencies that may not be available in all environments
- ❌ It's primarily for testing, examples, or demonstrations

## How to Mark a Resource Type

To mark a resource type for default registration, add the `defaultRegistration: true` field at the top level of your resource type YAML file:

```yaml
# This field indicates that this resource type should be automatically
# registered by default in Radius installations
defaultRegistration: true

# Standard resource type definition follows
namespace: MyCompany.Resources
types:
  myResourceType:
    description: A resource type description
    apiVersions:
      2023-10-01-preview:
        schema:
          type: object
          properties:
            environment:
              type: string
              description: The Radius environment
            application:
              type: string
              description: The Radius application
          required:
            - environment
            - application
```

## File Format Requirements

Your resource type file must follow the Radius manifest format:

### Required Fields

- `namespace` - The resource provider namespace (e.g., `MyCompany.Resources`)
- `types` - A map of resource type definitions

### Optional Fields

- `defaultRegistration` - Set to `true` to enable default registration
- `location` - Location mapping for the resource provider
- `description` - Description of the resource provider

### Resource Type Fields

Each resource type must have:

- `apiVersions` - Map of API version to schema
- `schema` - JSON Schema defining the resource properties

Optionally:

- `description` - Description of the resource type
- `capabilities` - List of capabilities (e.g., `ManualResourceProvisioning`)
- `defaultApiVersion` - The default API version to use

## Validation

Your resource type file will be validated during the sync process. The validation checks:

1. ✅ File is valid YAML
2. ✅ Required fields (`namespace`, `types`) are present
3. ✅ `types` field is a dictionary/map
4. ✅ Each resource type has valid `apiVersions`
5. ✅ Schemas follow JSON Schema format

If validation fails, the file will not be synced and an error will be reported in the sync workflow logs.

## Sync Process

Once you mark a resource type for default registration:

1. **Daily Sync**: A GitHub Actions workflow runs daily in the Radius repository to check for updates
2. **Detection**: The workflow identifies files with `defaultRegistration: true`
3. **Sync**: Marked files are copied to `deploy/manifest/built-in-providers/self-hosted/` with a `synced_` prefix
4. **PR Creation**: If changes are detected, a pull request is automatically created
5. **Review**: Radius maintainers review and approve the PR
6. **Merge**: Once merged, the resource type is available in the next Radius release

## Updating a Default Resource Type

To update a resource type that's already marked for default registration:

1. Make your changes in the resource-types-contrib repository
2. Ensure `defaultRegistration: true` is still set
3. Commit and push your changes
4. Wait for the next sync run (or manually trigger it)
5. Review the automatically created PR in the Radius repository

⚠️ **Important**: Never edit the `synced_*` files directly in the Radius repository. All changes must be made in resource-types-contrib.

## Removing Default Registration

To remove a resource type from default registration:

1. Change `defaultRegistration: true` to `defaultRegistration: false` (or remove the field)
2. Commit and push your changes
3. Manually create a PR in the Radius repository to remove the corresponding `synced_*` file

## Testing

Before marking a resource type for default registration:

1. **Test Locally**: Validate your YAML file locally using the Radius CLI:
   ```bash
   rad resource-type create --from-file your-resource-type.yaml
   ```

2. **Create Test Environment**: Test the resource type in a development Radius environment

3. **Write Documentation**: Ensure your resource type is well-documented with clear examples

4. **Follow Best Practices**: Review the [Radius resource type documentation](https://docs.radapp.io/) for best practices

## Example

Here's a complete example of a resource type marked for default registration:

```yaml
defaultRegistration: true

namespace: Example.Database
types:
  postgresInstances:
    description: |
      A PostgreSQL database instance resource type.
      Provisions and manages PostgreSQL databases in Radius environments.
    
    capabilities:
      - ManualResourceProvisioning
    
    defaultApiVersion: "2024-01-01-preview"
    
    apiVersions:
      "2024-01-01-preview":
        schema:
          type: object
          properties:
            environment:
              type: string
              description: The Radius environment resource ID
            application:
              type: string
              description: The Radius application resource ID
            host:
              type: string
              description: PostgreSQL server hostname
            port:
              type: integer
              description: PostgreSQL server port
              default: 5432
            database:
              type: string
              description: Database name
            username:
              type: string
              description: Database username
            version:
              type: string
              description: PostgreSQL version
              enum:
                - "13"
                - "14"
                - "15"
                - "16"
          required:
            - environment
            - application
            - host
            - database
            - username
```

## Troubleshooting

### My resource type isn't being synced

1. Verify `defaultRegistration: true` is set at the top level of the YAML
2. Check that the file is in the correct directory in resource-types-contrib
3. Review the sync workflow logs in the Radius repository for errors
4. Ensure the file passes validation

### Validation errors

Common validation errors and fixes:

| Error | Fix |
|-------|-----|
| "Missing required field: namespace" | Add `namespace:` field to your YAML |
| "Missing required field: types" | Add `types:` field with at least one resource type |
| "Field 'types' must be a dictionary" | Ensure `types:` is a map/dictionary, not a list |
| "YAML validation error" | Check for syntax errors in your YAML file |

### Changes not appearing

- The sync runs daily at 2 AM UTC - wait for the next run or manually trigger it
- Check if a PR already exists with your changes
- Review recent sync workflow runs for errors

## Getting Help

If you have questions or issues:

1. Check the [Resource Type Sync Documentation](https://github.com/radius-project/radius/blob/main/docs/contributing/resource-type-sync.md)
2. Review existing resource types in resource-types-contrib for examples
3. Ask in the [Radius Discord](https://discord.gg/SRG3ePMKNy) #dev channel
4. Open an issue in the [Radius repository](https://github.com/radius-project/radius/issues)

## References

- [Radius Resource Type Documentation](https://docs.radapp.io/)
- [Resource Type Sync Mechanism](https://github.com/radius-project/radius/blob/main/docs/contributing/resource-type-sync.md)
- [Resource Type Manifest Format](https://github.com/radius-project/radius/blob/main/pkg/cli/manifest/manifest.go)
