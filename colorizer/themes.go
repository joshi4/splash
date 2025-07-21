package colorizer

import (
	"github.com/charmbracelet/lipgloss"
)

// ColorTheme defines the color scheme for different log components
type ColorTheme struct {
	// Log levels
	Error   lipgloss.Style
	Warning lipgloss.Style
	Info    lipgloss.Style
	Debug   lipgloss.Style

	// HTTP status codes
	StatusOK    lipgloss.Style // 2xx
	StatusWarn  lipgloss.Style // 4xx
	StatusError lipgloss.Style // 5xx

	// General components
	Timestamp lipgloss.Style
	IP        lipgloss.Style
	URL       lipgloss.Style
	Method    lipgloss.Style

	// JSON/structured data
	JSONKey    lipgloss.Style
	JSONValue  lipgloss.Style
	JSONString lipgloss.Style
	JSONNumber lipgloss.Style

	// Logfmt
	LogfmtKey   lipgloss.Style
	LogfmtValue lipgloss.Style

	// System/process info
	Hostname lipgloss.Style
	PID      lipgloss.Style
	Service  lipgloss.Style

	// File references
	Filename lipgloss.Style
	LineNum  lipgloss.Style

	// Punctuation/structure
	Bracket lipgloss.Style
	Quote   lipgloss.Style
	Equals  lipgloss.Style

	// Search highlighting - Unified style for all log formats
	SearchHighlight        lipgloss.Style // Deprecated - use UnifiedSearchHighlight
	JSONSearchHighlight    lipgloss.Style // Deprecated - use UnifiedSearchHighlight
	UnifiedSearchHighlight lipgloss.Style // Bright Orange + Adaptive - used for all search highlighting
}

// NewAdaptiveTheme creates a color theme that adapts to the terminal
func NewAdaptiveTheme() *ColorTheme {
	return &ColorTheme{
		// Log levels with semantic colors using ANSI colors for better compatibility
		Error:   lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "1", Dark: "9"}).Bold(true), // Red/Bright red
		Warning: lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "130", Dark: "11"}),         // Dark orange for light, bright yellow for dark
		Info:    lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "4", Dark: "14"}),           // Blue/Bright cyan
		Debug:   lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "2", Dark: "10"}),           // Green/Bright green

		// HTTP status codes
		StatusOK:    lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "2", Dark: "10"}),           // Green/Bright green
		StatusWarn:  lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "130", Dark: "11"}),         // Dark orange for light, bright yellow for dark
		StatusError: lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "1", Dark: "9"}).Bold(true), // Red/Bright red

		// General components - ANSI colors for better compatibility
		Timestamp: lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "8", Dark: "245"}),                // Bright black/Medium gray
		IP:        lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "4", Dark: "12"}).Bold(true),      // Blue/Bright blue
		URL:       lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "6", Dark: "14"}).Underline(true), // Cyan/Bright cyan
		Method:    lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "5", Dark: "13"}).Bold(true),      // Magenta/Bright magenta

		// JSON/structured data - ANSI colors for better compatibility
		JSONKey:    lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "5", Dark: "13"}).Bold(true), // Magenta/Bright magenta
		JSONValue:  lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "0", Dark: "7"}),             // Black/White
		JSONString: lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "2", Dark: "10"}),            // Green/Bright green
		JSONNumber: lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "4", Dark: "12"}),            // Blue/Bright blue

		// Logfmt
		LogfmtKey:   lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "5", Dark: "13"}).Bold(true), // Magenta/Bright magenta
		LogfmtValue: lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "0", Dark: "7"}),             // Black/White

		// System/process info - ANSI colors for better compatibility
		Hostname: lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "3", Dark: "11"}).Bold(true), // Yellow/Bright yellow
		PID:      lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "5", Dark: "13"}),            // Magenta/Bright magenta
		Service:  lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "4", Dark: "12"}).Bold(true), // Blue/Bright blue

		// File references
		Filename: lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "3", Dark: "11"}), // Yellow/Bright yellow
		LineNum:  lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "5", Dark: "13"}), // Magenta/Bright magenta

		// Punctuation/structure (subtle but visible)
		Bracket: lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "8", Dark: "8"}), // Bright black/Bright black
		Quote:   lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "8", Dark: "8"}), // Bright black/Bright black
		Equals:  lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "8", Dark: "8"}), // Bright black/Bright black

		// Search highlighting (bold text, no background)
		SearchHighlight: lipgloss.NewStyle().
			Bold(true), // Bold for visibility

		// JSON-specific search highlighting with high visibility foreground colors
		JSONSearchHighlight: lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "3", Dark: "11"}). // Yellow: ANSI yellow for light, bright yellow for dark
			Bold(true),                                                 // Bold for extra visibility

		// Unified search highlighting - Yellow background for light, orange background for dark
		UnifiedSearchHighlight: lipgloss.NewStyle().
			Background(lipgloss.AdaptiveColor{Light: "3", Dark: "208"}). // Yellow for light, orange for dark
			Foreground(lipgloss.AdaptiveColor{Light: "1", Dark: "0"}).   // Red text for light, black for dark
			Bold(true),
	}
}

