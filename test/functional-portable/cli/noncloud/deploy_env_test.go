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

package resource_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/radius-project/radius/test/radcli"
	"github.com/radius-project/radius/test/rp"
	"github.com/radius-project/radius/test/testcontext"
	"github.com/stretchr/testify/require"
)

// Test_DeployEnvironmentTemplate verifies that an environment can be created
// by deploying a Bicep template that defines an environment resource, without
// specifying an existing environment via the --environment flag.
//
// This validates the fix for: https://github.com/radius-project/radius/issues/9453
//
// IMPORTANT NOTE: This is an end-to-end test that runs the actual `rad` CLI binary.
// If the test workspace has a default environment configured (workspace.Environment),
// the CLI will use it as a fallback even without the --environment flag. This means
// the test validates the overall deployment flow but may not exercise the specific
// templateCreatesEnvironment logic path if a workspace default exists.
//
// To verify the core fix logic (templateCreatesEnvironment check), see the unit test:
// Test_Validate/rad_deploy_-_template_creates_environment in pkg/cli/cmd/deploy/deploy_test.go
//
// This functional test still provides value by:
// 1. Verifying end-to-end deployment of environment templates works
// 2. Ensuring no regressions in the overall deployment flow
// 3. Testing the actual CLI binary behavior (not just mocked code paths)
func Test_DeployEnvironmentTemplate(t *testing.T) {
	ctx, cancel := testcontext.NewWithCancel(t)
	t.Cleanup(cancel)

	options := rp.NewRPTestOptions(t)
	cli := radcli.NewCLI(t, options.ConfigFilePath)

	// Generate a unique resource group name to avoid conflicts with parallel tests
	uniqueGroupName := fmt.Sprintf("test-deploy-env-%d", time.Now().Unix())
	envName := "deploy-env-test"

	// Ensure cleanup even if test fails
	t.Cleanup(func() {
		// Try to delete the test group if it still exists
		// Ignore errors as the group might have been successfully deleted
		_ = cli.GroupDelete(context.Background(), uniqueGroupName, radcli.DeleteOptions{Confirm: true})
	})

	// Create the unique resource group
	t.Logf("Creating resource group: %s", uniqueGroupName)
	err := cli.GroupCreate(ctx, uniqueGroupName)
	require.NoError(t, err, "Failed to create resource group")

	// Get the template file path
	cwd, err := os.Getwd()
	require.NoError(t, err)
	templateFilePath := filepath.Join(cwd, "testdata/corerp-deploy-env-test.bicep")

	// Deploy the environment template WITHOUT specifying --environment flag
	// This is the key test - before the fix, this would fail with:
	// "no environment name or ID provided, pass in an environment name or ID"
	t.Logf("Deploying environment template to group: %s without --environment flag", uniqueGroupName)
	err = cli.DeployWithGroup(ctx, templateFilePath, "", "", uniqueGroupName)
	require.NoError(t, err, "Failed to deploy environment template - the fix may not be working")

	// Set options for group-scoped operations
	showOpts := radcli.ShowOptions{Group: uniqueGroupName}

	// Verify environment was created successfully
	// Use ResourceShow instead of EnvShow since we're querying by group
	t.Logf("Verifying environment was created: %s", envName)
	output, err := cli.ResourceShow(ctx, "Applications.Core/environments", envName, showOpts)
	require.NoError(t, err, "Failed to show environment - it may not have been created")
	require.Contains(t, output, envName, "Environment should exist")

	t.Logf("Successfully verified environment %s was created from template without --environment flag", envName)

	// Clean up
	t.Logf("Cleaning up: deleting group %s", uniqueGroupName)
	deleteOpts := radcli.DeleteOptions{Group: uniqueGroupName, Confirm: true}
	err = cli.GroupDelete(ctx, uniqueGroupName, deleteOpts)
	require.NoError(t, err, "Failed to delete resource group")
}
