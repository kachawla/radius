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

package list

import (
	"context"
	"testing"

	"github.com/radius-project/radius/pkg/cli/clients"
	"github.com/radius-project/radius/pkg/cli/clierrors"
	"github.com/radius-project/radius/pkg/cli/config"
	"github.com/radius-project/radius/pkg/cli/connections"
	"github.com/radius-project/radius/pkg/cli/framework"
	"github.com/radius-project/radius/pkg/cli/objectformats"
	"github.com/radius-project/radius/pkg/cli/output"
	"github.com/radius-project/radius/pkg/cli/workspaces"
	"github.com/radius-project/radius/pkg/corerp/api/v20231001preview"
	"github.com/radius-project/radius/pkg/to"
	"github.com/radius-project/radius/test/radcli"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func Test_CommandValidation(t *testing.T) {
	radcli.SharedCommandValidation(t, NewCommand)
}

func Test_Validate(t *testing.T) {
	testcases := []radcli.ValidateInput{
		{
			Name:          "Show Command with default application",
			Input:         []string{},
			ExpectedValid: true,
			ConfigHolder: framework.ConfigHolder{
				ConfigFilePath: "",
				Config:         radcli.LoadConfigWithWorkspace(t),
				DirectoryConfig: &config.DirectoryConfig{
					Workspace: config.DirectoryWorkspaceConfig{
						Application: "test-application",
					},
				},
			},
		},
		{
			Name:          "Show Command with flag",
			Input:         []string{"-a", "test-app"},
			ExpectedValid: true,
			ConfigHolder: framework.ConfigHolder{
				ConfigFilePath: "",
				Config:         radcli.LoadConfigWithWorkspace(t),
			},
		},
		{
			Name:          "Show Command with positional arg",
			Input:         []string{"test-app"},
			ExpectedValid: true,
			ConfigHolder: framework.ConfigHolder{
				ConfigFilePath: "",
				Config:         radcli.LoadConfigWithWorkspace(t),
			},
		},
		{
			Name:          "Show Command with fallback workspace",
			Input:         []string{"--application", "test-app", "--group", "test-group"},
			ExpectedValid: true,
			ConfigHolder: framework.ConfigHolder{
				ConfigFilePath: "",
				Config:         radcli.LoadEmptyConfig(t),
			},
		},
		{
			Name:          "Show Command with incorrect args",
			Input:         []string{"foo", "bar"},
			ExpectedValid: false,
			ConfigHolder: framework.ConfigHolder{
				ConfigFilePath: "",
				Config:         radcli.LoadConfigWithWorkspace(t),
			},
		},
	}
	radcli.SharedValidateValidation(t, NewCommand, testcases)
}

func Test_Run(t *testing.T) {
	t.Run("Success: Application Found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		application := v20231001preview.ApplicationResource{
			Name: to.Ptr("test-app"),
		}

		appManagementClient := clients.NewMockApplicationsManagementClient(ctrl)
		appManagementClient.EXPECT().
			GetApplication(gomock.Any(), "test-app").
			Return(application, nil).
			Times(1)

		workspace := &workspaces.Workspace{
			Connection: map[string]any{
				"kind":    "kubernetes",
				"context": "kind-kind",
			},
			Name:  "kind-kind",
			Scope: "/planes/radius/local/resourceGroups/test-group",
		}
		outputSink := &output.MockOutput{}
		runner := &Runner{
			ConnectionFactory: &connections.MockFactory{ApplicationsManagementClient: appManagementClient},
			Workspace:         workspace,
			Format:            "table",
			Output:            outputSink,
			ApplicationName:   "test-app",
		}

		err := runner.Run(context.Background())
		require.NoError(t, err)

		expected := []any{
			output.FormattedOutput{
				Format:  "table",
				Obj:     application,
				Options: objectformats.GetResourceTableFormat(),
			},
		}

		require.Equal(t, expected, outputSink.Writes)
	})

	t.Run("Error: Application Not Found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		appManagementClient := clients.NewMockApplicationsManagementClient(ctrl)
		appManagementClient.EXPECT().
			GetApplication(gomock.Any(), "test-app").
			Return(v20231001preview.ApplicationResource{}, radcli.Create404Error()).
			Times(1)

		workspace := &workspaces.Workspace{
			Connection: map[string]any{
				"kind":    "kubernetes",
				"context": "kind-kind",
			},
			Name:  "kind-kind",
			Scope: "/planes/radius/local/resourceGroups/test-group",
		}
		outputSink := &output.MockOutput{}
		runner := &Runner{
			ConnectionFactory: &connections.MockFactory{ApplicationsManagementClient: appManagementClient},
			Workspace:         workspace,
			Format:            "table",
			Output:            outputSink,
			ApplicationName:   "test-app",
		}

		err := runner.Run(context.Background())
		require.Error(t, err)
		require.Equal(t, clierrors.Message("The application \"test-app\" was not found or has been deleted."), err)

		require.Empty(t, outputSink.Writes)
	})
}
