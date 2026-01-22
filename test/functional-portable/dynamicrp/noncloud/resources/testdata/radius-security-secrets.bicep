extension radius

@description('Specifies the location for resources.')
param location string = 'global'

@description('Secret value for testing')
@secure()
param secretValue string

resource env 'Applications.Core/environments@2023-10-01-preview' = {
  name: 'secrets-test-env'
  location: location
  properties: {
    compute: {
      kind: 'kubernetes'
      resourceId: 'self'
      namespace: 'secrets-test-env'
    }
  }
}

resource app 'Applications.Core/applications@2023-10-01-preview' = {
  name: 'secrets-test-app'
  location: location
  properties: {
    environment: env.id
  }
}

resource mySecret 'Radius.Security/secrets@2025-08-01-preview' = {
  name: 'test-secret'
  properties: {
    environment: env.id
    application: app.id
    data: {
      testKey: {
        value: secretValue
        encoding: 'string'
      }
      testKeyBase64: {
        value: 'SGVsbG8gV29ybGQh'
        encoding: 'base64'
      }
    }
  }
}

resource container 'Applications.Core/containers@2023-10-01-preview' = {
  name: 'secrets-test-cntr'
  properties: {
    application: app.id
    container: {
      image: 'ghcr.io/radius-project/samples/demo:latest'
      env: {
        SECRET_VALUE: {
          valueFrom: {
            secretRef: {
              source: mySecret.id
              key: 'testKey'
            }
          }
        }
      }
    }
  }
}
