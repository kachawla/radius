# Radius.Compute/routes Schema

- **Type**: `Radius.Compute/routes@2025-08-01-preview`
- **Source**: `radius-project/resource-types-contrib/Compute/routes/routes.yaml`

Only add a routes resource if the app needs external ingress from outside the cluster. Service-to-service communication does NOT require routes.

## Required properties

- `environment` (string): The Radius Environment ID.
- `application` (string): The Radius Application ID.
- `rules` (array): Each rule has `matches` (array) and `destinationContainer` (object).

## Optional properties

- `kind` (enum: HTTP, TCP, TLS, UDP): Assumed HTTP if not specified.
- `hostnames` (array of strings): Use only when kind is HTTP or TLS.

## Read-only outputs (set by recipe — do NOT set)

- `listener` (object): Contains `hostname` (string), `port` (integer), `protocol` (string). Set by the recipe.

## Valid Bicep structure

```bicep
resource myRoute 'Radius.Compute/routes@2025-08-01-preview' = {
  name: 'my-route'
  properties: {
    environment: environment          // REQUIRED
    application: app.id               // REQUIRED
    kind: 'HTTP'                      // optional — default is HTTP
    rules: [                          // REQUIRED — array of rule objects
      {
        matches: [                    // REQUIRED — array of match conditions
          {
            httpPath: '/'             // optional — match by path
          }
        ]
        destinationContainer: {       // REQUIRED — where to route traffic
          resourceId: myContainer.id  // REQUIRED — the container resource ID
          containerName: 'myapp'      // REQUIRED — specific container key within the resource
          containerPort: 3000         // REQUIRED — integer port number
        }
      }
    ]
  }
}
```

## Multiple rules example

```bicep
rules: [
  {
    matches: [
      { httpPath: '/' }
    ]
    destinationContainer: {
      resourceId: myContainer.id
      containerName: 'frontend'
      containerPort: 8080
    }
  }
  {
    matches: [
      { httpPath: '/api' }
    ]
    destinationContainer: {
      resourceId: myContainer.id
      containerName: 'backend'
      containerPort: 3000
    }
  }
]
```

## Common mistakes to avoid

- Do NOT use `target: { resource, port }` — this property does NOT exist
- Do NOT use `source`, `destination`, or `backend` — these are not valid properties
- Routes uses `rules` array with `matches` and `destinationContainer` — this is the ONLY valid structure
- `destinationContainer` requires ALL THREE fields: `resourceId`, `containerName`, `containerPort`
- `containerPort` is an integer, not a string
- `matches` is an array even if there is only one match condition
- `rules` is an array even if there is only one rule
- Do NOT add a route unless the app explicitly needs external ingress