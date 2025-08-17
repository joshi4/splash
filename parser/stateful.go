package parser

import (
	"context"
	"regexp"
)

// StatefulDetector defines the interface for multi-line log format detectors
// that need to maintain state across multiple lines
type StatefulDetector interface {
	FormatDetector
	// DetectStart returns true if this line starts a multi-line log entry
	DetectStart(ctx context.Context, line string) bool
	// DetectContinuation returns true if this line continues the current multi-line entry
	DetectContinuation(ctx context.Context, line string) bool
	// DetectEnd returns true if this line ends the current multi-line entry
	// If this returns true, the line is still considered part of this format
	DetectEnd(ctx context.Context, line string) bool
}

// StatefulRsyslogDetector handles multi-line rsyslog entries
type StatefulRsyslogDetector struct{}

const rsyslogStartPattern = `^\w{3}\s+\d{1,2}\s+\d{2}:\d{2}:\d{2}\s+\S+\s+(?:rsyslogd|syslogd)\[\d+\]:`

var rsyslogStartRegex = regexp.MustCompile(rsyslogStartPattern)

func (d *StatefulRsyslogDetector) DetectStart(ctx context.Context, line string) bool {
	done := make(chan bool, 1)
	go func() {
		done <- rsyslogStartRegex.MatchString(line)
	}()

	select {
	case result := <-done:
		return result
	case <-ctx.Done():
		return false
	}
}

func (d *StatefulRsyslogDetector) DetectContinuation(_ context.Context, line string) bool {
	// Continuation lines have leading whitespace
	return len(line) > 0 && (line[0] == ' ' || line[0] == '\t')
}

func (d *StatefulRsyslogDetector) DetectEnd(_ context.Context, _ string) bool {
	// Rsyslog entries don't have explicit end markers
	// They end when we encounter a non-continuation line
	return false
}

func (d *StatefulRsyslogDetector) Detect(ctx context.Context, line string) bool {
	// For backward compatibility with existing FormatDetector interface
	return d.DetectStart(ctx, line)
}

func (d *StatefulRsyslogDetector) Format() LogFormat {
	return RsyslogFormat
}

func (d *StatefulRsyslogDetector) Specificity() int {
	return 55 // Slightly higher than generic regex-based to prefer rsyslog over syslog when applicable
}

func (d *StatefulRsyslogDetector) PatternLength() int {
	return len(rsyslogStartPattern)
}

// StatefulJavaExceptionDetector handles multi-line Java exception traces
type StatefulJavaExceptionDetector struct{}

const javaExceptionStartPattern = `^(Exception in thread|Caused by:)`
const javaStackTraceLinePattern = `^\s+(at\s+|\.\.\.|\d+\s+more)`

var javaExceptionStartRegex = regexp.MustCompile(javaExceptionStartPattern)
var javaStackTraceLineRegex = regexp.MustCompile(javaStackTraceLinePattern)

func (d *StatefulJavaExceptionDetector) DetectStart(ctx context.Context, line string) bool {
	done := make(chan bool, 1)
	go func() {
		done <- javaExceptionStartRegex.MatchString(line)
	}()

	select {
	case result := <-done:
		return result
	case <-ctx.Done():
		return false
	}
}

func (d *StatefulJavaExceptionDetector) DetectContinuation(_ context.Context, line string) bool {
	// Java stack trace lines start with whitespace
	return len(line) > 0 && (line[0] == ' ' || line[0] == '\t')
}

func (d *StatefulJavaExceptionDetector) DetectEnd(_ context.Context, _ string) bool {
	// Java exceptions don't have explicit end markers
	// They end when we encounter a non-continuation line
	return false
}

func (d *StatefulJavaExceptionDetector) Detect(ctx context.Context, line string) bool {
	// Match exception headers OR stack trace lines for backward compatibility
	done := make(chan bool, 1)
	go func() {
		isStart := javaExceptionStartRegex.MatchString(line)
		isStackTrace := javaStackTraceLineRegex.MatchString(line)
		done <- isStart || isStackTrace
	}()

	select {
	case result := <-done:
		return result
	case <-ctx.Done():
		return false
	}
}

