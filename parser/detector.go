package parser

import (
	"context"
	"encoding/json"
	"regexp"
	"strings"
	"sync"
	"time"
)

// FormatDetector defines the interface for detecting log formats
type FormatDetector interface {
	Detect(ctx context.Context, line string) bool
	Format() LogFormat
	Specificity() int   // Basic tier (non-regex formats get manual scoring)
	PatternLength() int // Length of regex pattern for tie-breaking
}

// DetectionResult holds the result of format detection
type DetectionResult struct {
	Format        LogFormat
	Detected      bool
	Specificity   int
	PatternLength int
}

// Parser handles log format detection with optimization for repeated formats
type Parser struct {
	detectors        []FormatDetector
	previousFormat   LogFormat
	previousDetector FormatDetector
	mu               sync.RWMutex
}

// NewParser creates a new optimized parser with all supported detectors
func NewParser() *Parser {
	return &Parser{
		detectors: []FormatDetector{
			&JSONDetector{},
			&LogfmtDetector{},
			&JavaExceptionDetector{},   // High priority for Java exception headers
			&PythonExceptionDetector{}, // High priority for Python traceback headers
			&GoTestDetector{},          // High priority for specific go test patterns
			&KubernetesDetector{},      // Must be before DockerDetector
			&HerokuDetector{},
			&RsyslogDetector{}, // Before generic Syslog to be more specific
			&NginxDetector{},   // Must be before ApacheCommonDetector
			&ApacheCommonDetector{},
			&DockerDetector{},
			&RailsDetector{},
			&SyslogDetector{},
			&GoStandardDetector{},
		},
		previousFormat: UnknownFormat,
	}
}

// DetectFormat detects the log format for a given line with optimization
func (p *Parser) DetectFormat(line string) LogFormat {
	// Try previous detector first if we have one
	p.mu.RLock()
	previousDetector := p.previousDetector
	p.mu.RUnlock()

	if previousDetector != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()

		if previousDetector.Detect(ctx, line) {
			return previousDetector.Format()
		}
	}

	// Previous detector failed or doesn't exist, try all detectors concurrently
	return p.detectAllFormats(line)
}

// detectAllFormats runs all detectors concurrently and returns the most specific match
func (p *Parser) detectAllFormats(line string) LogFormat {
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	resultChan := make(chan DetectionResult, len(p.detectors))

	// Start all detectors concurrently
	for _, detector := range p.detectors {
		go func(d FormatDetector) {
			detected := d.Detect(ctx, line)
			resultChan <- DetectionResult{
				Format:        d.Format(),
				Detected:      detected,
				Specificity:   d.Specificity(),
				PatternLength: d.PatternLength(),
			}
		}(detector)
	}

	// Collect all results and find the most specific match
	var matches []DetectionResult
LOOP:
	for range len(p.detectors) {
		select {
		case result := <-resultChan:
			if result.Detected {
				matches = append(matches, result)
			}
		case <-ctx.Done():
			break LOOP
		}
	}

	if len(matches) == 0 {
		// Handle continuation lines: stick with previous format only if appropriate
		p.mu.RLock()
		prev := p.previousFormat
		p.mu.RUnlock()

		if prev == RsyslogFormat {
			return RsyslogFormat
		}
		if prev == SyslogFormat {
			return SyslogFormat
		}
		if prev == JavaExceptionFormat {
			// For Java exceptions, only continue if line has leading whitespace
			// Lines without leading whitespace that don't match exception headers should end the exception
			if len(line) > 0 && (line[0] == ' ' || line[0] == '\t') {
				return JavaExceptionFormat
			}
			// Line has no leading whitespace and didn't match exception header - end exception
		}
		if prev == PythonExceptionFormat {
			// For Python exceptions, only continue if line has leading whitespace
			// Lines without leading whitespace that don't match headers should end the exception
			if len(line) > 0 && (line[0] == ' ' || line[0] == '\t') {
				return PythonExceptionFormat
			}
			// Line has no leading whitespace and didn't match traceback/exception header - end exception
		}
		return UnknownFormat
	}

	// Find the most specific match using specificity, then pattern length, then format order
	bestMatch := matches[0]
	for _, match := range matches[1:] {
		if match.Specificity > bestMatch.Specificity {
			bestMatch = match
		} else if match.Specificity == bestMatch.Specificity {
			// First tie-breaker: longer regex pattern is more specific
			if match.PatternLength > bestMatch.PatternLength {
				bestMatch = match
			} else if match.PatternLength == bestMatch.PatternLength {
				// Final tie-breaker: use format enum order (lower enum value = higher priority)
				if match.Format < bestMatch.Format {
					bestMatch = match
				}
			}
		}
	}

	// Update previous format and detector
	p.mu.Lock()
	p.previousFormat = bestMatch.Format
	for _, detector := range p.detectors {
		if detector.Format() == bestMatch.Format {
			p.previousDetector = detector
			break
		}
	}
	p.mu.Unlock()

	return bestMatch.Format
}

