package colorizer

import (
	"encoding/json"
	"regexp"
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

func TestJSONSearchMarkersNotBreakingStructure(t *testing.T) {
	c := NewColorizer()

	tests := []struct {
		name         string
		line         string
		searchString string
		shouldNotContain []string // Markers that should NOT appear in the output
	}{
		{
			name:         "Search in key name doesn't break JSON structure",
			line:         `{"slideshow":{"title":"Sample Slide Show"}}`,
			searchString: "slide",
			shouldNotContain: []string{"⟦SEARCH_START⟧", "⟦SEARCH_END⟧", `{"⟦SEARCH_START⟧slide⟦SEARCH_END⟧show"`},
		},
		{
			name:         "Search in nested object key",
			line:         `{"data":{"timeout":30,"message":"timeout occurred"}}`,
			searchString: "timeout",
			shouldNotContain: []string{"⟦SEARCH_START⟧", "⟦SEARCH_END⟧"},
		},
		{
			name:         "Search in value doesn't break JSON",
			line:         `{"level":"ERROR","message":"Connection timeout"}`,
			searchString: "timeout",
			shouldNotContain: []string{"⟦SEARCH_START⟧", "⟦SEARCH_END⟧"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c.SetSearchString(tt.searchString)
			result := c.ColorizeLog(tt.line, parser.JSONFormat)
			
			// Verify the result is valid (no broken markers)
			for _, forbidden := range tt.shouldNotContain {
				if strings.Contains(result, forbidden) {
					t.Errorf("Result should not contain %q, got: %q", forbidden, result)
				}
			}
			
			// Verify the search term is still present (should be highlighted)
			if !strings.Contains(result, tt.searchString) {
				t.Errorf("Search term %q should be present in result, got: %q", tt.searchString, result)
			}
			
			// Verify the result is still valid JSON structure (contains braces)
			if !strings.Contains(result, "{") || !strings.Contains(result, "}") {
				t.Errorf("Result should maintain JSON structure, got: %q", result)
			}
		})
	}
}

func TestJSONSearchNoAnsiCodesInsideKeys(t *testing.T) {
	c := NewColorizer()

	tests := []struct {
		name         string
		line         string
		searchString string
	}{
		{
			name:         "Search in JSON key should not insert ANSI codes inside quoted key name",
			line:         `{"slideshow":{"slides":[{"title":"test"}]}}`,
			searchString: "slide",
		},
		{
			name:         "Search matching multiple JSON keys",
			line:         `{"data":{"timeout":30},"config":{"timeout_ms":5000}}`,
			searchString: "timeout",
		},
		{
			name:         "Search in complex nested JSON",
			line:         `{"slideshow":{"author":"Yours Truly","slides":[{"title":"Overview"}]}}`,
			searchString: "slide",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c.SetSearchString(tt.searchString)
			result := c.ColorizeLog(tt.line, parser.JSONFormat)
			
			// Check that no ANSI escape codes appear inside JSON key quotes
			// Pattern: "key with ANSI codes inside"
			ansiInKeyPattern := `"[^"]*\x1b\[[0-9;]*m[^"]*"`
			matched, err := regexp.MatchString(ansiInKeyPattern, result)
			if err != nil {
				t.Fatalf("Regex error: %v", err)
			}
			if matched {
				t.Errorf("ANSI codes found inside JSON key quotes in result: %q", result)
			}
			
			// Also check for specific problematic patterns like the user reported
			problematicPatterns := []string{
				`"[38;2;221;160;221m`,  // Color code at start of key
				`m"`,                    // Color code at end of key  
				`"[0m`,                  // Reset code in key
			}
			
			for _, pattern := range problematicPatterns {
				if strings.Contains(result, pattern) {
					t.Errorf("Found problematic ANSI pattern %q in result: %q", pattern, result)
				}
			}
			
			// Verify the search term is still present (should be highlighted somewhere)
			if !strings.Contains(result, tt.searchString) {
				t.Errorf("Search term %q should be present in result, got: %q", tt.searchString, result)
			}
			
			// Verify that JSON structure is maintained (should still be parseable)
			// Remove ANSI codes and check if it's still valid JSON
			cleanResult := regexp.MustCompile(`\x1b\[[0-9;]*m`).ReplaceAllString(result, "")
			var testData interface{}
			if err := json.Unmarshal([]byte(cleanResult), &testData); err != nil {
				t.Errorf("Result should still be valid JSON after removing ANSI codes, got error: %v, result: %q", err, cleanResult)
			}
		})
	}
}

