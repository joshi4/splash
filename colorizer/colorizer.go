package colorizer

import (
	"encoding/json"
	"fmt"
	"regexp"
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
			styledValue := c.applyStyleWithMarkers(v, c.theme.GetLogLevelStyle(v))
			return c.theme.Quote.Render(`"`) + styledValue + c.theme.Quote.Render(`"`)
		}
		if c.isTimestampKey(key) {
			styledValue := c.applyStyleWithMarkers(v, c.theme.Timestamp)
			return c.theme.Quote.Render(`"`) + styledValue + c.theme.Quote.Render(`"`)
		}
		if c.isServiceKey(key) {
			styledValue := c.applyStyleWithMarkers(v, c.theme.Service)
			return c.theme.Quote.Render(`"`) + styledValue + c.theme.Quote.Render(`"`)
		}
		styledValue := c.applyStyleWithMarkers(v, c.theme.JSONString)
		return c.theme.Quote.Render(`"`) + styledValue + c.theme.Quote.Render(`"`)
	case float64:
		return c.theme.JSONNumber.Render(fmt.Sprintf("%g", v))
	case bool:
		if v {
			return c.theme.StatusOK.Render("true")
		}
		return c.theme.StatusWarn.Render("false")
	default:
		// Fallback to JSON marshaling
		jsonBytes, _ := json.Marshal(v)
		return c.theme.JSONValue.Render(string(jsonBytes))
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
				result.WriteString(keyStyle.Render(key))
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
	result.WriteString(c.theme.IP.Render(ip))
	result.WriteString(" - - ")
	result.WriteString(c.theme.Bracket.Render("["))
	result.WriteString(c.theme.Timestamp.Render(timestamp))
	result.WriteString(c.theme.Bracket.Render("] "))
	result.WriteString(c.theme.Quote.Render(`"`))
	result.WriteString(c.theme.Method.Render(method))
	result.WriteString(" ")
	result.WriteString(c.theme.URL.Render(url))
	result.WriteString(" ")
	result.WriteString(protocol)
	result.WriteString(c.theme.Quote.Render(`" `))
	result.WriteString(c.theme.GetHTTPStatusStyle(status).Render(status))
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
	re := regexp.MustCompile(`^(\[[^\]]+\]) (\w+) (--) : (.*)`)
	matches := re.FindStringSubmatch(line)
	
	if len(matches) != 5 {
		return c.colorizeGenericLog(line)
	}
	
	timestamp := matches[1]
	level := matches[2]
	separator := matches[3]
	message := matches[4]
	
	result := strings.Builder{}
	result.WriteString(c.theme.Bracket.Render("["))
	result.WriteString(c.theme.Timestamp.Render(timestamp[1:len(timestamp)-1])) // Remove brackets
	result.WriteString(c.theme.Bracket.Render("] "))
	result.WriteString(c.theme.GetLogLevelStyle(level).Render(level))
	result.WriteString(" ")
	result.WriteString(separator)
	result.WriteString(" : ")
	result.WriteString(message)
	
	return result.String()
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
	result.WriteString(c.theme.Timestamp.Render(timestamp))
	result.WriteString(" ")
	result.WriteString(c.theme.GetLogLevelStyle(level).Render(level))
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
	result.WriteString(c.theme.Timestamp.Render(timestamp))
	result.WriteString(" ")
	result.WriteString(c.theme.PID.Render(severity))
	result.WriteString(" ")
	result.WriteString(c.theme.Filename.Render(filename))
	result.WriteString(":")
	result.WriteString(c.theme.LineNum.Render(lineNum))
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
	result.WriteString(c.theme.Timestamp.Render(timestamp))
	result.WriteString(" app")
	result.WriteString(c.theme.Bracket.Render("["))
	result.WriteString(c.theme.Service.Render(dyno))
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
				result.WriteString(c.theme.GetLogLevelStyle(cleanWord).Render(cleanWord))
				result.WriteString(":")
			} else {
				result.WriteString(c.theme.GetLogLevelStyle(cleanWord).Render(cleanWord))
			}
		} else {
			result.WriteString(word)
		}
	}
	
	return result.String()
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

// applySearchHighlight applies prominent highlighting to matching lines
func (c *Colorizer) applySearchHighlight(line string, format parser.LogFormat) string {
	// First, add search highlighting markers to the original line
	lineWithMarkers := c.addSearchMarkers(line)
	
	// Apply normal colorization to the line with markers
	var colorizedLine string
	switch format {
	case parser.JSONFormat:
		colorizedLine = c.colorizeJSON(lineWithMarkers)
	case parser.LogfmtFormat:
		colorizedLine = c.colorizeLogfmt(lineWithMarkers)
	case parser.ApacheCommonFormat:
		colorizedLine = c.colorizeApacheCommon(lineWithMarkers)
	case parser.NginxFormat:
		colorizedLine = c.colorizeNginx(lineWithMarkers)
	case parser.SyslogFormat:
		colorizedLine = c.colorizeSyslog(lineWithMarkers)
	case parser.GoStandardFormat:
		colorizedLine = c.colorizeGoStandard(lineWithMarkers)
	case parser.RailsFormat:
		colorizedLine = c.colorizeRails(lineWithMarkers)
	case parser.DockerFormat:
		colorizedLine = c.colorizeDocker(lineWithMarkers)
	case parser.KubernetesFormat:
		colorizedLine = c.colorizeKubernetes(lineWithMarkers)
	case parser.HerokuFormat:
		colorizedLine = c.colorizeHeroku(lineWithMarkers)
	default:
		colorizedLine = lineWithMarkers
	}
	
	// Convert markers to actual bold formatting
	return c.convertMarkersToFormatting(colorizedLine)
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
		
		// Combine original style with bold
		combinedStyle := style.Bold(true)
		result += combinedStyle.Render(matchText)
		
		// Continue with remaining text
		remaining = remaining[markerEnd:]
	}
	
	return result
}
