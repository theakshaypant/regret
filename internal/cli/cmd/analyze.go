package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/theakshaypant/regret"
	"github.com/theakshaypant/regret/internal/cli/output"
)

// analyzeCmd represents the analyze command
var analyzeCmd = &cobra.Command{
	Use:   "analyze <pattern>",
	Short: "Perform detailed complexity analysis on a regex pattern",
	Long: `Analyze performs comprehensive complexity analysis on a regex pattern.

This command provides:
  - Complexity scoring (0-100 scale)
  - Big O notation (O(1), O(n), O(n²), O(2^n))
  - EDA/IDA detection
  - Detailed metrics
  - Actionable suggestions

Use this for in-depth understanding of pattern performance.`,
	Example: `  # Analyze a pattern
  regret analyze "(a+)+"
  
  # Thorough analysis
  regret analyze "(a+)+" --mode=thorough
  
  # JSON output
  regret analyze "(a+)+" --output=json
  
  # Table format
  regret analyze "(a+)+" --output=table`,
	Args: cobra.ExactArgs(1),
	Run:  runAnalyze,
}

func init() {
	rootCmd.AddCommand(analyzeCmd)
}

func runAnalyze(cmd *cobra.Command, args []string) {
	pattern := args[0]

	formatter := output.NewFormatter(outputFormat, noColor)

	if verbose {
		formatter.PrintInfo("Analyzing pattern: %s", pattern)
		formatter.PrintInfo("Mode: %s", mode)
	}

	// Get validation issues
	opts := getOptions()
	issues, err := regret.ValidateWithOptions(pattern, opts)
	if err != nil {
		formatter.PrintError("Failed to validate pattern: %v", err)
		os.Exit(1)
	}

	// Get complexity analysis
	score, err := regret.AnalyzeComplexity(pattern)
	if err != nil {
		formatter.PrintError("Failed to analyze complexity: %v", err)
		os.Exit(1)
	}

	// Create result
	result := &output.AnalysisResult{
		Pattern: pattern,
		Score:   score,
		Issues:  issues,
	}

	// Format and print
	if err := formatter.FormatAnalysisResult(result); err != nil {
		formatter.PrintError("Failed to format output: %v", err)
		os.Exit(1)
	}

	if verbose {
		formatter.PrintInfo("Analysis complete")
	}
}
