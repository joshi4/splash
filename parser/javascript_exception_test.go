package parser

import (
	"testing"
)

func TestJavaScriptExceptionStatefulDetection(t *testing.T) {
	testCases := []struct {
		name            string
		lines           []string
		expectedFormats []LogFormat
	}{
		{
			name: "Basic JavaScript error with stack trace",
			lines: []string{
				`Error`,
				`    at sum (/home/dev/Documents/trace.js:2:17)`,
				`    at start (/home/dev/Documents/trace.js:11:13)`,
				`    at Object.<anonymous> (/home/dev/Documents/trace.js:16:1)`,
				`    at Module._compile (internal/modules/cjs/loader.js:959:30)`,
			},
			expectedFormats: []LogFormat{
				JavaScriptExceptionFormat,
				JavaScriptExceptionFormat,
				JavaScriptExceptionFormat,
				JavaScriptExceptionFormat,
				JavaScriptExceptionFormat,
			},
		},
		{
			name: "TypeError with stack trace",
			lines: []string{
				`TypeError: Cannot read property 'foo' of undefined`,
				`    at processData (/home/app/src/utils.js:45:12)`,
				`    at handleRequest (/home/app/src/server.js:123:5)`,
				`    at Server.<anonymous> (/home/app/src/server.js:89:3)`,
			},
			expectedFormats: []LogFormat{
				JavaScriptExceptionFormat,
				JavaScriptExceptionFormat,
				JavaScriptExceptionFormat,
				JavaScriptExceptionFormat,
			},
		},
		{
			name: "Trace message with stack trace",
			lines: []string{
				`Trace: add called with  2 and 3`,
				`    at sum (/home/dev/Documents/stacktrace.js:2:13)`,
				`    at start (/home/dev/Documents/stacktrace.js:11:13)`,
				`    at Object.<anonymous> (/home/dev/Documents/stacktrace.js:16:1)`,
			},
			expectedFormats: []LogFormat{
				JavaScriptExceptionFormat,
				JavaScriptExceptionFormat,
				JavaScriptExceptionFormat,
				JavaScriptExceptionFormat,
			},
		},
		{
			name: "ReferenceError with stack trace",
			lines: []string{
				`ReferenceError: undefinedVariable is not defined`,
				`    at calculateTotal (/home/app/src/calculator.js:15:20)`,
				`    at Object.exports.processOrder (/home/app/src/orders.js:67:8)`,
			},
			expectedFormats: []LogFormat{
				JavaScriptExceptionFormat,
				JavaScriptExceptionFormat,
				JavaScriptExceptionFormat,
			},
		},
		{
			name: "Mixed formats with JavaScript exception",
			lines: []string{
				`{"level":"INFO","message":"Starting application"}`,
				`Error`,
				`    at sum (/home/dev/Documents/trace.js:2:17)`,
				`    at start (/home/dev/Documents/trace.js:11:13)`,
				`2025/01/19 10:30:00 INFO: Application started`,
			},
			expectedFormats: []LogFormat{
				JSONFormat,
				JavaScriptExceptionFormat,
				JavaScriptExceptionFormat,
				JavaScriptExceptionFormat,
				GoStandardFormat,
			},
		},
		{
			name: "Stack trace with internal modules",
			lines: []string{
				`Error`,
				`    at internal/main/run_main_module.js:17:11`,
				`    at Function.Module.runMain (internal/modules/cjs/loader.js:1047:10)`,
			},
			expectedFormats: []LogFormat{
				JavaScriptExceptionFormat,
				JavaScriptExceptionFormat,
				JavaScriptExceptionFormat,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			parser := NewParser()

			for i, line := range tc.lines {
				format := parser.DetectFormat(line)
				if format != tc.expectedFormats[i] {
					t.Errorf("Line %d: expected format %s, got %s for line: %s",
						i, tc.expectedFormats[i], format, line)
				}
			}
		})
	}
}

