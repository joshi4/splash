![Splash Logo](splash_logo.svg)


Splash is a CLI tool that automatically detects log formats and adds beautiful, adaptive colors to make logs easier to read and parse. Perfect for developers who spend time analyzing logs but find plain text hard to scan.

## ✨ Features

- 🎨 **Auto-detection** of 10+ popular log formats
- 🔍 **Search highlighting** with string or regex patterns
- 🌓 **Adaptive colors** that work with both light and dark terminals
- ⚡ **Streaming performance** - processes logs in real-time
- 🛡️ **Graceful signal handling** (SIGINT, SIGTERM, SIGHUP)
- 📦 **Zero configuration** - just pipe and go!

## 🚀 Installation

### Install from GitHub (Recommended)

```bash
go install github.com/joshi4/splash@latest
```

### Build from Source

```bash
git clone https://github.com/joshi4/splash.git
cd splash
go build
```

## 🎯 Quick Start

The easiest way to use Splash is to pipe any log output through it:

```bash
# Stream system logs (Linux/macOS)
journalctl -f | splash

# Monitor network connectivity with ping
ping google.com | splash

# Generate and colorize JSON logs
echo '{"timestamp":"2025-01-19T10:30:00Z","level":"ERROR","message":"Connection failed","service":"api"}' | splash

# View system logs (macOS)
log stream --predicate 'eventMessage contains "error"' | splash -s "error"
```

## 💡 Usage Examples

### Basic Log Colorization

```bash
# Create and colorize JSON logs
echo '{"timestamp":"2025-01-19T10:30:00Z","level":"ERROR","message":"Database connection failed","service":"api"}' | splash

# Generate structured logs with printf
printf 'timestamp=2025-01-19T10:30:00Z level=info msg="Server started" port=8080\n' | splash

# Simulate Apache access logs
echo '192.168.1.100 - - [19/Jan/2025:10:30:15 +0000] "GET /api/users HTTP/1.1" 200 1234' | splash

# Create multiple log formats at once
{ echo '{"level":"INFO","msg":"App started"}'; echo '2025/01/19 10:30:00 ERROR: Connection failed'; } | splash
```

### Search Highlighting

```bash
# Generate logs and highlight errors
{ echo '{"level":"INFO","msg":"Starting up"}'; echo '{"level":"ERROR","msg":"Connection failed"}'; echo '{"level":"WARN","msg":"Slow query"}'; } | splash -s "ERROR"

# Create HTTP logs and highlight error status codes
{ echo '192.168.1.1 - - [19/Jan/2025:10:30:00 +0000] "GET /api HTTP/1.1" 200 1234'; echo '192.168.1.2 - - [19/Jan/2025:10:30:01 +0000] "POST /api HTTP/1.1" 404 567'; } | splash -r "[45]\d\d"

# Monitor real network activity with pattern highlighting
ping -c 5 google.com | splash -s "time"

# Stream system logs with keyword highlighting (Linux)
journalctl --lines=10 | splash -s "systemd"
```

### Real-world Examples

```bash
# Monitor system logs with splash (Linux/systemd systems)
journalctl -f --no-pager | splash

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

## 🎨 Color Themes

Splash uses adaptive color themes that automatically adjust based on your terminal's appearance:

- **Timestamps**: Subtle, muted colors
- **Log levels**: Color-coded (ERROR=red, WARN=yellow, INFO=cyan, DEBUG=gray)
- **HTTP status codes**: Green (2xx), yellow (3xx), orange (4xx), red (5xx)
- **IP addresses**: Blue highlighting
- **JSON keys/values**: Structured syntax highlighting
- **Search matches**: Prominent background highlighting

## 🛠️ Development

### Building

```bash
go build                 # Build binary
go build ./...          # Build all packages
```

### Testing

```bash
go test ./...                    # Run all tests
go test ./parser                 # Test specific package
go test -run TestName ./parser   # Run specific test
```

### Running

```bash
go run main.go          # Run from source
echo "test log" | go run main.go -s "test"
```

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feature-name`
3. Make your changes and add tests
4. Run tests: `go test ./...`
5. Submit a pull request
