import radius as radius

@description('Radius environment ID')
param environment string

@description('Radius application ID')
param application string

@description('Secret value for testing')
@secure()
param secretValue string

resource app 'Applications.Core/applications@2023-10-01-preview' existing = {
  name: application
}

resource mySecret 'Radius.Security/secrets@2025-08-01-preview' = {
  name: 'test-secret'
  properties: {
    environment: environment
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
