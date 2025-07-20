package colorizer

import (
	"encoding/json"
	"regexp"
	"strings"
	"testing"

	"github.com/joshi4/splash/parser"
)

// stripTestAnsiCodes removes ANSI escape codes for testing purposes
func stripTestAnsiCodes(text string) string {
	ansiRegex := regexp.MustCompile(`\x1b\[[0-9;]*m`)
	return ansiRegex.ReplaceAllString(text, "")
}

func TestSearchFunctionality(t *testing.T) {
	c := NewColorizer()

	tests := []struct {
		name        string
		line        string
		searchType  string // "string" or "regexp"
		pattern     string
		shouldMatch bool
		description string
	}{
		{
			name:        "String search match",
			line:        `{"level":"ERROR","message":"Database failed"}`,
			searchType:  "string",
			pattern:     "ERROR",
			shouldMatch: true,
			description: "Should match exact string",
		},
		{
			name:        "String search no match",
			line:        `{"level":"INFO","message":"All good"}`,
			searchType:  "string",
			pattern:     "ERROR",
			shouldMatch: false,
			description: "Should not match when string not present",
		},
		{
			name:        "Regexp search match",
			line:        `127.0.0.1 - - [19/Jan/2025:08:30:00 +0000] "GET / HTTP/1.1" 404 1234`,
			searchType:  "regexp",
			pattern:     `HTTP/1\.1" [45]\d\d`,
			shouldMatch: true,
			description: "Should match regex pattern for 4xx/5xx status codes",
		},
		{
			name:        "Regexp search no match",
			line:        `127.0.0.1 - - [19/Jan/2025:08:30:00 +0000] "GET / HTTP/1.1" 200 1234`,
			searchType:  "regexp",
			pattern:     `HTTP/1\.1" [45]\d\d`,
			shouldMatch: false,
			description: "Should not match 2xx status codes",
		},
		{
			name:        "Case sensitive string search",
			line:        `timestamp=2025-01-19T08:30:00Z level=error msg="test"`,
			searchType:  "string",
			pattern:     "ERROR",
			shouldMatch: false,
			description: "String search should be case sensitive",
		},
		{
			name:        "Regexp case insensitive",
			line:        `timestamp=2025-01-19T08:30:00Z level=error msg="test"`,
			searchType:  "regexp",
			pattern:     `(?i)ERROR`,
			shouldMatch: true,
			description: "Regexp can be case insensitive with flags",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup search pattern
			if tt.searchType == "string" {
				c.SetSearchString(tt.pattern)
			} else {
				err := c.SetSearchRegex(tt.pattern)
				if err != nil {
					t.Fatalf("Failed to set regex pattern: %v", err)
				}
			}

			// Test if line matches
			matches := c.matchesSearch(tt.line)
			if matches != tt.shouldMatch {
				t.Errorf("Expected match=%v, got match=%v. %s", tt.shouldMatch, matches, tt.description)
			}

			// Test colorization with search
			result := c.ColorizeLog(tt.line, parser.JSONFormat)
			
			if tt.shouldMatch {
				// For regex patterns, we need to verify the match differently
				// since the result contains the actual matched text, not the pattern
				if tt.searchType == "regexp" {
					// Use the same regex to verify the match exists in the result
					regex, err := regexp.Compile(tt.pattern)
					if err != nil {
						t.Fatalf("Invalid regex pattern in test: %v", err)
					}
					if !regex.MatchString(result) {
						t.Errorf("Expected highlighted result to match regex %q, got: %q", tt.pattern, result)
					}
				} else {
					// For string patterns, check that the exact string is present
					if !strings.Contains(result, tt.pattern) {
						t.Errorf("Expected highlighted result to contain pattern %q, got: %q", tt.pattern, result)
					}
				}
			} else {
				// For non-matches, result should still be valid but pattern shouldn't be specially highlighted
				// We'll just verify it doesn't crash and produces output
				if result == "" {
					t.Errorf("Expected non-empty result for non-matching line")
				}
			}

			// Clear search for next test
			c.SetSearchString("")
		})
	}
}

func TestSearchHighlightingWithDifferentFormats(t *testing.T) {
	c := NewColorizer()
	c.SetSearchString("ERROR")

	tests := []struct {
		format parser.LogFormat
		line   string
	}{
		{parser.JSONFormat, `{"level":"ERROR","message":"test"}`},
		{parser.LogfmtFormat, `level=ERROR msg="test"`},
		{parser.SyslogFormat, `Jan 19 08:30:00 host app[123]: ERROR: test`},
		{parser.GoStandardFormat, `2025/01/19 08:30:00 ERROR: test`},
	}

	for _, tt := range tests {
		t.Run(tt.format.String(), func(t *testing.T) {
			result := c.ColorizeLog(tt.line, tt.format)
			
			// Should contain "ERROR" since all test lines contain it
			if !strings.Contains(result, "ERROR") {
				t.Errorf("Expected result to contain 'ERROR' for format %v, got: %q", tt.format, result)
			}
		})
	}
}

func TestInvalidRegex(t *testing.T) {
	c := NewColorizer()
	
	// Test invalid regex pattern
	err := c.SetSearchRegex("[invalid regex")
	if err == nil {
		t.Error("Expected error for invalid regex pattern")
	}
}

func TestNoSearchPattern(t *testing.T) {
	c := NewColorizer()
	
	line := `{"level":"ERROR","message":"test"}`
	
	// Should not match when no search pattern is set
	if c.matchesSearch(line) {
		t.Error("Expected no match when no search pattern is set")
	}
	
	// Colorization should work normally without highlighting
	result := c.ColorizeLog(line, parser.JSONFormat)
	if strings.Contains(result, ">>>") || strings.Contains(result, "<<<") {
		t.Errorf("Expected no highlighting when no search pattern set, got: %q", result)
	}
}

