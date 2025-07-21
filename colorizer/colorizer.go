package colorizer

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/joshi4/splash/parser"
)

// Colorizer handles adding colors to log lines based on their format
type Colorizer struct {
	theme        *ColorTheme
	searchString string
	searchRegex  *regexp.Regexp
}

// NewColorizer creates a new colorizer with adaptive theming
func NewColorizer() *Colorizer {
	return &Colorizer{
		theme: NewAdaptiveTheme(),
	}
}

// SetTheme sets a custom theme for the colorizer
func (c *Colorizer) SetTheme(theme *ColorTheme) {
	c.theme = theme
}

// SetSearchString sets a literal string to search for and highlight
func (c *Colorizer) SetSearchString(pattern string) {
	c.searchString = pattern
	c.searchRegex = nil
}

// SetSearchRegex sets a regular expression to search for and highlight
func (c *Colorizer) SetSearchRegex(pattern string) error {
	regex, err := regexp.Compile(pattern)
	if err != nil {
		return err
	}
	c.searchRegex = regex
	c.searchString = ""
	return nil
}

// applyUnifiedSearchHighlighting applies the unified search highlighting to any colorized text
// DEPRECATED: No longer used since switching to single-pass colorization
// This preserves original styling for non-matching parts while highlighting matches
func (c *Colorizer) applyUnifiedSearchHighlighting(colorizedText string) string {
	if !c.hasSearchPattern() {
		return colorizedText
	}

	// Strip ANSI codes to work with plain text positions
	plainText := c.stripAnsiCodes(colorizedText)

	// Build result preserving original styling for non-matching parts
	result := strings.Builder{}
	lastPlainEnd := 0

	// Find matches directly in plain text to avoid complex position mapping
	var plainMatches []struct{ start, end int }

	if c.searchRegex != nil {
		regexMatches := c.searchRegex.FindAllStringIndex(plainText, -1)
		for _, match := range regexMatches {
			plainMatches = append(plainMatches, struct{ start, end int }{match[0], match[1]})
		}
	} else if c.searchString != "" {
		searchLen := len(c.searchString)
		startPos := 0
		for {
			pos := strings.Index(plainText[startPos:], c.searchString)
			if pos == -1 {
				break
			}
			actualPos := startPos + pos
			plainMatches = append(plainMatches, struct{ start, end int }{actualPos, actualPos + searchLen})
			startPos = actualPos + 1
		}
	}

	if len(plainMatches) == 0 {
		return colorizedText
	}

	for _, match := range plainMatches {
		plainStart := match.start
		plainEnd := match.end

		// Skip if this match is before our current position (overlapping matches)
		if plainStart < lastPlainEnd {
			continue
		}

		// Add text before match (preserving original styling)
		if plainStart > lastPlainEnd {
			styledBefore := c.findStyledTextForPlainRange(colorizedText, plainText, lastPlainEnd, plainStart)
			result.WriteString(styledBefore)
		}

		// Add highlighted match (just the plain text with highlight styling)
		matchText := plainText[plainStart:plainEnd]
		result.WriteString(c.theme.UnifiedSearchHighlight.Render(matchText))

		lastPlainEnd = plainEnd
	}

	// Add remaining text after last match (preserving original styling)
	if lastPlainEnd < len(plainText) {
		styledRemaining := c.findStyledTextForPlainRange(colorizedText, plainText, lastPlainEnd, len(plainText))
		result.WriteString(styledRemaining)
	}

	return result.String()
}

// findStyledTextForPlainRange extracts the styled text corresponding to a plain text range
func (c *Colorizer) findStyledTextForPlainRange(colorizedText, plainText string, plainStart, plainEnd int) string {
	// Handle edge cases
	if plainStart >= plainEnd || plainStart >= len(plainText) || plainEnd > len(plainText) {
		return ""
	}

	if plainStart == 0 && plainEnd == len(plainText) {
		return colorizedText
	}

	// If there are no ANSI codes, return the plain text range
	if !strings.Contains(colorizedText, "\x1b[") {
		return plainText[plainStart:plainEnd]
	}

	// Find the corresponding positions in the colorized text
	colorizedStart := c.findColorizedPosition(colorizedText, plainText, plainStart)
	colorizedEnd := c.findColorizedPosition(colorizedText, plainText, plainEnd)

	if colorizedStart >= 0 && colorizedEnd >= 0 && colorizedEnd <= len(colorizedText) {
		return colorizedText[colorizedStart:colorizedEnd]
	}

	// Fallback: return plain text
	return plainText[plainStart:plainEnd]
}

// SearchMatch represents a found search match with its position
type SearchMatch struct {
	start int
	end   int
	text  string
}

// findAllSearchMatches finds all occurrences of the search pattern in the text
func (c *Colorizer) findAllSearchMatches(text string) []SearchMatch {
	// Strip ANSI codes to find matches in the actual text content
	plainText := c.stripAnsiCodes(text)

	var matches []SearchMatch

	if c.searchRegex != nil {
		// Handle regex search
		regexMatches := c.searchRegex.FindAllStringIndex(plainText, -1)
		for _, match := range regexMatches {
			// Convert positions in plain text to positions in colorized text
			colorizedStart := c.findColorizedPosition(text, plainText, match[0])
			colorizedEnd := c.findColorizedPosition(text, plainText, match[1])

			matches = append(matches, SearchMatch{
				start: colorizedStart,
				end:   colorizedEnd,
				text:  plainText[match[0]:match[1]],
			})
		}
	} else if c.searchString != "" {
		// Handle string search (case-sensitive)
		searchLen := len(c.searchString)
		startPos := 0

		for {
			pos := strings.Index(plainText[startPos:], c.searchString)
			if pos == -1 {
				break
			}

			actualPos := startPos + pos
			colorizedStart := c.findColorizedPosition(text, plainText, actualPos)
			colorizedEnd := c.findColorizedPosition(text, plainText, actualPos+searchLen)

			matches = append(matches, SearchMatch{
				start: colorizedStart,
				end:   colorizedEnd,
				text:  c.searchString,
			})

			startPos = actualPos + 1
		}
	}

	return matches
}

// matchesSearchPattern checks if a string matches the current search pattern
func (c *Colorizer) matchesSearchPattern(text string) bool {
	if !c.hasSearchPattern() {
		return false
	}

	if c.searchRegex != nil {
		return c.searchRegex.MatchString(text)
	}

	if c.searchString != "" {
		return strings.Contains(text, c.searchString)
	}

	return false
}

// applySearchHighlighting applies search highlighting to any text, highlighting only matching parts
// This is used during single-pass colorization for all formats
func (c *Colorizer) applySearchHighlighting(text string, normalStyle lipgloss.Style) string {
	if c.searchString == "" && c.searchRegex == nil {
		return normalStyle.Render(text)
	}

	// Find all search matches in the plain text
	var matches []SearchMatch

	if c.searchRegex != nil {
		// Handle regex search
		regexMatches := c.searchRegex.FindAllStringIndex(text, -1)
		for _, match := range regexMatches {
			matches = append(matches, SearchMatch{
				start: match[0],
				end:   match[1],
				text:  text[match[0]:match[1]],
			})
		}
	} else if c.searchString != "" {
		// Handle string search (case-sensitive)
		searchLen := len(c.searchString)
		startPos := 0

		for {
			pos := strings.Index(text[startPos:], c.searchString)
			if pos == -1 {
				break
			}

			actualPos := startPos + pos
			matches = append(matches, SearchMatch{
				start: actualPos,
				end:   actualPos + searchLen,
				text:  c.searchString,
			})
			startPos = actualPos + searchLen
		}
	}

	if len(matches) == 0 {
		return normalStyle.Render(text)
	}

	// Build result with highlighted matches
	result := strings.Builder{}
	lastEnd := 0

	for _, match := range matches {
		// Add text before match with normal style
		if match.start > lastEnd {
			beforeText := text[lastEnd:match.start]
			result.WriteString(normalStyle.Render(beforeText))
		}

		// Add highlighted match
		matchText := text[match.start:match.end]
		result.WriteString(c.theme.UnifiedSearchHighlight.Render(matchText))

		lastEnd = match.end
	}

	// Add remaining text after last match
	if lastEnd < len(text) {
		remainingText := text[lastEnd:]
		result.WriteString(normalStyle.Render(remainingText))
	}

	return result.String()
}

