/*
Copyright 2026 The Radius Authors.

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

package terraform

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/radius-project/radius/pkg/components/kubernetesclient/kubernetesclientprovider"
	"github.com/radius-project/radius/pkg/recipes"
	"github.com/stretchr/testify/require"
)

// TestOpenTofuLocalLifecycle exercises the real CLI without cloud resources or external providers.
func TestOpenTofuLocalLifecycle(t *testing.T) {
	if os.Getenv("RADIUS_TEST_OPENTOFU") != "1" {
		t.Skip("set RADIUS_TEST_OPENTOFU=1 to download the pinned OpenTofu binary and run a local lifecycle")
	}
	cliConfig := filepath.Join(t.TempDir(), "empty.tfrc")
	require.NoError(t, os.WriteFile(cliConfig, nil, 0600))
	t.Setenv(envTFCLIConfigFile, cliConfig)
	t.Setenv("CHECKPOINT_DISABLE", "1")

	archive := testZIP(t, map[string]string{"main.tf": `
terraform {
  required_providers {
    terraform = {
      source = "terraform.io/builtin/terraform"
    }
  }
}
variable "name" {
  type = string
}
resource "terraform_data" "example" {
  input = var.name
}
output "endpoint" {
  value = terraform_data.example.output
}
output "token" {
  value = "secret-${var.name}"
  sensitive = true
}
`})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/module.zip" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write(archive)
	}))
	defer server.Close()

	cacheDir := t.TempDir()
	e := NewOpenTofuExecutor(nil, nil, kubernetesclientprovider.KubernetesClientProvider{}, cacheDir, "")
	options := Options{
		RootDir: t.TempDir(),
		EnvRecipe: &recipes.EnvironmentDefinition{
			Name: "test-recipe", Driver: recipes.TemplateKindOpenTofu, TemplatePath: server.URL + "/module.zip",
			Parameters: map[string]any{"name": "first"},
		},
		ResourceRecipe: &recipes.ResourceMetadata{},
		EnvConfig:      &recipes.Configuration{},
		LogLevel:       "OFF",
	}

	metadata, err := e.GetRecipeMetadata(t.Context(), options)
	require.NoError(t, err)
	parameters := metadata["parameters"].(map[string]any)
	require.Equal(t, true, parameters["name"].(map[string]any)["required"])

	tf, err := e.installCLI(t.Context(), options)
	require.NoError(t, err)
	require.Equal(t, filepath.Join(cacheDir, ".opentofu", openTofuVersion, openTofuBinaryName()), tf.ExecPath())
	for _, name := range []string{"first", "updated"} {
		options.EnvRecipe.Parameters["name"] = name
		cfg, err := getTerraformConfig(t.Context(), tf.WorkingDir(), options)
		require.NoError(t, err)
		module, err := downloadAndInspect(t.Context(), tf, options)
		require.NoError(t, err)
		require.NoError(t, cfg.AddProviders(t.Context(), module.RequiredProviders, nil, options.EnvConfig, nil))
		require.NoError(t, cfg.AddAllOutputs(options.EnvRecipe.Name, module.OutputSensitivity))
		// Keep this smoke test local. Production uses the shared Kubernetes or cloud backend.
		cfg.Terraform.Backend = map[string]any{"local": map[string]any{}}
		require.NoError(t, cfg.Save(t.Context(), tf.WorkingDir()))
		state, err := initAndApply(t.Context(), tf, DefaultStateLockTimeout)
		require.NoError(t, err)
		require.Equal(t, name, state.Values.Outputs["endpoint"].Value)
		require.Equal(t, "secret-"+name, state.Values.Outputs["token"].Value)
		require.True(t, state.Values.Outputs["token"].Sensitive)
		require.Len(t, state.Values.RootModule.ChildModules[0].Resources, 1)
	}

	require.NoError(t, initAndDestroy(t.Context(), tf, DefaultStateLockTimeout))
	state, err := tf.Show(t.Context())
	require.NoError(t, err)
	if state.Values != nil {
		require.Empty(t, state.Values.Outputs)
		if state.Values.RootModule != nil {
			require.Empty(t, state.Values.RootModule.Resources)
			require.Empty(t, state.Values.RootModule.ChildModules)
		}
	}
	require.FileExists(t, tf.ExecPath(), "destroy must not remove the shared executable")
}
