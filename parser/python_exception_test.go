package parser

import (
	"fmt"
	"testing"
)

func TestPythonExceptionStatefulDetection(t *testing.T) {

	testCases := []struct {
		name           string
		lines          []string
		expectedFormats []LogFormat
	}{
		{
			name: "Basic Python traceback",
			lines: []string{
				"Traceback (most recent call last):",
				`  File "example_trace.py", line 21, in <module>`,
				"    function_a()",
				`  File "example_trace.py", line 17, in function_a`,
				"    function_b()",
				`  File "example_trace.py", line 6, in function_c`,
				"    result = 10 / 0  # This line will cause an error",
				"ZeroDivisionError: division by zero",
			},
			expectedFormats: []LogFormat{
				PythonExceptionFormat,
				PythonExceptionFormat,
				PythonExceptionFormat,
				PythonExceptionFormat,
				PythonExceptionFormat,
				PythonExceptionFormat,
				PythonExceptionFormat,
				PythonExceptionFormat,
			},
		},
		{
			name: "Mixed formats with Python exception",
			lines: []string{
				`{"level":"INFO","message":"Starting application"}`,
				"Traceback (most recent call last):",
				`  File "example_trace.py", line 21, in <module>`,
				"    function_a()",
				`  File "example_trace.py", line 17, in function_a`,
				"    function_b()",
				"ZeroDivisionError: division by zero",
				`{"level":"ERROR","message":"Application crashed"}`,
				"2025/01/19 10:30:00 Application restarted",
			},
			expectedFormats: []LogFormat{
				JSONFormat,            // JSON log
				PythonExceptionFormat, // Traceback header
				PythonExceptionFormat, // File line (continuation)
				PythonExceptionFormat, // Code line (continuation)
				PythonExceptionFormat, // File line (continuation)
				PythonExceptionFormat, // Code line (continuation)
				PythonExceptionFormat, // Exception line (no leading whitespace but matches Error pattern)
				JSONFormat,            // JSON log (switches format)
				GoStandardFormat,      // Go timestamp format (switches format)
			},
		},
		{
			name: "Python exception ends when line has no leading whitespace and doesn't match header/exception",
			lines: []string{
				"Traceback (most recent call last):",
				`  File "example_trace.py", line 21, in <module>`,
				"    function_a()",
				"ZeroDivisionError: division by zero",
				"INFO: Application restarted", // No leading whitespace, doesn't match patterns - should end exception
				`{"level":"INFO","message":"JSON log"}`,
			},
			expectedFormats: []LogFormat{
				PythonExceptionFormat, // Traceback header
				PythonExceptionFormat, // File line with leading whitespace
				PythonExceptionFormat, // Code line with leading whitespace
				PythonExceptionFormat, // Exception line (matches Error pattern)
				UnknownFormat,         // No leading whitespace, not traceback/exception header - ends exception
				JSONFormat,            // JSON format
			},
		},
		{
			name: "Whitespace-based continuation logic for Python",
			lines: []string{
				"Traceback (most recent call last):",
				`  File "example_trace.py", line 21, in <module>`,
				"    some code line with spaces", // Leading spaces should continue exception
				"\t\tanother line with tabs",     // Leading tabs should continue exception
				"No leading whitespace ends it", // No leading whitespace should end exception
			},
			expectedFormats: []LogFormat{
				PythonExceptionFormat, // Traceback header
				PythonExceptionFormat, // File line with leading spaces
				PythonExceptionFormat, // Line with leading spaces - should continue
				PythonExceptionFormat, // Line with leading tabs - should continue
				UnknownFormat,         // No leading whitespace - should end exception
			},
		},
		{
			name: "Various Python exception types",
			lines: []string{
				"Traceback (most recent call last):",
				`  File "test.py", line 1, in <module>`,
				"    raise ValueError('test')",
				"ValueError: test",
				"Traceback (most recent call last):",
				`  File "test.py", line 1, in <module>`,
				"    import nonexistent",
				"ImportError: No module named 'nonexistent'",
				"Traceback (most recent call last):",
				`  File "test.py", line 1, in <module>`,
				"    d = {}; d['key']",
				"KeyError: 'key'",
			},
			expectedFormats: []LogFormat{
				PythonExceptionFormat, // Traceback header
				PythonExceptionFormat, // File line
				PythonExceptionFormat, // Code line
				PythonExceptionFormat, // ValueError
				PythonExceptionFormat, // New traceback header (continues as Python)
				PythonExceptionFormat, // File line
				PythonExceptionFormat, // Code line
				PythonExceptionFormat, // ImportError
				PythonExceptionFormat, // New traceback header (continues as Python)
				PythonExceptionFormat, // File line
				PythonExceptionFormat, // Code line
				PythonExceptionFormat, // KeyError
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Reset parser state for each test case
			parser := NewParser()

			for i, line := range tc.lines {
				actual := parser.DetectFormat(line)
				expected := tc.expectedFormats[i]

				if actual != expected {
					t.Errorf("Line %d: %q\n  Expected: %s\n  Actual: %s",
						i+1, line, expected.String(), actual.String())
				}
			}
		})
	}
}

