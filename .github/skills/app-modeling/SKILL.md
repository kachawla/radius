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

## Output Format

Your entire visible response must be EXACTLY this structure and nothing else. No headings, no step labels, no explanations, no source analysis, no "Let me read..." preamble. Just these lines in order:

```
I will create an application definition for <app-name>.

First, let me review the platform constitution.

> Read cloud-native application architecture patterns.
> Read available resource types.
> Read resource types composition rules.
> Read resource types dependencies.
> Read naming conventions.
> Read secret handling requirements.

I see this repository has a single application called <app-name>.

I will identify what abstract application resources this application uses.

It has these application resources:

1. Container: <name>
2. Container image: <name>
3. <Database type> database: <name>
4. Secret: <name>

An application definition has been created for <app-name>.

<.radius/app.bicep code block>

<.radius/bicepconfig.json code block>
```

After outputting EXACTLY the above, silently create a pull request adding `.radius/app.bicep` and `.radius/bicepconfig.json` to the repository. The PR title should be `Add Radius application definition`. The PR description should say `Add .radius/app.bicep and .radius/bicepconfig.json for <app-name>.` — nothing else. Do NOT re-narrate the workflow in the PR. Do NOT repeat the generated files in chat after creating the PR. Do NOT ask for confirmation.

## Internal Workflow (do NOT show these steps to the user)

Internally, before producing the output above:

1. Analyze the source repository (package manifest, Dockerfile/compose, entry point, persistence layer, env vars).
2. Classify into exactly one architecture pattern. Read [architecture-patterns.md](references/architecture-patterns.md).
3. Resolve resource types from `radius-project/resource-types-contrib` — MUST match existing schemas. Read only the relevant YAML schema files.
4. Read [bicep-structure-rules.md](references/bicep-structure-rules.md) for correct Bicep structure.
5. Read [naming-conventions.md](references/naming-conventions.md) and apply the Deterministic Naming Rules below.
6. Read [secrets-handling.md](references/secrets-handling.md).
7. Generate the Bicep and validate against the checklist.

Then produce the output in the exact format above.

## Deterministic Naming Rules

These rules eliminate ambiguity. Apply them exactly.

### Symbolic names (left side of `=` in Bicep)

| Resource | Symbolic name |
|---|---|
| Application | `<shortName>App` where `<shortName>` is the app name without hyphens, camelCase (e.g., `todo-list-app` → `todoApp`) |
| Container | `<shortName>Container` (e.g., `todoContainer`) |
| Container image | `demoImage` (always) |
| Database | `database` (always) |
| Database secret | `dbSecret` (always) |
| Route | `<shortName>Route` (e.g., `todoRoute`) |

### Resource `name` properties (string values in Bicep)

| Resource | Name value |
|---|---|
| Application | Repository name in kebab-case (e.g., `'todo-list-app'`) |
| Container | `'<app-name>-frontend'` for single-container apps (e.g., `'todo-list-frontend'`) |
| Container image | `'demo-image'` (always) |
| Database | Short name of the database engine: `'mysql'`, `'postgres'`, `'neo4j'` |
| Database secret | `'dbsecret'` (always) |

### Connection keys

| Connection | Key |
|---|---|
| Database | `mysqldb`, `postgresdb`, `neo4jdb` (engine name + `db`) |
| Container image | `demoContainerImage` (always) |

### Other fixed values

| Field | Value |
|---|---|
| Database secret USERNAME | `'admin'` (always) |
| Container key in `containers` map | Short name derived from app (e.g., `todo` for todo-list-app) |
| Port key in `ports` map | `web` (always, for HTTP) |

### Extension order

Always declare extensions in this exact order:
1. `extension radius`
2. `extension radiusCompute`
3. `extension radiusSecurity`
4. `extension radiusData`

## Resource Type Resolution

### Built-in types (from `radius-project/radius`)

| Need | Resource Type | API Version |
|---|---|---|
| Application grouping | `Applications.Core/applications` | `2023-10-01-preview` |

`Applications.Core/applications` is built into Radius. It uses `extension radius` — no additional extension needed. Its API version is `2023-10-01-preview`. Do NOT use `Radius.Core/applications` — it does not exist.

### Extensible types (from `radius-project/resource-types-contrib`)

Read the resource type YAML schema files from the `radius-project/resource-types-contrib` repository. Each resource type has a YAML file at `<Category>/<typeName>/<typeName>.yaml`.

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

This is the COMPLETE list. Do NOT use any type not listed above. Do NOT invent properties. All extensible types use API version `2025-08-01-preview`.

## Extension naming

Bicep extensions are named by namespace, NOT by individual type:

| Namespace | Extension name | Registry |
|---|---|---|
| `Applications.Core` | `radius` | `br:biceptypes.azurecr.io/radius:latest` |
| `Radius.Compute` | `radiusCompute` | `br:biceptypes.azurecr.io/radiuscompute:latest` |
| `Radius.Data` | `radiusData` | `br:biceptypes.azurecr.io/radiusdata:latest` |
| `Radius.Security` | `radiusSecurity` | `br:biceptypes.azurecr.io/radiussecurity:latest` |