func (d *StatefulJavaExceptionDetector) Format() LogFormat {
	return JavaExceptionFormat
}

func (d *StatefulJavaExceptionDetector) Specificity() int {
	return 70 // Higher than standard regex-based formats, same as GoTest
}

func (d *StatefulJavaExceptionDetector) PatternLength() int {
	return len(javaExceptionStartPattern) + len(javaStackTraceLinePattern)
}

// StatefulPythonExceptionDetector handles multi-line Python exception traces
type StatefulPythonExceptionDetector struct{}

const pythonExceptionStartPattern = `^Traceback \(most recent call last\):`
const pythonExceptionLinePattern = `^[A-Za-z][A-Za-z0-9]*Error:`

var pythonExceptionStartRegex = regexp.MustCompile(pythonExceptionStartPattern)
var pythonExceptionLineRegex = regexp.MustCompile(pythonExceptionLinePattern)

func (d *StatefulPythonExceptionDetector) DetectStart(ctx context.Context, line string) bool {
	done := make(chan bool, 1)
	go func() {
		isTracebackStart := pythonExceptionStartRegex.MatchString(line)
		isExceptionLine := pythonExceptionLineRegex.MatchString(line)
		done <- isTracebackStart || isExceptionLine
	}()

	select {
	case result := <-done:
		return result
	case <-ctx.Done():
		return false
	}
}

func (d *StatefulPythonExceptionDetector) DetectContinuation(_ context.Context, line string) bool {
	// Python traceback lines start with whitespace
	return len(line) > 0 && (line[0] == ' ' || line[0] == '\t')
}

func (d *StatefulPythonExceptionDetector) DetectEnd(_ context.Context, _ string) bool {
	// Python exceptions don't have explicit end markers
	// They end when we encounter a non-continuation line
	return false
}

func (d *StatefulPythonExceptionDetector) Detect(ctx context.Context, line string) bool {
	// Match traceback headers OR exception lines for backward compatibility
	done := make(chan bool, 1)
	go func() {
		isStart := pythonExceptionStartRegex.MatchString(line)
		isException := pythonExceptionLineRegex.MatchString(line)
		done <- isStart || isException
	}()

	select {
	case result := <-done:
		return result
	case <-ctx.Done():
		return false
	}
}

func (d *StatefulPythonExceptionDetector) Format() LogFormat {
	return PythonExceptionFormat
}

func (d *StatefulPythonExceptionDetector) Specificity() int {
	return 70 // Higher than standard regex-based formats, same as GoTest and Java
}

func (d *StatefulPythonExceptionDetector) PatternLength() int {
	return len(pythonExceptionStartPattern) + len(pythonExceptionLinePattern)
}

// StatefulGoroutineStackTraceDetector handles multi-line Go goroutine stack traces
type StatefulGoroutineStackTraceDetector struct{}

const goroutineStartPattern = `^goroutine \d+ \[.*\]:`
const goroutineStackTraceLinePattern = `^(\s+[a-zA-Z_][a-zA-Z0-9_]*\.|[a-zA-Z_][a-zA-Z0-9_]*\.[a-zA-Z_]|\s+/)`

var goroutineStartRegex = regexp.MustCompile(goroutineStartPattern)
var goroutineStackTraceLineRegex = regexp.MustCompile(goroutineStackTraceLinePattern)

func (d *StatefulGoroutineStackTraceDetector) DetectStart(ctx context.Context, line string) bool {
	done := make(chan bool, 1)
	go func() {
		done <- goroutineStartRegex.MatchString(line)
	}()

	if ctx != nil {
		select {
		case result := <-done:
			return result
		case <-ctx.Done():
			return false
		}
	} else {
		// Handle nil context case
		return <-done
	}
}

