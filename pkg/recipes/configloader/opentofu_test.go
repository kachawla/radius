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

package configloader

import (
	"context"
	"net/http"
	"testing"

	armpolicy "github.com/Azure/azure-sdk-for-go/sdk/azcore/arm/policy"
	azfake "github.com/Azure/azure-sdk-for-go/sdk/azcore/fake"
	v20231001 "github.com/radius-project/radius/pkg/corerp/api/v20231001preview"
	v20250801 "github.com/radius-project/radius/pkg/corerp/api/v20250801preview"
	"github.com/radius-project/radius/pkg/corerp/api/v20250801preview/fake"
	"github.com/radius-project/radius/pkg/recipes"
	"github.com/radius-project/radius/pkg/to"
	"github.com/stretchr/testify/require"
)

func TestOpenTofuEnvironmentRecipe(t *testing.T) {
	t.Parallel()
	for _, version := range []*string{nil, new("1.0.0")} {
		resource := &v20231001.EnvironmentResource{Properties: &v20231001.EnvironmentProperties{
			Recipes: map[string]map[string]v20231001.RecipePropertiesClassification{
				"Test.Resources/widgets": {
					"default": &v20231001.OpenTofuRecipeProperties{
						TemplateKind:    to.Ptr(recipes.TemplateKindOpenTofu),
						TemplatePath:    new("example/widgets/aws"),
						TemplateVersion: version,
					},
				},
			},
		}}
		definition, err := getRecipeDefinition(resource, &recipes.ResourceMetadata{
			Name: "default", ResourceID: "/planes/radius/local/resourceGroups/rg/providers/Test.Resources/widgets/widget",
		})
		require.NoError(t, err)
		require.Equal(t, recipes.TemplateKindOpenTofu, definition.Driver)
		require.Equal(t, to.String(version), definition.TemplateVersion)
	}
}

func TestModuleRecipePackDefinition(t *testing.T) {
	t.Parallel()
	for _, kind := range []v20250801.RecipeKind{v20250801.RecipeKindTerraform, v20250801.RecipeKindOpentofu} {
		for _, source := range []struct {
			location string
			path     string
			version  string
		}{
			{"example/widgets/aws:1.0.0", "example/widgets/aws", "1.0.0"},
			{"registry.opentofu.org/example/widgets/aws:2.0.0", "registry.opentofu.org/example/widgets/aws", "2.0.0"},
			{"git::https://example.com/modules.git//widgets?ref=v1", "git::https://example.com/modules.git//widgets?ref=v1", ""},
		} {
			t.Run(string(kind)+"/"+source.location, func(t *testing.T) {
				t.Parallel()
				const packID = "/planes/radius/local/resourceGroups/rg/providers/Radius.Core/recipePacks/modules"
				const resourceType = "Test.Resources/widgets"
				armOptions := &armpolicy.ClientOptions{Transport: fake.NewServerFactoryTransport(&fake.ServerFactory{
					RecipePacksServer: fake.RecipePacksServer{
						Get: func(_ context.Context, _ string, name string, _ *v20250801.RecipePacksClientGetOptions) (resp azfake.Responder[v20250801.RecipePacksClientGetResponse], errResp azfake.ErrorResponder) {
							require.Equal(t, "modules", name)
							resp.SetResponse(http.StatusOK, v20250801.RecipePacksClientGetResponse{
								Properties: &v20250801.RecipePackProperties{Recipes: map[string]*v20250801.RecipeDefinition{
									resourceType: {
										Kind: &kind, Source: &source.location,
										Parameters: map[string]any{"size": "small", "region": "west"},
										Outputs: map[string]any{
											"host":    "endpoint",
											"secrets": map[string]any{"password": "token"},
										},
									},
								}},
							}, nil)
							return
						},
					},
				})}
				env := minimalEnv("", "")
				env.Properties.RecipePacks = []*string{new(packID)}
				env.Properties.RecipeParameters = map[string]*v20250801.RecipeParameterValue{
					resourceType: {AdditionalProperties: map[string]any{"size": "large"}},
				}
				definition, err := getRecipeDefinitionFromEnvironmentV20250801(t.Context(), env, &recipes.ResourceMetadata{
					ResourceID: "/planes/radius/local/resourceGroups/rg/providers/Test.Resources/widgets/widget",
				}, armOptions)
				require.NoError(t, err)
				require.Equal(t, string(kind), definition.Driver)
				require.Equal(t, source.path, definition.TemplatePath)
				require.Equal(t, source.version, definition.TemplateVersion)
				require.Equal(t, map[string]any{"size": "large", "region": "west"}, definition.Parameters)
				require.Equal(t, map[string]string{"host": "endpoint"}, definition.Outputs)
				require.Equal(t, map[string]string{"password": "token"}, definition.SecretOutputs)
			})
		}
	}
}
