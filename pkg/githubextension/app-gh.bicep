extension radius
extension radiusCompute
extension radiusSecurity
extension radiusData

param environment string

@secure()
param password string

@description('The full container image reference to build and push. Must be lowercase.')
param image string

resource todoApp 'Radius.Core/applications@2025-08-01-preview' = {
  name: 'todo-list-app'
  properties: {
    environment: environment
  }
}

resource database 'Radius.Data/mySqlDatabases@2025-08-01-preview' = {
  name: 'mysql'
  properties: {
    environment: environment
    application: todoApp.id
    database: 'todos'
    version: '8.0'
    secretName: dbSecret.name
  }
}

resource dbSecret 'Radius.Security/secrets@2025-08-01-preview' = {
  name: 'dbsecret'
  properties: {
    environment: environment
    application: todoApp.id
    data: {
      USERNAME: {
        value: 'admin'
      }
      PASSWORD: {
        value: password
      }
    }
  }
}

resource demoImage 'Radius.Compute/containerImages@2025-08-01-preview' = {
  name: 'demo-image'
  properties: {
    environment: environment
    application: todoApp.id
    image: image
    build: {
      context: '/app/demo'
    }
  }
}

resource todoContainer 'Radius.Compute/containers@2025-08-01-preview' = {
  name: 'todo-list-frontend'
  properties: {
    environment: environment
    application: todoApp.id
    containers: {
      todo: {
        image: demoImage.properties.image
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
      demoContainerImage: {
        source: demoImage.id
      }
    }
  }
}