// ColorizeLog applies colors to a log line based on its detected format
func (c *Colorizer) ColorizeLog(line string, format parser.LogFormat) string {
	if line == "" {
		return line
	}

	var result string

	// Apply colorization with integrated search highlighting (single-pass)
	switch format {
	case parser.JSONFormat:
		result = c.colorizeJSON(line)
	case parser.LogfmtFormat:
		result = c.colorizeLogfmt(line)
	case parser.ApacheCommonFormat:
		result = c.colorizeApacheCommon(line)
	case parser.NginxFormat:
		result = c.colorizeNginx(line)
	case parser.SyslogFormat:
		result = c.colorizeSyslog(line)
	case parser.GoStandardFormat:
		result = c.colorizeGoStandard(line)
	case parser.RailsFormat:
		result = c.colorizeRails(line)
	case parser.DockerFormat:
		result = c.colorizeDocker(line)
	case parser.KubernetesFormat:
		result = c.colorizeKubernetes(line)
	case parser.HerokuFormat:
		result = c.colorizeHeroku(line)
	default:
		result = c.colorizeGenericLog(line)
	}

	return result
}

// colorizeJSON adds colors to JSON log lines
func (c *Colorizer) colorizeJSON(line string) string {
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(line), &data); err != nil {
		return line // Return original if not valid JSON
	}

	// Re-marshal with indentation to get clean formatting, then colorize
	result := strings.Builder{}
	result.WriteString(c.theme.Bracket.Render("{"))

	first := true
	for key, value := range data {
		if !first {
			result.WriteString(", ")
		}
		first = false

		// Colorize key
		result.WriteString(c.theme.Quote.Render(`"`))

		// Special handling for common log fields with search highlighting
		keyStyle := c.theme.JSONKey
		if c.isLogLevelKey(key) {
			keyStyle = c.theme.GetLogLevelStyle(key)
		}
		styledKey := c.applySearchHighlighting(key, keyStyle)
		result.WriteString(styledKey)
		result.WriteString(c.theme.Quote.Render(`"`))
		result.WriteString(c.theme.Equals.Render(":"))

		// Colorize value based on key and type
		result.WriteString(c.colorizeJSONValue(key, value))
	}

	result.WriteString(c.theme.Bracket.Render("}"))
	return result.String()
}

// colorizeJSONValue colors a JSON value based on context and type with integrated search highlighting
func (c *Colorizer) colorizeJSONValue(key string, value interface{}) string {
	switch v := value.(type) {
	case string:
		// Special handling for known fields
		if c.isLogLevelKey(key) {
			styledValue := c.applySearchHighlighting(v, c.theme.GetLogLevelStyle(v))
			return c.theme.Quote.Render(`"`) + styledValue + c.theme.Quote.Render(`"`)
		}
		if c.isTimestampKey(key) {
			styledValue := c.applySearchHighlighting(v, c.theme.Timestamp)
			return c.theme.Quote.Render(`"`) + styledValue + c.theme.Quote.Render(`"`)
		}
		if c.isServiceKey(key) {
			styledValue := c.applySearchHighlighting(v, c.theme.Service)
			return c.theme.Quote.Render(`"`) + styledValue + c.theme.Quote.Render(`"`)
		}
		// Regular string value
		styledValue := c.applySearchHighlighting(v, c.theme.JSONString)
		return c.theme.Quote.Render(`"`) + styledValue + c.theme.Quote.Render(`"`)
	case float64:
		numberStr := fmt.Sprintf("%g", v)
		return c.applySearchHighlighting(numberStr, c.theme.JSONNumber)
	case bool:
		if v {
			return c.applySearchHighlighting("true", c.theme.StatusOK)
		}
		return c.applySearchHighlighting("false", c.theme.StatusWarn)
	case map[string]interface{}:
		// Recursively colorize nested JSON objects
		return c.colorizeNestedJSONObject(v)
	case []interface{}:
		// Handle JSON arrays
		return c.colorizeJSONArray(v)
	default:
		// Fallback to JSON marshaling for other types (null, etc.)
		jsonBytes, _ := json.Marshal(v)
		return c.applySearchHighlighting(string(jsonBytes), c.theme.JSONValue)
	}
}

// colorizeNestedJSONObject recursively colorizes a nested JSON object
func (c *Colorizer) colorizeNestedJSONObject(obj map[string]interface{}) string {
	result := strings.Builder{}
	result.WriteString(c.theme.Bracket.Render("{"))

	first := true
	for key, value := range obj {
		if !first {
			result.WriteString(", ")
		}
		first = false

		// Colorize nested key with search highlighting
		result.WriteString(c.theme.Quote.Render(`"`))
		styledKey := c.applySearchHighlighting(key, c.theme.JSONKey)
		result.WriteString(styledKey)
		result.WriteString(c.theme.Quote.Render(`"`))
		result.WriteString(c.theme.Equals.Render(":"))

		// Recursively colorize nested value
		result.WriteString(c.colorizeJSONValue(key, value))
	}

	result.WriteString(c.theme.Bracket.Render("}"))
	return result.String()
}

// colorizeJSONArray colorizes a JSON array
func (c *Colorizer) colorizeJSONArray(arr []interface{}) string {
	result := strings.Builder{}
	result.WriteString(c.theme.Bracket.Render("["))

	for i, value := range arr {
		if i > 0 {
			result.WriteString(", ")
		}

		// Colorize array element (use empty key for array elements)
		result.WriteString(c.colorizeJSONValue("", value))
	}

	result.WriteString(c.theme.Bracket.Render("]"))
	return result.String()
}

// colorizeLogfmt adds colors to logfmt lines
func (c *Colorizer) colorizeLogfmt(line string) string {
	// Simple logfmt parsing - split by spaces and look for key=value pairs
	parts := strings.Fields(line)
	result := strings.Builder{}

	for i, part := range parts {
		if i > 0 {
			result.WriteString(" ")
		}

		if strings.Contains(part, "=") && !strings.HasPrefix(part, "=") && !strings.HasSuffix(part, "=") {
			// This is a key=value pair
			kv := strings.SplitN(part, "=", 2)
			if len(kv) == 2 {
				key, value := kv[0], kv[1]

				// Remove quotes from value if present
				cleanValue := strings.Trim(value, `"`)

				// Color the key with search highlighting
				keyStyle := c.theme.LogfmtKey
				if c.isLogLevelKey(key) {
					keyStyle = c.theme.GetLogLevelStyle(key)
				}
				result.WriteString(c.applySearchHighlighting(key, keyStyle))
				result.WriteString(c.theme.Equals.Render("="))

				// Color the value based on key with search highlighting
				if c.isLogLevelKey(key) {
					if strings.HasPrefix(value, `"`) && strings.HasSuffix(value, `"`) {
						result.WriteString(c.theme.Quote.Render(`"`))
						result.WriteString(c.applySearchHighlighting(cleanValue, c.theme.GetLogLevelStyle(cleanValue)))
						result.WriteString(c.theme.Quote.Render(`"`))
					} else {
						result.WriteString(c.applySearchHighlighting(value, c.theme.GetLogLevelStyle(cleanValue)))
					}
				} else if c.isTimestampKey(key) {
					result.WriteString(c.applySearchHighlighting(value, c.theme.Timestamp))
				} else if c.isServiceKey(key) {
					result.WriteString(c.applySearchHighlighting(value, c.theme.Service))
				} else {
					result.WriteString(c.applySearchHighlighting(value, c.theme.LogfmtValue))
				}
			} else {
				result.WriteString(part)
			}
		} else {
			// Not a key=value pair, check if it's a log level
			if c.looksLikeLogLevel(part) {
				result.WriteString(c.applySearchHighlighting(part, c.theme.GetLogLevelStyle(part)))
			} else {
				result.WriteString(c.applySearchHighlighting(part, lipgloss.NewStyle()))
			}
		}
	}

	return result.String()
}

