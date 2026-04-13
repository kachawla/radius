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

## Read-only outputs (set by recipe, do NOT set in app.bicep)

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