// Individual detector implementations

type JSONDetector struct{}

func (d *JSONDetector) Detect(ctx context.Context, line string) bool {
	done := make(chan bool, 1)
	go func() {
		var js json.RawMessage
		done <- json.Unmarshal([]byte(line), &js) == nil
	}()

	select {
	case result := <-done:
		return result
	case <-ctx.Done():
		return false
	}
}

func (d *JSONDetector) Format() LogFormat {
	return JSONFormat
}

func (d *JSONDetector) Specificity() int {
	return 100 // Tier 1: Structured formats (highest)
}

func (d *JSONDetector) PatternLength() int {
	return 0 // Non-regex based detection
}

type LogfmtDetector struct{}

func (d *LogfmtDetector) Detect(ctx context.Context, line string) bool {
	done := make(chan bool, 1)
	go func() {
		line = strings.TrimSpace(line)
		if len(line) == 0 {
			done <- false
			return
		}

		kvPairs := 0
		totalTokens := 0

		// Parse the line character by character to handle quoted values
		i := 0
		for i < len(line) {
			// Skip whitespace
			for i < len(line) && line[i] == ' ' {
				i++
			}
			if i >= len(line) {
				break
			}

			for i < len(line) && line[i] != '=' && line[i] != ' ' {
				i++
			}

			if i >= len(line) || line[i] != '=' {
				// Not a key=value pair, skip to next whitespace
				for i < len(line) && line[i] != ' ' {
					i++
				}
				totalTokens++
				continue
			}

			i++ // skip the '='

			if i >= len(line) {
				// Key with no value (key=)
				kvPairs++
				totalTokens++
				break
			}

			if line[i] == '"' {
				// Quoted value
				i++ // skip opening quote
				for i < len(line) && line[i] != '"' {
					if line[i] == '\\' && i+1 < len(line) {
						i += 2 // skip escaped character
					} else {
						i++
					}
				}
				if i < len(line) {
					i++ // skip closing quote
				}
			} else {
				// Unquoted value - read until whitespace
				for i < len(line) && line[i] != ' ' {
					i++
				}
			}

			kvPairs++
			totalTokens++
		}
		done <- kvPairs > 0 && totalTokens > 0 && float64(kvPairs)/float64(totalTokens) > 0.5
	}()

	select {
	case result := <-done:
		return result
	case <-ctx.Done():
		return false
	}
}

func (d *LogfmtDetector) Format() LogFormat {
	return LogfmtFormat
}

func (d *LogfmtDetector) Specificity() int {
	return 100 // Tier 1: Structured formats (highest)
}

func (d *LogfmtDetector) PatternLength() int {
	return 0 // Non-regex based detection
}

type ApacheCommonDetector struct{}

const apacheCommonPattern = `^\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3} - - \[\d{2}\/\w{3}\/\d{4}:\d{2}:\d{2}:\d{2} [+-]\d{4}\] "[A-Z]+ .* HTTP\/\d\.\d" \d{3} \d+$`

var apacheCommonRegex = regexp.MustCompile(apacheCommonPattern)

func (d *ApacheCommonDetector) Detect(ctx context.Context, line string) bool {
	done := make(chan bool, 1)
	go func() {
		done <- apacheCommonRegex.MatchString(line)
	}()

	select {
	case result := <-done:
		return result
	case <-ctx.Done():
		return false
	}
}

func (d *ApacheCommonDetector) Format() LogFormat {
	return ApacheCommonFormat
}

func (d *ApacheCommonDetector) Specificity() int {
	return 50 // Tier 2: Regex-based formats
}

func (d *ApacheCommonDetector) PatternLength() int {
	return len(apacheCommonPattern)
}

type NginxDetector struct{}

const nginxPattern = `^\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3} - - \[\d{2}\/\w{3}\/\d{4}:\d{2}:\d{2}:\d{2} [+-]\d{4}\] "[A-Z]+ .* HTTP\/\d\.\d" \d{3} \d+ ".*" ".*"$`

var nginxRegex = regexp.MustCompile(nginxPattern)

func (d *NginxDetector) Detect(ctx context.Context, line string) bool {
	done := make(chan bool, 1)
	go func() {
		done <- nginxRegex.MatchString(line)
	}()

	select {
	case result := <-done:
		return result
	case <-ctx.Done():
		return false
	}
}

func (d *NginxDetector) Format() LogFormat {
	return NginxFormat
}

func (d *NginxDetector) Specificity() int {
	return 50 // Tier 2: Regex-based formats
}

