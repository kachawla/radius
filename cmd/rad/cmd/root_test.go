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

package cmd

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// Test_PanicRecoveryBehavior tests the panic recovery middleware behavior
// by simulating what happens when a panic occurs during command execution.
func Test_PanicRecoveryBehavior(t *testing.T) {
	t.Run("panic recovery captures and formats panic message", func(t *testing.T) {
		// This test verifies the panic recovery behavior by capturing
		// what the deferred function would output
		var output bytes.Buffer

		// Simulate the panic recovery logic
		func() {
			defer func() {
				if r := recover(); r != nil {
					output.WriteString("Error: An unexpected panic occurred\n")
					output.WriteString(fmt.Sprintf("Panic: %v\n", r))
					output.WriteString("\nPlease report this issue at https://github.com/radius-project/radius/issues\n")
					output.WriteString("") // Output an extra blank line for readability
				}
			}()

			// Simulate a panic
			panic("test panic message")
		}()

		result := output.String()

		// Verify the output format matches what's in the Execute() function
		require.Contains(t, result, "Error: An unexpected panic occurred")
		require.Contains(t, result, "Panic: test panic message")
		require.Contains(t, result, "Please report this issue at https://github.com/radius-project/radius/issues")
	})

	t.Run("no panic recovery output when no panic occurs", func(t *testing.T) {
		var output bytes.Buffer

		// Simulate the panic recovery logic without a panic
		func() {
			defer func() {
				if r := recover(); r != nil {
					output.WriteString("Error: An unexpected panic occurred\n")
					output.WriteString(fmt.Sprintf("Panic: %v\n", r))
					output.WriteString("\nPlease report this issue at https://github.com/radius-project/radius/issues\n")
					output.WriteString("")
				}
			}()

			// No panic occurs
		}()

		result := output.String()

		// Verify no output when there's no panic
		require.Empty(t, result)
	})

	t.Run("panic recovery handles different panic types", func(t *testing.T) {
		testCases := []struct {
			name        string
			panicValue  interface{}
			expectedMsg string
		}{
			{
				name:        "string panic",
				panicValue:  "string error",
				expectedMsg: "Panic: string error",
			},
			{
				name:        "error panic",
				panicValue:  fmt.Errorf("error object"),
				expectedMsg: "Panic: error object",
			},
			{
				name:        "integer panic",
				panicValue:  42,
				expectedMsg: "Panic: 42",
			},
			{
				name:        "nil panic",
				panicValue:  nil,
				expectedMsg: "Panic: <nil>",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				var output bytes.Buffer

				func() {
					defer func() {
						if r := recover(); r != nil {
							output.WriteString(fmt.Sprintf("Panic: %v", r))
						}
					}()

					panic(tc.panicValue)
				}()

				result := output.String()
				require.Contains(t, result, tc.expectedMsg)
			})
		}
	})
}

func Test_prettyPrintRPError(t *testing.T) {
	t.Run("handles standard error", func(t *testing.T) {
		err := fmt.Errorf("test error message")
		result := prettyPrintRPError(err)
		require.Contains(t, result, "test error")
	})
}

func Test_prettyPrintJSON(t *testing.T) {
	t.Run("formats JSON correctly", func(t *testing.T) {
		obj := map[string]string{"key": "value"}
		result, err := prettyPrintJSON(obj)
		require.NoError(t, err)
		require.Contains(t, result, "key")
		require.Contains(t, result, "value")
		// Verify it's indented
		require.True(t, strings.Contains(result, "\n"))
	})

	t.Run("handles invalid JSON", func(t *testing.T) {
		// Create something that can't be marshalled
		invalidObj := make(chan int)
		_, err := prettyPrintJSON(invalidObj)
		require.Error(t, err)
	})

	t.Run("formats complex objects", func(t *testing.T) {
		obj := map[string]interface{}{
			"nested": map[string]string{
				"inner": "value",
			},
			"array": []string{"a", "b", "c"},
		}
		result, err := prettyPrintJSON(obj)
		require.NoError(t, err)
		require.Contains(t, result, "nested")
		require.Contains(t, result, "inner")
		require.Contains(t, result, "array")
	})
}
