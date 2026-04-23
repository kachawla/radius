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

package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestRemoveDefaultRegistrationField(t *testing.T) {
	input := []byte(`
defaultRegistration: true
namespace: Test.Provider
types:
  testType:
    apiVersions:
      "2023-10-01-preview":
        schema:
          type: object
`)

	result, err := removeDefaultRegistrationField(input)
	require.NoError(t, err)

	var data map[string]interface{}
	err = yaml.Unmarshal(result, &data)
	require.NoError(t, err)

	// Verify defaultRegistration field is removed
	_, exists := data["defaultRegistration"]
	require.False(t, exists, "defaultRegistration field should be removed")

	// Verify other fields remain
	require.Contains(t, data, "namespace")
	require.Contains(t, data, "types")
	require.Equal(t, "Test.Provider", data["namespace"])
}

func TestContentEqual(t *testing.T) {
	tests := []struct {
		name     string
		content1 string
		content2 string
		expected bool
	}{
		{
			name: "identical content",
			content1: `
namespace: Test.Provider
types:
  testType:
    apiVersions:
      "2023-10-01-preview":
        schema:
          type: object
`,
			content2: `
namespace: Test.Provider
types:
  testType:
    apiVersions:
      "2023-10-01-preview":
        schema:
          type: object
`,
			expected: true,
		},
		{
			name: "different formatting but same content",
			content1: `namespace: Test.Provider
types:
  testType:
    apiVersions:
      "2023-10-01-preview":
        schema: {type: object}`,
			content2: `
namespace: Test.Provider
types:
  testType:
    apiVersions:
      "2023-10-01-preview":
        schema:
          type: object
`,
			expected: true,
		},
		{
			name: "different content",
			content1: `
namespace: Test.Provider1
types:
  testType:
    apiVersions:
      "2023-10-01-preview":
        schema:
          type: object
`,
			content2: `
namespace: Test.Provider2
types:
  testType:
    apiVersions:
      "2023-10-01-preview":
        schema:
          type: object
`,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := contentEqual([]byte(tt.content1), []byte(tt.content2))
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestSyncResourceTypes_Integration(t *testing.T) {
	// Create temporary directories for the test
	tmpDir := t.TempDir()
	testSourceDir := filepath.Join(tmpDir, "source")
	testTargetDir := filepath.Join(tmpDir, "target")

	require.NoError(t, os.MkdirAll(testSourceDir, 0755))
	require.NoError(t, os.MkdirAll(testTargetDir, 0755))

	// Create test manifest files
	manifestWithDefault := `defaultRegistration: true
namespace: Test.WithDefault
types:
  testResource:
    apiVersions:
      "2023-10-01-preview":
        schema:
          type: object
          properties:
            name:
              type: string
`

	manifestWithoutDefault := `defaultRegistration: false
namespace: Test.WithoutDefault
types:
  anotherResource:
    apiVersions:
      "2023-10-01-preview":
        schema:
          type: object
          properties:
            value:
              type: string
`

	// Write test files
	require.NoError(t, os.WriteFile(filepath.Join(testSourceDir, "with_default.yaml"), []byte(manifestWithDefault), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(testSourceDir, "without_default.yaml"), []byte(manifestWithoutDefault), 0644))

	// Set up global variables for the sync
	sourceDir = testSourceDir
	targetDir = testTargetDir
	dryRun = false
	verbose = false

	// Run sync
	result := syncResourceTypes()

	// Verify results
	require.Empty(t, result.errors, "should not have errors")
	require.Len(t, result.addedFiles, 1, "should add one file")
	require.Contains(t, result.addedFiles, "with_default.yaml")
	require.NotContains(t, result.addedFiles, "without_default.yaml")

	// Verify the synced file exists and defaultRegistration is removed
	syncedFile := filepath.Join(testTargetDir, "with_default.yaml")
	require.FileExists(t, syncedFile)

	content, err := os.ReadFile(syncedFile)
	require.NoError(t, err)

	var data map[string]interface{}
	require.NoError(t, yaml.Unmarshal(content, &data))

	_, exists := data["defaultRegistration"]
	require.False(t, exists, "defaultRegistration should be removed from synced file")
	require.Equal(t, "Test.WithDefault", data["namespace"])
}

func TestSyncResourceTypes_Update(t *testing.T) {
	// Create temporary directories for the test
	tmpDir := t.TempDir()
	testSourceDir := filepath.Join(tmpDir, "source")
	testTargetDir := filepath.Join(tmpDir, "target")

	require.NoError(t, os.MkdirAll(testSourceDir, 0755))
	require.NoError(t, os.MkdirAll(testTargetDir, 0755))

	// Create initial manifest
	initialManifest := `defaultRegistration: true
namespace: Test.Provider
types:
  testResource:
    apiVersions:
      "2023-10-01-preview":
        schema:
          type: object
          properties:
            name:
              type: string
`

	// Create updated manifest with additional property
	updatedManifest := `defaultRegistration: true
namespace: Test.Provider
types:
  testResource:
    apiVersions:
      "2023-10-01-preview":
        schema:
          type: object
          properties:
            name:
              type: string
            description:
              type: string
`

	// Write initial version to both source and target
	sourceFile := filepath.Join(testSourceDir, "test.yaml")
	targetFile := filepath.Join(testTargetDir, "test.yaml")

	cleaned, err := removeDefaultRegistrationField([]byte(initialManifest))
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(targetFile, cleaned, 0644))

	// Write updated version to source
	require.NoError(t, os.WriteFile(sourceFile, []byte(updatedManifest), 0644))

	// Set up global variables for the sync
	sourceDir = testSourceDir
	targetDir = testTargetDir
	dryRun = false
	verbose = false

	// Run sync
	result := syncResourceTypes()

	// Verify results
	require.Empty(t, result.errors, "should not have errors")
	require.Len(t, result.updatedFiles, 1, "should update one file")
	require.Contains(t, result.updatedFiles, "test.yaml")

	// Verify the updated file has the new property
	content, err := os.ReadFile(targetFile)
	require.NoError(t, err)

	var data map[string]interface{}
	require.NoError(t, yaml.Unmarshal(content, &data))

	types := data["types"].(map[string]interface{})
	testResource := types["testResource"].(map[string]interface{})
	apiVersions := testResource["apiVersions"].(map[string]interface{})
	version := apiVersions["2023-10-01-preview"].(map[string]interface{})
	schema := version["schema"].(map[string]interface{})
	properties := schema["properties"].(map[string]interface{})

	require.Contains(t, properties, "name")
	require.Contains(t, properties, "description")
}
