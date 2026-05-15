package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRunGenerate_SingleFile(t *testing.T) {
	outputDir := t.TempDir()

	err := RunGenerate([]string{"testdata/containers.yaml"}, outputDir)
	if err != nil {
		t.Fatalf("RunGenerate returned error: %v", err)
	}

	// Verify all three output files are created.
	for _, filename := range []string{"types.json", "index.json", "index.md"} {
		path := filepath.Join(outputDir, filename)
		info, err := os.Stat(path)
		if err != nil {
			t.Errorf("expected %s to exist: %v", filename, err)
			continue
		}
		if info.Size() == 0 {
			t.Errorf("expected %s to be non-empty", filename)
		}
	}
}

func TestRunGenerate_MultipleFiles_SameNamespace(t *testing.T) {
	outputDir := t.TempDir()

	err := RunGenerate([]string{"testdata/containers.yaml", "testdata/routes.yaml"}, outputDir)
	if err != nil {
		t.Fatalf("RunGenerate returned error: %v", err)
	}

	// Verify all three output files are created.
	for _, filename := range []string{"types.json", "index.json", "index.md"} {
		path := filepath.Join(outputDir, filename)
		info, err := os.Stat(path)
		if err != nil {
			t.Errorf("expected %s to exist: %v", filename, err)
			continue
		}
		if info.Size() == 0 {
			t.Errorf("expected %s to be non-empty", filename)
		}
	}

	// Verify types.json contains definitions from both manifests by checking
	// that the merged output is larger than a single-manifest output.
	singleDir := t.TempDir()
	err = RunGenerate([]string{"testdata/containers.yaml"}, singleDir)
	if err != nil {
		t.Fatalf("RunGenerate (single) returned error: %v", err)
	}

	mergedTypes, _ := os.ReadFile(filepath.Join(outputDir, "types.json"))
	singleTypes, _ := os.ReadFile(filepath.Join(singleDir, "types.json"))

	if len(mergedTypes) <= len(singleTypes) {
		t.Errorf("merged types.json (%d bytes) should be larger than single-manifest types.json (%d bytes)",
			len(mergedTypes), len(singleTypes))
	}
}

func TestRunGenerate_MultipleFiles_DifferentNamespaces(t *testing.T) {
	outputDir := t.TempDir()

	err := RunGenerate([]string{"testdata/containers.yaml", "testdata/secrets.yaml"}, outputDir)
	if err == nil {
		t.Fatal("expected error when merging manifests with different namespaces, got nil")
	}

	expected := "all manifests must share the same namespace"
	if got := err.Error(); !contains(got, expected) {
		t.Errorf("expected error containing %q, got %q", expected, got)
	}
}

func TestRunGenerate_NonexistentFile(t *testing.T) {
	outputDir := t.TempDir()

	err := RunGenerate([]string{"testdata/nonexistent.yaml"}, outputDir)
	if err == nil {
		t.Fatal("expected error for nonexistent manifest file, got nil")
	}

	expected := "manifest file does not exist"
	if got := err.Error(); !contains(got, expected) {
		t.Errorf("expected error containing %q, got %q", expected, got)
	}
}

func TestRunGenerate_EmptyManifestList(t *testing.T) {
	outputDir := t.TempDir()

	err := RunGenerate([]string{}, outputDir)
	if err == nil {
		t.Fatal("expected error for empty manifest list, got nil")
	}

	expected := "at least one manifest file is required"
	if got := err.Error(); !contains(got, expected) {
		t.Errorf("expected error containing %q, got %q", expected, got)
	}
}

func TestMergeManifestFiles_DuplicateType(t *testing.T) {
	// Both files define "containers" in Radius.Compute — should be rejected.
	_, err := mergeManifestFiles([]string{"testdata/containers.yaml", "testdata/containers.yaml"})
	if err == nil {
		t.Fatal("expected error for duplicate resource type, got nil")
	}

	expected := "duplicate resource type"
	if got := err.Error(); !contains(got, expected) {
		t.Errorf("expected error containing %q, got %q", expected, got)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
