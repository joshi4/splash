package parser

import (
	"testing"
)

func TestJavaExceptionStatefulDetection(t *testing.T) {
	testCases := []struct {
		name            string
		lines           []string
		expectedFormats []LogFormat
	}{
		{
			name: "Basic Java exception with stack trace",
			lines: []string{
				`Exception in thread "main" java.lang.ArithmeticException: / by zero`,
				"\tat com.example.MyClass.divide(MyClass.java:10)",
				"\tat com.example.MyClass.calculate(MyClass.java:6)",
				"\tat com.example.MyClass.main(MyClass.java:3)",
			},
			expectedFormats: []LogFormat{
				JavaExceptionFormat,
				JavaExceptionFormat,
				JavaExceptionFormat,
				JavaExceptionFormat,
			},
		},
		{
			name: "Java exception with Caused by",
			lines: []string{
				`Exception in thread "main" java.lang.RuntimeException: Something went wrong`,
				"\tat com.example.service.PaymentService.processPayment(PaymentService.java:45)",
				"Caused by: java.sql.SQLException: Connection timeout",
				"\tat com.mysql.cj.jdbc.ConnectionImpl.createNewIO(ConnectionImpl.java:836)",
				"\t... 34 more",
			},
			expectedFormats: []LogFormat{
				JavaExceptionFormat,
				JavaExceptionFormat,
				JavaExceptionFormat,
				JavaExceptionFormat,
				JavaExceptionFormat,
			},
		},
		{
			name: "Mixed formats with Java exception",
			lines: []string{
				`{"level":"INFO","message":"Starting application"}`,
				`Exception in thread "main" java.lang.ArithmeticException: / by zero`,
				"\tat com.example.MyClass.divide(MyClass.java:10)",
				"\tat com.example.MyClass.calculate(MyClass.java:6)",
				`{"level":"ERROR","message":"Application crashed"}`,
				"Caused by: java.sql.SQLException: Connection timeout",
				"\tat com.mysql.cj.jdbc.ConnectionImpl.createNewIO(ConnectionImpl.java:836)",
				"\t... 5 more",
				"2025/01/19 10:30:00 Application restarted",
			},
			expectedFormats: []LogFormat{
				JSONFormat,          // JSON log
				JavaExceptionFormat, // Exception header
				JavaExceptionFormat, // Stack trace with leading whitespace
				JavaExceptionFormat, // Stack trace with leading whitespace
				JSONFormat,          // JSON log (switches format)
				JavaExceptionFormat, // Caused by header (new exception)
				JavaExceptionFormat, // Stack trace with leading whitespace
				JavaExceptionFormat, // ... more with leading whitespace
				GoStandardFormat,    // Go timestamp format (switches format)
			},
		},
		{
			name: "Java exception ends when line has no leading whitespace and doesn't match header",
			lines: []string{
				`Exception in thread "main" java.lang.ArithmeticException: / by zero`,
				"\tat com.example.MyClass.divide(MyClass.java:10)",
				"\tat com.example.MyClass.calculate(MyClass.java:6)",
				"INFO: Application restarted", // No leading whitespace - should end exception
				`{"level":"INFO","message":"JSON log"}`,
			},
			expectedFormats: []LogFormat{
				JavaExceptionFormat, // Exception header
				JavaExceptionFormat, // Stack trace with leading whitespace
				JavaExceptionFormat, // Stack trace with leading whitespace
				UnknownFormat,       // No leading whitespace, not exception header - ends exception
				JSONFormat,          // JSON format
			},
		},
		{
			name: "Whitespace-based continuation logic",
			lines: []string{
				`Exception in thread "main" java.lang.ArithmeticException: / by zero`,
				"\tat com.example.MyClass.divide(MyClass.java:10)",
				"    some continuation line with spaces", // Leading spaces should continue exception
				"\t\tanother continuation with tabs",     // Leading tabs should continue exception
				"No leading whitespace ends it",          // No leading whitespace should end exception
			},
			expectedFormats: []LogFormat{
				JavaExceptionFormat, // Exception header
				JavaExceptionFormat, // Stack trace with leading tab
				JavaExceptionFormat, // Line with leading spaces - should continue
				JavaExceptionFormat, // Line with leading tabs - should continue
				UnknownFormat,       // No leading whitespace - should end exception
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

func TestJavaExceptionWhitespaceHandling(t *testing.T) {
	parser := NewParser()

	testCases := []struct {
		line     string
		expected LogFormat
		desc     string
	}{
		{
			line:     `Exception in thread "main" java.lang.ArithmeticException: / by zero`,
			expected: JavaExceptionFormat,
			desc:     "Exception header with no leading whitespace",
		},
		{
			line:     "Caused by: java.sql.SQLException: Connection timeout",
			expected: JavaExceptionFormat,
			desc:     "Caused by header with no leading whitespace",
		},
		{
			line:     "\tat com.example.MyClass.divide(MyClass.java:10)",
			expected: JavaExceptionFormat,
			desc:     "Stack trace line with leading tab",
		},
		{
			line:     "    at com.example.MyClass.divide(MyClass.java:10)",
			expected: JavaExceptionFormat,
			desc:     "Stack trace line with leading spaces",
		},
		{
			line:     "\t... 34 more",
			expected: JavaExceptionFormat,
			desc:     "More line with leading tab",
		},
		{
			line:     "    ... 34 more",
			expected: JavaExceptionFormat,
			desc:     "More line with leading spaces",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			actual := parser.DetectFormat(tc.line)
			if actual != tc.expected {
				t.Errorf("Line: %q\n  Expected: %s\n  Actual: %s",
					tc.line, tc.expected.String(), actual.String())
			}
		})
	}
}

func TestJavaExceptionLongStackTrace(t *testing.T) {
	parser := NewParser()

	// Test with a very long stack trace to ensure no hardcoded limits
	lines := []string{
		`Exception in thread "main" java.lang.RuntimeException: Something went wrong`,
	}

	// Add 50 stack trace lines
	for i := 0; i < 50; i++ {
		lines = append(lines, "\tat com.example.service.Class"+string(rune('A'+i%26))+".method(File.java:"+string(rune('0'+i%10))+")")
	}

	// Add a caused by section
	lines = append(lines, "Caused by: java.sql.SQLException: Connection timeout")
	for i := 0; i < 20; i++ {
		lines = append(lines, "\tat com.mysql.cj.jdbc.Class"+string(rune('A'+i%26))+".method(File.java:"+string(rune('0'+i%10))+")")
	}
	lines = append(lines, "\t... 34 more")

	// Test that all lines are detected as Java exception format
	for i, line := range lines {
		actual := parser.DetectFormat(line)
		if actual != JavaExceptionFormat {
			t.Errorf("Line %d should be JavaExceptionFormat but got %s: %q",
				i+1, actual.String(), line)
		}
	}

	// Test that a non-whitespace line ends the exception
	endLine := "INFO: Application restarted"
	actual := parser.DetectFormat(endLine)
	if actual == JavaExceptionFormat {
		t.Errorf("Line without leading whitespace should not be JavaExceptionFormat: %q", endLine)
	}
}

func TestJavaExceptionWithTestData(t *testing.T) {
	parser := NewParser()

	// Test the mixed Java exception test data file
	testLines := []string{
		`{"level":"INFO","message":"Starting application"}`,
		`Exception in thread "main" java.lang.ArithmeticException: / by zero`,
		"\tat com.example.MyClass.divide(MyClass.java:10)",
		"\tat com.example.MyClass.calculate(MyClass.java:6)",
		"\tat com.example.MyClass.main(MyClass.java:3)",
		`{"level":"ERROR","message":"Application crashed"}`,
		"Caused by: java.sql.SQLException: Connection timeout",
		"\tat com.mysql.cj.jdbc.ConnectionImpl.createNewIO(ConnectionImpl.java:836)",
		"\t... 5 more",
		"2025/01/19 10:30:00 Application restarted",
	}

	expectedFormats := []LogFormat{
		JSONFormat,          // JSON log
		JavaExceptionFormat, // Exception header
		JavaExceptionFormat, // Stack trace
		JavaExceptionFormat, // Stack trace
		JavaExceptionFormat, // Stack trace
		JSONFormat,          // JSON log (switches format)
		JavaExceptionFormat, // Caused by header (new exception)
		JavaExceptionFormat, // Stack trace
		JavaExceptionFormat, // ... more
		GoStandardFormat,    // Go timestamp format (switches format)
	}

	for i, line := range testLines {
		actual := parser.DetectFormat(line)
		expected := expectedFormats[i]

		if actual != expected {
			t.Errorf("Mixed test data line %d: %q\n  Expected: %s\n  Actual: %s",
				i+1, line, expected.String(), actual.String())
		}
	}
}