func TestPythonExceptionPatterns(t *testing.T) {
	parser := NewParser()

	testCases := []struct {
		line     string
		expected LogFormat
		desc     string
	}{
		{
			line:     "Traceback (most recent call last):",
			expected: PythonExceptionFormat,
			desc:     "Traceback header with no leading whitespace",
		},
		{
			line:     "ValueError: invalid literal for int()",
			expected: PythonExceptionFormat,
			desc:     "ValueError exception line",
		},
		{
			line:     "ZeroDivisionError: division by zero",
			expected: PythonExceptionFormat,
			desc:     "ZeroDivisionError exception line",
		},
		{
			line:     "ImportError: No module named 'xyz'",
			expected: PythonExceptionFormat,
			desc:     "ImportError exception line",
		},
		{
			line:     "KeyError: 'missing_key'",
			expected: PythonExceptionFormat,
			desc:     "KeyError exception line",
		},
		{
			line:     "AttributeError: 'NoneType' object has no attribute 'method'",
			expected: PythonExceptionFormat,
			desc:     "AttributeError exception line",
		},
		{
			line:     `  File "script.py", line 42, in main`,
			expected: PythonExceptionFormat,
			desc:     "File line with leading spaces",
		},
		{
			line:     "\t\tFile \"script.py\", line 42, in main",
			expected: PythonExceptionFormat,
			desc:     "File line with leading tabs",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			// Start with a traceback header to establish Python exception context
			parser.DetectFormat("Traceback (most recent call last):")
			
			actual := parser.DetectFormat(tc.line)
			if actual != tc.expected {
				t.Errorf("Line: %q\n  Expected: %s\n  Actual: %s",
					tc.line, tc.expected.String(), actual.String())
			}
		})
	}
}

func TestPythonExceptionLongTraceback(t *testing.T) {
	parser := NewParser()

	// Test with a very long traceback to ensure no hardcoded limits
	lines := []string{
		"Traceback (most recent call last):",
	}

	// Add 50 file/code line pairs
	for i := 0; i < 50; i++ {
		filename := fmt.Sprintf("file%d.py", i)
		lines = append(lines, fmt.Sprintf(`  File "%s", line %d, in function_%d`, filename, i+1, i))
		lines = append(lines, fmt.Sprintf("    call_function_%d()", i+1))
	}

	// Add final exception
	lines = append(lines, "RuntimeError: deep recursion error")

	// Test that all lines are detected as Python exception format
	for i, line := range lines {
		actual := parser.DetectFormat(line)
		if actual != PythonExceptionFormat {
			t.Errorf("Line %d should be PythonExceptionFormat but got %s: %q",
				i+1, actual.String(), line)
		}
	}

	// Test that a non-whitespace line ends the exception
	endLine := "INFO: Application restarted"
	actual := parser.DetectFormat(endLine)
	if actual == PythonExceptionFormat {
		t.Errorf("Line without leading whitespace should not be PythonExceptionFormat: %q", endLine)
	}
}
