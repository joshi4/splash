package parser

import (
	"testing"
)

func TestGoroutineStackTraceStatefulDetection(t *testing.T) {
	testCases := []struct {
		name            string
		lines           []string
		expectedFormats []LogFormat
	}{
		{
			name: "Basic goroutine stack trace",
			lines: []string{
				`goroutine 1 [running]:`,
				`main.Example(0x2080c3f50, 0x2, 0x4, 0x425c0, 0x5, 0xa)`,
				`        /Users/bill/Spaces/Go/Projects/src/github.com/goinaction/code/temp/main.go:9 +0x64`,
				`main.main()`,
				`        /Users/bill/Spaces/Go/Projects/src/github.com/goinaction/code/temp/main.go:5 +0x85`,
			},
			expectedFormats: []LogFormat{
				GoroutineStackTraceFormat,
				GoroutineStackTraceFormat,
				GoroutineStackTraceFormat,
				GoroutineStackTraceFormat,
				GoroutineStackTraceFormat,
			},
		},
		{
			name: "Multiple goroutines",
			lines: []string{
				`goroutine 1 [running]:`,
				`main.Example(0x2080c3f50, 0x2, 0x4, 0x425c0, 0x5, 0xa)`,
				`        /Users/bill/Spaces/Go/Projects/src/github.com/goinaction/code/temp/main.go:9 +0x64`,
				``,
				`goroutine 2 [runnable]:`,
				`runtime.forcegchelper()`,
				`        /Users/bill/go/src/runtime/proc.go:90`,
				`runtime.goexit()`,
				`        /Users/bill/go/src/runtime/asm_amd64.s:2232 +0x1`,
			},
			expectedFormats: []LogFormat{
				GoroutineStackTraceFormat,
				GoroutineStackTraceFormat,
				GoroutineStackTraceFormat,
				UnknownFormat, // Empty line
				GoroutineStackTraceFormat,
				GoroutineStackTraceFormat,
				GoroutineStackTraceFormat,
				GoroutineStackTraceFormat,
				GoroutineStackTraceFormat,
			},
		},
		{
			name: "Mixed formats with goroutine stack trace",
			lines: []string{
				`{"level":"INFO","message":"Starting application"}`,
				`goroutine 1 [running]:`,
				`main.Example(0x2080c3f50, 0x2, 0x4, 0x425c0, 0x5, 0xa)`,
				`        /Users/bill/Spaces/Go/Projects/src/github.com/goinaction/code/temp/main.go:9 +0x64`,
				`2025/01/19 10:30:00 INFO: Application started`,
			},
			expectedFormats: []LogFormat{
				JSONFormat,
				GoroutineStackTraceFormat,
				GoroutineStackTraceFormat,
				GoroutineStackTraceFormat,
				GoStandardFormat,
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

func TestGoroutineStackTraceDetector(t *testing.T) {
	detector := &StatefulGoroutineStackTraceDetector{}

	testCases := []struct {
		name     string
		line     string
		expected bool
	}{
		{
			name:     "Goroutine header - running",
			line:     "goroutine 1 [running]:",
			expected: true,
		},
		{
			name:     "Goroutine header - runnable",
			line:     "goroutine 2 [runnable]:",
			expected: true,
		},
		{
			name:     "Goroutine header - IO wait",
			line:     "goroutine 3 [IO wait]:",
			expected: true,
		},
		{
			name:     "Goroutine header - chan receive",
			line:     "goroutine 4 [chan receive]:",
			expected: true,
		},
		{
			name:     "Function call line",
			line:     "main.Example(0x2080c3f50, 0x2, 0x4, 0x425c0, 0x5, 0xa)",
			expected: true, // Function calls are part of goroutine stack traces
		},
		{
			name:     "File path line",
			line:     "        /Users/bill/Spaces/Go/Projects/src/github.com/goinaction/code/temp/main.go:9 +0x64",
			expected: true, // File paths are part of goroutine stack traces
		},
		{
			name:     "Stack trace line with runtime",
			line:     "        runtime.forcegchelper()",
			expected: true, // Runtime functions with whitespace should match
		},
		{
			name:     "Not a goroutine line",
			line:     "This is not a goroutine stack trace",
			expected: false,
		},
		{
			name:     "JSON log line",
			line:     `{"level":"INFO","message":"Starting application"}`,
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
			result := detector.Detect(nil, tc.line)
			if result != tc.expected {
				t.Errorf("Expected %v, got %v for line: %s", tc.expected, result, tc.line)
			}
		})
	}
}

func TestGoroutineStackTraceDetectorContinuation(t *testing.T) {
	detector := &StatefulGoroutineStackTraceDetector{}

	testCases := []struct {
		name     string
		line     string
		expected bool
	}{
		{
			name:     "Line with spaces",
			line:     "        /Users/bill/go/src/runtime/proc.go:90",
			expected: true,
		},
		{
			name:     "Line with tabs",
			line:     "\t\truntime.goexit()",
			expected: true,
		},
		{
			name:     "Function call with spaces",
			line:     "    main.Example(0x2080c3f50, 0x2, 0x4, 0x425c0, 0x5, 0xa)",
			expected: true,
		},
		{
			name:     "Line without leading whitespace",
			line:     "goroutine 1 [running]:",
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

func TestGoroutineStackTraceDetectorStart(t *testing.T) {
	detector := &StatefulGoroutineStackTraceDetector{}

	testCases := []struct {
		name     string
		line     string
		expected bool
	}{
		{
			name:     "Goroutine 1 running",
			line:     "goroutine 1 [running]:",
			expected: true,
		},
		{
			name:     "Goroutine 42 runnable",
			line:     "goroutine 42 [runnable]:",
			expected: true,
		},
		{
			name:     "Goroutine 123 chan receive",
			line:     "goroutine 123 [chan receive]:",
			expected: true,
		},
		{
			name:     "Goroutine with spaces in status",
			line:     "goroutine 1 [IO wait]:",
			expected: true,
		},
		{
			name:     "Not a goroutine header",
			line:     "main.Example(0x2080c3f50, 0x2, 0x4, 0x425c0, 0x5, 0xa)",
			expected: false,
		},
		{
			name:     "Stack trace line",
			line:     "        /Users/bill/go/src/runtime/proc.go:90",
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

func TestGoroutineStackTraceProperties(t *testing.T) {
	detector := &StatefulGoroutineStackTraceDetector{}

	// Test Format() method
	if detector.Format() != GoroutineStackTraceFormat {
		t.Errorf("Expected format %v, got %v", GoroutineStackTraceFormat, detector.Format())
	}

	// Test Specificity() method
	expectedSpecificity := 70
	if detector.Specificity() != expectedSpecificity {
		t.Errorf("Expected specificity %d, got %d", expectedSpecificity, detector.Specificity())
	}

	// Test PatternLength() method
	expectedPatternLength := len(goroutineStartPattern) + len(goroutineStackTraceLinePattern)
	if detector.PatternLength() != expectedPatternLength {
		t.Errorf("Expected pattern length %d, got %d", expectedPatternLength, detector.PatternLength())
	}

	// Test DetectEnd() method - should always return false for goroutine stack traces
	if detector.DetectEnd(nil, "any line") {
		t.Error("Expected DetectEnd to return false for all lines")
	}
}
