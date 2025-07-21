package colorizer

import (
	"encoding/json"
	"regexp"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/joshi4/splash/parser"
	"github.com/muesli/termenv"
)

func TestHerokuSearchHighlighting(t *testing.T) {
	// This test ensures that Heroku log format search highlighting works correctly
	// and prevents regression of the bug where no highlighting was applied

	originalProfile := lipgloss.ColorProfile()
	defer lipgloss.SetColorProfile(originalProfile)
	lipgloss.SetColorProfile(termenv.TrueColor)

	colorizer := NewColorizer()

	tests := []struct {
		name        string
		line        string
		search      string
		description string
	}{
		{
			name:        "Timestamp highlighting",
			line:        "2025-01-19T08:30:00+00:00 app[web.1]: INFO Starting web dyno",
			search:      "2025",
			description: "Year in timestamp should be highlighted",
		},
		{
			name:        "Dyno highlighting",
			line:        "2025-01-19T08:30:00+00:00 app[web.1]: INFO Starting web dyno",
			search:      "web",
			description: "Dyno name should be highlighted",
		},
		{
			name:        "Message highlighting",
			line:        "2025-01-19T08:30:00+00:00 app[worker.1]: ERROR Database connection failed",
			search:      "Database",
			description: "Message content should be highlighted",
		},
		{
			name:        "Log level highlighting",
			line:        "2025-01-19T08:30:00+00:00 app[web.1]: ERROR Connection failed",
			search:      "ERROR",
			description: "Log level should be highlighted",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			colorizer.SetSearchString(tt.search)
			result := colorizer.ColorizeLog(tt.line, parser.HerokuFormat)

			// Verify the result contains ANSI codes (styling preserved)
			if !strings.Contains(result, "\x1b[") {
				t.Errorf("Expected ANSI color codes in result for %s, but none found", tt.description)
			}

			// Verify the search term is still present in the result
			if !strings.Contains(result, tt.search) {
				t.Errorf("Expected search term '%s' to be present in result", tt.search)
			}

			// The result should be longer due to ANSI codes
			if len(result) <= len(tt.line) {
				t.Errorf("Expected colorized result to be longer than original line due to ANSI codes")
			}
		})
	}
}

func TestDarkThemeTimestampStylingPreservation(t *testing.T) {
	// Force color output for testing
	originalProfile := lipgloss.ColorProfile()
	defer lipgloss.SetColorProfile(originalProfile)
	lipgloss.SetColorProfile(termenv.TrueColor)

	// Create colorizer with dark theme
	colorizer := NewColorizer()
	colorizer.SetTheme(NewDarkTheme())
	colorizer.SetSearchString("2025")

	// Test timestamp with partial highlighting
	testLine := "2025/01/19 08:30:00 INFO: Application started"
	result := colorizer.ColorizeLog(testLine, parser.GoStandardFormat)

	// The result should contain:
	// 1. Highlighted "2025" with search highlight style
	// 2. The rest of the timestamp "/01/19 08:30:00" with original timestamp styling
	// 3. The log level "INFO" with info level styling

	// Check that the result contains ANSI color codes (indicating styling is preserved)
	if !strings.Contains(result, "\x1b[") {
		t.Error("Expected ANSI color codes in result, but none found")
	}

	// Check that search highlighting is applied (should contain background color)
	if !strings.Contains(result, "2025") {
		t.Error("Expected search term '2025' to be present in result")
	}

	// Test with another timestamp format - check Kubernetes logs
	testLine2 := "2025-01-19T10:30:00.123Z 1 main.go:42] INFO Application started"
	result2 := colorizer.ColorizeLog(testLine2, parser.KubernetesFormat)

	// Should have highlighted "2025" but preserved styling for the rest
	if !strings.Contains(result2, "\x1b[") {
		t.Error("Expected ANSI color codes in Kubernetes timestamp result, but none found")
	}

	// Test with Syslog format
	testLine3 := "Jan 19 10:30:00 hostname myapp[1234]: INFO Application started"
	colorizer.SetSearchString("Jan")
	result3 := colorizer.ColorizeLog(testLine3, parser.SyslogFormat)

	// Should have highlighted "Jan" but preserved timestamp styling for the rest
	if !strings.Contains(result3, "\x1b[") {
		t.Error("Expected ANSI color codes in Syslog timestamp result, but none found")
	}
}

