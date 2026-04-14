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

## Valid Bicep structure

```bicep
// Option A (preferred): Let auto-injection handle everything.
// The app code should read CONNECTION_MYSQLDB_HOST, etc.
connections: {
  mysqldb: {
    source: database.id
  }
}

// Option B: Disable auto-injection and map env vars manually.
// Use ONLY when the app cannot change its env var names.
connections: {
  mysqldb: {
    source: database.id
    disableDefaultEnvVars: true
  }
}
```

When using Option B with `disableDefaultEnvVars: true`, you must add explicit `env` entries in the container. However, note that readOnly properties (like `host`, `port`, `password`) are only available at runtime — you cannot reference them as `database.properties.host` in Bicep. The recommended approach is always Option A.

## Rules

1. NEVER add manual `env` entries that duplicate auto-injected vars.
2. If the app uses different env var names (e.g. `MYSQL_HOST`), choose:
   - **Option A (preferred)**: Update app code to read `CONNECTION_*` vars. Let auto-injection handle everything.
   - **Option B (last resort)**: Set `disableDefaultEnvVars: true`. Note the limitation above.
3. Do NOT reference readOnly properties of other resources in Bicep (e.g. `database.properties.host`) — these are not available at compile time.

## Common mistakes to avoid

- Do NOT put `connections` inside `containers` — it is a top-level property under `properties`
- Do NOT use array syntax for `connections` — it is an object map
- Do NOT put `disableDefaultEnvVars` on the container — it goes on the individual connection entry
- Do NOT add `env` entries for `MYSQL_HOST`, `MYSQL_PASSWORD` etc. when auto-injection is enabled — they will conflict
- Do NOT use string interpolation like `'${database.properties.host}'` to reference readOnly outputs