// colorizeApacheCommon adds colors to Apache Common Log format
func (c *Colorizer) colorizeApacheCommon(line string) string {
	// Apache Common Log format: IP - - [timestamp] "method URL protocol" status size
	re := regexp.MustCompile(`^(\S+) (\S+) (\S+) \[([^\]]+)\] "([A-Z]+) ([^"]*) ([^"]*)" (\d+) (\S+)`)
	matches := re.FindStringSubmatch(line)

	if len(matches) != 10 {
		return c.colorizeGenericLog(line)
	}

	ip := matches[1]
	timestamp := matches[4]
	method := matches[5]
	url := matches[6]
	protocol := matches[7]
	status := matches[8]
	size := matches[9]

	result := strings.Builder{}
	// Apply search highlighting during colorization (single-pass)
	result.WriteString(c.applySearchHighlighting(ip, c.theme.IP))
	result.WriteString(" - - ")
	result.WriteString(c.theme.Bracket.Render("["))
	result.WriteString(c.applySearchHighlighting(timestamp, c.theme.Timestamp))
	result.WriteString(c.theme.Bracket.Render("] "))
	result.WriteString(c.theme.Quote.Render(`"`))
	result.WriteString(c.applySearchHighlighting(method, c.theme.Method))
	result.WriteString(" ")
	result.WriteString(c.applySearchHighlighting(url, c.theme.URL))
	result.WriteString(" ")
	result.WriteString(c.applySearchHighlighting(protocol, lipgloss.NewStyle()))
	result.WriteString(c.theme.Quote.Render(`" `))
	result.WriteString(c.applySearchHighlighting(status, c.theme.GetHTTPStatusStyle(status)))
	result.WriteString(" ")
	result.WriteString(c.applySearchHighlighting(size, lipgloss.NewStyle()))

	return result.String()
}

// colorizeNginx adds colors to Nginx log format (extends Apache)
func (c *Colorizer) colorizeNginx(line string) string {
	// Nginx format: IP - - [timestamp] "method URL protocol" status size "referer" "user-agent"
	re := regexp.MustCompile(`^(\S+) (\S+) (\S+) \[([^\]]+)\] "([A-Z]+) ([^"]*) ([^"]*)" (\d+) (\S+) "([^"]*)" "([^"]*)"`)
	matches := re.FindStringSubmatch(line)

	if len(matches) != 12 {
		return c.colorizeApacheCommon(line) // Fallback to Apache format
	}

	ip := matches[1]
	timestamp := matches[4]
	method := matches[5]
	url := matches[6]
	protocol := matches[7]
	status := matches[8]
	size := matches[9]
	referer := matches[10]
	userAgent := matches[11]

	result := strings.Builder{}
	// Apply search highlighting during colorization (single-pass)
	result.WriteString(c.applySearchHighlighting(ip, c.theme.IP))
	result.WriteString(" - - ")
	result.WriteString(c.theme.Bracket.Render("["))
	result.WriteString(c.applySearchHighlighting(timestamp, c.theme.Timestamp))
	result.WriteString(c.theme.Bracket.Render("] "))
	result.WriteString(c.theme.Quote.Render(`"`))
	result.WriteString(c.applySearchHighlighting(method, c.theme.Method))
	result.WriteString(" ")
	result.WriteString(c.applySearchHighlighting(url, c.theme.URL))
	result.WriteString(" ")
	result.WriteString(c.applySearchHighlighting(protocol, lipgloss.NewStyle()))
	result.WriteString(c.theme.Quote.Render(`" `))
	result.WriteString(c.applySearchHighlighting(status, c.theme.GetHTTPStatusStyle(status)))
	result.WriteString(" ")
	result.WriteString(c.applySearchHighlighting(size, lipgloss.NewStyle()))
	result.WriteString(" ")
	result.WriteString(c.theme.Quote.Render(`"`))
	result.WriteString(c.applySearchHighlighting(referer, lipgloss.NewStyle()))
	result.WriteString(c.theme.Quote.Render(`" "`))
	result.WriteString(c.applySearchHighlighting(userAgent, lipgloss.NewStyle()))
	result.WriteString(c.theme.Quote.Render(`"`))

	return result.String()
}

// Helper functions for identifying special keys
func (c *Colorizer) isLogLevelKey(key string) bool {
	lowerKey := strings.ToLower(key)
	return lowerKey == "level" || lowerKey == "severity" || lowerKey == "loglevel"
}

func (c *Colorizer) isTimestampKey(key string) bool {
	lowerKey := strings.ToLower(key)
	return lowerKey == "timestamp" || lowerKey == "time" || lowerKey == "ts" || lowerKey == "@timestamp"
}

func (c *Colorizer) isServiceKey(key string) bool {
	lowerKey := strings.ToLower(key)
	return lowerKey == "service" || lowerKey == "component" || lowerKey == "module" || lowerKey == "app"
}

func (c *Colorizer) looksLikeLogLevel(s string) bool {
	upper := strings.ToUpper(s)
	return upper == "ERROR" || upper == "WARN" || upper == "WARNING" ||
		upper == "INFO" || upper == "DEBUG" || upper == "TRACE" ||
		upper == "FATAL" || upper == "CRITICAL"
}

// colorizeSyslog adds colors to syslog format lines
func (c *Colorizer) colorizeSyslog(line string) string {
	// Syslog format: "Jan 19 10:30:00 hostname myapp[1234]: ERROR: Database connection failed"
	re := regexp.MustCompile(`^(\w{3} \d{1,2} \d{2}:\d{2}:\d{2}) (\S+) (\S+)\[(\d+)\]: (.*)`)
	matches := re.FindStringSubmatch(line)

	if len(matches) != 6 {
		return c.colorizeGenericLog(line) // Fallback
	}

	timestamp := matches[1]
	hostname := matches[2]
	process := matches[3]
	pid := matches[4]
	message := matches[5]

	result := strings.Builder{}
	// Apply search highlighting during colorization (single-pass)
	result.WriteString(c.applySearchHighlighting(timestamp, c.theme.Timestamp))
	result.WriteString(" ")
	result.WriteString(c.applySearchHighlighting(hostname, c.theme.Hostname))
	result.WriteString(" ")
	result.WriteString(c.applySearchHighlighting(process, c.theme.Service))
	result.WriteString(c.theme.Bracket.Render("["))
	result.WriteString(c.applySearchHighlighting(pid, c.theme.PID))
	result.WriteString(c.theme.Bracket.Render("]: "))
	result.WriteString(c.colorizeMessageWithHighlighting(message))

	return result.String()
}

func (c *Colorizer) colorizeGoStandard(line string) string {
	// Go standard format: "2025/01/19 10:30:00 ERROR: Database connection failed"
	re := regexp.MustCompile(`^(\d{4}/\d{2}/\d{2} \d{2}:\d{2}:\d{2}) (.*)`)
	matches := re.FindStringSubmatch(line)

	if len(matches) != 3 {
		return c.colorizeGenericLog(line)
	}

	timestamp := matches[1]
	message := matches[2]

	result := strings.Builder{}
	// Apply search highlighting during colorization (single-pass)
	result.WriteString(c.applySearchHighlighting(timestamp, c.theme.Timestamp))
	result.WriteString(" ")
	result.WriteString(c.colorizeMessageWithHighlighting(message))

	return result.String()
}