// NewLightTheme creates a theme optimized for light terminal backgrounds
func NewLightTheme() *ColorTheme {
	return &ColorTheme{
		// Log levels with darker ANSI colors for light backgrounds
		Error:   lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Bold(true),   // ANSI red
		Warning: lipgloss.NewStyle().Foreground(lipgloss.Color("130")).Bold(true), // ANSI dark orange (darker than yellow)
		Info:    lipgloss.NewStyle().Foreground(lipgloss.Color("4")).Bold(true),   // ANSI blue
		Debug:   lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Bold(true),   // ANSI green

		// HTTP status codes
		StatusOK:    lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Bold(true),   // ANSI green
		StatusWarn:  lipgloss.NewStyle().Foreground(lipgloss.Color("130")).Bold(true), // ANSI dark orange
		StatusError: lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Bold(true),   // ANSI red

		// General components - darker ANSI colors for light themes
		Timestamp: lipgloss.NewStyle().Foreground(lipgloss.Color("8")),                 // ANSI bright black
		IP:        lipgloss.NewStyle().Foreground(lipgloss.Color("4")).Bold(true),      // ANSI blue
		URL:       lipgloss.NewStyle().Foreground(lipgloss.Color("6")).Underline(true), // ANSI cyan underlined
		Method:    lipgloss.NewStyle().Foreground(lipgloss.Color("5")).Bold(true),      // ANSI magenta

		// JSON/structured data
		JSONKey:    lipgloss.NewStyle().Foreground(lipgloss.Color("5")).Bold(true), // ANSI magenta
		JSONValue:  lipgloss.NewStyle().Foreground(lipgloss.Color("0")),            // ANSI black
		JSONString: lipgloss.NewStyle().Foreground(lipgloss.Color("2")),            // ANSI green
		JSONNumber: lipgloss.NewStyle().Foreground(lipgloss.Color("4")),            // ANSI blue

		// Logfmt
		LogfmtKey:   lipgloss.NewStyle().Foreground(lipgloss.Color("5")).Bold(true), // ANSI magenta
		LogfmtValue: lipgloss.NewStyle().Foreground(lipgloss.Color("0")),            // ANSI black

		// System/process info
		Hostname: lipgloss.NewStyle().Foreground(lipgloss.Color("3")).Bold(true), // ANSI yellow
		PID:      lipgloss.NewStyle().Foreground(lipgloss.Color("5")),            // ANSI magenta
		Service:  lipgloss.NewStyle().Foreground(lipgloss.Color("4")).Bold(true), // ANSI blue

		// File references
		Filename: lipgloss.NewStyle().Foreground(lipgloss.Color("3")), // ANSI yellow
		LineNum:  lipgloss.NewStyle().Foreground(lipgloss.Color("5")), // ANSI magenta

		// Punctuation/structure
		Bracket: lipgloss.NewStyle().Foreground(lipgloss.Color("8")), // ANSI bright black
		Quote:   lipgloss.NewStyle().Foreground(lipgloss.Color("8")), // ANSI bright black
		Equals:  lipgloss.NewStyle().Foreground(lipgloss.Color("8")), // ANSI bright black

		// Search highlighting
		SearchHighlight: lipgloss.NewStyle().Bold(true),
		JSONSearchHighlight: lipgloss.NewStyle().
			Foreground(lipgloss.Color("3")).Bold(true), // ANSI yellow
		UnifiedSearchHighlight: lipgloss.NewStyle().
			Background(lipgloss.Color("3")). // ANSI yellow background
			Foreground(lipgloss.Color("1")). // ANSI red text
			Bold(true),
	}
}

