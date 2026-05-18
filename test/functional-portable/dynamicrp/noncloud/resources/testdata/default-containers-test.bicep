extension radius
extension containers

param environment string

resource app 'Radius.Core/applications@2025-08-01-preview' = {
  name: 'default-containers-app'
  properties: {
    environment: environment
  }
}

// Deploy a minimal Radius.Compute/containers resource using the default recipe.
// This validates the end-to-end path: manifest loaded at startup -> type
// registered -> recipe available -> container deployed.
resource container 'Radius.Compute/containers@2025-08-01-preview' = {
  name: 'default-container'
  properties: {
    environment: environment
    application: app.id
    containers: {
      web: {
        image: 'ghcr.io/radius-project/mirror/debian:latest'
        command: ['/bin/sh']
        args: ['-c', 'while true; do echo hello; sleep 10;done']
      }
    }
  }
}