func (c *Colorizer) colorizeRails(line string) string {
	// Rails format: "[2025-01-19 10:30:00] ERROR -- : Database connection failed"
	// WEBrick format: "[2025-01-19 10:30:00] INFO  WEBrick 1.4.4"

	// Try Rails format first
	railsRe := regexp.MustCompile(`^(\[[^\]]+\]) (\w+) (--) : (.*)`)
	matches := railsRe.FindStringSubmatch(line)

	if len(matches) == 5 {
		timestamp := matches[1]
		level := matches[2]
		separator := matches[3]
		message := matches[4]

		result := strings.Builder{}
		result.WriteString(c.theme.Bracket.Render("["))
		timestampContent := timestamp[1 : len(timestamp)-1] // Remove brackets
		// Apply search highlighting during colorization (single-pass)
		result.WriteString(c.applySearchHighlighting(timestampContent, c.theme.Timestamp))
		result.WriteString(c.theme.Bracket.Render("] "))
		result.WriteString(c.applySearchHighlighting(level, c.theme.GetLogLevelStyle(level)))
		result.WriteString(" ")
		result.WriteString(c.applySearchHighlighting(separator, lipgloss.NewStyle()))
		result.WriteString(" : ")
		result.WriteString(c.applySearchHighlighting(message, lipgloss.NewStyle()))

		return result.String()
	}

	// Try WEBrick format
	webrickRe := regexp.MustCompile(`^(\[[^\]]+\]) (\w+)\s+(.*)`)
	matches = webrickRe.FindStringSubmatch(line)

	if len(matches) == 4 {
		timestamp := matches[1]
		level := matches[2]
		message := matches[3]

		result := strings.Builder{}
		result.WriteString(c.theme.Bracket.Render("["))
		timestampContent := timestamp[1 : len(timestamp)-1] // Remove brackets
		// Apply search highlighting during colorization (single-pass)
		result.WriteString(c.applySearchHighlighting(timestampContent, c.theme.Timestamp))
		result.WriteString(c.theme.Bracket.Render("] "))
		result.WriteString(c.applySearchHighlighting(level, c.theme.GetLogLevelStyle(level)))
		result.WriteString(" ")
		result.WriteString(c.applySearchHighlighting(message, c.theme.JSONValue))

		return result.String()
	}

	return c.colorizeGenericLog(line)
}

func (c *Colorizer) colorizeDocker(line string) string {
	// Docker format: "2025-01-19T10:30:00.123456789Z ERROR Database connection failed"
	re := regexp.MustCompile(`^(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d+Z)\s+([A-Z]+)\s+(.*)`)
	matches := re.FindStringSubmatch(line)

	if len(matches) != 4 {
		return c.colorizeGenericLog(line)
	}

	timestamp := matches[1]
	level := matches[2]
	message := matches[3]

	result := strings.Builder{}
	// Apply search highlighting during colorization (single-pass)
	result.WriteString(c.applySearchHighlighting(timestamp, c.theme.Timestamp))
	result.WriteString(" ")
	result.WriteString(c.applySearchHighlighting(level, c.theme.GetLogLevelStyle(level)))
	result.WriteString(" ")
	result.WriteString(c.applySearchHighlighting(message, lipgloss.NewStyle()))

	return result.String()
}

func (c *Colorizer) colorizeKubernetes(line string) string {
	// Kubernetes format: "2025-01-19T10:30:00.123Z 1 main.go:42] ERROR Database connection failed"
	re := regexp.MustCompile(`^(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d+Z) (\d+) ([^:]+):(\d+)\] (.*)`)
	matches := re.FindStringSubmatch(line)

	if len(matches) != 6 {
		return c.colorizeGenericLog(line)
	}

	timestamp := matches[1]
	severity := matches[2]
	filename := matches[3]
	lineNum := matches[4]
	message := matches[5]

	result := strings.Builder{}
	// Apply search highlighting during colorization (single-pass)
	result.WriteString(c.applySearchHighlighting(timestamp, c.theme.Timestamp))
	result.WriteString(" ")
	result.WriteString(c.applySearchHighlighting(severity, c.theme.PID))
	result.WriteString(" ")
	result.WriteString(c.applySearchHighlighting(filename, c.theme.Filename))
	result.WriteString(":")
	result.WriteString(c.applySearchHighlighting(lineNum, c.theme.LineNum))
	result.WriteString(c.theme.Bracket.Render("] "))
	result.WriteString(c.colorizeMessageWithHighlighting(message))

	return result.String()
}

func (c *Colorizer) colorizeHeroku(line string) string {
	// Heroku format: "2025-01-19T10:30:00+00:00 app[web.1]: ERROR Database connection failed"
	re := regexp.MustCompile(`^(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}[+-]\d{2}:\d{2}) app\[([^\]]+)\]: (.*)`)
	matches := re.FindStringSubmatch(line)

	if len(matches) != 4 {
		return c.colorizeGenericLog(line)
	}

	timestamp := matches[1]
	dyno := matches[2]
	message := matches[3]

	result := strings.Builder{}
	// Apply search highlighting during colorization (single-pass)
	result.WriteString(c.applySearchHighlighting(timestamp, c.theme.Timestamp))
	result.WriteString(" app")
	result.WriteString(c.theme.Bracket.Render("["))
	result.WriteString(c.applySearchHighlighting(dyno, c.theme.Service))
	result.WriteString(c.theme.Bracket.Render("]: "))
	result.WriteString(c.colorizeMessageWithHighlighting(message))

	return result.String()
}

// colorizeMessage applies colors to message content, looking for log levels
func (c *Colorizer) colorizeMessage(message string) string {
	// Simple approach: look for log level at the beginning of the message
	parts := strings.Fields(message)
	if len(parts) == 0 {
		return message
	}

	firstWord := parts[0]
	// Remove trailing colon if present
	cleanWord := strings.TrimSuffix(firstWord, ":")

	if c.looksLikeLogLevel(cleanWord) {
		result := strings.Builder{}
		if strings.HasSuffix(firstWord, ":") {
			styledLevel := c.theme.GetLogLevelStyle(cleanWord).Render(cleanWord)
			result.WriteString(styledLevel)
			result.WriteString(":")
		} else {
			styledLevel := c.theme.GetLogLevelStyle(cleanWord).Render(cleanWord)
			result.WriteString(styledLevel)
		}

		if len(parts) > 1 {
			result.WriteString(" ")
			result.WriteString(strings.Join(parts[1:], " "))
		}
		return result.String()
	}

	return message
}

// colorizeMessageWithHighlighting colors message parts with integrated search highlighting
func (c *Colorizer) colorizeMessageWithHighlighting(message string) string {
	// Simple approach: look for log level at the beginning of the message
	parts := strings.Fields(message)
	if len(parts) == 0 {
		return c.applySearchHighlighting(message, lipgloss.NewStyle())
	}

	firstWord := parts[0]
	// Remove trailing colon if present
	cleanWord := strings.TrimSuffix(firstWord, ":")

	if c.looksLikeLogLevel(cleanWord) {
		result := strings.Builder{}
		if strings.HasSuffix(firstWord, ":") {
			styledLevel := c.applySearchHighlighting(cleanWord, c.theme.GetLogLevelStyle(cleanWord))
			result.WriteString(styledLevel)
			result.WriteString(":")
		} else {
			styledLevel := c.applySearchHighlighting(cleanWord, c.theme.GetLogLevelStyle(cleanWord))
			result.WriteString(styledLevel)
		}

		if len(parts) > 1 {
			result.WriteString(" ")
			remainingMessage := strings.Join(parts[1:], " ")
			result.WriteString(c.applySearchHighlighting(remainingMessage, lipgloss.NewStyle()))
		}
		return result.String()
	}

	return c.applySearchHighlighting(message, lipgloss.NewStyle())
}

// colorizeGenericLog provides basic colorization for unrecognized formats
func (c *Colorizer) colorizeGenericLog(line string) string {
	// Look for common patterns in any log line
	words := strings.Fields(line)
	result := strings.Builder{}

	for i, word := range words {
		if i > 0 {
			result.WriteString(" ")
		}

		cleanWord := strings.TrimSuffix(word, ":")
		if c.looksLikeLogLevel(cleanWord) {
			if strings.HasSuffix(word, ":") {
				styledLevel := c.applySearchHighlighting(cleanWord, c.theme.GetLogLevelStyle(cleanWord))
				result.WriteString(styledLevel)
				result.WriteString(":")
			} else {
				styledLevel := c.applySearchHighlighting(cleanWord, c.theme.GetLogLevelStyle(cleanWord))
				result.WriteString(styledLevel)
			}
		} else {
			styledWord := c.applySearchHighlighting(word, lipgloss.NewStyle())
			result.WriteString(styledWord)
		}
	}

	return result.String()
}

// Search highlighting markers
const (
	searchStartMarker = "⟦SEARCH_START⟧"
	searchEndMarker   = "⟦SEARCH_END⟧"
)