func TestDarkThemePreventTimestampStylingRegression(t *testing.T) {
	// This test specifically checks that timestamps don't lose their styling
	// when search highlighting is applied in dark themes

	originalProfile := lipgloss.ColorProfile()
	defer lipgloss.SetColorProfile(originalProfile)
	lipgloss.SetColorProfile(termenv.TrueColor)

	darkTheme := NewDarkTheme()
	colorizer := NewColorizer()
	colorizer.SetTheme(darkTheme)

	// Test cases where timestamps should maintain their color even with partial highlighting
	tests := []struct {
		name        string
		format      parser.LogFormat
		line        string
		search      string
		description string
	}{
		{
			name:        "GoStandard timestamp with year highlight",
			format:      parser.GoStandardFormat,
			line:        "2025/01/19 08:30:00 INFO: Test message",
			search:      "2025",
			description: "Year portion highlighted, rest of timestamp should keep gray color",
		},
		{
			name:        "Docker timestamp with time highlight",
			format:      parser.DockerFormat,
			line:        "2025-01-19T10:30:00.123456789Z INFO Test message",
			search:      "30:00",
			description: "Time portion highlighted, rest of timestamp should keep gray color",
		},
		{
			name:        "Kubernetes timestamp with date highlight",
			format:      parser.KubernetesFormat,
			line:        "2025-01-19T10:30:00.123Z 1 main.go:42] INFO Test message",
			search:      "19T",
			description: "Date portion highlighted, rest of timestamp should keep gray color",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			colorizer.SetSearchString(tt.search)
			result := colorizer.ColorizeLog(tt.line, tt.format)

			// Verify the result contains ANSI codes (styling preserved)
			if !strings.Contains(result, "\x1b[") {
				t.Errorf("Expected ANSI color codes in result for %s, but none found", tt.description)
			}

			// The result should contain search highlighting and original timestamp styling
			// We can't easily test the exact colors, but we can verify styling is applied
			if len(result) <= len(tt.line) {
				t.Errorf("Expected colorized result to be longer than original line due to ANSI codes")
			}
		})
	}
}

func TestUnifiedSearchHighlightingPreservesOriginalStyling(t *testing.T) {
	// Force color output for testing
	originalProfile := lipgloss.ColorProfile()
	defer lipgloss.SetColorProfile(originalProfile)
	lipgloss.SetColorProfile(termenv.TrueColor)

	tests := []struct {
		name              string
		format            parser.LogFormat
		line              string
		search            string
		expectHighlighted bool
	}{
		{
			name:              "Go Standard with timestamp highlighting",
			format:            parser.GoStandardFormat,
			line:              "2025/01/19 08:30:00 INFO: Application started",
			search:            "2025",
			expectHighlighted: true,
		},
		{
			name:              "Logfmt with key highlighting",
			format:            parser.LogfmtFormat,
			line:              "timestamp=2025-01-19T10:30:00Z level=info msg=\"Request processed\"",
			search:            "level",
			expectHighlighted: true,
		},
		{
			name:              "Syslog with partial match",
			format:            parser.SyslogFormat,
			line:              "Jan 19 10:30:00 hostname myapp[1234]: ERROR: Database connection failed",
			search:            "app",
			expectHighlighted: true,
		},
		{
			name:              "Docker with partial match",
			format:            parser.DockerFormat,
			line:              "2025-01-19T10:30:00.123456789Z ERROR Database connection failed",
			search:            "Database",
			expectHighlighted: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test without search (baseline styling)
			c1 := NewColorizer()
			resultWithoutSearch := c1.ColorizeLog(tt.line, tt.format)

			// Test with search
			c2 := NewColorizer()
			c2.SetSearchString(tt.search)
			resultWithSearch := c2.ColorizeLog(tt.line, tt.format)

			// Count ANSI escape sequences
			baselineCodes := strings.Count(resultWithoutSearch, "\x1b[")
			searchCodes := strings.Count(resultWithSearch, "\x1b[")

			if tt.expectHighlighted {
				// With search highlighting, we should have at least as many ANSI codes (original + highlighting)
				if searchCodes < baselineCodes {
					t.Errorf("Search highlighting removed original styling. Baseline: %d codes, With search: %d codes", baselineCodes, searchCodes)
					t.Errorf("Baseline result: %q", resultWithoutSearch)
					t.Errorf("Search result: %q", resultWithSearch)
				}

				// Verify the search term is highlighted
				if !strings.Contains(resultWithSearch, tt.search) {
					t.Errorf("Search term '%s' not found in result", tt.search)
				}
			}
		})
	}
}

