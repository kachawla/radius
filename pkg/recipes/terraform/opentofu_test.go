/*
Copyright 2026 The Radius Authors.

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

package terraform

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/radius-project/radius/internal/tooling"
	"github.com/radius-project/radius/pkg/components/kubernetesclient/kubernetesclientprovider"
	"github.com/radius-project/radius/pkg/recipes"
	"github.com/radius-project/radius/pkg/ucp/credentials"
	"github.com/stretchr/testify/require"
)

func TestOpenTofuDownloadURL(t *testing.T) {
	t.Parallel()
	for _, platform := range []string{"linux_amd64", "linux_arm64", "darwin_amd64", "darwin_arm64", "windows_amd64", "windows_arm64"} {
		t.Run(platform, func(t *testing.T) {
			t.Parallel()
			goos, goarch, _ := strings.Cut(platform, "_")
			source, err := openTofuDownloadURL(goos, goarch)
			require.NoError(t, err)
			u, err := url.Parse(source)
			require.NoError(t, err)
			require.Equal(t, "https", u.Scheme)
			require.Equal(t, "github.com", u.Host)
			require.Equal(t, fmt.Sprintf("/opentofu/opentofu/releases/download/v%s/tofu_%s_%s.zip", openTofuVersion, openTofuVersion, platform), u.Path)
			require.Equal(t, fmt.Sprintf("file:https://github.com/opentofu/opentofu/releases/download/v%s/tofu_%s_SHA256SUMS", openTofuVersion, openTofuVersion), u.Query().Get("checksum"))
		})
	}
	_, err := openTofuDownloadURL("unsupported", "amd64")
	require.ErrorContains(t, err, "configure opentofu.execPath")
}

func TestOpenTofuVersionMatchesManifest(t *testing.T) {
	t.Parallel()
	manifest, err := tooling.LoadManifest("../../../build/tools.yaml")
	require.NoError(t, err)
	for _, tool := range manifest.Tools {
		if tool.Name == "opentofu" {
			require.Equal(t, tool.Version, openTofuVersion)
			require.Equal(t, "opentofu/opentofu", tool.Source.Repository)
			return
		}
	}
	t.Fatal("OpenTofu is missing from the tool manifest")
}

func TestOpenTofuRejectsNativeOverrides(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"main.tofu", "main.tofu.json"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			directory := t.TempDir()
			require.NoError(t, os.WriteFile(filepath.Join(directory, "main.tf"), nil, 0600))
			require.NoError(t, os.WriteFile(filepath.Join(directory, name), nil, 0600))
			_, err := hasTerraformConfigFiles(directory, recipes.TemplateKindOpenTofu)
			require.ErrorContains(t, err, "currently support .tf and .tf.json files only")
			found, err := hasTerraformConfigFiles(directory, recipes.TemplateKindTerraform)
			require.NoError(t, err)
			require.True(t, found, "Terraform continues to ignore OpenTofu-only files")
		})
	}
}

func testZIP(t *testing.T, files map[string]string) []byte {
	t.Helper()
	var buf bytes.Buffer
	archive := zip.NewWriter(&buf)
	for name, contents := range files {
		file, err := archive.Create(name)
		require.NoError(t, err)
		_, err = file.Write([]byte(contents))
		require.NoError(t, err)
	}
	require.NoError(t, archive.Close())
	return buf.Bytes()
}

func TestDownloadOpenTofuVerifiesChecksum(t *testing.T) {
	t.Parallel()
	archive := testZIP(t, map[string]string{openTofuBinaryName(): "test executable"})
	for _, valid := range []bool{true, false} {
		t.Run(fmt.Sprintf("valid=%v", valid), func(t *testing.T) {
			t.Parallel()
			checksum := fmt.Sprintf("%x", sha256.Sum256(archive))
			if !valid {
				checksum = strings.Repeat("0", 64)
			}
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/tofu.zip":
					_, _ = w.Write(archive)
				case "/SHA256SUMS":
					_, _ = fmt.Fprintf(w, "%s  tofu.zip\n", checksum)
				default:
					http.NotFound(w, r)
				}
			}))
			defer server.Close()
			source := server.URL + "/tofu.zip?" + url.Values{"checksum": {"file:" + server.URL + "/SHA256SUMS"}}.Encode()
			destination := t.TempDir()
			err := downloadOpenTofu(t.Context(), source, destination)
			if !valid {
				require.Error(t, err)
				require.Contains(t, strings.ToLower(err.Error()), "checksums did not match")
				require.NoFileExists(t, filepath.Join(destination, openTofuBinaryName()))
				return
			}
			require.NoError(t, err)
			content, err := os.ReadFile(filepath.Join(destination, openTofuBinaryName()))
			require.NoError(t, err)
			require.Equal(t, "test executable", string(content))
		})
	}
}

func TestOpenTofuExecutorUsesOpenTofuForEveryOperation(t *testing.T) {
	snapshotsPath := installBackendTestTerraform(t)
	binaryPath := filepath.Join(t.TempDir(), openTofuBinaryName())
	require.NoError(t, os.Symlink(os.Args[0], binaryPath))
	e := NewOpenTofuExecutor(nil, nil, kubernetesclientprovider.KubernetesClientProvider{}, t.TempDir(), binaryPath)
	e.awsCredentials = &backendCredentialStub[credentials.AWSCredential]{value: backendTestAWSCredential(false)}
	e.deleteStateObject = noStateCleanup
	options := cloudExecutionOptions(t, "s3")
	options.EnvRecipe.Driver = recipes.TemplateKindOpenTofu

	for range 2 {
		state, err := e.Deploy(t.Context(), options)
		require.NoError(t, err)
		require.Equal(t, "declared", state.Values.Outputs["endpoint"].Value)
	}
	require.NoError(t, e.Delete(t.Context(), options))
	metadata, err := e.GetRecipeMetadata(t.Context(), options)
	require.NoError(t, err)
	require.Contains(t, metadata, "parameters")

	var commands []string
	for _, snapshot := range readBackendSnapshots(t, snapshotsPath) {
		require.Equal(t, binaryPath, snapshot.Executable)
		if snapshot.Command == "version" {
			continue
		}
		commands = append(commands, snapshot.Command)
		require.Contains(t, snapshot.CLIConfig, "registry-token", snapshot.Command)
		require.Equal(t, "env-secret", snapshot.Environment["TEST_SECRET_ENV"], snapshot.Command)
	}
	require.Equal(t, []string{"get", "init", "apply", "show", "get", "init", "apply", "show", "get", "init", "destroy", "get"}, commands)
}

func TestOpenTofuExecutorDoesNotFallBack(t *testing.T) {
	snapshotsPath := installBackendTestTerraform(t)
	for _, binaryPath := range []string{
		filepath.Join(t.TempDir(), "missing-tofu"),
		filepath.Join(os.Getenv("TERRAFORM_TEST_GLOBAL_DIR"), "terraform"),
		"tofu",
	} {
		t.Run(binaryPath, func(t *testing.T) {
			// The fake Terraform binary is available, but cannot satisfy an OpenTofu request.
			e := NewOpenTofuExecutor(nil, nil, kubernetesclientprovider.KubernetesClientProvider{}, t.TempDir(), binaryPath)
			operations := map[string]func(context.Context, Options) error{
				"deploy": func(ctx context.Context, opts Options) error { _, err := e.Deploy(ctx, opts); return err },
				"delete": e.Delete,
				"metadata": func(ctx context.Context, opts Options) error {
					_, err := e.GetRecipeMetadata(ctx, opts)
					return err
				},
			}
			for name, operation := range operations {
				t.Run(name, func(t *testing.T) {
					err := operation(t.Context(), Options{RootDir: t.TempDir()})
					require.Error(t, err)
					require.Contains(t, strings.ToLower(err.Error()), "opentofu")
				})
			}
		})
	}
	require.NoFileExists(t, snapshotsPath, "no Terraform lifecycle commands may run")
}
