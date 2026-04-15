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
2. Classify into exactly one architecture pattern. Read [architecture-patterns.md](references/architecture-patterns.md).
3. Resolve resource types from `radius-project/resource-types-contrib` — MUST match existing schemas before generating new ones. Read only the relevant YAML schema files.
4. Read [bicep-structure-rules.md](references/bicep-structure-rules.md) for correct Bicep structure.
5. Generate `.radius/app.bicep` and `.radius/bicepconfig.json`.
6. Validate against the checklist before outputting.

## Resource Type Resolution

Read the resource type YAML schema files from the `radius-project/resource-types-contrib` repository. Each resource type has a YAML file at `<Category>/<typeName>/<typeName>.yaml` that defines the exact properties, types, required fields, and readOnly annotations.

| Need | Resource Type | Schema file in `resource-types-contrib` |
|---|---|---|
| Container images (build from Dockerfile) | `Radius.Compute/containerImages` | `Compute/containerImages/containerImages.yaml` (PR #126 — read from `willdavsmith:containerimages-v2` branch until merged) |
| Containers | `Radius.Compute/containers` | `Compute/containers/containers.yaml` |
| MySQL | `Radius.Data/mySqlDatabases` | `Data/mySqlDatabases/mySqlDatabases.yaml` |
| PostgreSQL | `Radius.Data/postgreSqlDatabases` | `Data/postgreSqlDatabases/postgreSqlDatabases.yaml` |
| Neo4j | `Radius.Data/neo4jDatabases` | `Data/neo4jDatabases/neo4jDatabases.yaml` |
| Persistent storage | `Radius.Compute/persistentVolumes` | `Compute/persistentVolumes/persistentVolumes.yaml` |
| External ingress | `Radius.Compute/routes` | `Compute/routes/routes.yaml` |
| Secrets | `Radius.Security/secrets` | `Security/secrets/secrets.yaml` |

This is the COMPLETE list. There are no other resource types available. Do NOT use any type not listed above. Do NOT invent properties — use only what the YAML schema defines.

## app.bicep Structure (mandatory order)

```bicep
extension radius
extension containerImages               // only if building from Dockerfile
extension containers                    // only if using containerImages

param environment string

// 1. Application resource — always exactly one
// 2. Container image resources (if building from Dockerfile)
// 3. Data / infrastructure resources (databases, caches)
// 4. Container resources (with connections to images and infra)
// 5. Routes (only if external ingress needed)
```

Rules:
- Always start with `extension radius` then any additional extensions, then `param environment string`.
- Always declare exactly ONE `Radius.Core/applications` resource.
- If the app has a Dockerfile but no published image, add a `Radius.Compute/containerImages` resource. The container must reference the image via `myImage.properties.image` and have a connection to `myImage.id`.
- For each container service in the source app, emit exactly one `Radius.Compute/containers` resource.
- For each backing data store, emit the matching `Radius.Data/*` resource.
- Only add `Radius.Compute/routes` if the app needs external ingress. Service-to-service does NOT require routes.
- Only add `Radius.Compute/persistentVolumes` for file storage needs. Database persistence is handled by the database recipe.

## Bicep Configuration

Every project using `Radius.*` resource types needs a `.radius/bicepconfig.json` alongside `app.bicep`. Generate this when creating `app.bicep`:

```json
{
  "extensions": {
    "radius": "br:biceptypes.azurecr.io/radius:latest"
  },
  "experimentalFeaturesEnabled": {
    "extensibility": true,
    "dynamicTypeLoading": true
  }
}
```

Only include extensions actually used. The `radius` extension is always required. Add additional extensions for `Radius.*` resource types as needed.

## Connections

Wire containers to infrastructure via `connections`. Read [connection-auto-injection.md](references/connection-auto-injection.md) for the correct env var format.

**Critical**: `Radius.Compute/containers` injects a JSON blob (`CONNECTION_<NAME>_PROPERTIES`), NOT individual env vars. The app code must parse this JSON.

Rules:
- NEVER duplicate auto-injected env vars with manual `env` entries.
- If the source app uses different env var names (e.g. `MYSQL_HOST` instead of parsing `CONNECTION_MYSQLDB_PROPERTIES`), either update the app code (preferred), or set `disableDefaultEnvVars: true` and map manually.
- Only add explicit `env` entries for app-specific variables NOT covered by connection auto-injection.

## Bicep Structure Rules

Read [bicep-structure-rules.md](references/bicep-structure-rules.md) for all structural rules including valid Bicep patterns, common mistakes, and properties that do NOT exist.

## Naming

Read [naming-conventions.md](references/naming-conventions.md).

## Secrets

Read [secrets-handling.md](references/secrets-handling.md).

## Validation Checklist

Before outputting, verify ALL:
- [ ] Every resource type matches a YAML schema in `resource-types-contrib` — no invented types
- [ ] Every property used exists in that YAML schema — no invented properties
- [ ] Every apiVersion is `2025-08-01-preview`
- [ ] `extension radius` is the first line
- [ ] Additional extensions (`containerImages`, `containers`) declared if those resource types are used
- [ ] `param environment string` is declared
- [ ] Exactly one `Radius.Core/applications` resource
- [ ] Every container has `environment`, `application`, `containers`
- [ ] `connections` is a top-level property under `properties`, NOT inside `containers`
- [ ] `connections` is an object map, NOT an array
- [ ] Every `connections` source references a declared resource's `.id`
- [ ] `disableDefaultEnvVars` is on the connection entry, NOT on the container
- [ ] No manual `env` duplicates auto-injected connection vars
- [ ] No `readOnly` properties set on resource declarations
- [ ] Container images reference `containerImages.properties.image` or a real published image — never a bare base image
- [ ] Ports use `containerPort`, NOT `port`
- [ ] `.radius/bicepconfig.json` is generated alongside `app.bicep`
- [ ] No comments explaining skill rules appear in the output

## Guardrails

- ONLY use resource types listed in the Resource Type Resolution table above. If a type is not in that table, it does not exist. Do NOT invent resource types, do NOT invent properties on existing types, and do NOT reference schemas that are not in `resource-types-contrib`.
- Do NOT set readOnly properties — they are output by recipes at deploy time.
- Do NOT reference readOnly properties of other resources in Bicep (e.g. `database.properties.host`) — these are not available at compile time. Use connection auto-injection.
- Do NOT use array syntax where the schema specifies object maps (`connections`, `containers`, `ports`, `volumes`, `env` are all object maps).
- Do NOT place `connections` inside `containers` — it is a top-level property under `properties`.
- Do NOT include comments explaining skill rules or why properties are absent. The generated app.bicep must be clean, production-ready Bicep.
- Do NOT use a bare runtime base image (e.g. `node:22-alpine`) as the container image when the app has a Dockerfile. Use `Radius.Compute/containerImages` to build and push.
- Do NOT create a `Radius.Security/secrets` resource for database credentials. Database passwords are generated by the database recipe and auto-injected via connections.
- Ask for clarification if the app's architecture is ambiguous.

## Example

Read [todo-list-app-example.md](references/todo-list-app-example.md) for a complete worked example.