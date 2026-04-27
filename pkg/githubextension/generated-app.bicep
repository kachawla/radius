extension radius
param environment string

resource app 'Radius.Core/applications@2025-08-01-preview' = {
  name: 'todo-list-app'
  properties: {
    environment: environment
  }
}

// Backing data store (MySQL)
resource mysql 'Radius.Data/mySqlDatabases@2025-08-01-preview' = {
  name: 'todos-db'
  properties: {
    application: app.id
    environment: environment

    // Use the existing MySQL recipe from radius-project/resource-types-contrib.
    // (Do not set readOnly/output properties.)
  }
}

// Web/API container
resource web 'Radius.Compute/containers@2025-08-01-preview' = {
  name: 'web'
  properties: {
    application: app.id
    environment: environment

    containers: {
      web: {
        // Repo Dockerfile builds a Node app that listens on 3000.
        // Use a real/pullable image; replace with your published image.
        image: 'ghcr.io/dockersamples/todo-list-app:latest'
        ports: {
          http: {
            port: 3000
          }
        }

        // Connection to MySQL: Radius will auto-inject CONNECTION_* env vars.
        connections: [
          {
            name: 'mysql'
            source: mysql.id
          }
        ]

        // App expects MYSQL_* env vars (MYSQL_HOST/USER/PASSWORD/DB).
        // Per skill rules: do not duplicate auto-injected vars; instead disable defaults and map explicitly.
        disableDefaultEnvVars: true
        env: {
          MYSQL_HOST: '${mysql.properties.host}'
          MYSQL_USER: '${mysql.properties.username}'
          MYSQL_PASSWORD: '${mysql.properties.password}'
          MYSQL_DB: '${mysql.properties.database}'
        }
      }
    }
  }
}

// External ingress (the app is accessed on port 3000)
resource webRoute 'Radius.Compute/routes@2025-08-01-preview' = {
  name: 'web'
  properties: {
    application: app.id
    environment: environment

    // Route to the web container's HTTP port
    target: {
      resource: web.id
      port: 3000
    }
  }
}
