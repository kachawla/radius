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

package register

import (
	"context"

	"github.com/radius-project/radius/pkg/cli"
	"github.com/radius-project/radius/pkg/cli/bicep"
	"github.com/radius-project/radius/pkg/cli/clierrors"
	"github.com/radius-project/radius/pkg/cli/cmd/commonflags"
	"github.com/radius-project/radius/pkg/cli/connections"
	"github.com/radius-project/radius/pkg/cli/filesystem"
	"github.com/radius-project/radius/pkg/cli/framework"
	"github.com/radius-project/radius/pkg/cli/output"
	"github.com/radius-project/radius/pkg/cli/workspaces"
	corerp "github.com/radius-project/radius/pkg/corerp/api/v20231001preview"
	"github.com/radius-project/radius/pkg/recipes"
	"github.com/spf13/cobra"
)

// NewCommand creates an instance of the command and runner for the `rad recipe register` command.
//

// NewCommand creates a new Cobra command and a Runner object to register a recipe to an environment, with parameters
// specified using a JSON file or key-value-pairs.
func NewCommand(factory framework.Factory) (*cobra.Command, framework.Runner) {
	runner := NewRunner(factory)

	cmd := &cobra.Command{
		Use:   "register [recipe-name]",
		Short: "Add a recipe to an environment.",
		Long: `Add a recipe to an environment.
You can specify parameters using the '--parameter' flag ('-p' for short). Parameters can be passed as:
		
- A file containing a single value in JSON format
- A key-value-pair passed in the command line
		`,
		Example: `
# Add a recipe to an environment
rad recipe register cosmosdb -e env_name -w workspace --template-kind bicep --template-path template_path --resource-type Applications.Datastores/mongoDatabases
		
# Specify a parameter
rad recipe register cosmosdb -e env_name -w workspace --template-kind bicep --template-path template_path --resource-type Applications.Datastores/mongoDatabases --parameters throughput=400
		
# specify multiple parameters using a JSON parameter file
rad recipe register cosmosdb -e env_name -w workspace --template-kind bicep --template-path template_path --resource-type Applications.Datastores/mongoDatabases --parameters @myfile.json
		`,
		Args: cobra.ExactArgs(1),
		RunE: framework.RunCommand(runner),
	}

	commonflags.AddOutputFlag(cmd)
	commonflags.AddWorkspaceFlag(cmd)
	commonflags.AddResourceGroupFlag(cmd)
	commonflags.AddEnvironmentNameFlag(cmd)
	cmd.Flags().String("template-kind", "", "specify the kind for the template provided by the recipe.")
	_ = cmd.MarkFlagRequired("template-kind")
	cmd.Flags().String("template-version", "", "specify the version for the terraform module.")
	cmd.Flags().String("template-path", "", "specify the path to the template provided by the recipe.")
	_ = cmd.MarkFlagRequired("template-path")
	cmd.Flags().String("resource-type", "", "specify the type of the portable resource this recipe can be consumed by")
	_ = cmd.MarkFlagRequired("resource-type")
	cmd.Flags().Bool("plain-http", false, "Connect to the Bicep registry using HTTP (not-HTTPS). This should be used when the registry is known not to support HTTPS, for example in a locally-hosted registry. Defaults to false (use HTTPS/TLS).")
	commonflags.AddParameterFlag(cmd)

	return cmd, runner
}

// Runner is the runner implementation for the `rad recipe register` command.
type Runner struct {
	ConfigHolder      *framework.ConfigHolder
	ConnectionFactory connections.Factory
	Output            output.Interface
	Workspace         *workspaces.Workspace
	TemplateKind      string
	TemplatePath      string
	PlainHTTP         bool
	TemplateVersion   string
	ResourceType      string
	RecipeName        string
	Parameters        map[string]map[string]any
}

// NewRunner creates a new instance of the `rad recipe register` runner.
func NewRunner(factory framework.Factory) *Runner {
	return &Runner{
		ConfigHolder:      factory.GetConfigHolder(),
		ConnectionFactory: factory.GetConnectionFactory(),
		Output:            factory.GetOutput(),
	}
}

// Validate runs validation for the `rad recipe register` command.
//

