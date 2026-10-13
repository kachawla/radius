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
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	getter "github.com/hashicorp/go-getter/v2"
	"github.com/hashicorp/terraform-exec/tfexec"
	"github.com/radius-project/radius/pkg/ucp/ucplog"
)

// openTofuVersion is kept in sync with build/tools.yaml by the tool updater.
var openTofuVersion = "1.13.1"

var openTofuInstallMutex sync.Mutex

func installOpenTofu(ctx context.Context, opts InstallOptions, installDir, execPath string) (*tfexec.Terraform, error) {
	if execPath == "" {
		var err error
		execPath, err = ensureOpenTofuBinary(ctx, installDir)
		if err != nil {
			return nil, err
		}
	} else if err := verifyOpenTofuBinary(ctx, execPath); err != nil {
		// A configured executable must never trigger a download or fall back to Terraform.
		return nil, err
	}

	tf, err := NewTerraform(ctx, opts.RootDir, execPath)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize OpenTofu: %w", err)
	}
	configureTerraformLogs(ctx, tf, opts.LogLevel)
	return tf, nil
}

func ensureOpenTofuBinary(ctx context.Context, installDir string) (string, error) {
	if installDir == "" {
		return "", fmt.Errorf("an installation directory is required for OpenTofu")
	}
	source, err := openTofuDownloadURL(runtime.GOOS, runtime.GOARCH)
	if err != nil {
		return "", err
	}
	cacheDir, err := filepath.Abs(filepath.Join(installDir, ".opentofu", openTofuVersion))
	if err != nil {
		return "", fmt.Errorf("failed to resolve OpenTofu cache directory: %w", err)
	}
	binaryPath := filepath.Join(cacheDir, openTofuBinaryName())

	openTofuInstallMutex.Lock()
	defer openTofuInstallMutex.Unlock()

	if _, err := os.Stat(binaryPath); err == nil {
		if err := verifyOpenTofuBinary(ctx, binaryPath); err != nil {
			return "", err
		}
		return binaryPath, nil
	} else if !os.IsNotExist(err) {
		return "", fmt.Errorf("failed to inspect OpenTofu binary: %w", err)
	}

	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create OpenTofu cache directory: %w", err)
	}
	stagingDir, err := os.MkdirTemp(cacheDir, "download-")
	if err != nil {
		return "", fmt.Errorf("failed to create OpenTofu download directory: %w", err)
	}
	defer func() {
		if err := os.RemoveAll(stagingDir); err != nil {
			ucplog.FromContextOrDiscard(ctx).Error(err, "Failed to clean up OpenTofu download directory")
		}
	}()

	ucplog.FromContextOrDiscard(ctx).Info("Downloading OpenTofu", "version", openTofuVersion)
	if err := downloadOpenTofu(ctx, source, stagingDir); err != nil {
		return "", err
	}
	stagedBinary := filepath.Join(stagingDir, openTofuBinaryName())
	if err := os.Chmod(stagedBinary, 0755); err != nil {
		return "", fmt.Errorf("failed to make OpenTofu executable: %w", err)
	}
	if err := verifyOpenTofuBinary(ctx, stagedBinary); err != nil {
		return "", err
	}
	if err := os.Rename(stagedBinary, binaryPath); err != nil {
		return "", fmt.Errorf("failed to install OpenTofu binary: %w", err)
	}
	return binaryPath, nil
}

func openTofuBinaryName() string {
	if runtime.GOOS == "windows" {
		return "tofu.exe"
	}
	return "tofu"
}

func openTofuDownloadURL(goos, goarch string) (string, error) {
	switch goos + "_" + goarch {
	case "linux_amd64", "linux_arm64", "darwin_amd64", "darwin_arm64", "windows_amd64", "windows_arm64":
	default:
		return "", fmt.Errorf("OpenTofu downloads are not supported on %s/%s; configure opentofu.execPath", goos, goarch)
	}
	releaseURL := fmt.Sprintf("https://github.com/opentofu/opentofu/releases/download/v%s/", openTofuVersion)
	asset := fmt.Sprintf("tofu_%s_%s_%s.zip", openTofuVersion, goos, goarch)
	query := url.Values{
		"checksum": {"file:" + releaseURL + "tofu_" + openTofuVersion + "_SHA256SUMS"},
	}
	return releaseURL + asset + "?" + query.Encode(), nil
}

func downloadOpenTofu(ctx context.Context, source, destination string) error {
	client := getter.Client{
		DisableSymlinks: true,
		Getters: []getter.Getter{&getter.HttpGetter{
			Client:                &http.Client{Timeout: 5 * time.Minute},
			XTerraformGetDisabled: true,
		}},
	}
	if _, err := client.Get(ctx, &getter.Request{Src: source, Dst: destination, GetMode: getter.ModeDir}); err != nil {
		return fmt.Errorf("failed to download and verify OpenTofu: %w", err)
	}
	return nil
}

func verifyOpenTofuBinary(ctx context.Context, binaryPath string) error {
	if !filepath.IsAbs(binaryPath) {
		return fmt.Errorf("opentofu.execPath must be an absolute path: %q", binaryPath)
	}
	output, err := exec.CommandContext(ctx, binaryPath, "version").Output()
	if err != nil {
		return fmt.Errorf("OpenTofu version check failed for %q: %w", binaryPath, err)
	}
	if !strings.HasPrefix(strings.TrimSpace(string(output)), "OpenTofu v") {
		return fmt.Errorf("executable %q is not OpenTofu", binaryPath)
	}
	return nil
}
