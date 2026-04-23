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

package main

import (
	"bytes"
	"crypto/sha256"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/radius-project/radius/pkg/cli/manifest"
	"gopkg.in/yaml.v3"
)

var (
	sourceDir string
	targetDir string
	dryRun    bool
	verbose   bool
)

type syncResult struct {
	addedFiles   []string
	updatedFiles []string
	skippedFiles []string
	errors       []error
}

func main() {
	flag.StringVar(&sourceDir, "source", "", "Source directory containing resource type manifests (required)")
	flag.StringVar(&targetDir, "target", "", "Target directory for synced manifests (required)")
	flag.BoolVar(&dryRun, "dry-run", false, "Print actions without making changes")
	flag.BoolVar(&verbose, "verbose", false, "Enable verbose output")
	flag.Parse()

	if sourceDir == "" || targetDir == "" {
		fmt.Fprintf(os.Stderr, "Error: both --source and --target are required\n")
		flag.Usage()
		os.Exit(1)
	}

	// Ensure source directory exists
	if _, err := os.Stat(sourceDir); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: source directory does not exist: %s\n", sourceDir)
		os.Exit(1)
	}

	// Ensure target directory exists
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to create target directory: %v\n", err)
		os.Exit(1)
	}

	result := syncResourceTypes()

	// Print summary
	fmt.Println("\n=== Sync Summary ===")
	fmt.Printf("Added:   %d file(s)\n", len(result.addedFiles))
	fmt.Printf("Updated: %d file(s)\n", len(result.updatedFiles))
	fmt.Printf("Skipped: %d file(s)\n", len(result.skippedFiles))
	fmt.Printf("Errors:  %d\n", len(result.errors))

	if len(result.addedFiles) > 0 {
		fmt.Println("\nAdded files:")
		for _, f := range result.addedFiles {
			fmt.Printf("  + %s\n", f)
		}
	}

	if len(result.updatedFiles) > 0 {
		fmt.Println("\nUpdated files:")
		for _, f := range result.updatedFiles {
			fmt.Printf("  ~ %s\n", f)
		}
	}

	if verbose && len(result.skippedFiles) > 0 {
		fmt.Println("\nSkipped files (no changes):")
		for _, f := range result.skippedFiles {
			fmt.Printf("  = %s\n", f)
		}
	}

	if len(result.errors) > 0 {
		fmt.Println("\nErrors:")
		for _, err := range result.errors {
			fmt.Printf("  ! %v\n", err)
		}
		os.Exit(1)
	}

	if dryRun {
		fmt.Println("\n[DRY RUN] No changes were made")
	}

	// Exit with status code indicating if changes were detected
	if len(result.addedFiles) > 0 || len(result.updatedFiles) > 0 {
		os.Exit(0)
	}
	os.Exit(2) // No changes detected
}

func syncResourceTypes() syncResult {
	result := syncResult{}

	// Find all YAML files in source directory
	entries, err := os.ReadDir(sourceDir)
	if err != nil {
		result.errors = append(result.errors, fmt.Errorf("failed to read source directory: %w", err))
		return result
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// Only process YAML files
		if !strings.HasSuffix(entry.Name(), ".yaml") && !strings.HasSuffix(entry.Name(), ".yml") {
			if verbose {
				fmt.Printf("Skipping non-YAML file: %s\n", entry.Name())
			}
			continue
		}

		sourcePath := filepath.Join(sourceDir, entry.Name())

		// Parse the manifest
		resourceProvider, err := manifest.ReadFile(sourcePath)
		if err != nil {
			result.errors = append(result.errors, fmt.Errorf("failed to parse %s: %w", entry.Name(), err))
			continue
		}

		// Check if this manifest should be synced
		if !resourceProvider.DefaultRegistration {
			if verbose {
				fmt.Printf("Skipping %s (defaultRegistration: false)\n", entry.Name())
			}
			continue
		}

		fmt.Printf("Processing %s (namespace: %s, defaultRegistration: true)\n", entry.Name(), resourceProvider.Namespace)

		// Read source file content
		sourceContent, err := os.ReadFile(sourcePath)
		if err != nil {
			result.errors = append(result.errors, fmt.Errorf("failed to read %s: %w", entry.Name(), err))
			continue
		}

		// Remove the defaultRegistration field before writing to target
		// This field is only used for sync mechanism and not needed in Radius
		cleanedContent, err := removeDefaultRegistrationField(sourceContent)
		if err != nil {
			result.errors = append(result.errors, fmt.Errorf("failed to clean %s: %w", entry.Name(), err))
			continue
		}

		targetPath := filepath.Join(targetDir, entry.Name())

		// Check if target file exists
		targetContent, err := os.ReadFile(targetPath)
		if os.IsNotExist(err) {
			// File doesn't exist, add it
			if !dryRun {
				if err := os.WriteFile(targetPath, cleanedContent, 0644); err != nil {
					result.errors = append(result.errors, fmt.Errorf("failed to write %s: %w", entry.Name(), err))
					continue
				}
			}
			result.addedFiles = append(result.addedFiles, entry.Name())
		} else if err != nil {
			result.errors = append(result.errors, fmt.Errorf("failed to read target %s: %w", entry.Name(), err))
			continue
		} else {
			// File exists, check if content differs
			if !contentEqual(cleanedContent, targetContent) {
				if !dryRun {
					if err := os.WriteFile(targetPath, cleanedContent, 0644); err != nil {
						result.errors = append(result.errors, fmt.Errorf("failed to update %s: %w", entry.Name(), err))
						continue
					}
				}
				result.updatedFiles = append(result.updatedFiles, entry.Name())
			} else {
				result.skippedFiles = append(result.skippedFiles, entry.Name())
			}
		}
	}

	return result
}

// removeDefaultRegistrationField removes the defaultRegistration field from YAML content
func removeDefaultRegistrationField(content []byte) ([]byte, error) {
	var data map[string]interface{}
	if err := yaml.Unmarshal(content, &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal YAML: %w", err)
	}

	// Remove the defaultRegistration field
	delete(data, "defaultRegistration")

	// Marshal back to YAML
	var buf bytes.Buffer
	encoder := yaml.NewEncoder(&buf)
	encoder.SetIndent(2)
	if err := encoder.Encode(data); err != nil {
		return nil, fmt.Errorf("failed to marshal YAML: %w", err)
	}

	return buf.Bytes(), nil
}

// contentEqual checks if two byte slices represent the same YAML content
// It normalizes the YAML before comparing to avoid false positives from formatting differences
func contentEqual(content1, content2 []byte) bool {
	// Parse both contents
	var data1, data2 map[string]interface{}
	if err := yaml.Unmarshal(content1, &data1); err != nil {
		return false
	}
	if err := yaml.Unmarshal(content2, &data2); err != nil {
		return false
	}

	// Marshal both to a normalized form and compare hashes
	normalized1, err := yaml.Marshal(data1)
	if err != nil {
		return false
	}
	normalized2, err := yaml.Marshal(data2)
	if err != nil {
		return false
	}

	hash1 := sha256.Sum256(normalized1)
	hash2 := sha256.Sum256(normalized2)

	return bytes.Equal(hash1[:], hash2[:])
}
