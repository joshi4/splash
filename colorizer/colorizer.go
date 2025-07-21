package colorizer

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/joshi4/splash/parser"
)

const (
	// Light gray color used for default values
	defaultLightGray = "#F0F0F0"
	// Log level constants
	errorLevel    = "ERROR"
	warnLevel     = "WARN"
	warningLevel  = "WARNING"
	infoLevel     = "INFO"
	debugLevel    = "DEBUG"
	traceLevel    = "TRACE"
	fatalLevel    = "FATAL"
	criticalLevel = "CRITICAL"
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

// SearchMatch represents a found search match with its position
type SearchMatch struct {
	start int
	end   int
	text  string
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
				switch {
				case c.isLogLevelKey(key):
					if strings.HasPrefix(value, `"`) && strings.HasSuffix(value, `"`) {
						result.WriteString(c.theme.Quote.Render(`"`))
						result.WriteString(c.applySearchHighlighting(cleanValue, c.theme.GetLogLevelStyle(cleanValue)))
						result.WriteString(c.theme.Quote.Render(`"`))
					} else {
						result.WriteString(c.applySearchHighlighting(value, c.theme.GetLogLevelStyle(cleanValue)))
					}
				case c.isTimestampKey(key):
					result.WriteString(c.applySearchHighlighting(value, c.theme.Timestamp))
				case c.isServiceKey(key):
					result.WriteString(c.applySearchHighlighting(value, c.theme.Service))
				default:
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
	return upper == errorLevel || upper == warnLevel || upper == warningLevel ||
		upper == infoLevel || upper == debugLevel || upper == traceLevel ||
		upper == fatalLevel || upper == criticalLevel
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
