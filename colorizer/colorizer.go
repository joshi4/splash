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
	theme         *ColorTheme
	searchString  string
	searchRegex   *regexp.Regexp
}

// NewColorizer creates a new colorizer with adaptive theming
func NewColorizer() *Colorizer {
	return &Colorizer{
		theme: NewAdaptiveTheme(),
	}
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

// ColorizeLog applies colors to a log line based on its detected format
func (c *Colorizer) ColorizeLog(line string, format parser.LogFormat) string {
	if line == "" {
		return line
	}

	// Check if this line matches the search pattern
	if c.matchesSearch(line) {
		// Apply search highlighting to the entire line
		return c.applySearchHighlight(line, format)
	}

	switch format {
	case parser.JSONFormat:
		return c.colorizeJSON(line)
	case parser.LogfmtFormat:
		return c.colorizeLogfmt(line)
	case parser.ApacheCommonFormat:
		return c.colorizeApacheCommon(line)
	case parser.NginxFormat:
		return c.colorizeNginx(line)
	case parser.SyslogFormat:
		return c.colorizeSyslog(line)
	case parser.GoStandardFormat:
		return c.colorizeGoStandard(line)
	case parser.RailsFormat:
		return c.colorizeRails(line)
	case parser.DockerFormat:
		return c.colorizeDocker(line)
	case parser.KubernetesFormat:
		return c.colorizeKubernetes(line)
	case parser.HerokuFormat:
		return c.colorizeHeroku(line)
	default:
		// For unknown formats, still apply search highlighting if there's a match
		if c.matchesSearch(line) {
			return c.simpleSearchHighlight(line, line)
		}
		return line // No coloring for unknown formats
	}
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
		
		// Special handling for common log fields
		keyStyle := c.theme.JSONKey
		if c.isLogLevelKey(key) {
			keyStyle = c.theme.GetLogLevelStyle(key)
		}
		result.WriteString(keyStyle.Render(key))
		result.WriteString(c.theme.Quote.Render(`"`))
		result.WriteString(c.theme.Equals.Render(":"))
		
		// Colorize value based on key and type
		result.WriteString(c.colorizeJSONValue(key, value))
	}
	
	result.WriteString(c.theme.Bracket.Render("}"))
	return result.String()
}

