# Connection Auto-Injection

When a container has a `connections` entry, Radius automatically injects environment variables into the container for ALL properties of the connected resource.

## Naming pattern

```
CONNECTION_<CONNECTION-NAME>_<PROPERTY>
```

- `<CONNECTION-NAME>`: uppercase version of the connection key
- `<PROPERTY>`: uppercase version of the resource's property name

## Example

Connection named `mysqldb` to `Radius.Data/mySqlDatabases`:

```
CONNECTION_MYSQLDB_DATABASE
CONNECTION_MYSQLDB_USERNAME
CONNECTION_MYSQLDB_PASSWORD
CONNECTION_MYSQLDB_VERSION
CONNECTION_MYSQLDB_HOST
CONNECTION_MYSQLDB_PORT
```

## Rules

1. NEVER add manual `env` entries that duplicate auto-injected vars.
2. If the app uses different env var names (e.g. `MYSQL_HOST`), choose:
   - **Option A (preferred)**: Update app code to read `CONNECTION_*` vars. Remove `disableDefaultEnvVars`.
   - **Option B**: Set `disableDefaultEnvVars: true` and manually map env vars.
3. Sensitive properties (like `password`) are also auto-injected.