// Validate validates the command line args, sets the workspace, environment, template kind, template path, resource type,
// recipe name, and parameters, and returns an error if any of these fail.
func (r *Runner) Validate(cmd *cobra.Command, args []string) error {
	// Validate command line args
	workspace, err := cli.RequireWorkspace(cmd, r.ConfigHolder.Config, r.ConfigHolder.DirectoryConfig)
	if err != nil {
		return err
	}
	r.Workspace = workspace

	environment, err := cli.RequireEnvironmentName(cmd, args, *workspace)
	if err != nil {
		return err
	}
	r.Workspace.Environment = environment

	templateKind, templatePath, templateVersion, err := requireRecipeProperties(cmd)
	if err != nil {
		return err
	}
	r.TemplateKind = templateKind
	r.TemplatePath = templatePath
	r.TemplateVersion = templateVersion

	resourceType, err := cli.GetResourceType(cmd)
	if err != nil {
		return err
	}
	r.ResourceType = resourceType

	recipeName, err := cli.RequireRecipeNameArgs(cmd, args)
	if err != nil {
		return err
	}
	r.RecipeName = recipeName

	parameterArgs, err := cmd.Flags().GetStringArray("parameters")
	if err != nil {
		return err
	}

	parser := bicep.ParameterParser{FileSystem: filesystem.NewOSFS()}
	r.Parameters, err = parser.Parse(parameterArgs...)
	if err != nil {
		return err
	}

	plainHTTP, err := cmd.Flags().GetBool("plain-http")
	if err != nil {
		return err
	}
	r.PlainHTTP = plainHTTP

	return nil
}

// Run runs the `rad recipe register` command.
//

// Run function creates an ApplicationsManagementClient, gets the environment details, adds the recipe properties to the
// environment recipes, and creates the environment with the updated recipes. It returns an error if any of the steps fail.
func (r *Runner) Run(ctx context.Context) error {
	client, err := r.ConnectionFactory.CreateApplicationsManagementClient(ctx, *r.Workspace)
	if err != nil {
		return err
	}

	envResource, err := client.GetEnvironment(ctx, r.Workspace.Environment)
	if err != nil {
		return err
	}

	envRecipes := envResource.Properties.Recipes
	if envRecipes == nil {
		envRecipes = map[string]map[string]corerp.RecipePropertiesClassification{}
	}
	var properties corerp.RecipePropertiesClassification
	switch r.TemplateKind {
	case recipes.TemplateKindTerraform:
		properties = &corerp.TerraformRecipeProperties{
			TemplateKind:    &r.TemplateKind,
			TemplatePath:    &r.TemplatePath,
			TemplateVersion: &r.TemplateVersion,
			Parameters:      bicep.ConvertToMapStringInterface(r.Parameters),
		}
	case recipes.TemplateKindBicep:
		properties = &corerp.BicepRecipeProperties{
			TemplateKind: &r.TemplateKind,
			TemplatePath: &r.TemplatePath,
			PlainHTTP:    &r.PlainHTTP,
			Parameters:   bicep.ConvertToMapStringInterface(r.Parameters),
		}
	}
	if val, ok := envRecipes[r.ResourceType]; ok {
		val[r.RecipeName] = properties
	} else {
		envRecipes[r.ResourceType] = map[string]corerp.RecipePropertiesClassification{
			r.RecipeName: properties,
		}
	}
	envResource.Properties.Recipes = envRecipes

	err = client.CreateOrUpdateEnvironment(ctx, r.Workspace.Environment, &envResource)
	if err != nil {
		return clierrors.MessageWithCause(err, "Failed to register the recipe %q to the environment %q.", r.RecipeName, *envResource.ID)
	}

	r.Output.LogInfo("Successfully linked recipe %q to environment %q ", r.RecipeName, r.Workspace.Environment)
	return nil
}

func requireRecipeProperties(cmd *cobra.Command) (templateKind, templatePath, templateVersion string, err error) {
	templateKind, err = cmd.Flags().GetString("template-kind")
	if err != nil {
		return "", "", "", err
	}

	templatePath, err = cmd.Flags().GetString("template-path")
	if err != nil {
		return "", "", "", err
	}
	templateVersion, err = cmd.Flags().GetString("template-version")
	if err != nil {
		return "", "", "", err
	}
	return templateKind, templatePath, templateVersion, nil
}