// NewDarkTheme creates a theme optimized for dark terminal backgrounds
func NewDarkTheme() *ColorTheme {
	return &ColorTheme{
		// Log levels with brighter colors for dark backgrounds
		Error:   lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6B6B")).Bold(true), // Bright red
		Warning: lipgloss.NewStyle().Foreground(lipgloss.Color("#FFE66D")),            // Bright yellow
		Info:    lipgloss.NewStyle().Foreground(lipgloss.Color("#4ECDC4")),            // Bright cyan
		Debug:   lipgloss.NewStyle().Foreground(lipgloss.Color("#95E1D3")),            // Light green

		// HTTP status codes
		StatusOK:    lipgloss.NewStyle().Foreground(lipgloss.Color("#6BCF7F")),            // Bright green
		StatusWarn:  lipgloss.NewStyle().Foreground(lipgloss.Color("#FFD93D")),            // Bright yellow
		StatusError: lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6B6B")).Bold(true), // Bright red

		// General components
		Timestamp: lipgloss.NewStyle().Foreground(lipgloss.Color("245")),                     // ANSI medium gray (dimmed but visible)
		IP:        lipgloss.NewStyle().Foreground(lipgloss.Color("#74B9FF")),                 // Light blue
		URL:       lipgloss.NewStyle().Foreground(lipgloss.Color("#81ECEC")).Underline(true), // Light cyan underlined
		Method:    lipgloss.NewStyle().Foreground(lipgloss.Color("#FD79A8")).Bold(true),      // Light pink

		// JSON/structured data
		JSONKey:    lipgloss.NewStyle().Foreground(lipgloss.Color("#DDA0DD")), // Light plum
		JSONValue:  lipgloss.NewStyle().Foreground(lipgloss.Color("#F0F0F0")), // Light gray
		JSONString: lipgloss.NewStyle().Foreground(lipgloss.Color("#98FB98")), // Light green
		JSONNumber: lipgloss.NewStyle().Foreground(lipgloss.Color("#87CEEB")), // Light blue

		// Logfmt
		LogfmtKey:   lipgloss.NewStyle().Foreground(lipgloss.Color("#DDA0DD")), // Light plum
		LogfmtValue: lipgloss.NewStyle().Foreground(lipgloss.Color("#F0F0F0")), // Light gray

		// System/process info
		Hostname: lipgloss.NewStyle().Foreground(lipgloss.Color("#FFB347")),            // Light orange
		PID:      lipgloss.NewStyle().Foreground(lipgloss.Color("#B19CD9")),            // Light lavender
		Service:  lipgloss.NewStyle().Foreground(lipgloss.Color("#87CEEB")).Bold(true), // Light blue

		// File references
		Filename: lipgloss.NewStyle().Foreground(lipgloss.Color("#FFA07A")), // Light salmon
		LineNum:  lipgloss.NewStyle().Foreground(lipgloss.Color("#B19CD9")), // Light lavender

		// Punctuation/structure
		Bracket: lipgloss.NewStyle().Foreground(lipgloss.Color("#808080")), // Gray
		Quote:   lipgloss.NewStyle().Foreground(lipgloss.Color("#808080")), // Gray
		Equals:  lipgloss.NewStyle().Foreground(lipgloss.Color("#808080")), // Gray

		// Search highlighting
		SearchHighlight: lipgloss.NewStyle().Bold(true),
		JSONSearchHighlight: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFA500")).Bold(true),
		UnifiedSearchHighlight: lipgloss.NewStyle().
			Background(lipgloss.Color("214")). // ANSI bright orange background
			Foreground(lipgloss.Color("0")).   // ANSI black text
			Bold(true),
	}
}

// GetLogLevelStyle returns the appropriate style for a log level
func (t *ColorTheme) GetLogLevelStyle(level string) lipgloss.Style {
	switch level {
	case "ERROR", "error", "FATAL", "fatal", "CRIT", "critical":
		return t.Error
	case "WARN", "warn", "WARNING", "warning":
		return t.Warning
	case "INFO", "info":
		return t.Info
	case "DEBUG", "debug", "TRACE", "trace":
		return t.Debug
	default:
		return lipgloss.NewStyle() // No styling
	}
}

// GetHTTPStatusStyle returns the appropriate style for HTTP status codes
func (t *ColorTheme) GetHTTPStatusStyle(statusCode string) lipgloss.Style {
	if len(statusCode) >= 1 {
		switch statusCode[0] {
		case '2': // 2xx - Success
			return t.StatusOK
		case '4': // 4xx - Client Error
			return t.StatusWarn
		case '5': // 5xx - Server Error
			return t.StatusError
		case '3': // 3xx - Redirection
			return t.Info
		default:
			return lipgloss.NewStyle()
		}
	}
	return lipgloss.NewStyle()
}