// TestHerokuSearchMarkersNotLeftInOutput verifies that Heroku format search highlighting
// properly replaces markers with colors and doesn't leave search markers in output
func TestHerokuSearchMarkersNotLeftInOutput(t *testing.T) {
	c := NewColorizer()
	c.SetSearchString("FATAL")
	
	tests := []struct {
		name string
		line string
	}{
		{
			name: "Search in log level should not leave markers",
			line: "2025-01-19T08:32:00+00:00 app[worker.1]: FATAL Job processing failed",
		},
		{
			name: "Search in message part should not leave markers", 
			line: "2025-01-19T08:32:00+00:00 app[web.1]: ERROR FATAL connection timeout",
		},
		{
			name: "Search in dyno name should not leave markers",
			line: "2025-01-19T08:32:00+00:00 app[FATAL.1]: INFO Application started",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := c.ColorizeLog(tt.line, parser.HerokuFormat)
			
			// Check that no search markers are left in the output
			if strings.Contains(result, "⟦SEARCH_START⟧") || strings.Contains(result, "⟦SEARCH_END⟧") {
				t.Errorf("Search markers found in output for line: %s\nResult: %s", tt.line, result)
			}
			
			// If the search term is found in input, the output should contain it too
			if strings.Contains(tt.line, "FATAL") {
				if !strings.Contains(result, "FATAL") {
					t.Errorf("Search term 'FATAL' missing from result for line: %s\nResult: %s", tt.line, result)
				}
			}
		})
	}
}

// TestJSONSinglePassSearchHighlighting verifies that JSON search highlighting works in a single pass
func TestJSONSinglePassSearchHighlighting(t *testing.T) {
	c := NewColorizer()
	
	tests := []struct {
		name       string
		line       string
		searchTerm string
		expectMatch bool
	}{
		{
			name:        "Value highlighting in single pass",
			line:        `{"level":"ERROR","message":"Database failed"}`,
			searchTerm:  "ERROR",
			expectMatch: true,
		},
		{
			name:        "Key highlighting in single pass", 
			line:        `{"error_code":"E123","status":"ok"}`,
			searchTerm:  "error",
			expectMatch: true,
		},
		{
			name:        "Nested value highlighting",
			line:        `{"data":{"status":"ERROR","user":"admin"}}`,
			searchTerm:  "ERROR",
			expectMatch: true,
		},
		{
			name:        "Nested key highlighting",
			line:        `{"data":{"error_count":5,"user":"admin"}}`,
			searchTerm:  "error",
			expectMatch: true,
		},
		{
			name:        "No match should not highlight",
			line:        `{"level":"INFO","message":"All good"}`,
			searchTerm:  "ERROR",
			expectMatch: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c.SetSearchString(tt.searchTerm)
			result := c.ColorizeLog(tt.line, parser.JSONFormat)
			
			// The result should contain the search term
			if !strings.Contains(result, tt.searchTerm) && tt.expectMatch {
				t.Errorf("Expected search term '%s' to be present in result for line: %s\nResult: %s", 
					tt.searchTerm, tt.line, result)
			}
			
			// Should not contain search markers
			if strings.Contains(result, "⟦SEARCH_START⟧") || strings.Contains(result, "⟦SEARCH_END⟧") {
				t.Errorf("Found search markers in single-pass result for line: %s\nResult: %s", tt.line, result)
			}
			
			// JSON structure should remain valid (can be parsed after removing ANSI codes)
			stripped := stripTestAnsiCodes(result)
			var testData interface{}
			if err := json.Unmarshal([]byte(stripped), &testData); err != nil {
				t.Errorf("Result is not valid JSON after stripping ANSI codes: %v\nResult: %s\nStripped: %s", 
					err, result, stripped)
			}
		})
	}
}

// TestJSONSearchRegexSinglePass verifies that regex search works in single pass for JSON
func TestJSONSearchRegexSinglePass(t *testing.T) {
	c := NewColorizer()
	c.SetSearchRegex("ERR.*")
	
	line := `{"level":"ERROR","message":"Everything is working"}`
	result := c.ColorizeLog(line, parser.JSONFormat)
	
	// Should contain the matched pattern
	if !strings.Contains(result, "ERROR") {
		t.Errorf("Expected regex match 'ERROR' to be present in result\nResult: %s", result)
	}
	
	// Should not contain search markers
	if strings.Contains(result, "⟦SEARCH_START⟧") || strings.Contains(result, "⟦SEARCH_END⟧") {
		t.Errorf("Found search markers in single-pass regex result\nResult: %s", result)
	}
	
	// JSON should remain valid
	stripped := stripTestAnsiCodes(result)
	var testData interface{}
	if err := json.Unmarshal([]byte(stripped), &testData); err != nil {
		t.Errorf("Result is not valid JSON after regex highlighting: %v\nStripped: %s", err, stripped)
	}
}

func TestSearchPatternSwitching(t *testing.T) {
	c := NewColorizer()
	
	line := `{"level":"ERROR","message":"database connection failed"}`
	
	// Set string search
	c.SetSearchString("ERROR")
	if !c.matchesSearch(line) {
		t.Error("Expected match with string search")
	}
	
	// Switch to regexp search - should clear string search
	err := c.SetSearchRegex("database.*failed")
	if err != nil {
		t.Fatalf("Failed to set regexp: %v", err)
	}
	
	if !c.matchesSearch(line) {
		t.Error("Expected match with regexp search")
	}
	
	// Switch back to string search - should clear regexp
	c.SetSearchString("connection")
	if !c.matchesSearch(line) {
		t.Error("Expected match with new string search")
	}
}
