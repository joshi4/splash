package cmd

import (
	"context"
	"fmt"
	"os"

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
	// Get the path to the currently running executable
	executablePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	// Create a new upgrader instance
	upgrader := upgrade.NewUpgrader(owner, repo, executablePath)

	// Check if a new version is available
	ctx := context.Background()
	hasNewVersion, err := upgrader.IsNewVersionAvailable(ctx, version)
	if err != nil {
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

	if !hasNewVersion {
		fmt.Println("splash is already up to date")
		return nil
	}

	// Perform the upgrade
	fmt.Println("Upgrading splash to the latest version...")
	if err := upgrader.Upgrade(ctx, version); err != nil {
		return fmt.Errorf("failed to upgrade: %w", err)
	}

	fmt.Println("splash has been successfully upgraded to the latest version")
	return nil
}

// CheckForUpgradesOnExit checks for available upgrades and prompts the user
// This is called when splash is about to exit
func CheckForUpgradesOnExit() {
	// Get the path to the currently running executable
	executablePath, err := os.Executable()
	if err != nil {
		return // Silently fail to avoid disrupting normal exit
	}

	// Create a new upgrader instance
	upgrader := upgrade.NewUpgrader(owner, repo, executablePath)

	// Check if a new version is available
	ctx := context.Background()
	hasNewVersion, err := upgrader.IsNewVersionAvailable(ctx, version)
	if err != nil {
		// Silently fail if no releases exist or other errors occur
		// to avoid disrupting normal exit
		return
	}

	if !hasNewVersion {
		return // No update available
	}

	// Prompt user about available upgrade
	fmt.Fprintf(os.Stderr, "\n📦 A new version of splash is available! Run 'splash upgrade' to update.\n")
}

func init() {
	rootCmd.AddCommand(upgradeCmd)
}
