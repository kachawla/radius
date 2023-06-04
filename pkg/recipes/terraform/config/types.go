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

package config

type TerraformConfig struct {
	Terraform TerraformDefinition    `json:"terraform"`
	Provider  map[string]interface{} `json:"provider"`
	Module    map[string]interface{} `json:"module"`
}

type TerraformDefinition struct {
	RequiredProviders map[string]ProviderDefinition `json:"required_providers"`
	Backend           map[string]interface{}        `json:"backend"`
	RequiredVersion   string                        `json:"required_version"`
}

type ProviderDefinition struct {
	Source  string `json:"source"`
	Version string `json:"version"`
}

type TerraformProviderMetadata struct {
	Type       string
	Parameters map[string]interface{}
}