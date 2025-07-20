/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/joshi4/splash/colorizer"
	"github.com/joshi4/splash/parser"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "splash",
	Short: "Add color to your logs",
	Long:  "Add color to your logs",
	Run: func(cmd *cobra.Command, args []string) {
		runSplash()
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

// runSplash is the main function that reads from stdin and writes to stdout
func runSplash() {
	// Create a context that will be cancelled when we receive a signal
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	// Start a goroutine to handle signals
	go func() {
		sig := <-sigChan
		fmt.Fprintf(os.Stderr, "\nReceived signal: %v, shutting down gracefully...\n", sig)
		cancel()
	}()

	// Create optimized parser and colorizer
	logParser := parser.NewParser()
	logColorizer := colorizer.NewColorizer()

	// Read from stdin and write to stdout
	scanner := bufio.NewScanner(os.Stdin)
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
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.splash.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}


