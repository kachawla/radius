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
	"io"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// Test_HandlePanic tests the panic recovery handler directly
func Test_HandlePanic(t *testing.T) {
	t.Run("panic recovery captures and formats panic message with stack trace", func(t *testing.T) {
		// Capture stdout
		oldStdout := captureStdout(t)
		defer oldStdout.Restore()

		// Execute the panic recovery by calling handlePanic in a defer after a panic
		func() {
			defer handlePanic()
			panic("test panic message")
		}()

		result := oldStdout.String()

		// Verify the output format
		require.Contains(t, result, "Error: An unexpected internal error occurred: test panic message")
		require.Contains(t, result, "goroutine")
		require.Contains(t, result, "Please report this issue at https://github.com/radius-project/radius/issues")
	})

	t.Run("no panic recovery output when no panic occurs", func(t *testing.T) {
		// Capture stdout
		oldStdout := captureStdout(t)
		defer oldStdout.Restore()

		// Call handlePanic without a panic
		func() {
			defer handlePanic()
			// No panic occurs
		}()

		result := oldStdout.String()

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
				expectedMsg: "Error: An unexpected internal error occurred: string error",
			},
			{
				name:        "error panic",
				panicValue:  fmt.Errorf("error object"),
				expectedMsg: "Error: An unexpected internal error occurred: error object",
			},
			{
				name:        "integer panic",
				panicValue:  42,
				expectedMsg: "Error: An unexpected internal error occurred: 42",
			},
			{
				name:        "nil pointer panic",
				panicValue:  (*int)(nil),
				expectedMsg: "Error: An unexpected internal error occurred: <nil>",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				// Capture stdout
				oldStdout := captureStdout(t)
				defer oldStdout.Restore()

				func() {
					defer handlePanic()
					panic(tc.panicValue)
				}()

				result := oldStdout.String()
				require.Contains(t, result, tc.expectedMsg)
			})
		}
	})
}

// stdoutCapture helps capture stdout for testing
type stdoutCapture struct {
	buffer  *bytes.Buffer
	oldOut  *os.File
	reader  *os.File
	writer  *os.File
}

func (s *stdoutCapture) String() string {
	return s.buffer.String()
}

func (s *stdoutCapture) Restore() {
	os.Stdout = s.oldOut
	s.writer.Close()
	io.Copy(s.buffer, s.reader)
}

// captureStdout captures stdout for testing purposes
func captureStdout(t *testing.T) *stdoutCapture {
	t.Helper()
	
	oldOut := os.Stdout
	reader, writer, err := os.Pipe()
	require.NoError(t, err)
	
	os.Stdout = writer
	
	return &stdoutCapture{
		buffer: &bytes.Buffer{},
		oldOut: oldOut,
		reader: reader,
		writer: writer,
	}
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