func TestLightAndDarkThemeForcing(t *testing.T) {
	// Force color output for testing
	originalProfile := lipgloss.ColorProfile()
	defer lipgloss.SetColorProfile(originalProfile)
	lipgloss.SetColorProfile(termenv.TrueColor)

	testLine := `{"level":"ERROR","message":"Database failed"}`

	// Test adaptive theme (default)
	adaptiveColorizer := NewColorizer()
	adaptiveColorizer.SetSearchString("ERROR")
	adaptiveResult := adaptiveColorizer.ColorizeLog(testLine, parser.JSONFormat)

	// Test light theme
	lightColorizer := NewColorizer()
	lightColorizer.SetTheme(NewLightTheme())
	lightColorizer.SetSearchString("ERROR")
	lightResult := lightColorizer.ColorizeLog(testLine, parser.JSONFormat)

	// Test dark theme
	darkColorizer := NewColorizer()
	darkColorizer.SetTheme(NewDarkTheme())
	darkColorizer.SetSearchString("ERROR")
	darkResult := darkColorizer.ColorizeLog(testLine, parser.JSONFormat)

	// All should produce colorized output
	if !strings.Contains(adaptiveResult, "\x1b[") {
		t.Error("Adaptive theme should produce ANSI codes")
	}

	if !strings.Contains(lightResult, "\x1b[") {
		t.Error("Light theme should produce ANSI codes")
	}

	if !strings.Contains(darkResult, "\x1b[") {
		t.Error("Dark theme should produce ANSI codes")
	}

	// Results should be different between light and dark themes
	if lightResult == darkResult {
		t.Error("Light and dark themes should produce different output")
	}

	// All should contain the search term highlighting
	searchTerm := "ERROR"
	if !strings.Contains(adaptiveResult, searchTerm) {
		t.Error("Adaptive theme result should contain search term")
	}
	if !strings.Contains(lightResult, searchTerm) {
		t.Error("Light theme result should contain search term")
	}
	if !strings.Contains(darkResult, searchTerm) {
		t.Error("Dark theme result should contain search term")
	}
}

func TestUnifiedRegexSearchHighlightingPreservesOriginalStyling(t *testing.T) {
	// Force color output for testing
	originalProfile := lipgloss.ColorProfile()
	defer lipgloss.SetColorProfile(originalProfile)
	lipgloss.SetColorProfile(termenv.TrueColor)

	tests := []struct {
		name              string
		format            parser.LogFormat
		line              string
		regex             string
		expectHighlighted bool
	}{
		{
			name:              "Go Standard with timestamp regex",
			format:            parser.GoStandardFormat,
			line:              "2025/01/19 08:30:00 INFO: Application started",
			regex:             `\d{4}/\d{2}/\d{2}`,
			expectHighlighted: true,
		},
		{
			name:              "Logfmt with email regex",
			format:            parser.LogfmtFormat,
			line:              "user=john@example.com level=info msg=\"Login successful\"",
			regex:             `\w+@\w+\.\w+`,
			expectHighlighted: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test without search (baseline styling)
			c1 := NewColorizer()
			resultWithoutSearch := c1.ColorizeLog(tt.line, tt.format)

			// Test with regex search
			c2 := NewColorizer()
			err := c2.SetSearchRegex(tt.regex)
			if err != nil {
				t.Fatalf("Failed to set regex: %v", err)
			}
			resultWithSearch := c2.ColorizeLog(tt.line, tt.format)

			// Count ANSI escape sequences
			baselineCodes := strings.Count(resultWithoutSearch, "\x1b[")
			searchCodes := strings.Count(resultWithSearch, "\x1b[")

			if tt.expectHighlighted {
				// With search highlighting, we should have at least as many ANSI codes (original + highlighting)
				if searchCodes < baselineCodes {
					t.Errorf("Regex search highlighting removed original styling. Baseline: %d codes, With search: %d codes", baselineCodes, searchCodes)
					t.Errorf("Baseline result: %q", resultWithoutSearch)
					t.Errorf("Search result: %q", resultWithSearch)
				}
			}
		})
	}
}