// markSearchMatches marks search matches in the original text with special markers
func (c *Colorizer) markSearchMatches(line string) string {
	if !c.matchesSearch(line) {
		return line
	}

	var searchTexts []string
	if c.searchRegex != nil {
		// Find all regex matches
		matches := c.searchRegex.FindAllString(line, -1)
		searchTexts = matches
	} else if c.searchString != "" {
		searchTexts = []string{c.searchString}
	} else {
		return line
	}

	result := line

	// Mark each unique search text
	seen := make(map[string]bool)
	for _, searchText := range searchTexts {
		if seen[searchText] || searchText == "" {
			continue
		}
		seen[searchText] = true

		// Replace all occurrences with marked versions
		marked := searchStartMarker + searchText + searchEndMarker
		result = strings.ReplaceAll(result, searchText, marked)
	}

	return result
}

// stripAnsiCodes removes ANSI escape sequences from text
func (c *Colorizer) stripAnsiCodes(text string) string {
	// ANSI escape sequence pattern: \033[...m or \x1b[...m
	ansiRegex := regexp.MustCompile(`\033\[[0-9;]*m|\x1b\[[0-9;]*m`)
	return ansiRegex.ReplaceAllString(text, "")
}

// matchesSearch checks if a line matches the current search pattern
func (c *Colorizer) matchesSearch(line string) bool {
	if c.searchRegex != nil {
		return c.searchRegex.MatchString(line)
	}
	if c.searchString != "" {
		return strings.Contains(line, c.searchString)
	}
	return false
}

// hasSearchPattern checks if any search pattern is configured
func (c *Colorizer) hasSearchPattern() bool {
	return c.searchRegex != nil || c.searchString != ""
}

// applyPostColorizationSearchHighlight applies search highlighting to already colorized JSON
func (c *Colorizer) applyPostColorizationSearchHighlight(colorizedText, originalText string) string {
	if !c.matchesSearch(originalText) {
		return colorizedText
	}

	var searchTexts []string
	if c.searchRegex != nil {
		// Find all regex matches in the original text
		matches := c.searchRegex.FindAllString(originalText, -1)
		searchTexts = matches
	} else if c.searchString != "" {
		searchTexts = []string{c.searchString}
	}

	result := colorizedText

	// Apply highlighting to each unique search text
	seen := make(map[string]bool)
	for _, searchText := range searchTexts {
		if seen[searchText] || searchText == "" {
			continue
		}
		seen[searchText] = true

		// Use JSON-aware highlighting that preserves structure
		result = c.highlightInColorizedJSONText(result, searchText, originalText)
	}

	return result
}

// highlightInColorizedJSONText applies JSON-aware search highlighting that preserves structure
func (c *Colorizer) highlightInColorizedJSONText(colorizedText, searchText, originalText string) string {
	// Parse the original JSON to understand structure
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(originalText), &data); err != nil {
		// If JSON parsing fails, fall back to simple highlighting
		return c.highlightInColorizedText(colorizedText, searchText)
	}

	// Instead of trying to map positions (unreliable due to JSON reordering),
	// apply highlighting by finding and replacing all occurrences in the colorized text
	return c.highlightJSONTextSimple(colorizedText, searchText)
}

// highlightJSONTextSimple applies search highlighting to JSON text while preserving structure
func (c *Colorizer) highlightJSONTextSimple(colorizedText, searchText string) string {
	// Simple but effective approach: find all occurrences of the search text
	// and highlight them, but be careful about JSON structure
	plainText := c.stripAnsiCodes(colorizedText)

	// Find all positions where the search text appears
	var positions []int
	startPos := 0
	for {
		pos := strings.Index(plainText[startPos:], searchText)
		if pos == -1 {
			break
		}
		positions = append(positions, startPos+pos)
		startPos = startPos + pos + 1
	}

	if len(positions) == 0 {
		return colorizedText
	}

	// Apply highlighting from right to left to preserve positions
	result := colorizedText
	for i := len(positions) - 1; i >= 0; i-- {
		plainPos := positions[i]

		// Check if this match is inside a JSON key (between quotes followed by colon)
		if c.isPositionInJSONKey(plainText, plainPos, len(searchText)) {
			// Highlight the entire key
			result = c.highlightJSONKeyAtPosition(result, plainPos, searchText, plainText)
		} else {
			// Highlight the match in a JSON value by highlighting the entire quoted value
			result = c.highlightJSONValueAtPosition(result, plainPos, searchText, plainText)
		}
	}

	return result
}

// highlightJSONValueAtPosition highlights a JSON value containing a search match
func (c *Colorizer) highlightJSONValueAtPosition(colorizedText string, plainPos int, searchText, plainText string) string {
	// Find the bounds of the quoted value containing the match
	start, end := c.findJSONValueBounds(plainText, plainPos)
	if start == -1 || end == -1 {
		return colorizedText
	}

	// Find corresponding positions in colorized text
	colorizedStart := c.findColorizedPosition(colorizedText, c.stripAnsiCodes(colorizedText), start)
	colorizedEnd := c.findColorizedPosition(colorizedText, c.stripAnsiCodes(colorizedText), end)

	// Extract the value (including quotes)
	before := colorizedText[:colorizedStart]
	value := colorizedText[colorizedStart:colorizedEnd]
	after := colorizedText[colorizedEnd:]

	// Apply highlighting to the entire quoted value
	highlightedValue := c.theme.JSONSearchHighlight.Render(value)

	return before + highlightedValue + after
}

// isPositionInJSONKey checks if a position is inside a JSON key
func (c *Colorizer) isPositionInJSONKey(plainText string, pos, length int) bool {
	// Find the enclosing quotes
	beforePos := plainText[:pos]
	afterPos := plainText[pos+length:]

	// Find the last quote before the position
	lastQuote := strings.LastIndex(beforePos, `"`)
	if lastQuote == -1 {
		return false
	}

	// Find the next quote after the position
	nextQuote := strings.Index(afterPos, `"`)
	if nextQuote == -1 {
		return false
	}

	// Check what comes after the closing quote
	afterQuotePos := pos + length + nextQuote + 1
	if afterQuotePos >= len(plainText) {
		return false
	}

	remaining := strings.TrimSpace(plainText[afterQuotePos:])
	return strings.HasPrefix(remaining, ":")
}

// highlightJSONKeyAtPosition highlights a JSON key at a specific position
func (c *Colorizer) highlightJSONKeyAtPosition(colorizedText string, pos int, searchText, plainText string) string {
	// Find the bounds of the key in plain text
	keyStart, keyEnd := c.findKeyBoundsAtPosition(plainText, pos)
	if keyStart == -1 || keyEnd == -1 {
		// Fallback to simple highlighting
		colorizedPos := c.findColorizedPosition(colorizedText, c.stripAnsiCodes(colorizedText), pos)
		before := colorizedText[:colorizedPos]
		match := colorizedText[colorizedPos : colorizedPos+len(searchText)]
		after := colorizedText[colorizedPos+len(searchText):]
		return before + c.theme.JSONSearchHighlight.Render(match) + after
	}

	// Map to colorized text positions
	colorizedKeyStart := c.findColorizedPosition(colorizedText, c.stripAnsiCodes(colorizedText), keyStart)
	colorizedKeyEnd := c.findColorizedPosition(colorizedText, c.stripAnsiCodes(colorizedText), keyEnd)

	before := colorizedText[:colorizedKeyStart]
	keyText := colorizedText[colorizedKeyStart:colorizedKeyEnd]
	after := colorizedText[colorizedKeyEnd:]

	highlightedKey := c.theme.JSONSearchHighlight.Render(keyText)
	return before + highlightedKey + after
}

// findKeyBoundsAtPosition finds the start and end of a JSON key containing the given position
func (c *Colorizer) findKeyBoundsAtPosition(plainText string, pos int) (int, int) {
	// Find the opening quote
	keyStart := strings.LastIndex(plainText[:pos], `"`)
	if keyStart == -1 {
		return -1, -1
	}

	// Find the closing quote
	keyEnd := strings.Index(plainText[pos:], `"`)
	if keyEnd == -1 {
		return -1, -1
	}
	keyEnd = pos + keyEnd + 1 // Include the closing quote

	// Verify this is actually a key
	afterKey := strings.TrimSpace(plainText[keyEnd:])
	if !strings.HasPrefix(afterKey, ":") {
		return -1, -1
	}

	return keyStart, keyEnd
}

