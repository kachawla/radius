# Radius.Data/mySqlDatabases Schema

- **Type**: `Radius.Data/mySqlDatabases@2025-08-01-preview`
- **Source**: `radius-project/resource-types-contrib/Data/mySqlDatabases/mySqlDatabases.yaml`
- **Existing recipe**: `Data/mySqlDatabases/recipes/kubernetes/bicep/kubernetes-mysql.bicep`

## Required properties

- `environment` (string): The Radius Environment ID.

## Optional writable properties

- `application` (string): The Radius Application ID.
- `database` (string): The name of the database.
- `username` (string): The username for connecting to the database.
- `version` (string, enum: `5.7`, `8.0`, `8.4`): MySQL version. Assumed to be `8.4` if not specified.

## Read-only outputs (set by recipe — do NOT set in app.bicep)

- `password` (string): The password for connecting to the database.
- `host` (string): The host name used to connect to the database.
- `port` (integer): The port number used to connect to the database.

## Auto-injected env vars (via connection)

When a container connects with connection name `mysqldb`:

```
CONNECTION_MYSQLDB_DATABASE
CONNECTION_MYSQLDB_USERNAME
CONNECTION_MYSQLDB_PASSWORD
CONNECTION_MYSQLDB_VERSION
CONNECTION_MYSQLDB_HOST
CONNECTION_MYSQLDB_PORT
```

## Valid Bicep structure

```bicep
resource database 'Radius.Data/mySqlDatabases@2025-08-01-preview' = {
  name: 'my-database'
  properties: {
    environment: environment          // REQUIRED
    application: app.id               // optional but recommended
    version: '8.0'                    // optional — set when source app requires specific version
  }
}
```

## Common mistakes to avoid

- Do NOT set `host`, `port`, or `password` — these are readOnly, set by the recipe at deploy time
- Do NOT reference `database.properties.host` or `database.properties.password` in other resources — use connection auto-injection instead
- Do NOT add comments like "use existing recipe" or "do not set readOnly properties" — these are internal skill rules, not useful in generated Bicep
- Set `version` when the source app's compose.yaml or config specifies a MySQL version