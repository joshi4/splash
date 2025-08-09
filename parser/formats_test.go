package parser

import "testing"

func TestDetectFormat(t *testing.T) {
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
		{
			name:     "Java exception header",
			line:     `Exception in thread "main" java.lang.ArithmeticException: / by zero`,
			expected: JavaExceptionFormat,
		},
		{
			name:     "Java caused by line",
			line:     `Caused by: java.lang.NullPointerException: Cannot invoke method`,
			expected: JavaExceptionFormat,
		},
		{
			name:     "Python traceback header",
			line:     `Traceback (most recent call last):`,
			expected: PythonExceptionFormat,
		},

		{
			name:     "Python exception line",
			line:     `ZeroDivisionError: division by zero`,
			expected: PythonExceptionFormat,
		},
		{
			name:     "Unknown format",
			line:     `Some random log line that doesn't match any pattern`,
			expected: UnknownFormat,
		},
		{
			name:     "Empty line",
			line:     ``,
			expected: UnknownFormat,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a fresh parser for each test to avoid stateful interference
			p := NewParser()
			result := p.DetectFormat(tt.line)
			if result != tt.expected {
				t.Errorf("DetectFormat(%q) = %v, expected %v", tt.line, result, tt.expected)
			}
		})
	}
}
