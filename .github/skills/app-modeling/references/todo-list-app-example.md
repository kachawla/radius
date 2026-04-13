# Example: Todo-List-App (dockersamples/todo-list-app)

## Source analysis

- **Framework**: Node.js + Express.js
- **Port**: 3000
- **Persistence**: Swappable — SQLite (default) or MySQL (when `MYSQL_HOST` is set)
- **Env vars**: `MYSQL_HOST`, `MYSQL_USER`, `MYSQL_PASSWORD`, `MYSQL_DB`
- **Compose**: MySQL 8.0 with persistent volume
- **Pattern**: B — Stateful / Database-Backed Application

## Resource mapping

| Source component | Radius Resource Type | Exists in resource-types-contrib? |
|---|---|---|
| Node.js container | `Radius.Compute/containers` | Yes |
| MySQL 8.0 | `Radius.Data/mySqlDatabases` | Yes (with Kubernetes Bicep recipe) |

## Generated `.radius/app.bicep`

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
        image: 'node:22-alpine'
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

## Notes

- The app currently reads `MYSQL_HOST`, but Radius auto-injects `CONNECTION_MYSQLDB_HOST`.
  The app code at `src/persistence/index.js` should be updated to read `CONNECTION_MYSQLDB_HOST`
  instead of `MYSQL_HOST` (and similarly for all other DB env vars).
- MySQL version is set to `8.0` to match the `compose.yaml`.
- No persistent volume needed — database persistence is handled by the MySQL recipe.
- No routes resource needed unless external ingress is required.
- The existing `Data/mySqlDatabases/recipes/kubernetes/bicep/kubernetes-mysql.bicep` recipe is used — no new recipe needed.