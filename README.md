![Splash Logo](splash_logo.svg)


Splash adds beautiful, adaptive colors to make logs easier to read.

## Features

- **Auto-detection** of 10+ popular log formats
- **Mixed formats** - handles multiple log formats in a single stream
- **Search highlighting** with string or regex patterns
- **Adaptive colors** that work with both light and dark terminals
- **Streaming performance** - processes logs in real-time
- **Zero configuration** - No config. Just pipe.

## Screenshots

<table>
  <tr>
    <td><img src="./screenshots/go_test.jpeg" alt="output of go test for splash repo" width="1280"/></td>
    <td><img src="./screenshots/go_test_splash.jpeg" alt="output of go test | splash" width="1280"/></td>
  </tr>
  <tr>
    <td align="center"><b>go test -v ./... </b><br/></td>
    <td align="center"><b>go test -v ./... | splash </b><br/></td>
  </tr>
</table>

### Go Test
![Plain go test output](screenshots/go_test.jpeg)

## Installation

### Install from Homebrew (Recommended)

```bash
brew tap joshi4/splash
brew install splash
```

### For Go users

```bash
go install github.com/joshi4/splash@latest
```

## 🎯 Getting Started

The easiest way to use Splash is to pipe any log output through it:


### Basic Usage
```bash
echo '{"timestamp":"2025-01-19T10:30:00Z","level":"ERROR","message":"Connection failed","service":"api"}' | splash
```

### Highlight errors
```bash
{ echo '{"level":"INFO","msg":"Starting up"}'; echo '{"level":"ERROR","msg":"Connection failed"}'; echo '{"level":"WARN","msg":"Slow query"}'; } | splash -s "ERROR"
```

### Create HTTP logs and highlight error status codes
```bash
{ echo '192.168.1.1 - - [19/Jan/2025:10:30:00 +0000] "GET /api HTTP/1.1" 200 1234'; echo '192.168.1.2 - - [19/Jan/2025:10:30:01 +0000] "POST /api HTTP/1.1" 404 567'; } | splash -r "[45]\d\d"
```

# Test with sample data from different log formats
printf '{"ts":"2025-01-19T10:30:00Z","level":"INFO","msg":"User login","user":"alice"}\n{"ts":"2025-01-19T10:30:05Z","level":"ERROR","msg":"Auth failed","user":"bob"}\n' | splash -s "ERROR"

# Generate and monitor continuous output
while true; do echo "$(date -Iseconds) INFO Server is healthy"; sleep 2; done | splash

# Create a mix of log formats to test detection
{
  echo 'Jan 19 10:30:00 localhost myapp[1234]: INFO Application started'
  echo '{"timestamp":"2025-01-19T10:30:01Z","level":"WARN","message":"High memory usage"}'
  echo '127.0.0.1 - - [19/Jan/2025:10:30:02 +0000] "GET /health HTTP/1.1" 200 15'
} | splash

# Use with curl to monitor API responses (requires jq)
curl -s httpbin.org/json | jq -c . | splash
```

## 🔧 Command Line Options

```bash
splash [flags]

Flags:
  -s, --search string    Highlight lines containing this text
  -r, --regexp string    Highlight lines matching this regex pattern
  -h, --help            Show help information
```

**Note:** You cannot use both `-s` and `-r` flags simultaneously.

## 📋 Supported Log Formats

Splash automatically detects and colorizes these log formats:

| Format | Example |
|--------|---------|
| **JSON** | `{"timestamp":"2025-01-19T10:30:00Z","level":"ERROR","message":"DB failed"}` |
| **Logfmt** | `timestamp=2025-01-19T10:30:00Z level=error msg="DB failed" service=api` |
| **Apache Common** | `127.0.0.1 - - [19/Jan/2025:10:30:00 +0000] "GET /api HTTP/1.1" 200 1234` |
| **Nginx** | `127.0.0.1 - - [19/Jan/2025:10:30:00 +0000] "GET /api HTTP/1.1" 200 1234 "-" "Mozilla/5.0"` |
| **Syslog** | `Jan 19 10:30:00 hostname myapp[1234]: ERROR: Database connection failed` |
| **Go Standard** | `2025/01/19 10:30:00 ERROR: Database connection failed` |
| **Rails** | `[2025-01-19 10:30:00] ERROR -- : Database connection failed` |
| **Docker** | `2025-01-19T10:30:00.123456789Z ERROR Database connection failed` |
| **Kubernetes** | `2025-01-19T10:30:00.123Z 1 main.go:42] ERROR Database connection failed` |
| **Heroku** | `2025-01-19T10:30:00+00:00 app[web.1]: ERROR Database connection failed` |
| **Go Test** | `=== RUN   TestDatabaseConnection` |

## Build from Source

```bash
git clone https://github.com/joshi4/splash.git
cd splash
go build
```

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feature-name`
3. Make your changes and add tests
4. Run tests: `go test ./...`
5. Submit a pull request
