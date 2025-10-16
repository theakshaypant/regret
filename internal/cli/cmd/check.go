package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/theakshaypant/regret"
	"github.com/theakshaypant/regret/internal/cli/output"
)

// checkCmd represents the check command
var checkCmd = &cobra.Command{
	Use:   "check <pattern>",
	Short: "Quickly validate a regex pattern",
	Long: `Check validates a regex pattern for ReDoS vulnerabilities.

This command provides a quick safety check and returns:
  - Exit code 0: Pattern is safe
  - Exit code 1: Pattern is unsafe or error occurred

Perfect for CI/CD pipelines and quick validation.`,
	Example: `  # Check a pattern
  regret check "(a+)+"
  
  # Check with different modes
  regret check "(a+)+" --mode=fast
  regret check "(a+)+" --mode=thorough
  
  # JSON output for scripting
  regret check "(a+)+" --output=json`,
	Args: cobra.ExactArgs(1),
	Run:  runCheck,
}

func init() {
	rootCmd.AddCommand(checkCmd)
}

func runCheck(cmd *cobra.Command, args []string) {
	pattern := args[0]

	formatter := output.NewFormatter(outputFormat, noColor)

	// Validate pattern
	opts := getOptions()
	issues, err := regret.ValidateWithOptions(pattern, opts)
	if err != nil {
		formatter.PrintError("Failed to validate pattern: %v", err)
		os.Exit(1)
	}

	// Analyze complexity for scoring
	score, err := regret.AnalyzeComplexity(pattern)
	if err != nil {
		formatter.PrintError("Failed to analyze complexity: %v", err)
		os.Exit(1)
	}

	// Create result
	result := &output.CheckResult{
		Pattern:    pattern,
		Safe:       len(issues) == 0 && score.Safe,
		Complexity: score.TimeComplexity.String(),
		Score:      score.Overall,
		Issues:     issues,
	}

	// Format and print
	if err := formatter.FormatCheckResult(result); err != nil {
		formatter.PrintError("Failed to format output: %v", err)
		os.Exit(1)
	}

	// Exit with appropriate code
	if !result.Safe {
		os.Exit(1)
	}
}

func getOptions() *regret.Options {
	opts := regret.DefaultOptions()

	// Set validation mode
	switch mode {
	case "fast":
		opts.Mode = regret.Fast
	case "thorough":
		opts.Mode = regret.Thorough
	default:
		opts.Mode = regret.Balanced
	}

	return opts
}
