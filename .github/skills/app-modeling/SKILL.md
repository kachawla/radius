---
name: app-modeling
description: >
  Analyze a source code repository and generate a Radius application
  definition (.radius/app.bicep). Use when asked to create an application
  definition, model an application for Radius, or generate a Radius Bicep
  file. Resolves resource types from radius-project/resource-types-contrib
  and follows deterministic rules for validated output.
---

# Radius Application Modeling

Use this skill to generate a Radius application definition (`app.bicep`) from a source code repository.

## Workflow

1. Analyze the source repository (package manifest, Dockerfile/compose, entry point, persistence layer, env vars).
2. Classify into exactly one architecture pattern. Read [architecture-patterns](https://raw.githubusercontent.com/radius-project/radius/main/skills/app-modeling/references/architecture-patterns.md).
3. Resolve resource types from `radius-project/resource-types-contrib` — MUST match existing schemas before generating new ones. Read only the relevant schema references below.
4. Generate `.radius/app.bicep` following the structure and composition rules.
5. Validate against the checklist before outputting.

## Resource Type Resolution

For each infrastructure need identified in the source repo, search `radius-project/resource-types-contrib` for a matching type. Use the schema's exact namespace, apiVersion (`2025-08-01-preview`), property names, types, required fields, and readOnly annotations. Do NOT invent property names or types.

| Need | Resource Type | Schema Reference |
|---|---|---|
| Containers | `Radius.Compute/containers` | [containers-schema](https://raw.githubusercontent.com/radius-project/radius/main/skills/app-modeling/references/containers-schema.md) |
| MySQL | `Radius.Data/mySqlDatabases` | [mysql-schema](https://raw.githubusercontent.com/radius-project/radius/main/skills/app-modeling/references/mysql-schema.md) |
| Persistent storage | `Radius.Compute/persistentVolumes` | [persistent-volumes-schema](https://raw.githubusercontent.com/radius-project/radius/main/skills/app-modeling/references/persistent-volumes-schema.md) |
| External ingress | `Radius.Compute/routes` | [routes-schema](https://raw.githubusercontent.com/radius-project/radius/main/skills/app-modeling/references/routes-schema.md) |
| Secrets | `Radius.Security/secrets` | [secrets-schema](https://raw.githubusercontent.com/radius-project/radius/main/skills/app-modeling/references/secrets-schema.md) |

If no matching type exists in `resource-types-contrib`, generate a new schema following the [contribution guidelines](https://raw.githubusercontent.com/radius-project/resource-types-contrib/main/docs/contributing/contributing-resource-types-recipes.md).

## app.bicep Structure (mandatory order)

```bicep
extension radius
param environment string

resource app 'Radius.Core/applications@2025-08-01-preview' = {
  name: '<app-name-kebab-case>'
  properties: { environment: environment }
}
// Data/infra resources (databases, caches)
// Container resources (with connections to infra)
// Routes (only if external ingress needed)
```

Rules:
- Always start with `extension radius` then `param environment string`.
- Always declare exactly ONE `Radius.Core/applications` resource.
- For each container service in the source app, emit exactly one `Radius.Compute/containers` resource.
- For each backing data store, emit the matching `Radius.Data/*` resource.
- Only add `Radius.Compute/routes` if the app needs external ingress. Service-to-service does NOT require routes.
- Only add `Radius.Compute/persistentVolumes` for file storage needs. Database persistence is handled by the database recipe.

## Connections

Wire containers to infrastructure via `connections`. Auto-injection creates `CONNECTION_<NAME>_<PROPERTY>` env vars for all non-sensitive properties of the connected resource.

References:
- [connection-auto-injection](https://raw.githubusercontent.com/radius-project/radius/main/skills/app-modeling/references/connection-auto-injection.md)

Rules:
- NEVER duplicate auto-injected env vars with manual `env` entries.
- If the source app uses different env var names (e.g. `MYSQL_HOST` instead of `CONNECTION_MYSQLDB_HOST`), either update the app code to use `CONNECTION_*` names (preferred), or set `disableDefaultEnvVars: true` and map manually.
- Only add explicit `env` entries for app-specific variables NOT covered by connection auto-injection.

## Recipes

Check `radius-project/resource-types-contrib` for existing recipes at `<Category>/<typeName>/recipes/<platform>/<iac-language>/` before generating new ones. If a recipe exists, reference it. Do NOT regenerate.

## Naming

References:
- [naming-conventions](https://raw.githubusercontent.com/radius-project/radius/main/skills/app-modeling/references/naming-conventions.md)

## Secrets

References:
- [secrets-handling](https://raw.githubusercontent.com/radius-project/radius/main/skills/app-modeling/references/secrets-handling.md)

## Validation Checklist

Before outputting, verify ALL:
- [ ] Every resource type matches a schema in `resource-types-contrib` or is explicitly flagged as new
- [ ] Every apiVersion is `2025-08-01-preview`
- [ ] `extension radius` is the first line
- [ ] `param environment string` is declared
- [ ] Exactly one `Radius.Core/applications` resource
- [ ] Every container has `environment`, `application`, `containers`
- [ ] Every `connections` source references a declared resource's `.id`
- [ ] No manual `env` duplicates auto-injected connection vars (unless `disableDefaultEnvVars: true`)
- [ ] No `readOnly` properties set on resource declarations
- [ ] Container images are real, pullable images
- [ ] All output files go in `.radius/` directory

## Guardrails

- Do NOT invent property names not in the schema.
- Do NOT set readOnly properties — they are output by recipes at deploy time.
- Do NOT generate recipes if one already exists in `resource-types-contrib`.
- Ask for clarification if the app's architecture is ambiguous.

## Example

References:
- [todo-list-app-example](https://raw.githubusercontent.com/radius-project/radius/main/skills/app-modeling/references/todo-list-app-example.md)