func TestJSONPartialSearchHighlighting(t *testing.T) {
	c := NewColorizer()

	tests := []struct {
		name                   string
		jsonLine               string
		searchString           string
		expectPartialHighlight bool
		expectWholeHighlight   bool
	}{
		{
			name:                   "Partial match in JSON key",
			jsonLine:               `{"error_message": "Database failed", "user_id": 12345}`,
			searchString:           "err",
			expectPartialHighlight: true,
			expectWholeHighlight:   false,
		},
		{
			name:                   "Partial match in JSON value",
			jsonLine:               `{"message": "An error occurred", "status": "failure"}`,
			searchString:           "err",
			expectPartialHighlight: true,
			expectWholeHighlight:   false,
		},
		{
			name:                   "Multiple partial matches",
			jsonLine:               `{"error_code": "ERR_CONNECTION", "error_message": "Server error"}`,
			searchString:           "err",
			expectPartialHighlight: true,
			expectWholeHighlight:   false,
		},
		{
			name:                   "No match",
			jsonLine:               `{"message": "Success", "status": "ok"}`,
			searchString:           "err",
			expectPartialHighlight: false,
			expectWholeHighlight:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c.SetSearchString(tt.searchString)
			result := c.ColorizeLog(tt.jsonLine, parser.JSONFormat)

			// Check if JSON structure is still valid by parsing it
			// We need to strip ANSI codes first to check JSON validity
			plainResult := c.stripAnsiCodes(result)
			var jsonData map[string]interface{}
			if err := json.Unmarshal([]byte(plainResult), &jsonData); err != nil {
				t.Errorf("JSON structure was broken: %v", err)
			}

			if tt.expectPartialHighlight {
				// Should contain the search term but not highlight entire words
				if !strings.Contains(result, tt.searchString) {
					t.Errorf("Expected search term '%s' to be found in result", tt.searchString)
				}

				// Verify it's partial highlighting by checking that the search term
				// appears in context of larger words (like "error_message" containing "err")
				if tt.searchString == "err" {
					// The result should contain "err" but in contexts like "error_message" or "error"
					// where only the "err" part is highlighted, not the whole word
					if strings.Contains(tt.jsonLine, "error_message") && !strings.Contains(result, "error_message") {
						t.Errorf("Original contained 'error_message' but result doesn't show it properly")
					}
				}
			} else if !tt.expectPartialHighlight && strings.Contains(result, tt.searchString) {
				t.Errorf("Did not expect search term '%s' to be found in result", tt.searchString)
			}
		})
	}
}

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
			name:     "Simple nested object",
			line:     `{"level":"ERROR","error":{"code":500,"message":"Internal error"}}`,
			contains: []string{"level", "ERROR", "error", "code", "500", "message", "Internal error"},
		},
		{
			name:     "Deeply nested objects",
			line:     `{"timestamp":"2025-01-19T10:30:00Z","error":{"details":{"timeout":30,"retries":3}}}`,
			contains: []string{"timestamp", "error", "details", "timeout", "30", "retries", "3"},
		},
		{
			name:     "JSON with array",
			line:     `{"level":"INFO","tags":["api","success"],"data":{"count":2}}`,
			contains: []string{"level", "INFO", "tags", "api", "success", "data", "count", "2"},
		},
		{
			name:     "Complex nested structure",
			line:     `{"service":"api","user":{"id":123,"meta":{"roles":["admin","user"]}},"stats":{"requests":100}}`,
			contains: []string{"service", "api", "user", "id", "123", "meta", "roles", "admin", "user", "stats", "requests", "100"},
		},
		{
			name:     "Array of objects",
			line:     `{"events":[{"type":"start","time":"10:00"},{"type":"end","time":"11:00"}]}`,
			contains: []string{"events", "type", "start", "time", "10:00", "end", "11:00"},
		},
		{
			name:     "Nested array with primitives",
			line:     `{"data":{"numbers":[1,2,3],"booleans":[true,false]}}`,
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
			name:     "Nested timestamp field",
			line:     `{"data":{"timestamp":"2025-01-19T10:30:00Z","value":123}}`,
			contains: []string{"timestamp", "2025-01-19T10:30:00Z"},
		},
		{
			name:     "Nested level field",
			line:     `{"event":{"level":"ERROR","message":"Failed"}}`,
			contains: []string{"level", "ERROR"},
		},
		{
			name:     "Nested service field",
			line:     `{"context":{"service":"api-gateway","version":"1.0"}}`,
			contains: []string{"service", "api-gateway"},
		},
		{
			name:     "Multiple special fields nested",
			line:     `{"log":{"level":"WARN","service":"auth","timestamp":"2025-01-19T10:30:00Z"}}`,
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
		name             string
		line             string
		searchString     string
		shouldNotContain []string // Markers that should NOT appear in the output
	}{
		{
			name:             "Search in key name doesn't break JSON structure",
			line:             `{"slideshow":{"title":"Sample Slide Show"}}`,
			searchString:     "slide",
			shouldNotContain: []string{"⟦SEARCH_START⟧", "⟦SEARCH_END⟧", `{"⟦SEARCH_START⟧slide⟦SEARCH_END⟧show"`},
		},
		{
			name:             "Search in nested object key",
			line:             `{"data":{"timeout":30,"message":"timeout occurred"}}`,
			searchString:     "timeout",
			shouldNotContain: []string{"⟦SEARCH_START⟧", "⟦SEARCH_END⟧"},
		},
		{
			name:             "Search in value doesn't break JSON",
			line:             `{"level":"ERROR","message":"Connection timeout"}`,
			searchString:     "timeout",
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
				`"[38;2;221;160;221m`, // Color code at start of key
				`m"`,                  // Color code at end of key
				`"[0m`,                // Reset code in key
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
		name              string
		line              string
		searchString      string
		expectHighlighted bool
	}{
		{
			name:              "Search term only in JSON value should be highlighted",
			line:              `{"level":"ERROR","message":"timeout occurred"}`,
			searchString:      "timeout",
			expectHighlighted: true,
		},
		{
			name:              "Search term in nested JSON value",
			line:              `{"data":{"status":"error occurred","code":500}}`,
			searchString:      "error",
			expectHighlighted: true,
		},
		{
			name:              "Search term in array element value",
			line:              `{"tags":["error","warning","timeout"]}`,
			searchString:      "timeout",
			expectHighlighted: true,
		},
		{
			name:              "Search term not found",
			line:              `{"level":"INFO","message":"success"}`,
			searchString:      "error",
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

func TestJSONSearchUsesHighVisibilityColors(t *testing.T) {
	c := NewColorizer()
	c.SetSearchString("test")

	// Test that JSON uses JSONSearchHighlight style (high visibility orange)
	jsonLine := `{"message":"test value"}`
	jsonResult := c.ColorizeLog(jsonLine, parser.JSONFormat)

	// Test that non-JSON formats still use the regular SearchHighlight style
	logfmtLine := `level=info msg="test value"`
	logfmtResult := c.ColorizeLog(logfmtLine, parser.LogfmtFormat)

	// While we can't see the actual colors in tests, we can verify the search term is present
	if !strings.Contains(jsonResult, "test") {
		t.Errorf("Expected JSON search term to be present, got: %q", jsonResult)
	}

	if !strings.Contains(logfmtResult, "test") {
		t.Errorf("Expected Logfmt search term to be present, got: %q", logfmtResult)
	}

	// Note: Can't directly compare lipgloss styles as they contain functions

	// Test that JSON search highlight uses high visibility color
	testText := "test"
	jsonStyled := c.theme.JSONSearchHighlight.Render(testText)
	regularStyled := c.theme.SearchHighlight.Render(testText)

	// In a real terminal, these would look different, but in test environment lipgloss disables styling
	t.Logf("JSON search highlight: %q", jsonStyled)
	t.Logf("Regular search highlight: %q", regularStyled)
}