func (d *NginxDetector) PatternLength() int {
	return len(nginxPattern)
}

type SyslogDetector struct{}

const syslogPattern = `^\w{3} \d{1,2} \d{2}:\d{2}:\d{2} \S+ \S+\[\d+\]:`

var syslogRegex = regexp.MustCompile(syslogPattern)

func (d *SyslogDetector) Detect(ctx context.Context, line string) bool {
	done := make(chan bool, 1)
	go func() {
		done <- syslogRegex.MatchString(line)
	}()

	select {
	case result := <-done:
		return result
	case <-ctx.Done():
		return false
	}
}

func (d *SyslogDetector) Format() LogFormat {
	return SyslogFormat
}

func (d *SyslogDetector) Specificity() int {
	return 50 // Tier 2: Regex-based formats
}

func (d *SyslogDetector) PatternLength() int {
	return len(syslogPattern)
}

// RsyslogDetector matches classic rsyslog-style headers explicitly containing syslogd/rsyslogd
type RsyslogDetector struct{}

// Example: "Aug  8 00:15:23 Host syslogd[347]: ASL Sender Statistics"
// Also matches rsyslogd
const rsyslogPattern = `^\w{3}\s+\d{1,2}\s+\d{2}:\d{2}:\d{2}\s+\S+\s+(?:rsyslogd|syslogd)\[\d+\]:`

var rsyslogRegex = regexp.MustCompile(rsyslogPattern)

func (d *RsyslogDetector) Detect(ctx context.Context, line string) bool {
	done := make(chan bool, 1)
	go func() {
		done <- rsyslogRegex.MatchString(line)
	}()

	select {
	case result := <-done:
		return result
	case <-ctx.Done():
		return false
	}
}

func (d *RsyslogDetector) Format() LogFormat {
	return RsyslogFormat
}

func (d *RsyslogDetector) Specificity() int {
	return 55 // Slightly higher than generic regex-based to prefer rsyslog over syslog when applicable
}

func (d *RsyslogDetector) PatternLength() int {
	return len(rsyslogPattern)
}

type GoStandardDetector struct{}

const goStandardPattern = `^\d{4}\/\d{2}\/\d{2} \d{2}:\d{2}:\d{2}`

var goStandardRegex = regexp.MustCompile(goStandardPattern)

func (d *GoStandardDetector) Detect(ctx context.Context, line string) bool {
	done := make(chan bool, 1)
	go func() {
		done <- goStandardRegex.MatchString(line)
	}()

	select {
	case result := <-done:
		return result
	case <-ctx.Done():
		return false
	}
}

func (d *GoStandardDetector) Format() LogFormat {
	return GoStandardFormat
}

func (d *GoStandardDetector) Specificity() int {
	return 50 // Tier 2: Regex-based formats
}

func (d *GoStandardDetector) PatternLength() int {
	return len(goStandardPattern)
}

type RailsDetector struct{}

const railsPattern = `^\[\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}\] \w+( --|  \w)`

var railsRegex = regexp.MustCompile(railsPattern)

func (d *RailsDetector) Detect(ctx context.Context, line string) bool {
	done := make(chan bool, 1)
	go func() {
		done <- railsRegex.MatchString(line)
	}()

	select {
	case result := <-done:
		return result
	case <-ctx.Done():
		return false
	}
}

func (d *RailsDetector) Format() LogFormat {
	return RailsFormat
}

func (d *RailsDetector) Specificity() int {
	return 50 // Tier 2: Regex-based formats
}

func (d *RailsDetector) PatternLength() int {
	return len(railsPattern)
}

type DockerDetector struct{}

const dockerPattern = `^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d+Z\s+[A-Z]+`

var dockerRegex = regexp.MustCompile(dockerPattern)

func (d *DockerDetector) Detect(ctx context.Context, line string) bool {
	done := make(chan bool, 1)
	go func() {
		done <- dockerRegex.MatchString(line)
	}()

	select {
	case result := <-done:
		return result
	case <-ctx.Done():
		return false
	}
}

func (d *DockerDetector) Format() LogFormat {
	return DockerFormat
}

func (d *DockerDetector) Specificity() int {
	return 50 // Tier 2: Regex-based formats
}

func (d *DockerDetector) PatternLength() int {
	return len(dockerPattern)
}

type KubernetesDetector struct{}

const kubernetesPattern = `^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d+Z \d+ \S+:\d+\] `

var kubernetesRegex = regexp.MustCompile(kubernetesPattern)

func (d *KubernetesDetector) Detect(ctx context.Context, line string) bool {
	done := make(chan bool, 1)
	go func() {
		done <- kubernetesRegex.MatchString(line)
	}()

	select {
	case result := <-done:
		return result
	case <-ctx.Done():
		return false
	}
}