// JSONSearchMatch represents a search match with context information
type JSONSearchMatch struct {
	Position int  // Position in the original text
	InValue  bool // True if match is in a JSON value, false if in a key
}

// findJSONSearchMatches finds search matches and determines if they're in keys or values
func (c *Colorizer) findJSONSearchMatches(jsonText, searchText string) []JSONSearchMatch {
	var matches []JSONSearchMatch

	// Simple approach: find all occurrences and analyze context
	startPos := 0
	for {
		pos := strings.Index(jsonText[startPos:], searchText)
		if pos == -1 {
			break
		}
		absolutePos := startPos + pos

		// Determine if this match is in a key or value by analyzing surrounding context
		inValue := c.isSearchMatchInJSONValue(jsonText, absolutePos, len(searchText))

		matches = append(matches, JSONSearchMatch{
			Position: absolutePos,
			InValue:  inValue,
		})

		startPos = absolutePos + 1
	}

	return matches
}

// highlightJSONKey highlights a JSON key by wrapping the entire quoted key with ANSI codes
func (c *Colorizer) highlightJSONKey(colorizedText string, matchPos int, searchText, plainText string) string {
	// Find the start and end of the quoted key that contains our match
	keyStart, keyEnd := c.findJSONKeyBounds(plainText, matchPos)
	if keyStart == -1 || keyEnd == -1 {
		// Fallback: highlight just the match if we can't find key bounds
		colorizedPos := c.findColorizedPosition(colorizedText, plainText, matchPos)
		before := colorizedText[:colorizedPos]
		match := colorizedText[colorizedPos : colorizedPos+len(searchText)]
		after := colorizedText[colorizedPos+len(searchText):]
		return before + c.theme.SearchHighlight.Render(match) + after
	}

	// Map the key bounds from plain text to colorized text
	colorizedKeyStart := c.findColorizedPosition(colorizedText, plainText, keyStart)
	colorizedKeyEnd := c.findColorizedPosition(colorizedText, plainText, keyEnd)

	// Extract the entire quoted key from colorized text
	before := colorizedText[:colorizedKeyStart]
	keyWithQuotes := colorizedText[colorizedKeyStart:colorizedKeyEnd]
	after := colorizedText[colorizedKeyEnd:]

	// Apply highlighting to the entire quoted key
	highlightedKey := c.theme.SearchHighlight.Render(keyWithQuotes)

	return before + highlightedKey + after
}

// findJSONKeyBounds finds the start and end positions of a quoted JSON key containing the match
func (c *Colorizer) findJSONKeyBounds(plainText string, matchPos int) (int, int) {
	// Find the opening quote of the key that contains our match
	keyStart := strings.LastIndex(plainText[:matchPos], `"`)
	if keyStart == -1 {
		return -1, -1
	}

	// Find the closing quote of the key
	keyEnd := strings.Index(plainText[matchPos:], `"`)
	if keyEnd == -1 {
		return -1, -1
	}
	keyEnd = matchPos + keyEnd + 1 // Include the closing quote

	// Verify this is actually a key by checking what follows the closing quote
	afterKey := strings.TrimSpace(plainText[keyEnd:])
	if !strings.HasPrefix(afterKey, ":") {
		return -1, -1 // Not a key
	}

	return keyStart, keyEnd
}

// isSearchMatchInJSONValue determines if a search match is within a JSON value (not a key)
func (c *Colorizer) isSearchMatchInJSONValue(jsonText string, matchPos, matchLen int) bool {
	// Look backwards from match position to find the nearest quote and colon
	// If we find a colon before a quote, we're likely in a value
	// If we find a quote before a colon, we might be in a key

	beforeMatch := jsonText[:matchPos]
	afterMatch := jsonText[matchPos+matchLen:]

	// Find the last quote before our match
	lastQuotePos := strings.LastIndex(beforeMatch, `"`)
	if lastQuotePos == -1 {
		return false // Not in quotes at all
	}

	// Find the next quote after our match
	nextQuotePos := strings.Index(afterMatch, `"`)
	if nextQuotePos == -1 {
		return false // Not properly quoted
	}

	// Check what comes after the closing quote
	afterQuoteText := jsonText[matchPos+matchLen+nextQuotePos+1:]
	afterQuoteText = strings.TrimSpace(afterQuoteText)

	// If the character after the closing quote is ':', we're in a key
	// If it's ',' or '}' or ']', we're in a value
	if strings.HasPrefix(afterQuoteText, ":") {
		return false // This is a key
	}

	return true // This is likely a value
}

// highlightInColorizedText highlights search text in colorized output while preserving colors
func (c *Colorizer) highlightInColorizedText(colorizedText, searchText string) string {
	// Strip ANSI codes to find plain text positions
	plainText := c.stripAnsiCodes(colorizedText)

	// Find all occurrences of the search text in plain text
	var positions []int
	startPos := 0
	for {
		pos := strings.Index(plainText[startPos:], searchText)
		if pos == -1 {
			break
		}
		positions = append(positions, startPos+pos)
		startPos = startPos + pos + 1 // Move past this match for next search
	}

	if len(positions) == 0 {
		return colorizedText
	}

	// Apply highlighting from right to left to preserve positions
	result := colorizedText
	for i := len(positions) - 1; i >= 0; i-- {
		plainPos := positions[i]
		colorizedPos := c.findColorizedPosition(result, c.stripAnsiCodes(result), plainPos)

		// Extract the search text at this position in the colorized string
		before := result[:colorizedPos]
		matchedText := result[colorizedPos : colorizedPos+len(searchText)]
		after := result[colorizedPos+len(searchText):]

		// Apply search highlighting
		highlightedText := c.theme.SearchHighlight.Render(matchedText)
		result = before + highlightedText + after
	}

	return result
}

// simpleSearchHighlight applies bold highlighting to search matches in colorized text
func (c *Colorizer) simpleSearchHighlight(colorizedText, originalText string) string {
	var searchTexts []string

	// Strip ANSI codes from original text before finding matches
	cleanOriginalText := c.stripAnsiCodes(originalText)

	if c.searchRegex != nil {
		// Find all regex matches in the clean original text
		matches := c.searchRegex.FindAllString(cleanOriginalText, -1)
		searchTexts = matches
	} else if c.searchString != "" {
		searchTexts = []string{c.searchString}
	} else {
		return colorizedText
	}

	result := colorizedText

	// Apply highlighting to each unique search text while preserving colors
	seen := make(map[string]bool)
	for _, searchText := range searchTexts {
		if seen[searchText] || searchText == "" {
			continue
		}
		seen[searchText] = true

		result = c.highlightTextPreservingColors(result, searchText)
	}

	return result
}

// highlightTextPreservingColors highlights all occurrences of search text while preserving existing colors
func (c *Colorizer) highlightTextPreservingColors(colorizedText, searchText string) string {
	if searchText == "" {
		return colorizedText
	}

	// Work with clean text to find matches, but apply highlighting to colorized text
	cleanText := c.stripAnsiCodes(colorizedText)

	result := ""
	colorizedRemaining := colorizedText
	cleanRemaining := cleanText

	for {
		// Find the next occurrence of the search text in clean text
		cleanPos := strings.Index(cleanRemaining, searchText)
		if cleanPos == -1 {
			// No more matches, append remaining colorized text and break
			result += colorizedRemaining
			break
		}

		// Find the corresponding position in the colorized text
		colorizedPos := c.findColorizedPosition(colorizedRemaining, cleanRemaining, cleanPos)

		// Extract parts from colorized text
		before := colorizedRemaining[:colorizedPos]
		match := colorizedRemaining[colorizedPos : colorizedPos+len(searchText)]
		after := colorizedRemaining[colorizedPos+len(searchText):]

		// Add the text before the match
		result += before

		// Extract the active color context from the accumulated result
		activeColor := c.extractActiveColorFromBefore(result)

		// Create highlighted version that preserves the original color
		var highlightedMatch string
		if activeColor != "" {
			// Apply highlight with original color preserved
			highlightedMatch = c.createColorPreservingHighlight(match, activeColor)
		} else {
			// No active color, use default highlighting
			highlightedMatch = lipgloss.NewStyle().Bold(true).Background(lipgloss.Color("#4A4A4A")).Render(match)
		}

		// Add the highlighted match
		result += highlightedMatch

		// If there was an active color, we need to restore it for subsequent text
		if activeColor != "" && after != "" {
			result += activeColor
		}

		// Continue with the remaining text after this match
		colorizedRemaining = after
		cleanRemaining = cleanRemaining[cleanPos+len(searchText):]
	}

	return result
}

