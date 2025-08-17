package parser

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestDetectionWithTestData verifies format detection using real test log files
func TestDetectionWithTestData(t *testing.T) {
	testCases := []struct {
		filename       string
		expectedFormat LogFormat
		description    string
	}{
		{"json.log", JSONFormat, "JSON structured logs"},
		{"logfmt.log", LogfmtFormat, "Logfmt key=value logs"},
		{"apache_common.log", ApacheCommonFormat, "Apache Common Log format"},
		{"nginx.log", NginxFormat, "Nginx Combined Log format"},
		{"syslog.log", SyslogFormat, "System log format"},
		{"rsyslog.log", RsyslogFormat, "Rsyslog multi-line entries"},
		{"go_standard.log", GoStandardFormat, "Go standard log format"},
		{"rails.log", RailsFormat, "Rails application logs"},
		{"docker.log", DockerFormat, "Docker container logs"},
		{"kubernetes.log", KubernetesFormat, "Kubernetes pod logs"},
		{"heroku.log", HerokuFormat, "Heroku dyno logs"},
	}

	parser := NewParser()

	for _, tc := range testCases {
		t.Run(tc.filename, func(t *testing.T) {
			// Read the test file
			testdataPath := filepath.Join("..", "testdata", tc.filename)
			file, err := os.Open(testdataPath)
			if err != nil {
				t.Fatalf("Failed to open test file %s: %v", testdataPath, err)
			}
			defer file.Close()

			scanner := bufio.NewScanner(file)
			lineCount := 0
			correctDetections := 0

			for scanner.Scan() {
				line := scanner.Text()
				if strings.TrimSpace(line) == "" {
					continue // Skip empty lines
				}

				lineCount++
				detectedFormat := parser.DetectFormat(line)

				if detectedFormat == tc.expectedFormat {
					correctDetections++
				} else {
					t.Logf("Line %d: Expected %v, got %v for line: %s",
						lineCount, tc.expectedFormat, detectedFormat, line)
				}
			}

			if err := scanner.Err(); err != nil {
				t.Fatalf("Error reading file %s: %v", tc.filename, err)
			}

			// Require at least 80% accuracy (some edge cases may not match perfectly)
			accuracy := float64(correctDetections) / float64(lineCount)
			minAccuracy := 0.8

			if accuracy < minAccuracy {
				t.Errorf("%s: Detection accuracy too low: %.2f%% (%d/%d). Expected at least %.0f%%",
					tc.description, accuracy*100, correctDetections, lineCount, minAccuracy*100)
			} else {
				t.Logf("%s: Detection accuracy: %.2f%% (%d/%d)",
					tc.description, accuracy*100, correctDetections, lineCount)
			}
		})
	}
}

// TestMixedFormatFile tests the mixed format file to ensure different formats are detected
func TestMixedFormatFile(t *testing.T) {
	parser := NewParser()

	testdataPath := filepath.Join("..", "testdata", "mixed.log")
	file, err := os.Open(testdataPath)
	if err != nil {
		t.Fatalf("Failed to open mixed.log: %v", err)
	}
	defer file.Close()

	expectedFormats := []LogFormat{
		JSONFormat,         // JSON log entry
		LogfmtFormat,       // Logfmt entry
		ApacheCommonFormat, // Apache log
		SyslogFormat,       // Syslog entry
		GoStandardFormat,   // Go standard log
		RailsFormat,        // Rails log entry
		DockerFormat,       // Docker container log
		KubernetesFormat,   // Kubernetes pod log
		HerokuFormat,       // Heroku dyno log
		NginxFormat,        // Nginx log (has user agent)
		JSONFormat,         // Another JSON entry
		UnknownFormat,      // Unknown format line
	}

	scanner := bufio.NewScanner(file)
	lineIndex := 0

	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}

		if lineIndex >= len(expectedFormats) {
			t.Fatalf("More lines in mixed.log than expected formats")
		}

		detectedFormat := parser.DetectFormat(line)
		expectedFormat := expectedFormats[lineIndex]

		if detectedFormat != expectedFormat {
			t.Errorf("Line %d: Expected format %v, got %v for line: %s",
				lineIndex+1, expectedFormat, detectedFormat, line)
		}

		lineIndex++
	}

	if lineIndex != len(expectedFormats) {
		t.Errorf("Expected %d lines, got %d", len(expectedFormats), lineIndex)
	}
}

// BenchmarkTestDataDetection benchmarks format detection on real test data
func BenchmarkTestDataDetection(b *testing.B) {
	parser := NewParser()

	// Read a sample of lines from different files
	testLines := []string{
		`{"timestamp":"2025-01-19T08:30:00Z","level":"INFO","service":"api"}`,
		`timestamp=2025-01-19T08:30:00Z level=info msg="test" service=api`,
		`192.168.1.100 - - [19/Jan/2025:08:30:00 +0000] "GET / HTTP/1.1" 200 1234`,
		`Jan 19 08:30:00 server myapp[1234]: INFO: test message`,
		`2025-01-19T08:30:00.123Z 1 main.go:42] INFO test message`,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, line := range testLines {
			parser.DetectFormat(line)
		}
	}
}