// colorizeJSONValue colors a JSON value based on context and type
func (c *Colorizer) colorizeJSONValue(key string, value interface{}) string {
	switch v := value.(type) {
	case string:
		// Special handling for known fields
		if c.isLogLevelKey(key) {
			return c.theme.Quote.Render(`"`) + c.theme.GetLogLevelStyle(v).Render(v) + c.theme.Quote.Render(`"`)
		}
		if c.isTimestampKey(key) {
			return c.theme.Quote.Render(`"`) + c.theme.Timestamp.Render(v) + c.theme.Quote.Render(`"`)
		}
		if c.isServiceKey(key) {
			return c.theme.Quote.Render(`"`) + c.theme.Service.Render(v) + c.theme.Quote.Render(`"`)
		}
		return c.theme.Quote.Render(`"`) + c.theme.JSONString.Render(v) + c.theme.Quote.Render(`"`)
	case float64:
		numberStr := fmt.Sprintf("%g", v)
		styledNumber := c.applyStyleWithMarkers(numberStr, c.theme.JSONNumber)
		return styledNumber
	case bool:
		if v {
			styledBool := c.applyStyleWithMarkers("true", c.theme.StatusOK)
			return styledBool
		}
		styledBool := c.applyStyleWithMarkers("false", c.theme.StatusWarn)
		return styledBool
	default:
		// Fallback to JSON marshaling
		jsonBytes, _ := json.Marshal(v)
		styledValue := c.applyStyleWithMarkers(string(jsonBytes), c.theme.JSONValue)
		return styledValue
	}
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
				
				// Color the key
				keyStyle := c.theme.LogfmtKey
				if c.isLogLevelKey(key) {
					keyStyle = c.theme.GetLogLevelStyle(key)
				}
				styledKey := c.applyStyleWithMarkers(key, keyStyle)
				result.WriteString(styledKey)
				result.WriteString(c.theme.Equals.Render("="))
				
				// Color the value based on key
				if c.isLogLevelKey(key) {
					if strings.HasPrefix(value, `"`) && strings.HasSuffix(value, `"`) {
						result.WriteString(c.theme.Quote.Render(`"`))
						styledValue := c.applyStyleWithMarkers(cleanValue, c.theme.GetLogLevelStyle(cleanValue))
						result.WriteString(styledValue)
						result.WriteString(c.theme.Quote.Render(`"`))
					} else {
						styledValue := c.applyStyleWithMarkers(value, c.theme.GetLogLevelStyle(cleanValue))
						result.WriteString(styledValue)
					}
				} else if c.isTimestampKey(key) {
					styledValue := c.applyStyleWithMarkers(value, c.theme.Timestamp)
					result.WriteString(styledValue)
				} else if c.isServiceKey(key) {
					styledValue := c.applyStyleWithMarkers(value, c.theme.Service)
					result.WriteString(styledValue)
				} else {
					styledValue := c.applyStyleWithMarkers(value, c.theme.LogfmtValue)
					result.WriteString(styledValue)
				}
			} else {
				result.WriteString(part)
			}
		} else {
			// Not a key=value pair, check if it's a log level
			if c.looksLikeLogLevel(part) {
				result.WriteString(c.theme.GetLogLevelStyle(part).Render(part))
			} else {
				result.WriteString(part)
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
		return line // Fallback if regex doesn't match
	}
	
	ip := matches[1]
	timestamp := matches[4]
	method := matches[5]
	url := matches[6]
	protocol := matches[7]
	status := matches[8]
	size := matches[9]
	
	result := strings.Builder{}
	result.WriteString(c.applyStyleWithMarkers(ip, c.theme.IP))
	result.WriteString(" - - ")
	result.WriteString(c.theme.Bracket.Render("["))
	result.WriteString(c.applyStyleWithMarkers(timestamp, c.theme.Timestamp))
	result.WriteString(c.theme.Bracket.Render("] "))
	result.WriteString(c.theme.Quote.Render(`"`))
	result.WriteString(c.applyStyleWithMarkers(method, c.theme.Method))
	result.WriteString(" ")
	result.WriteString(c.applyStyleWithMarkers(url, c.theme.URL))
	result.WriteString(" ")
	result.WriteString(protocol)
	result.WriteString(c.theme.Quote.Render(`" `))
	result.WriteString(c.applyStyleWithMarkers(status, c.theme.GetHTTPStatusStyle(status)))
	result.WriteString(" ")
	result.WriteString(size)
	
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
	result.WriteString(c.applyStyleWithMarkers(ip, c.theme.IP))
	result.WriteString(" - - ")
	result.WriteString(c.theme.Bracket.Render("["))
	result.WriteString(c.applyStyleWithMarkers(timestamp, c.theme.Timestamp))
	result.WriteString(c.theme.Bracket.Render("] "))
	result.WriteString(c.theme.Quote.Render(`"`))
	result.WriteString(c.applyStyleWithMarkers(method, c.theme.Method))
	result.WriteString(" ")
	result.WriteString(c.applyStyleWithMarkers(url, c.theme.URL))
	result.WriteString(" ")
	result.WriteString(protocol)
	result.WriteString(c.theme.Quote.Render(`" `))
	result.WriteString(c.applyStyleWithMarkers(status, c.theme.GetHTTPStatusStyle(status)))
	result.WriteString(" ")
	result.WriteString(size)
	result.WriteString(" ")
	result.WriteString(c.theme.Quote.Render(`"`))
	result.WriteString(referer)
	result.WriteString(c.theme.Quote.Render(`" "`))
	result.WriteString(userAgent)
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
	result.WriteString(c.applyStyleWithMarkers(timestamp, c.theme.Timestamp))
	result.WriteString(" ")
	result.WriteString(c.applyStyleWithMarkers(hostname, c.theme.Hostname))
	result.WriteString(" ")
	result.WriteString(c.applyStyleWithMarkers(process, c.theme.Service))
	result.WriteString(c.theme.Bracket.Render("["))
	result.WriteString(c.applyStyleWithMarkers(pid, c.theme.PID))
	result.WriteString(c.theme.Bracket.Render("]: "))
	result.WriteString(c.colorizeMessage(message))
	
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
	result.WriteString(c.applyStyleWithMarkers(timestamp, c.theme.Timestamp))
	result.WriteString(" ")
	result.WriteString(c.colorizeMessage(message))
	
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
		timestampContent := timestamp[1:len(timestamp)-1] // Remove brackets
		result.WriteString(c.applyStyleWithMarkers(timestampContent, c.theme.Timestamp))
		result.WriteString(c.theme.Bracket.Render("] "))
		result.WriteString(c.applyStyleWithMarkers(level, c.theme.GetLogLevelStyle(level)))
		result.WriteString(" ")
		result.WriteString(separator)
		result.WriteString(" : ")
		result.WriteString(message)
		
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
		timestampContent := timestamp[1:len(timestamp)-1] // Remove brackets
		result.WriteString(c.applyStyleWithMarkers(timestampContent, c.theme.Timestamp))
		result.WriteString(c.theme.Bracket.Render("] "))
		result.WriteString(c.applyStyleWithMarkers(level, c.theme.GetLogLevelStyle(level)))
		result.WriteString(" ")
		result.WriteString(c.applyStyleWithMarkers(message, c.theme.JSONValue))
		
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
	result.WriteString(c.applyStyleWithMarkers(timestamp, c.theme.Timestamp))
	result.WriteString(" ")
	result.WriteString(c.applyStyleWithMarkers(level, c.theme.GetLogLevelStyle(level)))
	result.WriteString(" ")
	result.WriteString(message)
	
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
	result.WriteString(c.applyStyleWithMarkers(timestamp, c.theme.Timestamp))
	result.WriteString(" ")
	result.WriteString(c.applyStyleWithMarkers(severity, c.theme.PID))
	result.WriteString(" ")
	result.WriteString(c.applyStyleWithMarkers(filename, c.theme.Filename))
	result.WriteString(":")
	result.WriteString(c.applyStyleWithMarkers(lineNum, c.theme.LineNum))
	result.WriteString(c.theme.Bracket.Render("] "))
	result.WriteString(c.colorizeMessage(message))
	
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
	result.WriteString(c.applyStyleWithMarkers(timestamp, c.theme.Timestamp))
	result.WriteString(" app")
	result.WriteString(c.theme.Bracket.Render("["))
	result.WriteString(c.applyStyleWithMarkers(dyno, c.theme.Service))
	result.WriteString(c.theme.Bracket.Render("]: "))
	result.WriteString(c.colorizeMessage(message))
	
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
			styledLevel := c.applyStyleWithMarkers(cleanWord, c.theme.GetLogLevelStyle(cleanWord))
			result.WriteString(styledLevel)
			result.WriteString(":")
		} else {
			styledLevel := c.applyStyleWithMarkers(cleanWord, c.theme.GetLogLevelStyle(cleanWord))
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
				styledLevel := c.applyStyleWithMarkers(cleanWord, c.theme.GetLogLevelStyle(cleanWord))
				result.WriteString(styledLevel)
				result.WriteString(":")
			} else {
				styledLevel := c.applyStyleWithMarkers(cleanWord, c.theme.GetLogLevelStyle(cleanWord))
				result.WriteString(styledLevel)
			}
		} else {
			result.WriteString(word)
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

// matchesSearch checks if a line matches the current search pattern (ignoring ANSI codes)
func (c *Colorizer) matchesSearch(line string) bool {
	// Strip ANSI codes before checking for matches
	cleanLine := c.stripAnsiCodes(line)
	
	if c.searchRegex != nil {
		return c.searchRegex.MatchString(cleanLine)
	}
	if c.searchString != "" {
		return strings.Contains(cleanLine, c.searchString)
	}
	return false
}

// applySearchHighlight applies prominent highlighting to matching lines
func (c *Colorizer) applySearchHighlight(line string, format parser.LogFormat) string {
	// Apply normal colorization to the original line first
	var colorizedLine string
	switch format {
	case parser.JSONFormat:
		colorizedLine = c.colorizeJSON(line)
	case parser.LogfmtFormat:
		colorizedLine = c.colorizeLogfmt(line)
	case parser.ApacheCommonFormat:
		colorizedLine = c.colorizeApacheCommon(line)
	case parser.NginxFormat:
		colorizedLine = c.colorizeNginx(line)
	case parser.SyslogFormat:
		colorizedLine = c.colorizeSyslog(line)
	case parser.GoStandardFormat:
		colorizedLine = c.colorizeGoStandard(line)
	case parser.RailsFormat:
		colorizedLine = c.colorizeRails(line)
	case parser.DockerFormat:
		colorizedLine = c.colorizeDocker(line)
	case parser.KubernetesFormat:
		colorizedLine = c.colorizeKubernetes(line)
	case parser.HerokuFormat:
		colorizedLine = c.colorizeHeroku(line)
	default:
		colorizedLine = line
	}
	
	// Apply simple search highlighting to the colorized result
	return c.simpleSearchHighlight(colorizedLine, line)
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
	
	result := ""
	remaining := colorizedText
	
	for {
		// Find the next occurrence of the search text
		pos := strings.Index(remaining, searchText)
		if pos == -1 {
			// No more matches, append remaining text and break
			result += remaining
			break
		}
		
		// Extract parts: before match, match, after match
		before := remaining[:pos]
		match := remaining[pos:pos+len(searchText)]
		after := remaining[pos+len(searchText):]
		
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
		remaining = after
	}
	
	return result
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
		"#4A1F1F": "52",   // Dark red
		"#4A4A1F": "58",   // Dark yellow
		"#1F4A4A": "23",   // Dark cyan
		"#2F4A3F": "22",   // Dark green
		"#2F2F2F": "236",  // Dark gray
		"#1F2F4A": "17",   // Dark blue
		"#4A1F3F": "53",   // Dark pink
		"#3F1F3F": "54",   // Dark plum
		"#4A4A4A": "238",  // Medium gray (default)
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
		match := result[pos:pos+len(searchText)]
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
	// Check if the value contains search markers
	if !strings.Contains(value, "<<<SPLASHBOLD:") {
		// No markers, apply normal styling
		return style.Render(value)
	}
	
	// Split the value at marker boundaries and apply styling appropriately
	result := ""
	remaining := value
	
	for {
		markerStart := strings.Index(remaining, "<<<SPLASHBOLD:")
		if markerStart == -1 {
			// No more markers, apply style to remaining text
			if remaining != "" {
				result += style.Render(remaining)
			}
			break
		}
		
		// Apply style to text before marker
		if markerStart > 0 {
			result += style.Render(remaining[:markerStart])
		}
		
		// Find the end of the marker
		markerEnd := strings.Index(remaining[markerStart:], ":SPLASHBOLD>>>")
		if markerEnd == -1 {
			// Malformed marker, just style the rest normally
			result += style.Render(remaining[markerStart:])
			break
		}
		markerEnd += markerStart + len(":SPLASHBOLD>>>")
		
		// Extract the matched text from the marker
		matchText := remaining[markerStart+len("<<<SPLASHBOLD:") : markerEnd-len(":SPLASHBOLD>>>")]
		
		// Combine original style with bold and computed background
		highlightStyle := c.createHighlightStyle(style)
		result += highlightStyle.Render(matchText)
		
		// Continue with remaining text
		remaining = remaining[markerEnd:]
	}
	
	return result
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
		c.theme.Error.Render("test"):      "#FF6B6B", // Red
		c.theme.Warning.Render("test"):    "#FFE66D", // Yellow  
		c.theme.Info.Render("test"):       "#4ECDC4", // Cyan
		c.theme.Debug.Render("test"):      "#95E1D3", // Light green
		c.theme.Timestamp.Render("test"):  "#A8A8A8", // Gray
		c.theme.IP.Render("test"):         "#74B9FF", // Blue
		c.theme.URL.Render("test"):        "#81ECEC", // Cyan
		c.theme.Method.Render("test"):     "#FD79A8", // Pink
		c.theme.JSONKey.Render("test"):    "#DDA0DD", // Plum
		c.theme.JSONString.Render("test"): "#98FB98", // Pale green
		c.theme.JSONNumber.Render("test"): "#87CEEB", // Sky blue
		c.theme.JSONValue.Render("test"):  "#F0F0F0", // Light gray
		c.theme.LogfmtKey.Render("test"):  "#DDA0DD", // Plum
		c.theme.LogfmtValue.Render("test"): "#F0F0F0", // Light gray
		c.theme.Service.Render("test"):    "#87CEEB", // Sky blue
		c.theme.Hostname.Render("test"):   "#FFB347", // Peach
		c.theme.PID.Render("test"):        "#B19CD9", // Lavender
		c.theme.Filename.Render("test"):   "#FFA07A", // Light salmon
		c.theme.LineNum.Render("test"):    "#B19CD9", // Lavender
		c.theme.StatusOK.Render("test"):   "#6BCF7F", // Green
		c.theme.StatusWarn.Render("test"): "#FFD93D", // Yellow
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
	if r < 0x1F { r = 0x1F }
	if g < 0x1F { g = 0x1F }
	if b < 0x1F { b = 0x1F }
	
	return fmt.Sprintf("#%02X%02X%02X", r, g, b)
}
