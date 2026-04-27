extension radius

param environment string

resource app 'Radius.Core/applications@2025-08-01-preview' = {
  name: 'todo-list-app'
  properties: {
    environment: environment
  }
}

resource database 'Radius.Data/mySqlDatabases@2025-08-01-preview' = {
  name: 'todos-db'
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
