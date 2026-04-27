extension radius
extension radiusCompute
extension radiusData
extension radiusSecurity

param environment string

@secure()
param password string

@description('The full container image reference to build and push. Must be lowercase.')
param image string

resource app 'Applications.Core/applications@2023-10-01-preview' = {
  name: 'todo-list-app'
  properties: {
    environment: environment
  }
}

resource mysqlDatabase 'Radius.Data/mySqlDatabases@2025-08-01-preview' = {
  name: 'mysql'
  properties: {
    environment: environment
    application: app.id
    database: 'todos'
    version: '8.0'
    secretName: dbSecret.name
  }
}

resource dbSecret 'Radius.Security/secrets@2025-08-01-preview' = {
  name: 'db-secret'
  properties: {
    environment: environment
    application: app.id
    data: {
      USERNAME: {
        value: 'root'
      }
      PASSWORD: {
        value: password
      }
    }
  }
}

resource todoImage 'Radius.Compute/containerImages@2025-08-01-preview' = {
  name: 'todo-list-app-image'
  properties: {
    environment: environment
    application: app.id
    image: image
    build: {
      context: '/app/src/todo-list-app'
    }
  }
}

resource todoContainer 'Radius.Compute/containers@2025-08-01-preview' = {
  name: 'todo-list-app'
  properties: {
    environment: environment
    application: app.id
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
      mysqldb: {
        source: mysqlDatabase.id
      }
      todoContainerImage: {
        source: todoImage.id
      }
    }
  }
}
