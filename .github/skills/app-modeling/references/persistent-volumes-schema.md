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

## Usage in a container

Reference via `volumes` and `volumeMounts`:

```bicep
volumes: {
  data: {
    persistentVolume: {
      resourceId: myPersistentVolume.id
      accessMode: 'ReadWriteOnce'
    }
  }
}
```