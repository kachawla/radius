# Example: Todo-List-App (dockersamples/todo-list-app)

## Source analysis

- **Framework**: Node.js + Express.js
- **Port**: 3000
- **Persistence**: Swappable — SQLite (default) or MySQL (when `MYSQL_HOST` is set)
- **Env vars read by app**: `MYSQL_HOST`, `MYSQL_USER`, `MYSQL_PASSWORD`, `MYSQL_DB`
- **Compose**: MySQL 8.0 with persistent volume
- **Pattern**: B — Stateful / Database-Backed Application

## Resource mapping

| Source component | Radius Resource Type | Exists in resource-types-contrib? |
|---|---|---|
| Node.js container | `Radius.Compute/containers` | Yes |
| MySQL 8.0 | `Radius.Data/mySqlDatabases` | Yes (with Kubernetes Bicep recipe) |

## Correct generated `.radius/app.bicep`

```bicep
extension radius

param environment string

resource app 'Radius.Core/applications@2025-08-01-preview' = {
  name: 'todo-list-app'
  properties: {
    environment: environment
  }
}

resource database 'Radius.Data/mySqlDatabases@2025-08-01-preview' = {
  name: 'todo-database'
  properties: {
    environment: environment
    application: app.id
    version: '8.0'
  }
}

resource todoContainer 'Radius.Compute/containers@2025-08-01-preview' = {
  name: 'todo-app'
  properties: {
    environment: environment
    application: app.id
    containers: {
      todo: {
        image: 'ghcr.io/dockersamples/todo-list-app:latest'
        ports: {
          web: {
            containerPort: 3000
          }
        }
      }
    }
    connections: {
      mysqldb: {
        source: database.id
      }
    }
  }
}
```

## Key decisions explained

1. **No `version` comment** — `version: '8.0'` matches the compose.yaml. Clean, no explanation needed.
2. **No routes resource** — not added unless external ingress is explicitly required.
3. **No persistent volume** — database persistence is handled by the MySQL recipe.
4. **No explicit `env` mapping** — connection auto-injection handles `CONNECTION_MYSQLDB_*` env vars. The app's `src/persistence/index.js` should be updated to read `CONNECTION_MYSQLDB_HOST` instead of `MYSQL_HOST`.
5. **`connections` is at top level** — sibling of `containers`, NOT inside it.
6. **`containerPort: 3000`** — NOT `port: 3000`.
7. **No readOnly properties set** — `host`, `port`, `password` are output by the recipe.
8. **No skill-internal comments** — no "do not set readOnly" or "use existing recipe" comments in the output.