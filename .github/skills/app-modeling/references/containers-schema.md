# Radius.Compute/containers Schema

- **Type**: `Radius.Compute/containers@2025-08-01-preview`
- **Source**: `radius-project/resource-types-contrib/Compute/containers/containers.yaml`

## Required properties

- `environment` (string): The Radius Environment ID. Value should be `environment`.
- `application` (string): The Radius Application ID. e.g. `app.id`.
- `containers` (object map): Map of one or more containers.

## Container properties (each entry in `containers`)

- `image` (string, **required**): Container image. e.g. `node:22-alpine`.
- `command` (array of strings, optional): Overrides ENTRYPOINT. e.g. `['/bin/sh', '-c']`.
- `args` (array of strings, optional): Overrides CMD.
- `env` (object map, optional): Each entry has `value` (string) or `valueFrom.secretKeyRef` with `secretName` and `key`.
- `workingDir` (string, optional): Working directory inside container.
- `ports` (object map, optional): Each entry has `containerPort` (integer, required) and `protocol` (enum: TCP, UDP, optional).
- `volumeMounts` (array, optional): Each entry has `volumeName` (string, required) and `mountPath` (string, required).
- `resources` (object, optional): `requests` and `limits` each with `cpu` (string) and `memoryInMib` (integer).
- `readinessProbe` (object, optional): `exec`, `httpGet`, or `tcpSocket` with timing fields.
- `livenessProbe` (object, optional): Same structure as readinessProbe.
- `initContainer` (boolean, optional): Set true if container should run and succeed before others start.

## Top-level optional properties

These go under `properties`, as siblings of `containers` — NOT inside a container entry:

- `connections` (object map): Each entry has `source` (string, required — resource ID) and `disableDefaultEnvVars` (boolean, optional).
- `volumes` (object map): Each entry supports `persistentVolume` (with `resourceId` and optional `accessMode`), `secretName`, or `emptyDir` (with optional `medium`).
- `restartPolicy` (enum: Always, OnFailure, Never)
- `replicas` (integer)
- `autoScaling` (object): `maxReplicas` (integer) and `metrics` (array with `kind`, optional `customMetric`, and `target`).
- `extensions` (object): `daprSidecar` with `appId`, `appPort`, `config`.
- `platformOptions` (object): Platform-specific pass-through properties.

## Valid Bicep structure

```bicep
resource myContainer 'Radius.Compute/containers@2025-08-01-preview' = {
  name: 'my-container'
  properties: {
    environment: environment          // REQUIRED
    application: app.id               // REQUIRED
    containers: {                     // REQUIRED — object map, NOT array
      myapp: {                        // key = container name (camelCase)
        image: 'node:22-alpine'       // REQUIRED — real pullable image
        ports: {                      // object map, NOT array
          web: {                      // key = port name
            containerPort: 3000       // REQUIRED — NOT "port"
            protocol: 'TCP'           // optional
          }
        }
        env: {                        // object map — only for vars NOT auto-injected
          MY_CUSTOM_VAR: {
            value: 'some-value'       // must use { value: '...' } syntax
          }
        }
      }
    }
    connections: {                    // TOP-LEVEL — sibling of "containers", NOT inside it
      mysqldb: {                     // object map, NOT array — key = connection name
        source: database.id          // REQUIRED — must be a declared resource's .id
        disableDefaultEnvVars: true  // optional — on the CONNECTION, not the container
      }
    }
  }
}
```

## Common mistakes to avoid

- `connections` is NOT inside `containers` — it is a sibling of `containers` under `properties`
- `connections` is an **object map**, NOT an array. Keys are connection names.
- `disableDefaultEnvVars` is a property of the **connection entry**, NOT the container
- Port property is `containerPort`, NOT `port`
- `env` values use `{ value: 'string' }` syntax, NOT bare strings or interpolation
- `containers` is an **object map**, NOT an array. Keys are container names.
- `ports` is an **object map**, NOT an array. Keys are port names.
- Do NOT reference readOnly properties of other resources (like `mysql.properties.host`) — these are not available at Bicep compile time. Use connection auto-injection instead.