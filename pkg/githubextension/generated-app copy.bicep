extension radius
extension radiusCompute
extension radiusSecurity

param environment string

resource todoApp 'Radius.Core/applications@2025-08-01-preview' = {
  name: 'todo-list-app'
  properties: {
    environment: environment
  }
}

resource todoImage 'Radius.Compute/containerImages@2025-08-01-preview' = {
  name: 'todo-list'
  properties: {
    environment: environment
    application: todoApp.id
    image: 'ghcr.io/kachawla/docker-demo:latest'
    build: {
      context: '/app/src/todo-list-app'
    }
  }
}

resource database 'Radius.Data/mySqlDatabases@2025-08-01-preview' = {
  name: 'mysql'
  properties: {
    environment: environment
    application: todoApp.id
    database: 'todos'
    version: '8.0'
  }
}

resource todoContainer 'Radius.Compute/containers@2025-08-01-preview' = {
  name: 'todo-list-frontend'
  properties: {
    environment: environment
    application: todoApp.id
    containers: {
      todo: {
        image: todoImage.properties.image
        ports: {
          web: {
            containerPort: 3000
          }
        }
      }
    }
    connections: {
      image: {
        source: todoImage.id
      }
      mysqldb: {
        source: database.id
      }
    }
  }
}
