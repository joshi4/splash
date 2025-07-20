# Test Data for Splash

This directory contains sample log files for testing and demonstrating Splash's log format detection and colorization capabilities.

## Log Format Files

### Structured Formats
- **`json.log`** - JSON structured logs with various log levels and services
- **`logfmt.log`** - Logfmt key=value format logs

### Web Server Logs  
- **`apache_common.log`** - Apache Common Log Format
- **`nginx.log`** - Nginx Combined Log Format (includes user agent and referer)

### System Logs
- **`syslog.log`** - Standard Unix syslog format
- **`go_standard.log`** - Go's standard log package format
- **`rails.log`** - Ruby on Rails application logs

### Container & Cloud Logs
- **`docker.log`** - Docker container logs
- **`kubernetes.log`** - Kubernetes pod logs with file references
- **`heroku.log`** - Heroku dyno logs

### Mixed Format
- **`mixed.log`** - Multiple log formats in a single file for testing format switching

## Usage Examples

### Test individual formats:
```bash
cat testdata/json.log | ./splash
cat testdata/nginx.log | ./splash
cat testdata/kubernetes.log | ./splash
```

### Test mixed format handling:
```bash
cat testdata/mixed.log | ./splash
```

### Test all formats at once:
```bash
cat testdata/*.log | ./splash
```

## Log Characteristics

Each file contains logs with:
- **Multiple log levels** (INFO, DEBUG, WARN, ERROR, FATAL)
- **Various services** and components  
- **Realistic timestamps** and data
- **Different scenarios** (success, errors, warnings)
- **Format-specific elements** (HTTP status codes, file references, etc.)

These files are designed to thoroughly test Splash's format detection accuracy and colorization quality.
