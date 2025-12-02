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

package bicep

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_ContainsEnvironmentResource(t *testing.T) {
	t.Run("Template with environment resource", func(t *testing.T) {
		template := map[string]any{
			"resources": []interface{}{
				map[string]interface{}{
					"type": EnvironmentResourceType,
					"name": "my-env",
				},
			},
		}
		result := ContainsEnvironmentResource(template)
		require.True(t, result)
	})

	t.Run("Template with environment resource - case insensitive", func(t *testing.T) {
		template := map[string]any{
			"resources": []interface{}{
				map[string]interface{}{
					"type": "applications.core/environments",
					"name": "my-env",
				},
			},
		}
		result := ContainsEnvironmentResource(template)
		require.True(t, result)
	})

	t.Run("Template with multiple resources including environment", func(t *testing.T) {
		template := map[string]any{
			"resources": []interface{}{
				map[string]interface{}{
					"type": "Applications.Core/applications",
					"name": "my-app",
				},
				map[string]interface{}{
					"type": "Applications.Core/environments",
					"name": "my-env",
				},
				map[string]interface{}{
					"type": "Applications.Core/containers",
					"name": "my-container",
				},
			},
		}
		result := ContainsEnvironmentResource(template)
		require.True(t, result)
	})

	t.Run("Template without environment resource", func(t *testing.T) {
		template := map[string]any{
			"resources": []interface{}{
				map[string]interface{}{
					"type": "Applications.Core/applications",
					"name": "my-app",
				},
				map[string]interface{}{
					"type": "Applications.Core/containers",
					"name": "my-container",
				},
			},
		}
		result := ContainsEnvironmentResource(template)
		require.False(t, result)
	})

	t.Run("Template with no resources", func(t *testing.T) {
		template := map[string]any{
			"resources": []interface{}{},
		}
		result := ContainsEnvironmentResource(template)
		require.False(t, result)
	})

	t.Run("Template with missing resources field", func(t *testing.T) {
		template := map[string]any{
			"parameters": map[string]any{},
		}
		result := ContainsEnvironmentResource(template)
		require.False(t, result)
	})

	t.Run("Nil template", func(t *testing.T) {
		result := ContainsEnvironmentResource(nil)
		require.False(t, result)
	})

	t.Run("Template with invalid resources format", func(t *testing.T) {
		template := map[string]any{
			"resources": "not an array",
		}
		result := ContainsEnvironmentResource(template)
		require.False(t, result)
	})

	t.Run("Template with invalid resource format", func(t *testing.T) {
		template := map[string]any{
			"resources": []interface{}{
				"not a map",
				map[string]interface{}{
					"type": "Applications.Core/environments",
					"name": "my-env",
				},
			},
		}
		result := ContainsEnvironmentResource(template)
		require.True(t, result)
	})
}
