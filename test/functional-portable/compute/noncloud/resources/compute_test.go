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

	"github.com/radius-project/radius/test/rp"
	"github.com/radius-project/radius/test/step"
	"github.com/radius-project/radius/test/testutil"
	"github.com/radius-project/radius/test/validation"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Test_RadiusComputeContainers_Terraform tests creation of a Radius.Compute/containers resource using Terraform recipe.
// The test verifies:
//   - Deployment of the Radius.Compute/containers resource type with Terraform recipe
//   - Creation of required Kubernetes resources (Deployment and Service)
//   - Proper configuration of the container with specified image and ports
func Test_RadiusComputeContainers_Terraform(t *testing.T) {
	template := "testdata/compute-containers-terraform.bicep"
	appName := "compute-containers-app"
	envName := "compute-containers-env"
	containerName := "compute-container"
	resourceTypeName := "Radius.Compute/containers"

	test := rp.NewRPTest(t, appName, []rp.TestStep{
		{
			Executor: step.NewDeployExecutor(template, testutil.GetTerraformRecipeModuleServerURL()),
			RPResources: &validation.RPResourceSet{
				Resources: []validation.RPResource{
					{
						Name: envName,
						Type: validation.EnvironmentsResource,
					},
					{
						Name: appName,
						Type: validation.ApplicationsResource,
					},
					{
						Name: containerName,
						Type: resourceTypeName,
					},
				},
			},
			K8sObjects: &validation.K8sObjectSet{
				Namespaces: map[string][]validation.K8sObject{
					envName: {
						validation.NewK8sDeploymentForResource(containerName, containerName),
						validation.NewK8sServiceForResource(containerName, containerName+"-demo"),
					},
				},
			},
			PostStepVerify: func(ctx context.Context, t *testing.T, test rp.RPTest) {
				// Verify the deployment was created with correct configuration
				deploy, err := test.Options.K8sClient.AppsV1().Deployments(envName).Get(ctx, containerName, metav1.GetOptions{})
				require.NoError(t, err)
				require.NotNil(t, deploy)

				// Verify container configuration
				require.Len(t, deploy.Spec.Template.Spec.Containers, 1)
				container := deploy.Spec.Template.Spec.Containers[0]
				require.Equal(t, "demo", container.Name)
				require.Equal(t, "ghcr.io/radius-project/samples/demo:latest", container.Image)

				// Verify port configuration
				require.Len(t, container.Ports, 1)
				require.Equal(t, "web", container.Ports[0].Name)
				require.Equal(t, int32(3000), container.Ports[0].ContainerPort)

				// Verify service was created
				svc, err := test.Options.K8sClient.CoreV1().Services(envName).Get(ctx, containerName+"-demo", metav1.GetOptions{})
				require.NoError(t, err)
				require.NotNil(t, svc)

				// Verify service port configuration
				require.Len(t, svc.Spec.Ports, 1)
				require.Equal(t, "web", svc.Spec.Ports[0].Name)
				require.Equal(t, int32(3000), svc.Spec.Ports[0].Port)
			},
		},
	})

	test.Test(t)
}
