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
	JSONKey   lipgloss.Style
	JSONValue lipgloss.Style
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
	Filename  lipgloss.Style
	LineNum   lipgloss.Style
	
	// Punctuation/structure
	Bracket lipgloss.Style
	Quote   lipgloss.Style
	Equals  lipgloss.Style
	
	// Search highlighting - Unified style for all log formats
	SearchHighlight     lipgloss.Style // Deprecated - use UnifiedSearchHighlight
	JSONSearchHighlight lipgloss.Style // Deprecated - use UnifiedSearchHighlight  
	UnifiedSearchHighlight lipgloss.Style // Bright Orange + Adaptive - used for all search highlighting
}

// NewAdaptiveTheme creates a color theme that adapts to the terminal
func NewAdaptiveTheme() *ColorTheme {
	return &ColorTheme{
		// Log levels with semantic colors using adaptive ANSI colors
		Error:   lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#D63031", Dark: "#FF6B6B"}).Bold(true), // Red
		Warning: lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#E17000", Dark: "#FFE66D"}),             // Yellow
		Info:    lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#00B894", Dark: "#4ECDC4"}),             // Cyan
		Debug:   lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#00A085", Dark: "#95E1D3"}),             // Light green
		
		// HTTP status codes
		StatusOK:    lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#00B894", Dark: "#6BCF7F"}),         // Green
		StatusWarn:  lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#E17000", Dark: "#FFD93D"}),         // Yellow
		StatusError: lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#D63031", Dark: "#FF6B6B"}).Bold(true), // Red
		
		// General components
		Timestamp: lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#636E72", Dark: "#A8A8A8"}),           // Gray
		IP:        lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#0984E3", Dark: "#74B9FF"}),           // Blue
		URL:       lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#00CEC9", Dark: "#81ECEC"}).Underline(true), // Cyan underlined
		Method:    lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#E84393", Dark: "#FD79A8"}).Bold(true), // Pink
		
		// JSON/structured data
		JSONKey:    lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#8E44AD", Dark: "#DDA0DD"}),          // Plum
		JSONValue:  lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#2D3436", Dark: "#F0F0F0"}),          // Light gray
		JSONString: lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#00A085", Dark: "#98FB98"}),          // Pale green
		JSONNumber: lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#0984E3", Dark: "#87CEEB"}),          // Sky blue
		
		// Logfmt
		LogfmtKey:   lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#8E44AD", Dark: "#DDA0DD"}),         // Plum
		LogfmtValue: lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#2D3436", Dark: "#F0F0F0"}),         // Light gray
		
		// System/process info
		Hostname: lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#E17000", Dark: "#FFB347"}),            // Peach
		PID:      lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#6C5CE7", Dark: "#B19CD9"}),            // Lavender
		Service:  lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#0984E3", Dark: "#87CEEB"}).Bold(true), // Sky blue
		
		// File references
		Filename: lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#E17000", Dark: "#FFA07A"}),            // Light salmon
		LineNum:  lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#6C5CE7", Dark: "#B19CD9"}),            // Lavender
		
		// Punctuation/structure (subtle)
		Bracket: lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#636E72", Dark: "#808080"}),             // Gray
		Quote:   lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#636E72", Dark: "#808080"}),             // Gray
		Equals:  lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#636E72", Dark: "#808080"}),             // Gray
		
		// Search highlighting (bold text, no background)
		SearchHighlight: lipgloss.NewStyle().
			Bold(true),                             // Bold for visibility
		
		// JSON-specific search highlighting with high visibility foreground colors
		JSONSearchHighlight: lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#E17000", Dark: "#FFA500"}).  // Orange: dark orange for light theme, bright orange for dark theme
			Bold(true),                                                              // Bold for extra visibility
		// Alternative options (comment/uncomment to change):
		// Red variant:    Foreground(lipgloss.AdaptiveColor{Light: "#D63031", Dark: "#FF6B6B"}).Bold(true)
		// Yellow variant: Foreground(lipgloss.AdaptiveColor{Light: "#B8860B", Dark: "#FFFF00"}).Bold(true)
		
		// Unified search highlighting - Bright Orange + Adaptive (replaces above styles)
		UnifiedSearchHighlight: lipgloss.NewStyle().
			Background(lipgloss.AdaptiveColor{Light: "#FB923C", Dark: "#DC2626"}).
			Foreground(lipgloss.AdaptiveColor{Light: "#000000", Dark: "#FEF3C7"}).
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
