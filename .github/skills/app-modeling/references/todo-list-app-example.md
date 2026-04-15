# Example: Todo-List-App (dockersamples/todo-list-app)

## Source analysis

- **Framework**: Node.js + Express.js
- **Port**: 3000
- **Persistence**: Swappable — SQLite (default) or MySQL (when `MYSQL_HOST` is set)
- **Env vars read by app**: `MYSQL_HOST`, `MYSQL_USER`, `MYSQL_PASSWORD`, `MYSQL_DB`
- **Compose**: MySQL 8.0 with persistent volume
- **Dockerfile**: Yes — builds from `node:22-alpine`, runs `node src/index.js`
- **Published image**: No — must be built from Dockerfile
- **Pattern**: B — Stateful / Database-Backed Application

## Resource mapping

| Source component | Radius Resource Type | Exists in resource-types-contrib? |
|---|---|---|
| Dockerfile (build image) | `Radius.Compute/containerImages` | Yes |
| Node.js container | `Radius.Compute/containers` | Yes |
| MySQL 8.0 | `Radius.Data/mySqlDatabases` | Yes |
| DB credentials | `Radius.Security/secrets` | Yes |

## Key decisions explained

1. **`containerImages` resource** — the app has a Dockerfile but no published image. The `containerImages` resource builds and pushes it.
2. **`param image string`** — image reference is parameterized, not hardcoded.
3. **`build.context: '/app/src/todo-list-app'`** — the filesystem path where the repo source is volume-mounted on the Kubernetes node.
4. **`Radius.Security/secrets`** — database credentials (username + password) are stored in a secret resource. The database references it via `secretName: dbSecret.name`.
5. **`@secure() param password string`** — password is passed at deploy time, never hardcoded.
6. **`database: 'todos'`** — matches the `MYSQL_DATABASE: todos` from compose.yaml.
7. **`version: '8.0'`** — matches `mysql:8.0` from compose.yaml.
8. **Two connections on container** — `mysqldb` for database auto-injection; `demoContainerImage` for build ordering.
9. **No routes** — not added unless external ingress is explicitly required.
10. **App code change required** — `Radius.Compute/containers` injects a JSON blob via `CONNECTION_MYSQLDB_PROPERTIES`, not individual vars. The app's `src/persistence/index.js` must be updated to parse this JSON. See [connection-conventions.md](connection-conventions.md) for helper code.