Use `extension radiusCompute` — NOT `extension containerImages` or `extension containers`.

## app.bicep Structure (mandatory order)

```bicep
extension radius
extension radiusCompute               // if using Radius.Compute/* types
extension radiusSecurity              // if using Radius.Security/* types
extension radiusData                  // if using Radius.Data/* types

param environment string

@secure()
param password string                 // if database credentials needed

@description('The full container image reference to build and push. Must be lowercase.')
param image string                    // if building container images

// 1. Application resource — always exactly one (Applications.Core/applications@2023-10-01-preview)
// 2. Data / infrastructure resources (databases, caches)
// 3. Secret resources (database credentials, API keys)
// 4. Container image resources (if building from Dockerfile)
// 5. Container resources (with connections to images and infra)
// 6. Routes (only if external ingress needed)
```

Rules:
- Always start with `extension radius` then namespace-level extensions in the fixed order, then params.
- Always declare exactly ONE `Applications.Core/applications@2023-10-01-preview` resource.
- If the app has a Dockerfile but no published image, add a `Radius.Compute/containerImages` resource. Use a `param image string` for the image reference. The container must reference the image via `demoImage.properties.image` and have a connection to `demoImage.id`.
- For database credentials, create a `Radius.Security/secrets` resource and reference it via `secretName` on the database resource.
- Use `@secure() param` for passwords — NEVER hardcode them.
- For each container service, emit exactly one `Radius.Compute/containers` resource.
- For each backing data store, emit the matching `Radius.Data/*` resource.
- Only add `Radius.Compute/routes` if the app needs external ingress.

## Bicep Configuration

Output this exactly — no modifications:

```json
{
  "experimentalFeaturesEnabled": {
    "extensibility": true
  },
  "extensions": {
    "radius": "br:biceptypes.azurecr.io/radius:latest",
    "radiusCompute": "br:biceptypes.azurecr.io/radiuscompute:latest",
    "radiusData": "br:biceptypes.azurecr.io/radiusdata:latest",
    "radiusSecurity": "br:biceptypes.azurecr.io/radiussecurity:latest",
    "aws": "br:biceptypes.azurecr.io/aws:latest"
  }
}
```

## Connections

Wire containers to infrastructure via `connections`. Read [connection-conventions.md](references/connection-conventions.md) for the correct env var format.

Rules:
- NEVER duplicate auto-injected env vars with manual `env` entries.
- Only add explicit `env` entries for app-specific variables NOT covered by connection auto-injection.

## Secrets

Read [secrets-handling.md](references/secrets-handling.md).

Database resources reference secrets via `secretName: dbSecret.name`. Username is always `'admin'`. Use `@secure() param` for the password.

## Bicep Structure Rules

Read [bicep-structure-rules.md](references/bicep-structure-rules.md) for all structural rules.

## Validation Checklist

Before outputting, verify ALL:
- [ ] Application resource uses `Applications.Core/applications@2023-10-01-preview`
- [ ] Every `Radius.*` type matches a YAML schema in `resource-types-contrib`
- [ ] `Radius.*` types use `2025-08-01-preview`; `Applications.Core` uses `2023-10-01-preview`
- [ ] Extensions are in order: `radius`, `radiusCompute`, `radiusSecurity`, `radiusData`
- [ ] All names follow the Deterministic Naming Rules exactly
- [ ] `param environment string` is declared
- [ ] `@secure() param password string` declared if database credentials are needed
- [ ] `param image string` declared if building container images
- [ ] Exactly one `Applications.Core/applications` resource
- [ ] Database resources have `secretName` referencing `dbSecret.name`
- [ ] Secret USERNAME is `'admin'`
- [ ] `connections` is a top-level property under `properties`, NOT inside `containers`
- [ ] `connections` is an object map, NOT an array
- [ ] Container images use `param image string`, not hardcoded
- [ ] Ports use `containerPort`, NOT `port`
- [ ] `bicepconfig.json` is exactly as shown above
- [ ] No comments or explanations in the generated Bicep
- [ ] No source analysis, step headings, or reasoning shown in chat

## Guardrails

- Use `Applications.Core/applications@2023-10-01-preview` — NOT `Radius.Core/applications`.
- Do NOT set readOnly properties.
- Do NOT reference readOnly properties of other resources in Bicep.
- Do NOT use array syntax where the schema specifies object maps.
- Do NOT place `connections` inside `containers`.
- Do NOT include comments in generated Bicep.
- Do NOT use a bare runtime base image when the app has a Dockerfile.
- Do NOT use `extension containerImages` or `extension containers` — use `extension radiusCompute`.
- ALWAYS create `Radius.Security/secrets` for database credentials.
- ALWAYS use `@secure() param` for passwords.
- ALWAYS use `param image string` for container image references when building from Dockerfile.

## Example

Read [todo-list-app-example.md](references/todo-list-app-example.md) for a complete worked example. The generated Bicep in that example is the **expected correct output** for `dockersamples/todo-list-app`.