// findJSONValueBounds finds the start and end of a JSON value containing the given position
func (c *Colorizer) findJSONValueBounds(plainText string, pos int) (int, int) {
	// Find the opening quote of the value
	valueStart := strings.LastIndex(plainText[:pos], `"`)
	if valueStart == -1 {
		return -1, -1
	}

	// Find the closing quote of the value
	valueEnd := strings.Index(plainText[pos:], `"`)
	if valueEnd == -1 {
		return -1, -1
	}
	valueEnd = pos + valueEnd + 1 // Include the closing quote

	// Verify this is actually a value (not a key)
	beforeValue := strings.TrimSpace(plainText[:valueStart])
	if !strings.HasSuffix(beforeValue, ":") {
		return -1, -1
	}

	return valueStart, valueEnd
}

// findColorizedPosition maps a position in clean text to the corresponding position in colorized text
func (c *Colorizer) findColorizedPosition(colorizedText, cleanText string, cleanPos int) int {
	if cleanPos == 0 {
		return 0
	}

	colorizedPos := 0
	cleanIndex := 0

	// Iterate through colorized text character by character
	for colorizedPos < len(colorizedText) && cleanIndex < cleanPos {
		// Check if we're at the start of an ANSI sequence
		if colorizedPos < len(colorizedText) && colorizedText[colorizedPos] == '\033' {
			// Skip the entire ANSI sequence
			ansiEnd := colorizedPos
			for ansiEnd < len(colorizedText) && colorizedText[ansiEnd] != 'm' {
				ansiEnd++
			}
			if ansiEnd < len(colorizedText) {
				ansiEnd++ // Include the 'm'
			}
			colorizedPos = ansiEnd
		} else {
			// Regular character, advance both positions
			colorizedPos++
			cleanIndex++
		}
	}

	return colorizedPos
}

// extractActiveColorFromBefore extracts the active color from the text before a match
func (c *Colorizer) extractActiveColorFromBefore(beforeText string) string {
	// Look for the last ANSI color escape sequence
	lastEscPos := strings.LastIndex(beforeText, "\x1b[")
	if lastEscPos == -1 {
		return ""
	}

	// Find the end of the escape sequence
	escSeq := beforeText[lastEscPos:]
	mPos := strings.Index(escSeq, "m")
	if mPos == -1 {
		return ""
	}

	ansiCode := escSeq[:mPos+1]

	// Skip reset codes
	if ansiCode == "\x1b[0m" || ansiCode == "\x1b[m" {
		return ""
	}

	return ansiCode
}

// createColorPreservingHighlight creates a highlight that preserves the original color
func (c *Colorizer) createColorPreservingHighlight(text, activeColor string) string {
	// Extract color information to determine appropriate background
	fgColor := c.mapAnsiToHexColor(activeColor)
	bgColor := c.computeHighlightBackground(fgColor)

	// Create highlight with original foreground color, bold, and computed background
	return activeColor + "\x1b[1m\x1b[48;5;" + c.hexToAnsi256(bgColor) + "m" + text + "\x1b[0m"
}

// hexToAnsi256 converts a hex color to ANSI 256 color code (simplified)
func (c *Colorizer) hexToAnsi256(hexColor string) string {
	// This is a simplified mapping - a full implementation would do proper color space conversion
	// For now, we'll map common background colors to ANSI 256 codes
	colorMap := map[string]string{
		"#4A1F1F": "52",  // Dark red
		"#4A4A1F": "58",  // Dark yellow
		"#1F4A4A": "23",  // Dark cyan
		"#2F4A3F": "22",  // Dark green
		"#2F2F2F": "236", // Dark gray
		"#1F2F4A": "17",  // Dark blue
		"#4A1F3F": "53",  // Dark pink
		"#3F1F3F": "54",  // Dark plum
		"#4A4A4A": "238", // Medium gray (default)
	}

	if code, exists := colorMap[hexColor]; exists {
		return code
	}
	return "238" // Default to medium gray
}

// mapAnsiToHexColor maps ANSI color codes to hex colors (simplified)
func (c *Colorizer) mapAnsiToHexColor(ansiCode string) string {
	// Common ANSI to hex mappings for our theme colors
	ansiMap := map[string]string{
		"\x1b[38;5;248m": "#A8A8A8", // Gray (typical timestamp color)
		"\x1b[38;5;75m":  "#74B9FF", // Blue
		"\x1b[38;5;203m": "#FF6B6B", // Red
		"\x1b[38;5;227m": "#FFE66D", // Yellow
		"\x1b[38;5;86m":  "#4ECDC4", // Cyan
	}

	if color, exists := ansiMap[ansiCode]; exists {
		return color
	}

	return "#F0F0F0" // Default light gray
}

// applySearchHighlightToColorizedText applies highlighting to already colorized text
func (c *Colorizer) applySearchHighlightToColorizedText(colorizedText, originalText string) string {
	if c.searchRegex != nil {
		// Find all regex matches in the original text
		matches := c.searchRegex.FindAllStringIndex(originalText, -1)
		if len(matches) == 0 {
			return colorizedText
		}

		// Apply highlighting from right to left to preserve positions
		result := colorizedText
		for i := len(matches) - 1; i >= 0; i-- {
			start, end := matches[i][0], matches[i][1]
			matchText := originalText[start:end]
			// Find and highlight this match in the colorized text
			result = c.highlightTextInColorizedString(result, matchText)
		}
		return result
	}
	if c.searchString != "" {
		// Highlight all occurrences of the search string
		return c.highlightTextInColorizedString(colorizedText, c.searchString)
	}
	return colorizedText
}

// highlightTextInColorizedString finds and highlights a specific text in colorized string
func (c *Colorizer) highlightTextInColorizedString(colorizedText, searchText string) string {
	// This is more complex because we need to find the text while ignoring ANSI codes
	// For now, we'll use a simple approach that works for most cases

	// Split the colorized text to find plain text portions
	result := colorizedText

	// Find all occurrences of the search text in the original
	for {
		// Find the next occurrence of search text
		pos := strings.Index(result, searchText)
		if pos == -1 {
			break
		}

		// Extract the matched text and apply highlighting
		before := result[:pos]
		match := result[pos : pos+len(searchText)]
		after := result[pos+len(searchText):]

		// Determine the style to use based on surrounding context
		highlightStyle := c.determineHighlightStyleFromContext(before, match, after)
		highlightedMatch := highlightStyle.Render(match)

		result = before + highlightedMatch + after

		// Move past this match to avoid infinite loop
		// We add len(highlightedMatch) instead of len(match) because the highlighted version is longer
		if len(highlightedMatch) <= len(match) {
			// Safety check - if highlighted version isn't longer, advance by original length
			result = before + highlightedMatch + after
			break
		}
	}

	return result
}

// determineHighlightStyleFromContext determines the appropriate highlight style based on context
func (c *Colorizer) determineHighlightStyleFromContext(before, match, after string) lipgloss.Style {
	// Try to extract the foreground color from the context before the match
	// This is a simplified approach - look for ANSI color codes immediately before

	// Look for the last ANSI escape sequence in the 'before' text
	lastEscPos := strings.LastIndex(before, "\x1b[")
	if lastEscPos != -1 {
		// Extract the ANSI sequence
		escSeq := before[lastEscPos:]
		mPos := strings.Index(escSeq, "m")
		if mPos != -1 {
			ansiCode := escSeq[:mPos+1]
			// Try to map this ANSI code to a color
			color := c.mapAnsiToColor(ansiCode)
			bgColor := c.computeHighlightBackground(color)
			return lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Background(lipgloss.Color(bgColor)).Bold(true)
		}
	}

	// Default highlighting if we can't determine context
	return lipgloss.NewStyle().Bold(true).Background(lipgloss.Color("#3F3F3F"))
}

