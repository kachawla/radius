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
	"strings"

	"github.com/radius-project/radius/pkg/ucp/ucplog"
)

const (
	// DeprecationWarningHeader is the header name for deprecation warnings
	DeprecationWarningHeader = "Warning"
	
	// ApplicationsCoreDeprecationMessage is the deprecation message for Applications.Core resource types
	ApplicationsCoreDeprecationMessage = `299 - "The Applications.Core namespace is deprecated. Please migrate to Radius.Core namespace with API version 2025-08-01-preview. For more information, see https://docs.radapp.io/guides/migration/"`
)

// DeprecationMiddleware returns an HTTP middleware that adds deprecation warnings for Applications.Core resource types
func DeprecationMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger := ucplog.FromContextOrDiscard(r.Context())
			
			// Check if the request path contains Applications.Core
			if strings.Contains(r.URL.Path, "/Applications.Core/") {
				// Add deprecation warning header
				w.Header().Add(DeprecationWarningHeader, ApplicationsCoreDeprecationMessage)
				
				// Log the deprecation warning
				logger.Info("Applications.Core resource type is deprecated. Please migrate to Radius.Core namespace.",
					"path", r.URL.Path,
					"namespace", "Applications.Core",
					"migration_target", "Radius.Core",
					"new_api_version", "2025-08-01-preview")
			}
			
			// Call the next handler
			next.ServeHTTP(w, r)
		})
	}
}
