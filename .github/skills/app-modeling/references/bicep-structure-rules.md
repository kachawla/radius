# Bicep Structure Rules

These rules apply to ALL generated app.bicep files. Read the resource type YAML schema from `radius-project/resource-types-contrib` for property names and types. This file covers structural patterns only.

## General

- `extension radius` is always the first line
- Namespace-level extensions: `radiusCompute`, `radiusData`, `radiusSecurity` — declared only if those namespaces are used
- Do NOT use `extension containerImages` or `extension containers` — use `extension radiusCompute`
- `param environment string` always declared
- `@secure() param password string` declared if database credentials are needed
- `param image string` declared if building container images
- Exactly ONE `Applications.Core/applications@2023-10-01-preview` resource
- All output files go in `.radius/` directory

## Radius.Compute/containers structure

```bicep
resource myContainer 'Radius.Compute/containers@2025-08-01-preview' = {
  name: 'my-container'
  properties: {
    environment: environment
    application: app.id
    containers: {                     // object map, NOT array
      myapp: {                        // key = container name (camelCase)
        image: myImage.properties.image
        ports: {                      // object map, NOT array
          web: {
            containerPort: 3000       // NOT "port"
          }
        }
        env: {                        // only for vars NOT auto-injected
          MY_VAR: {
            value: 'some-value'       // must use { value: '...' } syntax
          }
        }
      }
    }
    connections: {                    // TOP-LEVEL — sibling of "containers"
      mysqldb: {                     // object map, NOT array
        source: database.id
      }
      demoContainerImage: {
        source: myImage.id
      }
    }
  }
}
```

Rules:
- `containers` is an object map, NOT an array
- `ports` is an object map, NOT an array
- `connections` is an object map, NOT an array
- `connections` is a TOP-LEVEL property under `properties` — NOT inside `containers`
- `disableDefaultEnvVars` goes on the connection entry, NOT on the container
- Port property is `containerPort`, NOT `port`
- `env` values use `{ value: 'string' }` syntax, NOT bare strings
- Do NOT reference readOnly properties of other resources (e.g. `database.properties.host`)

## Radius.Compute/containerImages structure

```bicep
@description('The full container image reference to build and push. Must be lowercase.')
param image string

resource myImage 'Radius.Compute/containerImages@2025-08-01-preview' = {
  name: 'demo-image'
  properties: {
    environment: environment
    application: app.id
    image: image
    build: {
      context: '/app/demo'
    }
  }
}
```

Rules:
- Uses `extension radiusCompute` (NOT `extension containerImages`)
- Image reference comes from `param image string` — NOT hardcoded
- Container must reference image as `myImage.properties.image`
- Container must have a connection to `myImage.id` for dependency ordering
- `image` must be lowercase
- `build.context` is the filesystem path where the repo source is volume-mounted on the Kubernetes node

## Radius.Data/* structure

```bicep
resource database 'Radius.Data/mySqlDatabases@2025-08-01-preview' = {
  name: 'mysql'
  properties: {
    environment: environment
    application: app.id
    database: 'todos'
    version: '8.0'
    secretName: dbSecret.name
  }
}
```

Rules:
- Uses `extension radiusData` (NOT individual type extensions)
- `secretName` references a `Radius.Security/secrets` resource for credentials
- Do NOT set readOnly properties (`host`, `port`) — these are output by the recipe

## Radius.Security/secrets structure

```bicep
@secure()
param password string

resource dbSecret 'Radius.Security/secrets@2025-08-01-preview' = {
  name: 'dbsecret'
  properties: {
    environment: environment
    application: app.id
    data: {
      USERNAME: {
        value: 'todo_list_app_user'
      }
      PASSWORD: {
        value: password
      }
    }
  }
}
```

Rules:
- Uses `extension radiusSecurity`
- ALWAYS create for database credentials — referenced via `secretName` on the database resource
- NEVER hardcode passwords — use `@secure() param`
- `data` is an object map, NOT an array
- Keys in `data` are UPPERCASE (`USERNAME`, `PASSWORD`)

## Radius.Compute/routes structure

```bicep
resource myRoute 'Radius.Compute/routes@2025-08-01-preview' = {
  name: 'my-route'
  properties: {
    environment: environment
    application: app.id
    rules: [
      {
        matches: [
          { httpPath: '/' }
        ]
        destinationContainer: {
          resourceId: myContainer.id
          containerName: 'myapp'
          containerPort: 3000
        }
      }
    ]
  }
}
```

Rules:
- Do NOT use `target`, `source`, `destination`, or `backend` — these do NOT exist
- `rules` is the ONLY valid structure — array of objects with `matches` and `destinationContainer`
- `destinationContainer` requires ALL THREE: `resourceId`, `containerName`, `containerPort`
- Only add routes if external ingress is needed

## Image resolution

1. If the repo publishes a pre-built container image, use it directly
2. If the repo has a Dockerfile but no published image, use `Radius.Compute/containerImages` with `param image string`
3. Do NOT use a bare runtime base image (e.g. `node:22-alpine`) — it runs without app code

## Properties that do NOT exist

These are commonly hallucinated. They will cause deployment errors:

| Resource Type | Invalid property |
|---|---|
| `Radius.Compute/containers` | `port` (use `containerPort`), `image` at top level |
| `Radius.Compute/routes` | `target`, `source`, `destination`, `backend` |

## Output rules

- Do NOT include comments explaining skill rules in generated Bicep
- Do NOT set readOnly properties
- Do NOT add `@description` decorators unless the user asks for them (exception: `param image string` always gets a description)