// mapAnsiToColor maps ANSI color codes to hex colors (simplified)
func (c *Colorizer) mapAnsiToColor(ansiCode string) string {
	// This is a simplified mapping - a full implementation would parse all ANSI codes
	// For now, we'll return a default color
	return "#F0F0F0" // Light gray default
}

// addSearchMarkers adds special markers around search matches
func (c *Colorizer) addSearchMarkers(line string) string {
	if c.searchRegex != nil {
		// Replace all regex matches with marked versions
		return c.searchRegex.ReplaceAllStringFunc(line, func(match string) string {
			return "<<<SPLASHBOLD:" + match + ":SPLASHBOLD>>>"
		})
	}
	if c.searchString != "" {
		// Replace all string matches with marked versions
		return strings.ReplaceAll(line, c.searchString, "<<<SPLASHBOLD:"+c.searchString+":SPLASHBOLD>>>")
	}
	return line
}

// convertMarkersToFormatting converts special markers to bold formatting
func (c *Colorizer) convertMarkersToFormatting(colorizedLine string) string {
	// Find all remaining markers and replace them with bold formatting
	// (Timestamps are handled separately in applyTimestampStyleWithMarkers)
	result := colorizedLine
	for {
		start := strings.Index(result, "<<<SPLASHBOLD:")
		if start == -1 {
			break
		}
		end := strings.Index(result[start:], ":SPLASHBOLD>>>")
		if end == -1 {
			break
		}
		end += start + len(":SPLASHBOLD>>>") // Adjust for absolute position and include the marker

		// Extract the text between markers
		matchText := result[start+len("<<<SPLASHBOLD:") : end-len(":SPLASHBOLD>>>")]

		// Replace the marked text with bold version
		boldText := c.theme.SearchHighlight.Render(matchText)
		result = result[:start] + boldText + result[end:]
	}
	return result
}

// applyStyleWithMarkers applies a style while preserving search markers
func (c *Colorizer) applyStyleWithMarkers(value string, style lipgloss.Style) string {
	// Simplified: just apply the style directly since we now use unified highlighting
	return style.Render(value)
}

// createHighlightStyle creates a style for search highlighting with computed background
func (c *Colorizer) createHighlightStyle(originalStyle lipgloss.Style) lipgloss.Style {
	// Get the foreground color from the original style
	fgColor := c.extractForegroundColor(originalStyle)

	// Compute an appropriate background color
	bgColor := c.computeHighlightBackground(fgColor)

	// Create highlight style with bold, original foreground, and computed background
	return originalStyle.Bold(true).Background(lipgloss.Color(bgColor))
}

// extractForegroundColor extracts the foreground color from a lipgloss style
func (c *Colorizer) extractForegroundColor(style lipgloss.Style) string {
	// This is a bit tricky since lipgloss doesn't expose style properties directly
	// We'll use a heuristic approach by rendering a test string and checking common colors

	// Check if this matches any of our known theme colors
	testRender := style.Render("test")

	// If the style has ANSI color codes, we can try to extract them
	// For now, we'll use a mapping based on our theme colors
	return c.mapStyleToColor(style, testRender)
}

// mapStyleToColor maps a style to its likely foreground color
func (c *Colorizer) mapStyleToColor(style lipgloss.Style, rendered string) string {
	// Compare against known theme styles to determine color
	// This is a simplified approach - in a more sophisticated version,
	// we would parse ANSI codes from the rendered string

	testStyles := map[string]string{
		c.theme.Error.Render("test"):       "#FF6B6B", // Red
		c.theme.Warning.Render("test"):     "#FFE66D", // Yellow
		c.theme.Info.Render("test"):        "#4ECDC4", // Cyan
		c.theme.Debug.Render("test"):       "#95E1D3", // Light green
		c.theme.Timestamp.Render("test"):   "#A8A8A8", // Gray
		c.theme.IP.Render("test"):          "#74B9FF", // Blue
		c.theme.URL.Render("test"):         "#81ECEC", // Cyan
		c.theme.Method.Render("test"):      "#FD79A8", // Pink
		c.theme.JSONKey.Render("test"):     "#DDA0DD", // Plum
		c.theme.JSONString.Render("test"):  "#98FB98", // Pale green
		c.theme.JSONNumber.Render("test"):  "#87CEEB", // Sky blue
		c.theme.JSONValue.Render("test"):   "#F0F0F0", // Light gray
		c.theme.LogfmtKey.Render("test"):   "#DDA0DD", // Plum
		c.theme.LogfmtValue.Render("test"): "#F0F0F0", // Light gray
		c.theme.Service.Render("test"):     "#87CEEB", // Sky blue
		c.theme.Hostname.Render("test"):    "#FFB347", // Peach
		c.theme.PID.Render("test"):         "#B19CD9", // Lavender
		c.theme.Filename.Render("test"):    "#FFA07A", // Light salmon
		c.theme.LineNum.Render("test"):     "#B19CD9", // Lavender
		c.theme.StatusOK.Render("test"):    "#6BCF7F", // Green
		c.theme.StatusWarn.Render("test"):  "#FFD93D", // Yellow
		c.theme.StatusError.Render("test"): "#FF6B6B", // Red
	}

	for styleRender, color := range testStyles {
		if styleRender == rendered {
			return color
		}
	}

	// Default to a neutral color if we can't determine the foreground
	return "#F0F0F0" // Light gray
}

// computeHighlightBackground computes an appropriate background color for highlighting
func (c *Colorizer) computeHighlightBackground(fgColor string) string {
	// Define background colors that work well with different foreground colors
	colorMap := map[string]string{
		"#FF6B6B": "#4A1F1F", // Red text -> Dark red background
		"#FFE66D": "#4A4A1F", // Yellow text -> Dark yellow background
		"#4ECDC4": "#1F4A4A", // Cyan text -> Dark cyan background
		"#95E1D3": "#2F4A3F", // Light green text -> Dark green background
		"#A8A8A8": "#2F2F2F", // Gray text -> Dark gray background
		"#74B9FF": "#1F2F4A", // Blue text -> Dark blue background
		"#81ECEC": "#1F4A4A", // Cyan text -> Dark cyan background
		"#FD79A8": "#4A1F3F", // Pink text -> Dark pink background
		"#DDA0DD": "#3F1F3F", // Plum text -> Dark plum background
		"#98FB98": "#2F4A2F", // Pale green text -> Dark green background
		"#87CEEB": "#1F3F4A", // Sky blue text -> Dark blue background
		"#FFB347": "#4A3F1F", // Peach text -> Dark orange background
		"#F0F0F0": "#3F3F3F", // Light gray text -> Dark gray background
		"#B19CD9": "#2F1F3F", // Lavender text -> Dark purple background
		"#FFA07A": "#4A2F1F", // Light salmon text -> Dark salmon background
		"#6BCF7F": "#1F3F2F", // Green text -> Dark green background
		"#FFD93D": "#4A4A1F", // Yellow text -> Dark yellow background
	}

	if bgColor, exists := colorMap[fgColor]; exists {
		return bgColor
	}

	// If we don't have a specific mapping, try to compute a darker version
	return c.computeDarkerVersion(fgColor)
}

// computeDarkerVersion computes a darker version of a hex color for background
func (c *Colorizer) computeDarkerVersion(hexColor string) string {
	// Remove # if present
	hex := strings.TrimPrefix(hexColor, "#")

	// Parse RGB components
	if len(hex) != 6 {
		return "#3F3F3F" // Default dark gray
	}

	r, err1 := strconv.ParseInt(hex[0:2], 16, 64)
	g, err2 := strconv.ParseInt(hex[2:4], 16, 64)
	b, err3 := strconv.ParseInt(hex[4:6], 16, 64)

	if err1 != nil || err2 != nil || err3 != nil {
		return "#3F3F3F" // Default dark gray
	}

	// Make it much darker (divide by 4) for background
	r = r / 4
	g = g / 4
	b = b / 4

	// Ensure minimum darkness
	if r < 0x1F {
		r = 0x1F
	}
	if g < 0x1F {
		g = 0x1F
	}
	if b < 0x1F {
		b = 0x1F
	}

	return fmt.Sprintf("#%02X%02X%02X", r, g, b)
}
