# Radius.Compute/persistentVolumes Schema

- **Type**: `Radius.Compute/persistentVolumes@2025-08-01-preview`
- **Source**: `radius-project/resource-types-contrib/Compute/persistentVolumes/persistentVolumes.yaml`
- **Existing recipe**: `Compute/persistentVolumes/recipes/kubernetes/bicep/kubernetes-volumes.bicep`

## Required properties

- `environment` (string): The Radius Environment ID.
- `sizeInGib` (integer): The size of the volume in gibibytes. `1` = 1024 MiB.

## Optional properties

- `application` (string): The Radius Application ID.
- `allowedAccessModes` (enum: ReadWriteOnce, ReadOnlyMany, ReadWriteMany): Assumed to be all modes if not specified.

## Valid Bicep structure

```bicep
resource myVolume 'Radius.Compute/persistentVolumes@2025-08-01-preview' = {
  name: 'my-volume'
  properties: {
    environment: environment          // REQUIRED
    application: app.id               // optional
    sizeInGib: 1                      // REQUIRED — integer, not string
  }
}
```

## Usage in a container

Reference via the container's `volumes` and `containers.*.volumeMounts`:

```bicep
resource myContainer 'Radius.Compute/containers@2025-08-01-preview' = {
  name: 'my-container'
  properties: {
    environment: environment
    application: app.id
    containers: {
      myapp: {
        image: 'nginx:alpine'
        volumeMounts: [
          {
            volumeName: 'data'
            mountPath: '/app/data'
          }
        ]
      }
    }
    volumes: {                        // TOP-LEVEL — sibling of "containers"
      data: {                         // key must match volumeName in volumeMounts
        persistentVolume: {
          resourceId: myVolume.id
          accessMode: 'ReadWriteOnce' // optional
        }
      }
    }
  }
}
```

## Common mistakes to avoid

- Do NOT add a persistent volume for database storage — the database recipe handles its own persistence
- `volumes` is a top-level property on the container resource, NOT inside `containers`
- `sizeInGib` is an integer, not a string