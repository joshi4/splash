package parser

import (
	"encoding/json"
	"regexp"
	"strings"
)

// LogFormat represents the detected log format
type LogFormat int

const (
	UnknownFormat LogFormat = iota
	JSONFormat
	LogfmtFormat
	ApacheCommonFormat
	NginxFormat
	SyslogFormat
    RsyslogFormat
	GoStandardFormat
	RailsFormat
	DockerFormat
	KubernetesFormat
	HerokuFormat
	GoTestFormat
)

// String returns the string representation of the log format
func (f LogFormat) String() string {
	switch f {
	case JSONFormat:
		return "JSON"
	case LogfmtFormat:
		return "Logfmt"
	case ApacheCommonFormat:
		return "Apache Common"
	case NginxFormat:
		return "Nginx"
	case SyslogFormat:
		return "Syslog"
    case RsyslogFormat:
        return "Rsyslog"
	case GoStandardFormat:
		return "Go Standard"
	case RailsFormat:
		return "Rails"
	case DockerFormat:
		return "Docker"
	case KubernetesFormat:
		return "Kubernetes"
	case HerokuFormat:
		return "Heroku"
	case GoTestFormat:
		return "Go Test"
	default:
		return "Unknown"
	}
}

var (
	// Legacy regex patterns - kept for backward compatibility
	legacyApacheCommonRegex = regexp.MustCompile(`^\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3} - - \[\d{2}\/\w{3}\/\d{4}:\d{2}:\d{2}:\d{2} [+-]\d{4}\] "[A-Z]+ .* HTTP\/\d\.\d" \d{3} \d+$`)
	legacyNginxRegex        = regexp.MustCompile(`^\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3} - - \[\d{2}\/\w{3}\/\d{4}:\d{2}:\d{2}:\d{2} [+-]\d{4}\] "[A-Z]+ .* HTTP\/\d\.\d" \d{3} \d+ ".*" ".*"$`)
	legacySyslogRegex       = regexp.MustCompile(`^\w{3} \d{1,2} \d{2}:\d{2}:\d{2} \S+ \S+\[\d+\]:`)
	legacyGoStandardRegex   = regexp.MustCompile(`^\d{4}\/\d{2}\/\d{2} \d{2}:\d{2}:\d{2}`)
	legacyRailsRegex        = regexp.MustCompile(`^\[\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}\] \w+ --`)
	legacyDockerRegex       = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d+Z`)
	legacyKubernetesRegex   = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d+Z \d+ \S+:\d+\]`)
	legacyHerokuRegex       = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}[+-]\d{2}:\d{2} app\[\S+\]:`)
	legacyGoTestRegex       = regexp.MustCompile(`^(=== RUN|--- PASS:|--- FAIL:|--- SKIP:|=== NAME|=== CONT|\? .* \[no test files\]|PASS$|FAIL$|ok .* [\d\.]+[a-z]*$|FAIL .*)`)
)

// DetectFormat analyzes a log line and returns the detected format
func DetectFormat(line string) LogFormat {
	line = strings.TrimSpace(line)

	if line == "" {
		return UnknownFormat
	}

	// Check for JSON format
	if isJSONFormat(line) {
		return JSONFormat
	}

	// Check for logfmt format
	if isLogfmtFormat(line) {
		return LogfmtFormat
	}

	// Check for GoTest format first (highly specific patterns)
	if legacyGoTestRegex.MatchString(line) {
		return GoTestFormat
	}

	// Check regex patterns in order of specificity
	if legacyKubernetesRegex.MatchString(line) {
		return KubernetesFormat
	}

	if legacyHerokuRegex.MatchString(line) {
		return HerokuFormat
	}

	if legacyDockerRegex.MatchString(line) {
		return DockerFormat
	}

	if legacyNginxRegex.MatchString(line) {
		return NginxFormat
	}

	if legacyApacheCommonRegex.MatchString(line) {
		return ApacheCommonFormat
	}

	if legacyRailsRegex.MatchString(line) {
		return RailsFormat
	}

	if legacySyslogRegex.MatchString(line) {
		return SyslogFormat
	}

	if legacyGoStandardRegex.MatchString(line) {
		return GoStandardFormat
	}

	return UnknownFormat
}

// isJSONFormat checks if the line is valid JSON
func isJSONFormat(line string) bool {
	var js json.RawMessage
	return json.Unmarshal([]byte(line), &js) == nil
}

// isLogfmtFormat checks if the line follows logfmt key=value format
func isLogfmtFormat(line string) bool {
	// Simple heuristic: look for key=value patterns
	// Should contain at least one key=value pair and mostly consist of such pairs
	parts := strings.Fields(line)
	if len(parts) == 0 {
		return false
	}

	kvPairs := 0
	for _, part := range parts {
		if strings.Contains(part, "=") && !strings.HasPrefix(part, "=") && !strings.HasSuffix(part, "=") {
			kvPairs++
		}
	}

	// Consider it logfmt if more than half the parts are key=value pairs
	return kvPairs > 0 && float64(kvPairs)/float64(len(parts)) > 0.5
}
