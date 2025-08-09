# Agent Instructions for Splash

## Project
Splash: Add color to your logs.

Splash takes logs from stdin and auto detects well known log formats and adds colors to it. We will use lipgloss for this (https://github.com/charmbracelet/lipgloss). Colors should adapt to the users terminal theme. The goal is to make logs easy to parse and scan for a human. Splash also handles sigabort and sigkill gracefully. Major use case is for users to pipe input into and out of splash.

**Key Feature: Search & Highlight** - Users can search for specific patterns (strings or regex) and matching lines will be prominently highlighted to stand out among all other lines.

## Supported Log Formats
**Structured Formats:**
- **JSON**: `{"timestamp":"2025-01-19T10:30:00Z","level":"ERROR","message":"Database connection failed","service":"api"}`
- **Logfmt**: `timestamp=2025-01-19T10:30:00Z level=error msg="Database connection failed" service=api`

**Standard Application Logs:**
- **Apache Common**: `127.0.0.1 - - [19/Jan/2025:10:30:00 +0000] "GET /api/users HTTP/1.1" 200 1234`
- **Nginx**: `127.0.0.1 - - [19/Jan/2025:10:30:00 +0000] "GET /api/users HTTP/1.1" 200 1234 "-" "Mozilla/5.0"`
- **Syslog**: `Jan 19 10:30:00 hostname myapp[1234]: ERROR: Database connection failed`

**Framework/Runtime Logs:**
- **Go standard**: `2025/01/19 10:30:00 ERROR: Database connection failed`
- **Rails**: `[2025-01-19 10:30:00] ERROR -- : Database connection failed`
- **Docker**: `2025-01-19T10:30:00.123456789Z ERROR Database connection failed`
- **Kubernetes**: `2025-01-19T10:30:00.123Z 1 main.go:42] ERROR Database connection failed`

**Cloud/Service Logs:**
- **Heroku**: `2025-01-19T10:30:00+00:00 app[web.1]: ERROR Database connection failed`

## Build/Test/Lint Commands
- **Build:** `go build` or `go build ./...`
- **Build with version:** `go build -ldflags "-X github.com/joshi4/splash/cmd.version=v1.0.0"`
- **Test:** `go test ./...` (all tests) or `go test ./path/to/package` (single package)
- **Single test:** `go test -run TestName ./path/to/package`
- **Lint:** Use `go vet ./...` or `gofmt -s -w .`
- **Run:** `go run main.go` or `go run ./cmd`
- **Testing:** You should thoroughly test functionality with unit and integration tests

## Architecture & Structure
- Go CLI application using Cobra framework
- Entry point: `main.go` → calls `cmd.Execute()`
- Commands defined in `cmd/` directory (currently just `root.go`)
- Module: `github.com/joshi4/splash`
- Dependencies: Cobra CLI framework (`github.com/spf13/cobra`), Lipgloss for styling (`github.com/charmbracelet/lipgloss`)

## Code Style Guidelines
- Always run: `'gofmt -s -w .'` to format code before you finish your work.
- Imports: standard library first, then third-party, then local imports
- Error handling: explicit error checking, return errors up the call stack
- Naming: camelCase for private, PascalCase for public, descriptive names
- Package naming: lowercase, single word when possible
- Comments: exported functions/types must have doc comments starting with the name
- Commits: Do not include Amp thread links in commit messages.
- Run Style checks at the end of your development cycle to ensure code quality.


## Development Principles: Generalization & Modularity

**CRITICAL: When fixing issues or implementing features, always generalize across ALL log formats.**

### Generalization Rules
1. **Audit All Formats**: When fixing a bug in one log format colorizer (e.g., Heroku), immediately check and fix the same issue in ALL other format colorizers (JSON, Logfmt, Syslog, Apache, Nginx, Rails, Docker, Kubernetes, etc.)

2. **Create Reusable Functions**: Instead of duplicating code across colorizers, create shared utility functions:
   - Example: `applyStyleWithMarkers()` for consistent search highlighting
   - Example: `computeHighlightBackground()` for smart background colors
   - Example: `mapStyleToColor()` for color extraction

3. **Modular Design**: Structure code so that:
   - Each colorizer follows the same interface pattern
   - Common functionality is extracted to shared methods
   - New log formats can be added easily
   - Features work consistently across all formats

4. **Extensible Architecture**: When adding features:
   - Design them to work with ALL supported log formats from day one
   - Use interface-based approaches where possible
   - Make color themes and styling configurable
   - Ensure new features integrate seamlessly with existing ones

### Examples of Good Generalization
- **Search Highlighting**: When implementing smart background colors, ensure ALL colorizers (JSON keys, logfmt values, timestamps, IP addresses, etc.) use the same `applyStyleWithMarkers()` function
- **Color Themes**: When adding new theme colors, update ALL relevant colorizers and the color mapping functions
- **Performance Optimizations**: Apply optimizations uniformly across all format parsers

### Anti-Patterns to Avoid
- ❌ Fixing a bug in only one colorizer without checking others
- ❌ Copy-pasting similar code instead of creating shared functions
- ❌ Hard-coding values instead of making them configurable
- ❌ Implementing features that only work with specific log formats

### Testing Generalization
- Test new features across multiple log formats before considering them complete
- Use the provided test data files (`testdata/*.log`) to verify consistency
- Ensure regex patterns and string searches work identically across all formats

## Test-Driven Bug Fixes

**CRITICAL: When user provides a failing test case or bug report, always add Go tests to verify the fix.**

### Bug Fix Protocol
1. **Reproduce the Issue**: When a user reports a bug with a specific input/output example, first reproduce it manually
2. **Add Regression Test**: Create a Go test that captures the failing case before implementing the fix
3. **Implement Fix**: Develop the solution that makes the test pass
4. **Verify Fix**: Ensure the new test passes and all existing tests continue to pass
5. **Prevent Regression**: The test ensures the bug cannot reoccur without detection

### Test Case Guidelines
- **Specific Test Cases**: When user provides exact input/output examples, create dedicated test functions
- **Edge Cases**: Include tests for boundary conditions and error scenarios
- **Multiple Formats**: If the bug affects multiple log formats, test all relevant formats
- **Search Functionality**: When fixing search-related bugs, test both string and regex search patterns

### Examples of Good Test Coverage
- **JSON Structure Integrity**: `TestJSONSearchMarkersNotBreakingStructure()` prevents search markers from breaking JSON parsing
- **Nested Object Handling**: `TestJSONNestedObjects()` ensures recursive colorization works correctly
- **Cross-Format Consistency**: Tests that verify identical behavior across different log formats