func (d *StatefulGoroutineStackTraceDetector) DetectContinuation(_ context.Context, line string) bool {
	// Goroutine stack trace lines start with whitespace
	return len(line) > 0 && (line[0] == ' ' || line[0] == '\t')
}

func (d *StatefulGoroutineStackTraceDetector) DetectEnd(_ context.Context, _ string) bool {
	// Goroutine stack traces don't have explicit end markers
	// They end when we encounter a non-continuation line
	return false
}

func (d *StatefulGoroutineStackTraceDetector) Detect(ctx context.Context, line string) bool {
	// Match goroutine headers OR stack trace lines for backward compatibility
	done := make(chan bool, 1)
	go func() {
		isStart := goroutineStartRegex.MatchString(line)
		isStackTrace := goroutineStackTraceLineRegex.MatchString(line)
		done <- isStart || isStackTrace
	}()

	if ctx != nil {
		select {
		case result := <-done:
			return result
		case <-ctx.Done():
			return false
		}
	} else {
		// Handle nil context case
		return <-done
	}
}

func (d *StatefulGoroutineStackTraceDetector) Format() LogFormat {
	return GoroutineStackTraceFormat
}

func (d *StatefulGoroutineStackTraceDetector) Specificity() int {
	return 70 // Higher than standard regex-based formats, same as GoTest and others
}

func (d *StatefulGoroutineStackTraceDetector) PatternLength() int {
	return len(goroutineStartPattern) + len(goroutineStackTraceLinePattern)
}

// StatefulJavaScriptExceptionDetector handles multi-line JavaScript exception traces
type StatefulJavaScriptExceptionDetector struct{}

const jsExceptionStartPattern = `^(Error$|Trace:|TypeError:|ReferenceError:|SyntaxError:|RangeError:|EvalError:|URIError:|InternalError:|[A-Z][a-zA-Z]*Exception:)`
const jsStackTraceLinePattern = `^\s+at\s+`

var jsExceptionStartRegex = regexp.MustCompile(jsExceptionStartPattern)
var jsStackTraceLineRegex = regexp.MustCompile(jsStackTraceLinePattern)

func (d *StatefulJavaScriptExceptionDetector) DetectStart(ctx context.Context, line string) bool {
	done := make(chan bool, 1)
	go func() {
		done <- jsExceptionStartRegex.MatchString(line)
	}()

	if ctx != nil {
		select {
		case result := <-done:
			return result
		case <-ctx.Done():
			return false
		}
	} else {
		// Handle nil context case
		return <-done
	}
}

func (d *StatefulJavaScriptExceptionDetector) DetectContinuation(_ context.Context, line string) bool {
	// JavaScript stack trace lines start with whitespace followed by "at"
	return jsStackTraceLineRegex.MatchString(line)
}

func (d *StatefulJavaScriptExceptionDetector) DetectEnd(_ context.Context, _ string) bool {
	// JavaScript exceptions don't have explicit end markers
	// They end when we encounter a non-continuation line
	return false
}

func (d *StatefulJavaScriptExceptionDetector) Detect(ctx context.Context, line string) bool {
	// Match exception headers OR stack trace lines for backward compatibility
	done := make(chan bool, 1)
	go func() {
		isStart := jsExceptionStartRegex.MatchString(line)
		isStackTrace := jsStackTraceLineRegex.MatchString(line)
		done <- isStart || isStackTrace
	}()

	if ctx != nil {
		select {
		case result := <-done:
			return result
		case <-ctx.Done():
			return false
		}
	} else {
		// Handle nil context case
		return <-done
	}
}

func (d *StatefulJavaScriptExceptionDetector) Format() LogFormat {
	return JavaScriptExceptionFormat
}

func (d *StatefulJavaScriptExceptionDetector) Specificity() int {
	return 70 // Higher than standard regex-based formats, same as other exception formats
}

func (d *StatefulJavaScriptExceptionDetector) PatternLength() int {
	return len(jsExceptionStartPattern) + len(jsStackTraceLinePattern)
}
