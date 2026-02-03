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
	// environmentResourceType is the resource type for Radius environments
	environmentResourceType = "radius.core/environments"

	// legacyEnvironmentResourceType is the legacy resource type for Radius environments
	legacyEnvironmentResourceType = "applications.core/environments"

	// legacyApplicationsResourcePrefix is the legacy Applications resource type prefix.
	legacyApplicationsResourcePrefix = "applications."

	// LegacyApplicationsAPIVersion is the deprecated Applications API version.
	LegacyApplicationsAPIVersion = "2023-10-01-preview"

	legacyApplicationsAPIVersionSuffix = "@" + LegacyApplicationsAPIVersion
)

// TemplateResourceInfo contains metadata about the resource types in a template.
type TemplateResourceInfo struct {
	HasEnvironmentResource          bool
	HasLegacyApplicationsAPIVersion bool
}

// AnalyzeTemplateResources inspects the compiled Radius Bicep template's resources and returns resource metadata.
func AnalyzeTemplateResources(template map[string]any) TemplateResourceInfo {
	info := TemplateResourceInfo{}
	resources := extractResourcesMap(template)
	if resources == nil {
		return info
	}

	for _, resourceValue := range resources {
		resource, ok := resourceValue.(map[string]any)
		if !ok {
			continue
		}

		resourceType, ok := resource["type"].(string)
		if !ok {
			continue
		}

		resourceTypeLower := strings.ToLower(resourceType)
		if strings.HasPrefix(resourceTypeLower, environmentResourceType) ||
			strings.HasPrefix(resourceTypeLower, legacyEnvironmentResourceType) {
			info.HasEnvironmentResource = true
		}

		if strings.HasPrefix(resourceTypeLower, legacyApplicationsResourcePrefix) &&
			strings.HasSuffix(resourceTypeLower, legacyApplicationsAPIVersionSuffix) {
			info.HasLegacyApplicationsAPIVersion = true
		}

		if info.HasEnvironmentResource && info.HasLegacyApplicationsAPIVersion {
			return info
		}
	}

	return info
}

// ContainsEnvironmentResource inspects the compiled Radius Bicep template's resources to determine if an
// environment resource will be created as part of the deployment.
//
// The expected structure of resource in the template is:
// {"resources": {"resourceName": {"type": "Applications.Core/environments@2023-10-01-preview", ...}}}
func ContainsEnvironmentResource(template map[string]any) bool {
	return AnalyzeTemplateResources(template).HasEnvironmentResource
}

// ContainsLegacyApplicationsAPIVersion checks if the template contains legacy Applications.* resources
// using the deprecated 2023-10-01-preview API version.
func ContainsLegacyApplicationsAPIVersion(template map[string]any) bool {
	return AnalyzeTemplateResources(template).HasLegacyApplicationsAPIVersion
}

func extractResourcesMap(template map[string]any) map[string]any {
	if template == nil {
		return nil
	}

	resourcesValue, ok := template["resources"]
	if !ok {
		return nil
	}

	resources, ok := resourcesValue.(map[string]any)
	if !ok {
		return nil
	}

	return resources
}
