package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/getsavvyinc/upgrade-cli"
	"github.com/spf13/cobra"
)

const (
	owner = "joshi4"
	repo  = "splash"
)

// Version should be set at build time via ldflags
// For development, we use a placeholder version
var version = "v0.1.0-dev"

// upgradeCmd represents the upgrade command
var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Upgrade splash to the latest version",
	Long: `Upgrade splash to the latest version available on GitHub.

This command will check for the latest release on GitHub and upgrade
your splash installation if a newer version is available.`,
	Example: `  splash upgrade`,
	Run: func(_ *cobra.Command, _ []string) {
		if err := runUpgrade(); err != nil {
			fmt.Fprintf(os.Stderr, "Upgrade failed: %v\n", err)
			os.Exit(1)
		}
	},
}

// runUpgrade performs the upgrade check and execution
func runUpgrade() error {
	return runUpgradeWithOptions(true, true) // verbose=true, interactive=true
}

// runUpgradeWithOptions performs the upgrade check and execution with configurable options
func runUpgradeWithOptions(verbose, interactive bool) error {
	// Get the path to the currently running executable
	executablePath, err := os.Executable()
	if err != nil {
		if verbose {
			return fmt.Errorf("failed to get executable path: %w", err)
		}
		return nil // Silently fail for non-verbose calls
	}

	// Create a new upgrader instance
	upgrader := upgrade.NewUpgrader(owner, repo, executablePath)

	// Check if a new version is available
	ctx := context.Background()
	hasNewVersion, err := upgrader.IsNewVersionAvailable(ctx, version)
	if err != nil {
		if verbose {
			// Check if this is due to no releases being available
			errStr := err.Error()
			if errStr == "failed to parse latest version:  with err Malformed version: " ||
				errStr == "failed to parse latest version: not found" ||
				errStr == "no releases found" {
				fmt.Println("No releases are available for splash yet")
				return nil
			}
			return fmt.Errorf("failed to check for new version: %w", err)
		}
		return nil // Silently fail for non-verbose calls
	}

	if !hasNewVersion {
		if verbose {
			fmt.Println("splash is already up to date")
		}
		return nil
	}

	// If interactive, prompt user
	if interactive {
		fmt.Fprintf(os.Stderr, "\n📦 A new version of splash is available!\n")
		fmt.Fprintf(os.Stderr, "Would you like to upgrade now? (y/N): ")

		// Read user input
		reader := bufio.NewReader(os.Stdin)
		response, err := reader.ReadString('\n')
		if err != nil {
			// If we can't read input, fallback to showing the manual command
			fmt.Fprintf(os.Stderr, "Run 'splash upgrade' to update.\n")
			return nil
		}

		response = strings.ToLower(strings.TrimSpace(response))
		if response != "y" && response != "yes" {
			// User declined, show them the manual command
			fmt.Fprintf(os.Stderr, "You can upgrade later by running 'splash upgrade'\n")
			return nil
		}
	}

	// Perform the upgrade
	if verbose || interactive {
		if interactive {
			fmt.Fprintf(os.Stderr, "Upgrading splash to the latest version...\n")
		} else {
			fmt.Println("Upgrading splash to the latest version...")
		}
	}
	
	if err := upgrader.Upgrade(ctx, version); err != nil {
		if verbose || interactive {
			if interactive {
				fmt.Fprintf(os.Stderr, "Upgrade failed: %v\n", err)
				fmt.Fprintf(os.Stderr, "You can try again with 'splash upgrade'\n")
			}
			return fmt.Errorf("failed to upgrade: %w", err)
		}
		return nil // Silently fail for non-verbose calls
	}

	if verbose || interactive {
		if interactive {
			fmt.Fprintf(os.Stderr, "splash has been successfully upgraded to the latest version!\n")
			fmt.Fprintf(os.Stderr, "Please restart splash to use the new version.\n")
		} else {
			fmt.Println("splash has been successfully upgraded to the latest version")
		}
	}
	return nil
}

// CheckForUpgradesOnExit checks for available upgrades and prompts the user
// This is called when splash is about to exit
func CheckForUpgradesOnExit() {
	// Reuse the upgrade logic but with silent error handling and interactive mode
	err := runUpgradeWithOptions(false, true) // verbose=false, interactive=true
	if err != nil {
		// Exit with error code on upgrade failure during interactive mode
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(upgradeCmd)
}