func (d *KubernetesDetector) Format() LogFormat {
	return KubernetesFormat
}

func (d *KubernetesDetector) Specificity() int {
	return 50 // Tier 2: Regex-based formats
}

func (d *KubernetesDetector) PatternLength() int {
	return len(kubernetesPattern)
}

type HerokuDetector struct{}

const herokuPattern = `^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}[+-]\d{2}:\d{2} app\[\S+\]:`

var herokuRegex = regexp.MustCompile(herokuPattern)

func (d *HerokuDetector) Detect(ctx context.Context, line string) bool {
	done := make(chan bool, 1)
	go func() {
		done <- herokuRegex.MatchString(line)
	}()

	select {
	case result := <-done:
		return result
	case <-ctx.Done():
		return false
	}
}

func (d *HerokuDetector) Format() LogFormat {
	return HerokuFormat
}

func (d *HerokuDetector) Specificity() int {
	return 50 // Tier 2: Regex-based formats
}

func (d *HerokuDetector) PatternLength() int {
	return len(herokuPattern)
}

type GoTestDetector struct{}

const goTestPattern = `^(=== RUN|--- PASS:|--- FAIL:|--- SKIP:|=== NAME|=== CONT|\? .* \[no test files\]|PASS$|FAIL$|ok .* [\d\.]+[a-z]*$|FAIL .*)`

var goTestRegex = regexp.MustCompile(goTestPattern)

func (d *GoTestDetector) Detect(ctx context.Context, line string) bool {
	done := make(chan bool, 1)
	go func() {
		done <- goTestRegex.MatchString(line)
	}()

	select {
	case result := <-done:
		return result
	case <-ctx.Done():
		return false
	}
}

func (d *GoTestDetector) Format() LogFormat {
	return GoTestFormat
}

func (d *GoTestDetector) Specificity() int {
	return 70 // Higher than standard regex-based formats but lower than structured formats
}

func (d *GoTestDetector) PatternLength() int {
	return len(goTestPattern)
}

type JavaExceptionDetector struct{}

// Match Java exception headers (no leading whitespace) and stack trace lines (with leading whitespace)
const javaExceptionHeaderPattern = `^(Exception in thread|Caused by:)`
const javaStackTracePattern = `^\s+(at\s+|\.\.\.|\d+\s+more)`

var javaExceptionHeaderRegex = regexp.MustCompile(javaExceptionHeaderPattern)
var javaStackTraceRegex = regexp.MustCompile(javaStackTracePattern)

func (d *JavaExceptionDetector) Detect(ctx context.Context, line string) bool {
	done := make(chan bool, 1)
	go func() {
		// Match exception headers OR stack trace lines
		isHeader := javaExceptionHeaderRegex.MatchString(line)
		isStackTrace := javaStackTraceRegex.MatchString(line)
		done <- isHeader || isStackTrace
	}()

	select {
	case result := <-done:
		return result
	case <-ctx.Done():
		return false
	}
}

func (d *JavaExceptionDetector) Format() LogFormat {
	return JavaExceptionFormat
}

func (d *JavaExceptionDetector) Specificity() int {
	return 70 // Higher than standard regex-based formats, same as GoTest
}

func (d *JavaExceptionDetector) PatternLength() int {
	return len(javaExceptionHeaderPattern) + len(javaStackTracePattern)
}

type PythonExceptionDetector struct{}

// Only match distinctive Python traceback headers and exception lines, not stack trace lines
const pythonTracebackHeaderPattern = `^Traceback \(most recent call last\):`
const pythonExceptionPattern = `^[A-Za-z][A-Za-z0-9]*Error:`

var pythonTracebackHeaderRegex = regexp.MustCompile(pythonTracebackHeaderPattern)
var pythonExceptionRegex = regexp.MustCompile(pythonExceptionPattern)

func (d *PythonExceptionDetector) Detect(ctx context.Context, line string) bool {
	done := make(chan bool, 1)
	go func() {
		// Only match traceback headers OR exception lines (let continuation logic handle the rest)
		isHeader := pythonTracebackHeaderRegex.MatchString(line)
		isException := pythonExceptionRegex.MatchString(line)
		done <- isHeader || isException
	}()

	select {
	case result := <-done:
		return result
	case <-ctx.Done():
		return false
	}
}

func (d *PythonExceptionDetector) Format() LogFormat {
	return PythonExceptionFormat
}

func (d *PythonExceptionDetector) Specificity() int {
	return 70 // Higher than standard regex-based formats, same as GoTest and Java
}

func (d *PythonExceptionDetector) PatternLength() int {
	return len(pythonTracebackHeaderPattern) + len(pythonExceptionPattern)
}
