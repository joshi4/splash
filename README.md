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
# Colorize application logs
tail -f /var/log/app.log | splash

# Colorize Docker container logs
docker logs mycontainer | splash

# Monitor Kubernetes pod logs with error highlighting
kubectl logs pod-name -f | splash -s "ERROR"

# Colorize web server access logs
cat /var/log/nginx/access.log | splash
```

## 💡 Usage Examples

### Basic Log Colorization

```bash
# JSON logs
cat app.log | splash
# {"timestamp":"2025-01-19T10:30:00Z","level":"ERROR","message":"DB failed"}

# Apache access logs
tail -f access.log | splash
# 192.168.1.1 - - [19/Jan/2025:10:30:00 +0000] "GET /api HTTP/1.1" 200 1234

# Docker logs
docker logs -f myapp | splash
# 2025-01-19T10:30:00.123Z INFO Application started on port 8080
```

### Search Highlighting

```bash
# Highlight all lines containing "ERROR"
tail -f app.log | splash -s "ERROR"

# Highlight lines matching a regex pattern (HTTP 4xx/5xx status codes)
cat access.log | splash -r "[45]\d\d"

# Highlight database-related entries
kubectl logs db-pod | splash -s "database"
```

### Real-world Examples

```bash
# Monitor application logs for errors
journalctl -f -u myapp | splash -s "ERROR"

# Analyze web server logs for failed requests
zcat access.log.gz | splash -r "\" [45]\d\d "

# Debug Kubernetes deployments
kubectl logs -f deployment/myapp | splash -s "panic"

# Monitor multiple log files
tail -f /var/log/*.log | splash
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
