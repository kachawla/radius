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
	"testing"

	"github.com/radius-project/radius/test/rp"
	"github.com/radius-project/radius/test/step"
	"github.com/radius-project/radius/test/validation"
)

func Test_Environment(t *testing.T) {
	template := "testdata/corerp-resources-environment.bicep"
	name := "corerp-resources-environment"

	test := rp.NewRPTest(t, name, []rp.TestStep{
		{
			Executor: step.NewDeployExecutor(template),
			RPResources: &validation.RPResourceSet{
				Resources: []validation.RPResource{
					{
						Name: "corerp-resources-environment-env",
						Type: validation.EnvironmentsResource,
					},
				},
			},
			// Environment should not render any K8s Objects directly
			K8sObjects: &validation.K8sObjectSet{},
		},
	})

	test.Test(t)
}

// Test_Environment_CreateFromTemplate verifies that an environment can be created
// by deploying a Bicep template that defines an environment resource, without
// specifying an existing environment via the --environment flag.
// This validates the fix for: https://github.com/radius-project/radius/issues/9453
func Test_Environment_CreateFromTemplate(t *testing.T) {
	template := "testdata/corerp-resources-environment-create.bicep"
	name := "corerp-resources-environment-create"

	test := rp.NewRPTest(t, name, []rp.TestStep{
		{
			// Deploy without specifying --environment flag
			// The environment will be created from the template
			Executor: step.NewDeployExecutor(template),
			RPResources: &validation.RPResourceSet{
				Resources: []validation.RPResource{
					{
						Name: "corerp-resources-environment-create-env",
						Type: validation.EnvironmentsResource,
					},
				},
			},
			// Environment should not render any K8s Objects directly
			K8sObjects: &validation.K8sObjectSet{},
		},
	})

	test.Test(t)
}
