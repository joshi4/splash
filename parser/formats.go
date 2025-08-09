package parser

// Only LogFormat enum and NewParser() are needed now
// The actual detection logic is in detector.go

// LogFormat represents the detected log format
type LogFormat int

const (
	UnknownFormat LogFormat = iota
	JSONFormat
	LogfmtFormat
	ApacheCommonFormat
	NginxFormat
	SyslogFormat
	RsyslogFormat
	GoStandardFormat
	RailsFormat
	DockerFormat
	KubernetesFormat
	HerokuFormat
	GoTestFormat
	JavaExceptionFormat
	PythonExceptionFormat
)

// String returns the string representation of the log format
func (f LogFormat) String() string {
	switch f {
	case JSONFormat:
		return "JSON"
	case LogfmtFormat:
		return "Logfmt"
	case ApacheCommonFormat:
		return "Apache Common"
	case NginxFormat:
		return "Nginx"
	case SyslogFormat:
		return "Syslog"
	case RsyslogFormat:
		return "Rsyslog"
	case GoStandardFormat:
		return "Go Standard"
	case RailsFormat:
		return "Rails"
	case DockerFormat:
		return "Docker"
	case KubernetesFormat:
		return "Kubernetes"
	case HerokuFormat:
		return "Heroku"
	case GoTestFormat:
		return "Go Test"
	case JavaExceptionFormat:
		return "Java Exception"
	case PythonExceptionFormat:
		return "Python Exception"
	default:
		return "Unknown"
	}
}



// DetectFormat is deprecated. Use NewParser().DetectFormat() for stateful detection with better performance and accuracy.
// This function is kept only for backward compatibility and will be removed in a future version.
func DetectFormat(line string) LogFormat {
	// For backward compatibility, create a temporary parser and use it
	// This is less efficient than using a persistent parser but maintains compatibility
	parser := NewParser()
	return parser.DetectFormat(line)
}
