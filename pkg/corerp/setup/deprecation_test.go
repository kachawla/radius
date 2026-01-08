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

package setup

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_DeprecationMiddleware_ApplicationsCore(t *testing.T) {
	// Create a test handler
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Wrap with deprecation middleware
	middleware := DeprecationMiddleware()
	handler := middleware(testHandler)

	// Test with Applications.Core path
	req := httptest.NewRequest("GET", "/planes/radius/local/resourceGroups/test-rg/providers/Applications.Core/containers/test-container", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	// Check response
	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, "OK", w.Body.String())

	// Check deprecation warning header
	warningHeader := w.Header().Get(DeprecationWarningHeader)
	require.NotEmpty(t, warningHeader, "Warning header should be present for Applications.Core requests")
	require.Contains(t, warningHeader, "Applications.Core namespace is deprecated")
	require.Contains(t, warningHeader, "Radius.Core")
	require.Contains(t, warningHeader, "2025-08-01-preview")
}

func Test_DeprecationMiddleware_RadiusCore(t *testing.T) {
	// Create a test handler
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Wrap with deprecation middleware
	middleware := DeprecationMiddleware()
	handler := middleware(testHandler)

	// Test with Radius.Core path (should not have deprecation warning)
	req := httptest.NewRequest("GET", "/planes/radius/local/resourceGroups/test-rg/providers/Radius.Core/environments/test-env", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	// Check response
	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, "OK", w.Body.String())

	// Check that deprecation warning header is NOT present
	warningHeader := w.Header().Get(DeprecationWarningHeader)
	require.Empty(t, warningHeader, "Warning header should not be present for Radius.Core requests")
}

func Test_DeprecationMiddleware_OtherNamespace(t *testing.T) {
	// Create a test handler
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Wrap with deprecation middleware
	middleware := DeprecationMiddleware()
	handler := middleware(testHandler)

	// Test with other namespace path
	req := httptest.NewRequest("GET", "/planes/radius/local/resourceGroups/test-rg/providers/Applications.Dapr/stateStores/test-store", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	// Check response
	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, "OK", w.Body.String())

	// Check that deprecation warning header is NOT present
	warningHeader := w.Header().Get(DeprecationWarningHeader)
	require.Empty(t, warningHeader, "Warning header should not be present for other namespace requests")
}

func Test_DeprecationMiddleware_MultipleResourceTypes(t *testing.T) {
	// Test different Applications.Core resource types
	resourcePaths := []string{
		"/planes/radius/local/resourceGroups/test-rg/providers/Applications.Core/containers/test-container",
		"/planes/radius/local/resourceGroups/test-rg/providers/Applications.Core/applications/test-app",
		"/planes/radius/local/resourceGroups/test-rg/providers/Applications.Core/environments/test-env",
		"/planes/radius/local/resourceGroups/test-rg/providers/Applications.Core/gateways/test-gateway",
		"/planes/radius/local/resourceGroups/test-rg/providers/Applications.Core/secretStores/test-secret",
		"/planes/radius/local/resourceGroups/test-rg/providers/Applications.Core/extenders/test-extender",
		"/planes/radius/local/resourceGroups/test-rg/providers/Applications.Core/volumes/test-volume",
	}

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := DeprecationMiddleware()
	handler := middleware(testHandler)

	for _, path := range resourcePaths {
		t.Run(path, func(t *testing.T) {
			req := httptest.NewRequest("GET", path, nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			require.Equal(t, http.StatusOK, w.Code)

			warningHeader := w.Header().Get(DeprecationWarningHeader)
			require.NotEmpty(t, warningHeader, "Warning header should be present for path: %s", path)
			require.Contains(t, warningHeader, "Applications.Core namespace is deprecated")
		})
	}
}
