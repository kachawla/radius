extension radius

@description('The URL of the server hosting test Terraform modules.')
param moduleServer string

@description('Name of the Radius Application.')
param appName string = 'compute-containers-app'

@description('Name of the Radius Environment.')
param envName string = 'compute-containers-env'

@description('Name of the container resource.')
param containerName string = 'compute-container'

resource env 'Applications.Core/environments@2023-10-01-preview' = {
  name: envName
  properties: {
    compute: {
      kind: 'kubernetes'
      resourceId: 'self'
      namespace: envName
    }
    recipes: {
      'Radius.Compute/containers': {
        default: {
          templateKind: 'terraform'
          templatePath: '${moduleServer}/radius-compute-containers.zip//terraform'
        }
      }
    }
  }
}

resource app 'Applications.Core/applications@2023-10-01-preview' = {
  name: appName
  properties: {
    environment: env.id
    extensions: [
      {
        kind: 'kubernetesNamespace'
        namespace: appName
      }
    ]
  }
}

resource container 'Radius.Compute/containers@2025-08-01-preview' = {
  name: containerName
  properties: {
    application: app.id
    environment: env.id
    containers: {
      demo: {
        image: 'ghcr.io/radius-project/samples/demo:latest'
        ports: {
          web: {
            containerPort: 3000
          }
        }
      }
    }
  }
}
