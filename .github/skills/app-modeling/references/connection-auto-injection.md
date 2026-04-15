# Connection Auto-Injection

When a container has a `connections` entry, Radius automatically injects environment variables into the container. **The format differs by container type.**

## Radius.Compute/containers (recipe-based) — JSON Properties Blob

All properties are packed into a single JSON environment variable:

```
CONNECTION_<NAME>_PROPERTIES={"host":"...","port":"...","database":"...","username":"...","password":"..."}
CONNECTION_<NAME>_ID=<resource-id>
CONNECTION_<NAME>_NAME=<connection-name>
CONNECTION_<NAME>_TYPE=<resource-type>
```

> Property names in the JSON blob are **lowercase** (matching the resource type schema). The connection name in the env var prefix is **UPPERCASE**.

### Example

Connection named `mysqldb` to `Radius.Data/mySqlDatabases`:

```
CONNECTION_MYSQLDB_PROPERTIES={"database":"todos","username":"root","password":"abc123","version":"8.0","host":"mysql-svc.default.svc.cluster.local","port":"3306"}
CONNECTION_MYSQLDB_ID=/planes/radius/local/.../Radius.Data/mySqlDatabases/todo-database
CONNECTION_MYSQLDB_NAME=mysqldb
CONNECTION_MYSQLDB_TYPE=Radius.Data/mySqlDatabases
```

### Application code must parse the JSON

The app must read `CONNECTION_<NAME>_PROPERTIES` and parse it as JSON. Portable helpers:

#### Node.js

```javascript
function getConnProp(connName, prop) {
  const propsJson = process.env[`CONNECTION_${connName}_PROPERTIES`];
  if (propsJson) {
    try {
      const props = JSON.parse(propsJson);
      return props[prop.toLowerCase()] || '';
    } catch (e) { /* fall through */ }
  }
  return process.env[`CONNECTION_${connName}_${prop}`] || '';
}

// Usage:
// const host = getConnProp('MYSQLDB', 'HOST');
// const port = getConnProp('MYSQLDB', 'PORT');
```

#### Go

```go
func getConnProp(connName, prop string) string {
    propsJSON := os.Getenv("CONNECTION_" + connName + "_PROPERTIES")
    if propsJSON != "" {
        var props map[string]interface{}
        if err := json.Unmarshal([]byte(propsJSON), &props); err == nil {
            if val, ok := props[strings.ToLower(prop)]; ok {
                return fmt.Sprintf("%v", val)
            }
        }
    }
    return os.Getenv("CONNECTION_" + connName + "_" + prop)
}
```

#### Python

```python
import json, os

def get_conn_prop(conn_name: str, prop: str) -> str:
    props_json = os.getenv(f"CONNECTION_{conn_name}_PROPERTIES", "")
    if props_json:
        try:
            props = json.loads(props_json)
            return str(props.get(prop.lower(), ""))
        except json.JSONDecodeError:
            pass
    return os.getenv(f"CONNECTION_{conn_name}_{prop}", "")
```

## Applications.Core/containers (built-in) — Individual Env Vars

Each readOnly property becomes a separate environment variable:

```
CONNECTION_<NAME>_HOST
CONNECTION_<NAME>_PORT
CONNECTION_<NAME>_DATABASE
CONNECTION_<NAME>_USERNAME
CONNECTION_<NAME>_PASSWORD
```

> This format is NOT used by `Radius.Compute/containers`.

## Valid Bicep structure

```bicep
// Option A (preferred): Let auto-injection handle everything.
// The app code must parse CONNECTION_MYSQLDB_PROPERTIES JSON.
connections: {
  mysqldb: {
    source: database.id
  }
}

// Option B: Disable auto-injection and map env vars manually.
// Use ONLY when the app cannot be changed to parse the JSON blob.
connections: {
  mysqldb: {
    source: database.id
    disableDefaultEnvVars: true
  }
}
```

## Rules

1. NEVER add manual `env` entries that duplicate auto-injected vars.
2. When using `Radius.Compute/containers`, the app must parse `CONNECTION_<NAME>_PROPERTIES` as JSON.
3. When using `disableDefaultEnvVars: true`, note that readOnly properties are only available at runtime — you cannot reference them as `database.properties.host` in Bicep.
4. Do NOT reference readOnly properties of other resources in Bicep.

## Common mistakes to avoid

- Do NOT put `connections` inside `containers` — it is a top-level property under `properties`
- Do NOT use array syntax for `connections` — it is an object map
- Do NOT put `disableDefaultEnvVars` on the container — it goes on the individual connection entry
- Do NOT assume individual env vars (`CONNECTION_MYSQLDB_HOST`) work with `Radius.Compute/containers` — they do NOT. The app must parse the `_PROPERTIES` JSON blob.
- Do NOT use string interpolation like `'${database.properties.host}'` to reference readOnly outputs