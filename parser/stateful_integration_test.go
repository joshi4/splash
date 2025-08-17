package parser

import (
	"testing"
)

// TestStatefulIntegration tests the comprehensive integration of stateful parsers
// with complex mixed-format scenarios
func TestStatefulIntegration(t *testing.T) {
	parser := NewParser()

	// Complex scenario: Mixed formats with nested multi-line entries
	lines := []string{
		// Start with JSON
		`{"level":"INFO","message":"Application starting"}`,

		// Java exception starts
		`Exception in thread "main" java.lang.RuntimeException: Database connection failed`,
		"\tat com.example.service.DatabaseService.connect(DatabaseService.java:45)",
		"\tat com.example.service.DatabaseService.init(DatabaseService.java:23)",
		"Caused by: java.sql.SQLException: Connection timeout after 30 seconds",
		"\tat com.mysql.cj.jdbc.ConnectionImpl.connectOneTryOnly(ConnectionImpl.java:956)",
		"\tat com.mysql.cj.jdbc.ConnectionImpl.createNewIO(ConnectionImpl.java:836)",
		"\t... 15 more",

		// Switch to Python exception
		"Traceback (most recent call last):",
		`  File "backup_service.py", line 42, in backup_data`,
		"    connection.execute(query)",
		`  File "database.py", line 18, in execute`,
		"    cursor.execute(sql, params)",
		"DatabaseError: connection already closed",

		// Rsyslog multi-line entry
		"Aug  8 10:30:45 server rsyslogd[1234]: System notification:",
		"        Multiple database connection failures detected.",
		"        Automatic failover initiated to backup server.",
		"        Expected downtime: 30 seconds.",

		// Back to single-line formats
		`{"level":"ERROR","message":"Failover completed","downtime":"28s"}`,
		"2025/01/19 10:31:15 Services restored",
	}

	expectedFormats := []LogFormat{
		JSONFormat,            // JSON
		JavaExceptionFormat,   // Java exception header
		JavaExceptionFormat,   // Stack trace
		JavaExceptionFormat,   // Stack trace
		JavaExceptionFormat,   // Caused by
		JavaExceptionFormat,   // Stack trace
		JavaExceptionFormat,   // Stack trace
		JavaExceptionFormat,   // ... more
		PythonExceptionFormat, // Python traceback header
		PythonExceptionFormat, // File line
		PythonExceptionFormat, // Code line
		PythonExceptionFormat, // File line
		PythonExceptionFormat, // Code line
		PythonExceptionFormat, // Exception
		RsyslogFormat,         // Rsyslog header
		RsyslogFormat,         // Continuation
		RsyslogFormat,         // Continuation
		RsyslogFormat,         // Continuation
		JSONFormat,            // JSON
		GoStandardFormat,      // Go standard
	}

	if len(lines) != len(expectedFormats) {
		t.Fatalf("Test setup error: lines and expected formats must have same length")
	}

	for i, line := range lines {
		actual := parser.DetectFormat(line)
		expected := expectedFormats[i]

		if actual != expected {
			t.Errorf("Line %d: %q\n  Expected: %s\n  Actual: %s",
				i+1, line, expected.String(), actual.String())
		}
	}
}

// TestStatefulParserStateTransitions tests proper state management
// when switching between different multi-line formats
func TestStatefulParserStateTransitions(t *testing.T) {
	testCases := []struct {
		name         string
		lines        []string
		expectedLast LogFormat
	}{
		{
			name: "Java to Python transition",
			lines: []string{
				`Exception in thread "main" java.lang.RuntimeException: Error`,
				"\tat com.example.Main.main(Main.java:10)",
				"Traceback (most recent call last):", // Should start Python, end Java
				`  File "script.py", line 1, in <module>`,
			},
			expectedLast: PythonExceptionFormat,
		},
		{
			name: "Python to Rsyslog transition",
			lines: []string{
				"Traceback (most recent call last):",
				`  File "script.py", line 1, in <module>`,
				"Aug  8 10:30:45 server rsyslogd[1234]: Log entry:", // Should start Rsyslog, end Python
				"        Continuation line.",
			},
			expectedLast: RsyslogFormat,
		},
		{
			name: "Rsyslog to single-line transition",
			lines: []string{
				"Aug  8 10:30:45 server rsyslogd[1234]: Multi-line entry:",
				"        First continuation.",
				"        Second continuation.",
				`{"level":"INFO","message":"JSON entry"}`, // Should end Rsyslog, start JSON
			},
			expectedLast: JSONFormat,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Reset parser for each test case
			testParser := NewParser()

			var lastFormat LogFormat
			for _, line := range tc.lines {
				lastFormat = testParser.DetectFormat(line)
			}

			if lastFormat != tc.expectedLast {
				t.Errorf("Expected final format %s, got %s",
					tc.expectedLast.String(), lastFormat.String())
			}
		})
	}
}

// TestStatefulParserEdgeCases tests edge cases in stateful parsing
func TestStatefulParserEdgeCases(t *testing.T) {
	testCases := []struct {
		name     string
		lines    []string
		expected []LogFormat
	}{
		{
			name: "Empty continuation lines",
			lines: []string{
				`Exception in thread "main" java.lang.RuntimeException: Error`,
				"\tat com.example.Main.main(Main.java:10)",
				"", // Empty line should end Java exception
				"Next log line",
			},
			expected: []LogFormat{
				JavaExceptionFormat,
				JavaExceptionFormat,
				UnknownFormat, // Empty line
				UnknownFormat, // No specific format
			},
		},
		{
			name: "Mixed whitespace types",
			lines: []string{
				"Traceback (most recent call last):",
				"  File with spaces",      // Spaces
				"\tFile with tab",         // Tab
				"    Mixed    whitespace", // Multiple spaces
			},
			expected: []LogFormat{
				PythonExceptionFormat,
				PythonExceptionFormat,
				PythonExceptionFormat,
				PythonExceptionFormat,
			},
		},
		{
			name: "Rapid format switches",
			lines: []string{
				`{"level":"INFO"}`,
				`Exception in thread "main" java.lang.Error`,
				`{"level":"ERROR"}`,
				"Traceback (most recent call last):",
				`{"level":"DEBUG"}`,
			},
			expected: []LogFormat{
				JSONFormat,
				JavaExceptionFormat,
				JSONFormat,
				PythonExceptionFormat,
				JSONFormat,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Reset parser for each test case
			testParser := NewParser()

			for i, line := range tc.lines {
				if i >= len(tc.expected) {
					t.Fatalf("More lines than expected formats in test %s", tc.name)
				}

				actual := testParser.DetectFormat(line)
				expected := tc.expected[i]

				if actual != expected {
					t.Errorf("Line %d in %s: %q\n  Expected: %s\n  Actual: %s",
						i+1, tc.name, line, expected.String(), actual.String())
				}
			}
		})
	}
}