func TestJavaScriptExceptionDetector(t *testing.T) {
	detector := &StatefulJavaScriptExceptionDetector{}

	testCases := []struct {
		name     string
		line     string
		expected bool
	}{
		{
			name:     "Simple Error",
			line:     "Error",
			expected: true,
		},
		{
			name:     "TypeError with message",
			line:     "TypeError: Cannot read property 'foo' of undefined",
			expected: true,
		},
		{
			name:     "ReferenceError with message",
			line:     "ReferenceError: undefinedVariable is not defined",
			expected: true,
		},
		{
			name:     "SyntaxError with message",
			line:     "SyntaxError: Unexpected token '}'",
			expected: true,
		},
		{
			name:     "Custom Exception",
			line:     "ValidationException: Invalid input data",
			expected: true,
		},
		{
			name:     "Trace message",
			line:     "Trace: add called with  2 and 3",
			expected: true,
		},
		{
			name:     "Stack trace line with function",
			line:     "    at sum (/home/dev/Documents/trace.js:2:17)",
			expected: true,
		},
		{
			name:     "Stack trace line with Object method",
			line:     "    at Object.<anonymous> (/home/dev/Documents/trace.js:16:1)",
			expected: true,
		},
		{
			name:     "Stack trace line with internal module",
			line:     "    at internal/main/run_main_module.js:17:11",
			expected: true,
		},
		{
			name:     "Stack trace line with Module method",
			line:     "    at Module._compile (internal/modules/cjs/loader.js:959:30)",
			expected: true,
		},
		{
			name:     "Not a JavaScript exception",
			line:     "This is not a JavaScript exception",
			expected: false,
		},
		{
			name:     "JSON log line",
			line:     `{"level":"INFO","message":"Starting application"}`,
			expected: false,
		},
		{
			name:     "Go log line",
			line:     `2025/01/19 10:30:00 ERROR: Something went wrong`,
			expected: false,
		},
		{
			name:     "Empty line",
			line:     "",
			expected: false,
		},
		{
			name:     "Line starting with Error but not an exception",
			line:     "ErrorCode: 404 - Not Found",
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := detector.Detect(nil, tc.line)
			if result != tc.expected {
				t.Errorf("Expected %v, got %v for line: %s", tc.expected, result, tc.line)
			}
		})
	}
}

func TestJavaScriptExceptionDetectorContinuation(t *testing.T) {
	detector := &StatefulJavaScriptExceptionDetector{}

	testCases := []struct {
		name     string
		line     string
		expected bool
	}{
		{
			name:     "Stack trace line with spaces",
			line:     "    at sum (/home/dev/Documents/trace.js:2:17)",
			expected: true,
		},
		{
			name:     "Stack trace line with tabs",
			line:     "\t\tat processData (/home/app/src/utils.js:45:12)",
			expected: true,
		},
		{
			name:     "Internal module stack trace",
			line:     "    at internal/main/run_main_module.js:17:11",
			expected: true,
		},
		{
			name:     "Line without leading whitespace",
			line:     "Error",
			expected: false,
		},
		{
			name:     "Line with at but no whitespace",
			line:     "at sum (/home/dev/Documents/trace.js:2:17)",
			expected: false,
		},
		{
			name:     "Empty line",
			line:     "",
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := detector.DetectContinuation(nil, tc.line)
			if result != tc.expected {
				t.Errorf("Expected %v, got %v for line: %s", tc.expected, result, tc.line)
			}
		})
	}
}

func TestJavaScriptExceptionDetectorStart(t *testing.T) {
	detector := &StatefulJavaScriptExceptionDetector{}

	testCases := []struct {
		name     string
		line     string
		expected bool
	}{
		{
			name:     "Simple Error",
			line:     "Error",
			expected: true,
		},
		{
			name:     "TypeError",
			line:     "TypeError: Cannot read property 'foo' of undefined",
			expected: true,
		},
		{
			name:     "ReferenceError",
			line:     "ReferenceError: undefinedVariable is not defined",
			expected: true,
		},
		{
			name:     "SyntaxError",
			line:     "SyntaxError: Unexpected token",
			expected: true,
		},
		{
			name:     "RangeError",
			line:     "RangeError: Maximum call stack size exceeded",
			expected: true,
		},
		{
			name:     "Custom Exception",
			line:     "ValidationException: Invalid data",
			expected: true,
		},
		{
			name:     "Trace message",
			line:     "Trace: debugging information",
			expected: true,
		},
		{
			name:     "Stack trace line",
			line:     "    at sum (/home/dev/Documents/trace.js:2:17)",
			expected: false,
		},
		{
			name:     "Not an exception header",
			line:     "ErrorCode: 404",
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := detector.DetectStart(nil, tc.line)
			if result != tc.expected {
				t.Errorf("Expected %v, got %v for line: %s", tc.expected, result, tc.line)
			}
		})
	}
}

func TestJavaScriptExceptionProperties(t *testing.T) {
	detector := &StatefulJavaScriptExceptionDetector{}

	// Test Format() method
	if detector.Format() != JavaScriptExceptionFormat {
		t.Errorf("Expected format %v, got %v", JavaScriptExceptionFormat, detector.Format())
	}

	// Test Specificity() method
	expectedSpecificity := 70
	if detector.Specificity() != expectedSpecificity {
		t.Errorf("Expected specificity %d, got %d", expectedSpecificity, detector.Specificity())
	}

	// Test PatternLength() method
	expectedPatternLength := len(jsExceptionStartPattern) + len(jsStackTraceLinePattern)
	if detector.PatternLength() != expectedPatternLength {
		t.Errorf("Expected pattern length %d, got %d", expectedPatternLength, detector.PatternLength())
	}

	// Test DetectEnd() method - should always return false for JavaScript exceptions
	if detector.DetectEnd(nil, "any line") {
		t.Error("Expected DetectEnd to return false for all lines")
	}
}
