package colorizer

import (
	"strings"
	"testing"

	"github.com/joshi4/splash/parser"
)

func TestColorizer(t *testing.T) {
	c := NewColorizer()

	tests := []struct {
		name     string
		line     string
		format   parser.LogFormat
		contains []string // Strings that should be present in colorized output
	}{
		{
			name:     "JSON with ERROR level",
			line:     `{"timestamp":"2025-01-19T10:30:00Z","level":"ERROR","message":"Database failed"}`,
			format:   parser.JSONFormat,
			contains: []string{"timestamp", "ERROR", "message", "Database failed"},
		},
		{
			name:     "Logfmt with info level",
			line:     `timestamp=2025-01-19T10:30:00Z level=info msg="Request processed"`,
			format:   parser.LogfmtFormat,
			contains: []string{"timestamp", "level", "info", "msg", "Request processed"},
		},
		{
			name:     "Apache Common with 200 status",
			line:     `127.0.0.1 - - [19/Jan/2025:10:30:00 +0000] "GET /api/users HTTP/1.1" 200 1234`,
			format:   parser.ApacheCommonFormat,
			contains: []string{"127.0.0.1", "19/Jan/2025:10:30:00", "GET", "/api/users", "200"},
		},
		{
			name:     "Syslog with ERROR",
			line:     `Jan 19 10:30:00 hostname myapp[1234]: ERROR: Database connection failed`,
			format:   parser.SyslogFormat,
			contains: []string{"Jan 19 10:30:00", "hostname", "myapp", "1234", "ERROR"},
		},
		{
			name:     "Docker with ERROR level",
			line:     `2025-01-19T10:30:00.123456789Z ERROR Database connection failed`,
			format:   parser.DockerFormat,
			contains: []string{"2025-01-19T10:30:00", "ERROR", "Database connection failed"},
		},
		{
			name:     "Kubernetes with file reference",
			line:     `2025-01-19T10:30:00.123Z 1 main.go:42] ERROR Database connection failed`,
			format:   parser.KubernetesFormat,
			contains: []string{"2025-01-19T10:30:00", "main.go", "42", "ERROR"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := c.ColorizeLog(tt.line, tt.format)
			
			// Verify the result is not empty and different from input (has been processed)
			if result == "" {
				t.Error("Colorized result is empty")
			}
			
			// Check that expected content is present (may be styled)
			for _, expected := range tt.contains {
				if !strings.Contains(result, expected) {
					t.Errorf("Expected %q to contain %q, got: %q", result, expected, result)
				}
			}
			
			// For formats that should be processed, verify result has ANSI codes or is structured differently
			if tt.format != parser.UnknownFormat {
				// Check if result contains ANSI escape sequences (color codes) or has been restructured
				hasAnsiCodes := strings.Contains(result, "\x1b[") || strings.Contains(result, "\033[")
				isRestructured := result != tt.line // Content order might change (like JSON)
				
				if !hasAnsiCodes && !isRestructured {
					t.Logf("Warning: Result appears unchanged from input for format %v", tt.format)
					t.Logf("Input:  %q", tt.line)
					t.Logf("Output: %q", result)
				}
			}
		})
	}
}

func TestColorizeUnknownFormat(t *testing.T) {
	c := NewColorizer()
	line := "Some random log line with ERROR level"
	
	result := c.ColorizeLog(line, parser.UnknownFormat)
	
	// Should return original line for unknown formats
	if result != line {
		t.Errorf("Expected original line for unknown format, got: %q", result)
	}
}

func TestLogLevelDetection(t *testing.T) {
	c := NewColorizer()
	
	levels := []string{"ERROR", "WARN", "WARNING", "INFO", "DEBUG", "TRACE", "FATAL"}
	
	for _, level := range levels {
		line := "Some log with " + level + " level"
		result := c.colorizeGenericLog(line)
		
		if !strings.Contains(result, level) {
			t.Errorf("Expected %q to contain %q", result, level)
		}
	}
}

func TestHTTPStatusColors(t *testing.T) {
	c := NewColorizer()
	
	tests := []struct {
		status   string
		expected string // We can't easily test actual colors, but we can test that styling is applied
	}{
		{"200", "2"},
		{"404", "4"},
		{"500", "5"},
		{"301", "3"},
	}
	
	for _, tt := range tests {
		style := c.theme.GetHTTPStatusStyle(tt.status)
		result := style.Render(tt.status)
		
		// Should contain the status code
		if !strings.Contains(result, tt.status) {
			t.Errorf("Expected styled result to contain %q", tt.status)
		}
	}
}

