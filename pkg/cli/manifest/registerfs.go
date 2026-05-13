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

package manifest

import (
	"bytes"
	"context"
	"fmt"
	"io/fs"

	yaml "github.com/goccy/go-yaml"
)

// DefaultsConfig represents the structure of the defaults.yaml configuration file
// that lists which resource type manifests should be registered by default.
type DefaultsConfig struct {
	// DefaultRegistration is a list of manifest file paths to register.
	DefaultRegistration []string `yaml:"defaultRegistration"`
}

// RegisterFS reads defaults.yaml from the provided fs.FS, parses and validates each
// listed manifest, merges manifests sharing a namespace into a single ResourceProvider,
// and returns the merged providers.
func RegisterFS(ctx context.Context, fsys fs.FS) ([]ResourceProvider, error) {
	data, err := fs.ReadFile(fsys, "defaults.yaml")
	if err != nil {
		return nil, fmt.Errorf("failed to read defaults.yaml: %w", err)
	}

	config := DefaultsConfig{}
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	if err := decoder.Decode(&config); err != nil {
		return nil, fmt.Errorf("failed to parse defaults.yaml: %w", err)
	}

	if len(config.DefaultRegistration) == 0 {
		return nil, nil
	}

	// Parse and validate each manifest, merging by namespace.
	merged := map[string]*ResourceProvider{}
	for _, path := range config.DefaultRegistration {
		manifestData, err := fs.ReadFile(fsys, path)
		if err != nil {
			return nil, fmt.Errorf("failed to read manifest %s listed in defaults.yaml: %w", path, err)
		}

		provider, err := ReadBytes(manifestData)
		if err != nil {
			return nil, fmt.Errorf("failed to parse manifest %s: %w", path, err)
		}

		if err := validateManifestSchemas(ctx, provider); err != nil {
			return nil, fmt.Errorf("failed to validate manifest %s: %w", path, err)
		}

		if existing, ok := merged[provider.Namespace]; ok {
			for typeName, resourceType := range provider.Types {
				existing.Types[typeName] = resourceType
			}
		} else {
			merged[provider.Namespace] = provider
		}
	}

	result := make([]ResourceProvider, 0, len(merged))
	for _, provider := range merged {
		result = append(result, *provider)
	}

	return result, nil
}
