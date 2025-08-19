package cmd

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	"github.com/spf13/cobra"

	"github.com/joshi4/splash/colorizer"
	"github.com/joshi4/splash/parser"
)

// Command flags
var (
	searchPattern string
	regexPattern  string
	lightTheme    bool
	darkTheme     bool
	noColor       bool
)

// createSplashHeader creates a colorful SPLASH header using log colors
func createSplashHeader() string {
	theme := colorizer.NewAdaptiveTheme()

	// ASCII art for SPLASH using block characters - each letter is 5 lines tall
	sArt := []string{
		"█████",
		"█    ",
		"█████",
		"    █",
		"█████",
	}

	pArt := []string{
		"█████",
		"█   █",
		"█████",
		"█    ",
		"█    ",
	}

	lArt := []string{
		"█    ",
		"█    ",
		"█    ",
		"█    ",
		"█████",
	}

	aArt := []string{
		" ███ ",
		"█   █",
		"█████",
		"█   █",
		"█   █",
	}

	s2Art := []string{
		"█████",
		"█    ",
		"█████",
		"    █",
		"█████",
	}

	hArt := []string{
		"█   █",
		"█   █",
		"█████",
		"█   █",
		"█   █",
	}

	// Combine all letters with colors
	var lines []string
	for i := 0; i < 5; i++ {
		line := "  " +
			theme.Error.Render(sArt[i]) + " " + // Red
			theme.Warning.Render(pArt[i]) + " " + // Yellow
			theme.Info.Render(lArt[i]) + " " + // Cyan
			theme.StatusOK.Render(aArt[i]) + " " + // Green
			theme.IP.Render(s2Art[i]) + " " + // Blue
			theme.Method.Render(hArt[i]) // Pink
		lines = append(lines, line)
	}

	header := "\n"
	for _, line := range lines {
		header += line + "\n"
	}

	subtitle := lipgloss.NewStyle().
		Bold(true).
		Render("  Add color to your logs")

	return header + "\n" + subtitle + "\n"
}

// createColorizerWithTheme creates a colorizer with proper theme detection
func createColorizerWithTheme() *colorizer.Colorizer {
	c := colorizer.NewColorizer()

	// Handle explicit theme flags
	if lightTheme && darkTheme {
		fmt.Fprintf(os.Stderr, "Cannot use both --light and --dark flags simultaneously\n")
		os.Exit(1)
	}

	if lightTheme {
		// Force light theme colors by setting the colorizer to use light theme
		c.SetTheme(colorizer.NewLightTheme())
	} else if darkTheme {
		// Force dark theme colors
		c.SetTheme(colorizer.NewDarkTheme())
	}
	// Otherwise use default adaptive theme

	return c
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "splash",
	Short: "Add color to your logs",
	Long:  createSplashHeader() + "\nSplash transforms streams of boring plaintext into colorful and easy to read logs.\nUse Splash to easily scan and debug issues from your logs.\n\nSupported formats: JSON, Logfmt, Syslog, Apache, Nginx, Rails, Docker,\nKubernetes, Heroku, Go standard logs, and more.",
	Example: `  tail -f /var/log/app.log | splash
  docker logs mycontainer | splash
  kubectl logs pod-name | splash -s "ERROR"
  cat access.log | splash -r "[45]\d\d"`,
	Run: func(cmd *cobra.Command, _ []string) {
		// If stdin is not a pipe and no search flags are provided, show usage
		if !isStdinFromPipe() && searchPattern == "" && regexPattern == "" {
			_ = cmd.Help()
			return
		}

		if err := runSplash(); err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

// isStdinFromPipe checks if stdin is from a pipe (not a terminal)
func isStdinFromPipe() bool {
	stat, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return (stat.Mode() & os.ModeCharDevice) == 0
}

// runSplash is the main function that reads from stdin and writes to stdout
func runSplash() error {
	// Handle color profile and theme detection
	if noColor {
		lipgloss.SetColorProfile(termenv.Ascii)
	} else {
		// Force color output even when stdout is not a TTY (when piping output)
		// This ensures colors work when doing: echo "log" | splash | less -R
		lipgloss.SetColorProfile(termenv.TrueColor)
	}

	// Create a context that will be canceled when we receive a signal
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	// Start a goroutine to handle signals
	go func() {
		<-sigChan
		cancel()
	}()

	// Create optimized parser and colorizer with theme detection
	logParser := parser.NewParser()
	logColorizer := createColorizerWithTheme()

	// Set search patterns if provided
	if searchPattern != "" && regexPattern != "" {
		return fmt.Errorf("cannot use both --search and --regexp flags simultaneously")
	}

	if searchPattern != "" {
		logColorizer.SetSearchString(searchPattern)
	} else if regexPattern != "" {
		err := logColorizer.SetSearchRegex(regexPattern)
		if err != nil {
			return fmt.Errorf("invalid regex pattern: %v", err)
		}
	}

	// Read from stdin and write to stdout
	scanner := bufio.NewScanner(os.Stdin)

	// Channel to signal when reading is done
	done := make(chan bool)

	go func() {
		defer close(done)
		for scanner.Scan() {
			select {
			case <-ctx.Done():
				return
			default:
				line := scanner.Text()
				// Detect log format for this line using optimized parser
				format := logParser.DetectFormat(line)
				// Apply colors based on detected format
				colorizedLine := logColorizer.ColorizeLog(line, format)
				fmt.Println(colorizedLine)
			}
		}

		// Check for scanner errors
		if err := scanner.Err(); err != nil && err != io.EOF {
			fmt.Fprintf(os.Stderr, "Error reading from stdin: %v\n", err)
		}
	}()

	// Wait for either the context to be canceled or reading to complete
	select {
	case <-ctx.Done():
		// Check for upgrades before exiting due to signal
		CheckForUpgradesOnExit()
		return nil
	case <-done:
		// Check for upgrades before normal exit
		CheckForUpgradesOnExit()
		return nil
	}
}

func init() {
	// Search flags
	rootCmd.Flags().StringVarP(&searchPattern, "search", "s", "", "search for all instances of a string")
	rootCmd.Flags().StringVarP(&regexPattern, "regexp", "r", "", "search for text that matches a regexp")

	// Theme flags
	rootCmd.Flags().BoolVar(&lightTheme, "light", false, "force light theme colors (for light terminal backgrounds)")
	rootCmd.Flags().BoolVar(&darkTheme, "dark", false, "force dark theme colors (for dark terminal backgrounds)")
	rootCmd.Flags().BoolVar(&noColor, "no-color", false, "disable all colors")
}
