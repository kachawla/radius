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
	"testing"

	"github.com/radius-project/radius/test/radcli"
	"github.com/radius-project/radius/test/rp"
	"github.com/stretchr/testify/require"
)

// Test_DefaultResourceTypes_Registered verifies that all default resource types
// copied from resource-types-contrib are registered and visible after Radius
// startup. These types are loaded from per-type manifest files in
// deploy/manifest/built-in-providers/ which have no explicit location field
// (UCP routes them via DefaultDownstreamEndpoint to dynamic-rp).
//
// This test catches regressions where:
// - Manifest files fail to load at startup
// - Multiple files sharing a namespace overwrite each other's types
// - A type is removed from defaults.yaml but not noticed
func Test_DefaultResourceTypes_Registered(t *testing.T) {
	// These are the resource types listed in deploy/manifest/defaults.yaml.
	// If defaults.yaml is updated, this list should be updated to match.
	defaultTypes := []string{
		"Radius.Compute/containers",
		"Radius.Compute/persistentVolumes",
		"Radius.Compute/routes",
		"Radius.Data/mySqlDatabases",
		"Radius.Security/secrets",
	}

	options := rp.NewRPTestOptions(t)
	cli := radcli.NewCLI(t, options.ConfigFilePath)
	ctx := context.Background()

	for _, resourceType := range defaultTypes {
		t.Run(resourceType, func(t *testing.T) {
			output, err := cli.RunCommand(ctx, []string{"resource-type", "show", resourceType, "--output", "json"})
			require.NoErrorf(t, err, "resource type %s should be registered after startup", resourceType)
			require.Contains(t, output, resourceType, "resource type %s should appear in show output", resourceType)
		})
	}
}
