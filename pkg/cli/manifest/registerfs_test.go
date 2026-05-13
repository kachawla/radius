/*
Copyright 2023 The Radius Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package manifest

import (
	"context"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegisterFS(t *testing.T) {
	t.Parallel()

	validManifest1 := `
namespace: Radius.Compute
types:
  containers:
    apiVersions:
      "2025-08-01-preview":
        schema: {}`

	validManifest2 := `
namespace: Radius.Compute
types:
  routes:
    apiVersions:
      "2025-08-01-preview":
        schema: {}`

	validManifest3 := `
namespace: Radius.Security
types:
  secrets:
    apiVersions:
      "2025-08-01-preview":
        schema: {}`

	t.Run("merges manifests sharing a namespace", func(t *testing.T) {
		t.Parallel()

		fsys := fstest.MapFS{
			"defaults.yaml": &fstest.MapFile{
				Data: []byte(`defaultRegistration:
  - Compute/containers/containers.yaml
  - Compute/routes/routes.yaml
  - Security/secrets/secrets.yaml
`),
			},
			"Compute/containers/containers.yaml": &fstest.MapFile{Data: []byte(validManifest1)},
			"Compute/routes/routes.yaml":         &fstest.MapFile{Data: []byte(validManifest2)},
			"Security/secrets/secrets.yaml":       &fstest.MapFile{Data: []byte(validManifest3)},
		}

		providers, err := RegisterFS(context.Background(), fsys)
		require.NoError(t, err)
		require.Len(t, providers, 2)

		var computeProvider *ResourceProvider
		var securityProvider *ResourceProvider
		for i := range providers {
			switch providers[i].Namespace {
			case "Radius.Compute":
				computeProvider = &providers[i]
			case "Radius.Security":
				securityProvider = &providers[i]
			}
		}

		require.NotNil(t, computeProvider)
		assert.Len(t, computeProvider.Types, 2)
		assert.Contains(t, computeProvider.Types, "containers")
		assert.Contains(t, computeProvider.Types, "routes")

		require.NotNil(t, securityProvider)
		assert.Len(t, securityProvider.Types, 1)
		assert.Contains(t, securityProvider.Types, "secrets")
	})

	t.Run("returns error when defaults.yaml is missing", func(t *testing.T) {
		t.Parallel()

		fsys := fstest.MapFS{}

		providers, err := RegisterFS(context.Background(), fsys)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to read defaults.yaml")
		assert.Nil(t, providers)
	})

	t.Run("returns nil when defaults.yaml has no entries", func(t *testing.T) {
		t.Parallel()

		fsys := fstest.MapFS{
			"defaults.yaml": &fstest.MapFile{
				Data: []byte("defaultRegistration:\n"),
			},
		}

		providers, err := RegisterFS(context.Background(), fsys)
		require.NoError(t, err)
		assert.Nil(t, providers)
	})

	t.Run("returns error when manifest path is missing from FS", func(t *testing.T) {
		t.Parallel()

		fsys := fstest.MapFS{
			"defaults.yaml": &fstest.MapFile{
				Data: []byte(`defaultRegistration:
  - nonexistent.yaml
`),
			},
		}

		providers, err := RegisterFS(context.Background(), fsys)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to read manifest nonexistent.yaml listed in defaults.yaml")
		assert.Nil(t, providers)
	})

	t.Run("returns error for invalid manifest YAML", func(t *testing.T) {
		t.Parallel()

		fsys := fstest.MapFS{
			"defaults.yaml": &fstest.MapFile{
				Data: []byte(`defaultRegistration:
  - bad.yaml
`),
			},
			"bad.yaml": &fstest.MapFile{Data: []byte("invalid: yaml: [")},
		}

		providers, err := RegisterFS(context.Background(), fsys)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse manifest bad.yaml")
		assert.Nil(t, providers)
	})

	t.Run("returns single provider without merging", func(t *testing.T) {
		t.Parallel()

		fsys := fstest.MapFS{
			"defaults.yaml": &fstest.MapFile{
				Data: []byte(`defaultRegistration:
  - Security/secrets/secrets.yaml
`),
			},
			"Security/secrets/secrets.yaml": &fstest.MapFile{Data: []byte(validManifest3)},
		}

		providers, err := RegisterFS(context.Background(), fsys)
		require.NoError(t, err)
		require.Len(t, providers, 1)
		assert.Equal(t, "Radius.Security", providers[0].Namespace)
		assert.Contains(t, providers[0].Types, "secrets")
	})
}
