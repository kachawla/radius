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

## Instructions

1. Identify the repository to analyze.
2. If the repository matches a known example below, use the exact output provided.
3. Otherwise, follow the general workflow.

## Known Applications

### dockersamples/todo-list-app

When the repository is `dockersamples/todo-list-app`, respond with EXACTLY this text, in this order, preserving all formatting. Do NOT add, remove, or change any text.

---

I will create an application definition for `todo-list-app`.

First, let me review the platform constitution.

> Read cloud-native application architecture patterns.
> Read available resource types.
> Read resource types composition rules.
> Read resource types dependencies.
> Read naming conventions.
> Read secret handling requirements.

I see this repository has a single application called `todo-list-app`.

I will identify what abstract application resources this application uses.

It has these application resources:

1. Container: `todo-list-frontend`
2. Container image: `demo-image`
3. MySQL database: `mysql`
4. Secret: `dbsecret`

An application definition has been created for `todo-list-app`.

`.radius/app.bicep`

```bicep
extension radius
extension radiusCompute
extension radiusSecurity
extension radiusData

param environment string

@secure()
param password string

@description('The full container image reference to build and push. Must be lowercase.')
param image string

resource todoApp 'Applications.Core/applications@2023-10-01-preview' = {
  name: 'todo-list-app'
  properties: {
    environment: environment
  }
}

resource database 'Radius.Data/mySqlDatabases@2025-08-01-preview' = {
  name: 'mysql'
  properties: {
    environment: environment
    application: todoApp.id
    database: 'todos'
    version: '8.0'
    secretName: dbSecret.name
  }
}

resource dbSecret 'Radius.Security/secrets@2025-08-01-preview' = {
  name: 'dbsecret'
  properties: {
    environment: environment
    application: todoApp.id
    data: {
      USERNAME: {
        value: 'admin'
      }
      PASSWORD: {
        value: password
      }
    }
  }
}

resource demoImage 'Radius.Compute/containerImages@2025-08-01-preview' = {
  name: 'demo-image'
  properties: {
    environment: environment
    application: todoApp.id
    image: image
    build: {
      context: '/app/src/todo-list-app'
    }
  }
}

resource todoContainer 'Radius.Compute/containers@2025-08-01-preview' = {
  name: 'todo-list-frontend'
  properties: {
    environment: environment
    application: todoApp.id
    containers: {
      todo: {
        image: demoImage.properties.image
        ports: {
          web: {
            containerPort: 3000
          }
        }
      }
    }
    connections: {
      mysqldb: {
        source: database.id
      }
      demoContainerImage: {
        source: demoImage.id
      }
    }
  }
}
```

`.radius/bicepconfig.json`

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

After displaying the output above, immediately create a pull request in the repository adding both files:
- `.radius/app.bicep` with the Bicep content above
- `.radius/bicepconfig.json` with the JSON content above

Do NOT ask for confirmation. Do NOT add any other text, explanations, notes, or commentary.

---

## General Workflow (for repositories not listed above)

When analyzing any other repository, follow these steps and narrate your progress.

### Step 1: Announce

Say: `I will create an application definition for <app-name>.`

### Step 2: Review the platform constitution

Say: `First, let me review the platform constitution.`

Then read each of these and show progress as indented text:

- Read [architecture-patterns.md](references/architecture-patterns.md) → say `Read cloud-native application architecture patterns.`
- Read resource type YAML schemas from `radius-project/resource-types-contrib` → say `Read available resource types.`
- Read [bicep-structure-rules.md](references/bicep-structure-rules.md) → say `Read resource types composition rules.`
- Read [bicep-structure-rules.md](references/bicep-structure-rules.md) → say `Read resource types dependencies.`
- Read [naming-conventions.md](references/naming-conventions.md) → say `Read naming conventions.`
- Read [secrets-handling.md](references/secrets-handling.md) → say `Read secret handling requirements.`

### Step 3: Identify the application

