# Naming Conventions

| Element | Convention | Example |
|---|---|---|
| Bicep symbolic name | camelCase, descriptive | `todoApp`, `mysqlDatabase`, `webContainer` |
| Resource `name` property | kebab-case, matches app/repo name | `'todo-list-app'`, `'my-database'` |
| Connection keys | camelCase, short, describes the target | `mysqldb`, `redis`, `storage` |
| Application name | kebab-case, matches repository name | `'todo-list-app'` |
| Container keys | camelCase, describes the container role | `todo`, `frontend`, `api` |