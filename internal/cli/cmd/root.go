package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	// Global flags
	outputFormat string
	mode         string
	verbose      bool
	quiet        bool
	noColor      bool
	configFile   string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "regret",
	Short: "Regex threat detection and analysis tool",
	Long: `Regret (Regex Threat) is a static analysis tool for detecting
dangerous regex patterns that can lead to catastrophic backtracking (ReDoS).

It provides validation, complexity analysis, and adversarial input generation
to help you write safe and performant regular expressions.`,
	Version: "0.1.0",
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "text", "Output format (text|json|table)")
	rootCmd.PersistentFlags().StringVarP(&mode, "mode", "m", "balanced", "Validation mode (fast|balanced|thorough)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "Quiet mode (errors only)")
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "Disable color output")
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "", "Config file path")
}

func initConfig() {
	if noColor {
		os.Setenv("NO_COLOR", "1")
	}

	if configFile != "" {
		// TODO: Load config file
	}
}

// getValidationMode converts string mode to regret validation mode
func getValidationMode() string {
	switch mode {
	case "fast":
		return "fast"
	case "thorough":
		return "thorough"
	default:
		return "balanced"
	}
}

// exitWithError prints error and exits with code 1
func exitWithError(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "Error: "+format+"\n", args...)
	os.Exit(1)
}
