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
	"strings"
)

const (
	// EnvironmentResourceType is the resource type for Radius environments
	EnvironmentResourceType = "Applications.Core/environments"
)

// ContainsEnvironmentResource checks if the template contains an Applications.Core/environments resource.
// This function inspects the compiled ARM template's resources array to determine if an environment
// resource will be created as part of the deployment.
func ContainsEnvironmentResource(template map[string]any) bool {
	if template == nil {
		return false
	}

	// ARM templates have a "resources" array
	resourcesInterface, ok := template["resources"]
	if !ok {
		return false
	}

	resources, ok := resourcesInterface.([]interface{})
	if !ok {
		return false
	}

	// Check each resource in the array
	for _, resourceInterface := range resources {
		resource, ok := resourceInterface.(map[string]interface{})
		if !ok {
			continue
		}

		// Check the "type" field of the resource
		resourceType, ok := resource["type"].(string)
		if !ok {
			continue
		}

		// Check if this is an environment resource
		// We check case-insensitively to handle both:
		// - Applications.Core/environments
		// - applications.core/environments
		if strings.EqualFold(resourceType, EnvironmentResourceType) {
			return true
		}
	}

	return false
}