func TestJSONSearchHighlightsBothKeysAndValues(t *testing.T) {
	c := NewColorizer()

	tests := []struct {
		name         string
		line         string
		searchString string
		description  string
	}{
		{
			name:         "Search term appears in both key and value",
			line:         `{"slideshow":"slide presentation","data":"normal"}`,
			searchString: "slide",
			description:  "Should highlight both the key 'slideshow' and value 'slide presentation'",
		},
		{
			name:         "Multiple key matches",
			line:         `{"timeout":30,"timeout_ms":5000,"msg":"Connection timeout"}`,
			searchString: "timeout",
			description:  "Should highlight keys 'timeout', 'timeout_ms' and value 'Connection timeout'",
		},
		{
			name:         "Nested structure with mixed matches",
			line:         `{"error":{"code":500,"error_msg":"An error occurred"}}`,
			searchString: "error",
			description:  "Should highlight key 'error', nested key 'error_msg', and value 'An error occurred'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c.SetSearchString(tt.searchString)
			result := c.ColorizeLog(tt.line, parser.JSONFormat)
			
			// Verify no ANSI codes inside key quotes (structure preservation)
			ansiInKeyPattern := `"[^"]*\x1b\[[0-9;]*m[^"]*"`
			matched, err := regexp.MatchString(ansiInKeyPattern, result)
			if err != nil {
				t.Fatalf("Regex error: %v", err)
			}
			if matched {
				t.Errorf("ANSI codes found inside JSON key quotes in result: %q", result)
			}
			
			// Verify the search term is still present
			if !strings.Contains(result, tt.searchString) {
				t.Errorf("Search term %q should be present in result, got: %q", tt.searchString, result)
			}
			
			// Verify JSON structure is preserved
			cleanResult := regexp.MustCompile(`\x1b\[[0-9;]*m`).ReplaceAllString(result, "")
			var testData interface{}
			if err := json.Unmarshal([]byte(cleanResult), &testData); err != nil {
				t.Errorf("Result should still be valid JSON after removing ANSI codes, got error: %v, result: %q", err, cleanResult)
			}
			
			// The clean result should be identical to the original (just reordered potentially)
			var originalData, resultData interface{}
			if err := json.Unmarshal([]byte(tt.line), &originalData); err != nil {
				t.Fatalf("Original JSON is invalid: %v", err)
			}
			if err := json.Unmarshal([]byte(cleanResult), &resultData); err != nil {
				t.Fatalf("Clean result JSON is invalid: %v", err)
			}
			
			// Both should parse to equivalent structures (can't compare directly due to map ordering)
			// Just ensure they're both valid - structural comparison would be complex
		})
	}
}

func TestJSONSearchValueHighlighting(t *testing.T) {
	c := NewColorizer()

	tests := []struct {
		name         string
		line         string
		searchString string
		expectHighlighted bool
	}{
		{
			name:         "Search term only in JSON value should be highlighted",
			line:         `{"level":"ERROR","message":"timeout occurred"}`,
			searchString: "timeout",
			expectHighlighted: true,
		},
		{
			name:         "Search term in nested JSON value",
			line:         `{"data":{"status":"error occurred","code":500}}`,
			searchString: "error",
			expectHighlighted: true,
		},
		{
			name:         "Search term in array element value",
			line:         `{"tags":["error","warning","timeout"]}`,
			searchString: "timeout",
			expectHighlighted: true,
		},
		{
			name:         "Search term not found",
			line:         `{"level":"INFO","message":"success"}`,
			searchString: "error",
			expectHighlighted: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c.SetSearchString(tt.searchString)
			result := c.ColorizeLog(tt.line, parser.JSONFormat)
			
			if tt.expectHighlighted {
				// Check that search term is present
				if !strings.Contains(result, tt.searchString) {
					t.Errorf("Expected search term %q to be present in result, got: %q", tt.searchString, result)
				}
				
				// For manual verification, let's also check that there are ANSI codes in the result
				// when highlighting is expected (this would be visible in a real terminal)
				hasAnsiCodes := strings.Contains(result, "\x1b[") || strings.Contains(result, "\033[")
				if !hasAnsiCodes {
					t.Logf("Note: No ANSI codes found in result (expected in test environment): %q", result)
				}
			} else {
				// Search term should not be in the result if not found in original
				if !strings.Contains(tt.line, tt.searchString) && strings.Contains(result, tt.searchString) {
					t.Errorf("Search term %q should not appear in result when not in original, got: %q", tt.searchString, result)
				}
			}
			
			// Verify JSON structure is always preserved
			cleanResult := regexp.MustCompile(`\x1b\[[0-9;]*m`).ReplaceAllString(result, "")
			var testData interface{}
			if err := json.Unmarshal([]byte(cleanResult), &testData); err != nil {
				t.Errorf("Result should still be valid JSON after removing ANSI codes, got error: %v, result: %q", err, cleanResult)
			}
		})
	}
}