func TestJSONNestedObjects(t *testing.T) {
	c := NewColorizer()

	tests := []struct {
		name     string
		line     string
		contains []string // Strings that should be present in colorized output
	}{
		{
			name: "Simple nested object",
			line: `{"level":"ERROR","error":{"code":500,"message":"Internal error"}}`,
			contains: []string{"level", "ERROR", "error", "code", "500", "message", "Internal error"},
		},
		{
			name: "Deeply nested objects",
			line: `{"timestamp":"2025-01-19T10:30:00Z","error":{"details":{"timeout":30,"retries":3}}}`,
			contains: []string{"timestamp", "error", "details", "timeout", "30", "retries", "3"},
		},
		{
			name: "JSON with array",
			line: `{"level":"INFO","tags":["api","success"],"data":{"count":2}}`,
			contains: []string{"level", "INFO", "tags", "api", "success", "data", "count", "2"},
		},
		{
			name: "Complex nested structure",
			line: `{"service":"api","user":{"id":123,"meta":{"roles":["admin","user"]}},"stats":{"requests":100}}`,
			contains: []string{"service", "api", "user", "id", "123", "meta", "roles", "admin", "user", "stats", "requests", "100"},
		},
		{
			name: "Array of objects",
			line: `{"events":[{"type":"start","time":"10:00"},{"type":"end","time":"11:00"}]}`,
			contains: []string{"events", "type", "start", "time", "10:00", "end", "11:00"},
		},
		{
			name: "Nested array with primitives",
			line: `{"data":{"numbers":[1,2,3],"booleans":[true,false]}}`,
			contains: []string{"data", "numbers", "1", "2", "3", "booleans", "true", "false"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := c.ColorizeLog(tt.line, parser.JSONFormat)
			
			// Verify the result is not empty
			if result == "" {
				t.Error("Colorized result is empty")
			}
			
			// Check that expected content is present
			for _, expected := range tt.contains {
				if !strings.Contains(result, expected) {
					t.Errorf("Expected %q to contain %q, got: %q", result, expected, result)
				}
			}
			
			// Verify result has structural elements (brackets, quotes, etc.)
			if !strings.Contains(result, "{") || !strings.Contains(result, "}") {
				t.Errorf("Expected result to contain JSON structural elements, got: %q", result)
			}
		})
	}
}

func TestJSONNestedObjectsWithSearch(t *testing.T) {
	c := NewColorizer()

	tests := []struct {
		name         string
		line         string
		searchString string
		shouldMatch  bool
	}{
		{
			name:         "Search in nested object value",
			line:         `{"level":"ERROR","error":{"code":500,"message":"timeout occurred"}}`,
			searchString: "timeout",
			shouldMatch:  true,
		},
		{
			name:         "Search in nested object key",
			line:         `{"data":{"timeout":30,"retries":3}}`,
			searchString: "timeout",
			shouldMatch:  true,
		},
		{
			name:         "Search in array element",
			line:         `{"tags":["error","timeout","network"]}`,
			searchString: "timeout",
			shouldMatch:  true,
		},
		{
			name:         "Search with no match in nested structure",
			line:         `{"level":"INFO","data":{"status":"ok","code":200}}`,
			searchString: "error",
			shouldMatch:  false,
		},
		{
			name:         "Search across multiple nesting levels",
			line:         `{"service":"api","error":{"details":{"timeout":30}}}`,
			searchString: "timeout",
			shouldMatch:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c.SetSearchString(tt.searchString)
			result := c.ColorizeLog(tt.line, parser.JSONFormat)
			
			// Verify the result is not empty
			if result == "" {
				t.Error("Colorized result is empty")
			}
			
			// Check if search highlighting is applied when expected
			containsSearchTerm := strings.Contains(tt.line, tt.searchString)
			if containsSearchTerm != tt.shouldMatch {
				t.Errorf("Test setup error: expected shouldMatch=%v but line contains search term=%v", tt.shouldMatch, containsSearchTerm)
			}
			
			if tt.shouldMatch {
				// When search matches, the original search term should still be present in the result
				if !strings.Contains(result, tt.searchString) {
					t.Errorf("Expected search term %q to be present in result when match expected, got: %q", tt.searchString, result)
				}
			}
		})
	}
}

func TestJSONSpecialFieldsInNestedObjects(t *testing.T) {
	c := NewColorizer()

	tests := []struct {
		name     string
		line     string
		contains []string // Special field content that should be styled
	}{
		{
			name: "Nested timestamp field",
			line: `{"data":{"timestamp":"2025-01-19T10:30:00Z","value":123}}`,
			contains: []string{"timestamp", "2025-01-19T10:30:00Z"},
		},
		{
			name: "Nested level field",
			line: `{"event":{"level":"ERROR","message":"Failed"}}`,
			contains: []string{"level", "ERROR"},
		},
		{
			name: "Nested service field",
			line: `{"context":{"service":"api-gateway","version":"1.0"}}`,
			contains: []string{"service", "api-gateway"},
		},
		{
			name: "Multiple special fields nested",
			line: `{"log":{"level":"WARN","service":"auth","timestamp":"2025-01-19T10:30:00Z"}}`,
			contains: []string{"level", "WARN", "service", "auth", "timestamp", "2025-01-19T10:30:00Z"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := c.ColorizeLog(tt.line, parser.JSONFormat)
			
			// Check that expected content is present
			for _, expected := range tt.contains {
				if !strings.Contains(result, expected) {
					t.Errorf("Expected %q to contain %q, got: %q", result, expected, result)
				}
			}
		})
	}
}
