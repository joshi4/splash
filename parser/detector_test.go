package parser

import (
	"testing"
	"time"
)

func TestOptimizedParser(t *testing.T) {
	parser := NewParser()

	// Test cases for various log formats
	tests := []struct {
		name     string
		line     string
		expected LogFormat
	}{
		{
			name:     "JSON format",
			line:     `{"timestamp":"2025-01-19T10:30:00Z","level":"ERROR","message":"Database connection failed","service":"api"}`,
			expected: JSONFormat,
		},
		{
			name:     "Logfmt format",
			line:     `timestamp=2025-01-19T10:30:00Z level=error msg="Database connection failed" service=api`,
			expected: LogfmtFormat,
		},
		{
			name:     "Apache Common format",
			line:     `127.0.0.1 - - [19/Jan/2025:10:30:00 +0000] "GET /api/users HTTP/1.1" 200 1234`,
			expected: ApacheCommonFormat,
		},
		{
			name:     "Nginx format",
			line:     `127.0.0.1 - - [19/Jan/2025:10:30:00 +0000] "GET /api/users HTTP/1.1" 200 1234 "-" "Mozilla/5.0"`,
			expected: NginxFormat,
		},
		{
			name:     "Syslog format",
			line:     `Jan 19 10:30:00 hostname myapp[1234]: ERROR: Database connection failed`,
			expected: SyslogFormat,
		},
		{
			name:     "Go standard format",
			line:     `2025/01/19 10:30:00 ERROR: Database connection failed`,
			expected: GoStandardFormat,
		},
		{
			name:     "Rails format",
			line:     `[2025-01-19 10:30:00] ERROR -- : Database connection failed`,
			expected: RailsFormat,
		},
		{
			name:     "Docker format",
			line:     `2025-01-19T10:30:00.123456789Z ERROR Database connection failed`,
			expected: DockerFormat,
		},
		{
			name:     "Kubernetes format",
			line:     `2025-01-19T10:30:00.123Z 1 main.go:42] ERROR Database connection failed`,
			expected: KubernetesFormat,
		},
		{
			name:     "Heroku format",
			line:     `2025-01-19T10:30:00+00:00 app[web.1]: ERROR Database connection failed`,
			expected: HerokuFormat,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.DetectFormat(tt.line)
			if result != tt.expected {
				t.Errorf("DetectFormat(%q) = %v, expected %v", tt.line, result, tt.expected)
			}
		})
	}
}

func TestParserOptimization(t *testing.T) {
	parser := NewParser()

	// First detection should set the previous format
	jsonLine := `{"timestamp":"2025-01-19T10:30:00Z","level":"ERROR","message":"Test"}`
	format1 := parser.DetectFormat(jsonLine)
	if format1 != JSONFormat {
		t.Errorf("Expected JSON format, got %v", format1)
	}

	// Second detection of same format should be faster (using previous detector)
	jsonLine2 := `{"timestamp":"2025-01-19T10:30:01Z","level":"INFO","message":"Another test"}`
	start := time.Now()
	format2 := parser.DetectFormat(jsonLine2)
	duration := time.Since(start)

	if format2 != JSONFormat {
		t.Errorf("Expected JSON format, got %v", format2)
	}

	// Should be relatively fast since it uses the cached detector
	if duration > 10*time.Millisecond {
		t.Logf("Detection took %v, might be slower than expected", duration)
	}

	// Test fallback: different format should reset and detect correctly
	syslogLine := `Jan 19 10:30:00 hostname myapp[1234]: ERROR: Database connection failed`
	format3 := parser.DetectFormat(syslogLine)
	if format3 != SyslogFormat {
		t.Errorf("Expected Syslog format, got %v", format3)
	}

	// Verify the parser updated its previous format
	syslogLine2 := `Jan 19 10:31:00 hostname myapp[1234]: INFO: Connection restored`
	format4 := parser.DetectFormat(syslogLine2)
	if format4 != SyslogFormat {
		t.Errorf("Expected Syslog format, got %v", format4)
	}
}

func TestConcurrentDetection(t *testing.T) {
	parser := NewParser()

	// Test concurrent access to the parser
	lines := []string{
		`{"timestamp":"2025-01-19T10:30:00Z","level":"ERROR","message":"Test1"}`,
		`{"timestamp":"2025-01-19T10:30:01Z","level":"INFO","message":"Test2"}`,
		`{"timestamp":"2025-01-19T10:30:02Z","level":"WARN","message":"Test3"}`,
	}

	done := make(chan bool, len(lines))

	// Start multiple goroutines
	for i, line := range lines {
		go func(index int, l string) {
			format := parser.DetectFormat(l)
			if format != JSONFormat {
				t.Errorf("Goroutine %d: Expected JSON format, got %v", index, format)
			}
			done <- true
		}(i, line)
	}

	// Wait for all goroutines to complete
	for i := 0; i < len(lines); i++ {
		<-done
	}
}

func BenchmarkOptimizedParser(b *testing.B) {
	parser := NewParser()
	jsonLine := `{"timestamp":"2025-01-19T10:30:00Z","level":"ERROR","message":"Database connection failed","service":"api"}`

	// Prime the parser with first detection
	parser.DetectFormat(jsonLine)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parser.DetectFormat(jsonLine)
	}
}

func BenchmarkOriginalDetection(b *testing.B) {
	jsonLine := `{"timestamp":"2025-01-19T10:30:00Z","level":"ERROR","message":"Database connection failed","service":"api"}`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		DetectFormat(jsonLine) // Using the original function from formats.go
	}
}

func TestDeterministicDetection(t *testing.T) {
	parser := NewParser()

	// Test line that could potentially match multiple formats
	line := `{"timestamp":"2025-01-19T10:30:00Z","level":"ERROR","message":"Test"}`

	// Run detection multiple times - should always return the same result
	firstResult := parser.DetectFormat(line)
	for i := 0; i < 20; i++ {
		result := parser.DetectFormat(line)
		if result != firstResult {
			t.Errorf("Non-deterministic detection: first=%v, iteration %d=%v", firstResult, i, result)
		}
	}
}

func TestSpecificityOrdering(t *testing.T) {
	parser := NewParser()

	tests := []struct {
		name     string
		line     string
		expected LogFormat
		reason   string
	}{
		{
			name:     "JSON beats everything",
			line:     `{"timestamp":"2025-01-19T10:30:00Z","level":"ERROR"}`,
			expected: JSONFormat,
			reason:   "JSON has highest specificity (90)",
		},
		{
			name:     "Kubernetes beats Docker",
			line:     `2025-01-19T10:30:00.123Z 1 main.go:42] ERROR Database connection failed`,
			expected: KubernetesFormat,
			reason:   "Kubernetes (85) beats Docker (55) in specificity",
		},
		{
			name:     "Nginx beats Apache",
			line:     `127.0.0.1 - - [19/Jan/2025:10:30:00 +0000] "GET /api/users HTTP/1.1" 200 1234 "-" "Mozilla/5.0"`,
			expected: NginxFormat,
			reason:   "Nginx (75) beats Apache (70) in specificity",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test multiple times to ensure deterministic behavior
			for i := 0; i < 5; i++ {
				result := parser.DetectFormat(tt.line)
				if result != tt.expected {
					t.Errorf("Expected %v, got %v. Reason: %s", tt.expected, result, tt.reason)
				}
			}
		})
	}
}