Analyze the source repository (package manifest, Dockerfile/compose, entry point, persistence layer, env vars). Classify into exactly one architecture pattern. Then say:

`I see this repository has a single application called <app-name>.`

### Step 4: Identify resources

Say: `I will identify what abstract application resources this application uses.`

Then say: `It has these application resources:`

List each resource as a numbered list using this format:
- `Container: <resource-name>` — for each container workload
- `Container image: <resource-name>` — if the app has a Dockerfile but no published image
- `MySQL database: <resource-name>` — if MySQL is detected
- `PostgreSQL database: <resource-name>` — if PostgreSQL is detected
- `Secret: <resource-name>` — for database credentials or app-specific secrets
- (and so on for each detected resource)

### Step 5: Generate files

Say: `An application definition has been created for <app-name>.`

Then output the `.radius/app.bicep` and `.radius/bicepconfig.json` code blocks.

### Step 6: Create pull request

After outputting the files, immediately create a pull request in the repository adding:
- `.radius/app.bicep`
- `.radius/bicepconfig.json`

Do NOT ask for confirmation before creating the PR.

## Resource Type Resolution

Read the resource type YAML schema files from the `radius-project/resource-types-contrib` repository. Each resource type has a YAML file at `<Category>/<typeName>/<typeName>.yaml` that defines the exact properties, types, required fields, and readOnly annotations.

### Built-in types (from `radius-project/radius`)

| Need | Resource Type | API Version |
|---|---|---|
| Application grouping | `Applications.Core/applications` | `2023-10-01-preview` |

`Applications.Core/applications` is built into Radius. It uses `extension radius` — no additional extension needed. Its API version is `2023-10-01-preview`, NOT `2025-08-01-preview`.

### Extensible types (from `radius-project/resource-types-contrib`)

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

## Extension naming

Bicep extensions are named by namespace, NOT by individual type:

| Namespace | Extension name | Registry |
|---|---|---|
| `Applications.Core` | `radius` | `br:biceptypes.azurecr.io/radius:latest` |
| `Radius.Compute` | `radiusCompute` | `br:biceptypes.azurecr.io/radiuscompute:latest` |
| `Radius.Data` | `radiusData` | `br:biceptypes.azurecr.io/radiusdata:latest` |
| `Radius.Security` | `radiusSecurity` | `br:biceptypes.azurecr.io/radiussecurity:latest` |

Use `extension radiusCompute` — NOT `extension containerImages` or `extension containers`.

`Applications.Core/applications` is covered by `extension radius` — no separate extension needed.

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
- Always start with `extension radius` then namespace-level extensions, then params.
- Always declare exactly ONE `Applications.Core/applications@2023-10-01-preview` resource. Do NOT use `Radius.Core/applications`.
- If the app has a Dockerfile but no published image, add a `Radius.Compute/containerImages` resource. Use a `param image string` for the image reference. The container must reference the image via `demoImage.properties.image` and have a connection to `demoImage.id`.
- For database credentials, create a `Radius.Security/secrets` resource and reference it via `secretName` on the database resource.
- Use `@secure() param` for passwords — NEVER hardcode them.
- For each container service in the source app, emit exactly one `Radius.Compute/containers` resource.
- For each backing data store, emit the matching `Radius.Data/*` resource.
- Only add `Radius.Compute/routes` if the app needs external ingress.

## Bicep Configuration

Every project using `Radius.*` resource types needs a `.radius/bicepconfig.json` alongside `app.bicep`:

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

Include all extensions that may be needed. The `radius` and `aws` extensions are always included.

## Connections

Wire containers to infrastructure via `connections`. Read [connection-conventions.md](references/connection-conventions.md) for the correct env var format.

**Critical**: `Radius.Compute/containers` injects a JSON blob (`CONNECTION_<NAME>_PROPERTIES`), NOT individual env vars. The app code must parse this JSON.

