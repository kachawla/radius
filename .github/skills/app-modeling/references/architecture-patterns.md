# Architecture Patterns

Classify the application into exactly ONE of these patterns. This determines the resource composition.

## Pattern A: Stateless Web/API Service
- **Signals**: HTTP server, no database dependency, serves API or static content
- **Resources**: `Radius.Compute/containers` + optional `Radius.Compute/routes` for external ingress

## Pattern B: Stateful / Database-Backed Application
- **Signals**: HTTP server + database client library (mysql, mysql2, pg, sqlite3, mongoose, sequelize, prisma, etc.)
- **Resources**: `Radius.Compute/containers` + `Radius.Data/*` (matching database type) + optional `Radius.Compute/routes`

## Pattern C: Event-Driven Application
- **Signals**: Message queue client (amqplib, kafkajs, sqs-consumer, bull, etc.), pub/sub patterns
- **Resources**: `Radius.Compute/containers` + messaging resource type

## Pattern D: Batch Job
- **Signals**: No HTTP server, runs a task and exits, cron-like behavior
- **Resources**: `Radius.Compute/containers` with `restartPolicy: 'OnFailure'` or `'Never'`

## Pattern E: Streaming / Real-Time Processing Application
- **Signals**: WebSocket server (ws, socket.io), stream processing libraries
- **Resources**: `Radius.Compute/containers` + streaming resource type + optional `Radius.Compute/routes`

## How to classify

1. Check the package manifest for database client libraries → if present, **Pattern B**
2. Check for message queue libraries → if present, **Pattern C**
3. Check for HTTP server/framework → if present without DB, **Pattern A**
4. Check for streaming/WebSocket libraries → if present, **Pattern E**
5. If no HTTP server and runs to completion → **Pattern D**

## Valid resource composition per pattern

### Pattern B example (most common)

```bicep
extension radius

param environment string

resource app 'Radius.Core/applications@2025-08-01-preview' = {
  name: 'my-app'
  properties: {
    environment: environment
  }
}

resource database 'Radius.Data/mySqlDatabases@2025-08-01-preview' = {
  name: 'my-database'
  properties: {
    environment: environment
    application: app.id
  }
}

resource webContainer 'Radius.Compute/containers@2025-08-01-preview' = {
  name: 'web'
  properties: {
    environment: environment
    application: app.id
    containers: {
      web: {
        image: 'myapp:latest'
        ports: {
          http: {
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