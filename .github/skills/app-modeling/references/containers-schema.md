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
- `env` (object map, optional): Environment variables. Each entry has `value` (string) or `valueFrom.secretKeyRef` with `secretName` and `key`.
- `workingDir` (string, optional): Working directory inside container.
- `ports` (object map, optional): Each entry has `containerPort` (integer, required) and `protocol` (enum: TCP, UDP, optional).
- `volumeMounts` (array, optional): Each entry has `volumeName` (string, required) and `mountPath` (string, required).
- `resources` (object, optional): `requests` and `limits` each with `cpu` (string) and `memoryInMib` (integer).
- `readinessProbe` (object, optional): `exec`, `httpGet`, or `tcpSocket` with timing fields.
- `livenessProbe` (object, optional): Same structure as readinessProbe.
- `initContainer` (boolean, optional): Set true if container should run and succeed before others start.

## Top-level optional properties

- `connections` (object map): Each entry has `source` (string, required — resource ID) and `disableDefaultEnvVars` (boolean, optional).
- `volumes` (object map): Each entry supports `persistentVolume` (with `resourceId` and optional `accessMode`), `secretName`, or `emptyDir` (with optional `medium`).
- `restartPolicy` (enum: Always, OnFailure, Never)
- `replicas` (integer)
- `autoScaling` (object): `maxReplicas` (integer) and `metrics` (array with `kind`, optional `customMetric`, and `target`).
- `extensions` (object): `daprSidecar` with `appId`, `appPort`, `config`.
- `platformOptions` (object): Platform-specific pass-through properties.