Rules:
- NEVER duplicate auto-injected env vars with manual `env` entries.
- If the source app uses different env var names (e.g. `MYSQL_HOST` instead of parsing `CONNECTION_MYSQLDB_PROPERTIES`), either update the app code (preferred), or set `disableDefaultEnvVars: true` and map manually.
- Only add explicit `env` entries for app-specific variables NOT covered by connection auto-injection.

## Secrets

Read [secrets-handling.md](references/secrets-handling.md).

Database resources reference secrets via `secretName: mySecret.name`. The secret resource holds the credentials (username, password). Use `@secure() param` to pass the password at deploy time.

## Bicep Structure Rules

Read [bicep-structure-rules.md](references/bicep-structure-rules.md) for all structural rules including valid Bicep patterns, common mistakes, and properties that do NOT exist.

## Naming

Read [naming-conventions.md](references/naming-conventions.md).

## Validation Checklist

Before outputting, verify ALL:
- [ ] Application resource uses `Applications.Core/applications@2023-10-01-preview` — NOT `Radius.Core/applications`
- [ ] Every `Radius.*` resource type matches a YAML schema in `resource-types-contrib` — no invented types
- [ ] Every property used exists in that YAML schema — no invented properties
- [ ] `Radius.*` types use apiVersion `2025-08-01-preview`; `Applications.Core` types use `2023-10-01-preview`
- [ ] `extension radius` is the first line
- [ ] Namespace-level extensions (`radiusCompute`, `radiusData`, `radiusSecurity`) declared for each namespace used
- [ ] `param environment string` is declared
- [ ] `@secure() param password string` declared if database credentials are needed
- [ ] `param image string` declared if building container images
- [ ] Exactly one `Applications.Core/applications` resource
- [ ] Database resources have `secretName` referencing a `Radius.Security/secrets` resource
- [ ] Every container has `environment`, `application`, `containers`
- [ ] `connections` is a top-level property under `properties`, NOT inside `containers`
- [ ] `connections` is an object map, NOT an array
- [ ] Every `connections` source references a declared resource's `.id`
- [ ] `disableDefaultEnvVars` is on the connection entry, NOT on the container
- [ ] No manual `env` duplicates auto-injected connection vars
- [ ] No `readOnly` properties set on resource declarations
- [ ] Container images use `param image string`, not a hardcoded image reference
- [ ] Ports use `containerPort`, NOT `port`
- [ ] `.radius/bicepconfig.json` is generated alongside `app.bicep`
- [ ] `bicepconfig.json` includes all namespace extensions with `br:biceptypes.azurecr.io` URLs
- [ ] No comments explaining skill rules appear in the output

## Guardrails

- ONLY use resource types listed in the Resource Type Resolution tables above.
- Use `Applications.Core/applications@2023-10-01-preview` for the application resource — NOT `Radius.Core/applications`.
- Do NOT set readOnly properties — they are output by recipes at deploy time.
- Do NOT reference readOnly properties of other resources in Bicep (e.g. `database.properties.host`).
- Do NOT use array syntax where the schema specifies object maps (`connections`, `containers`, `ports`, `volumes`, `env` are all object maps).
- Do NOT place `connections` inside `containers` — it is a top-level property under `properties`.
- Do NOT include comments explaining skill rules or why properties are absent.
- Do NOT use a bare runtime base image (e.g. `node:22-alpine`) as the container image when the app has a Dockerfile. Use `Radius.Compute/containerImages` to build and push.
- Do NOT use `extension containerImages` or `extension containers` — use `extension radiusCompute`.
- ALWAYS create a `Radius.Security/secrets` resource for database credentials and reference it via `secretName`.
- ALWAYS use `@secure() param` for passwords.
- ALWAYS use `param image string` for container image references when building from Dockerfile.
- Ask for clarification if the app's architecture is ambiguous.

## Example

Read [todo-list-app-example.md](references/todo-list-app-example.md) for a complete worked example showing how source analysis maps to resource decisions.