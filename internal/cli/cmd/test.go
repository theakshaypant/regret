package cmd

import (
	"fmt"
	"os"
	"regexp"
	"time"

	"github.com/spf13/cobra"
	"github.com/theakshaypant/regret"
	"github.com/theakshaypant/regret/internal/cli/output"
)

var (
	testSize int
)

// testCmd represents the test command
var testCmd = &cobra.Command{
	Use:   "test <pattern>",
	Short: "Test pattern with adversarial inputs",
	Long: `Test generates adversarial inputs (pump patterns) and tests the regex
pattern against them to detect actual ReDoS behavior.

This command:
  - Generates pattern-specific pump inputs
  - Measures actual matching time
  - Detects exponential/polynomial growth
  - Validates theoretical analysis

Use this to confirm ReDoS vulnerabilities with real testing.`,
	Example: `  # Test with default size
  regret test "(a+)+"
  
  # Test with specific size
  regret test "(a+)+" --size=20
  
  # Test with verbose output
  regret test "(a+)+" --size=30 --verbose`,
	Args: cobra.ExactArgs(1),
	Run:  runTest,
}

func init() {
	rootCmd.AddCommand(testCmd)
	testCmd.Flags().IntVarP(&testSize, "size", "s", 20, "Pump size (number of repetitions)")
}

func runTest(cmd *cobra.Command, args []string) {
	pattern := args[0]

	formatter := output.NewFormatter(outputFormat, noColor)

	if verbose {
		formatter.PrintInfo("Analyzing pattern for: %s", pattern)
	}

	// Analyze complexity to get pump pattern
	score, err := regret.AnalyzeComplexity(pattern)
	if err != nil {
		formatter.PrintError("Failed to analyze pattern: %v", err)
		os.Exit(1)
	}

	// Check if pump pattern was generated
	if len(score.PumpPattern) == 0 {
		formatter.PrintError("No pump pattern available for this pattern (may be safe or score too low)")
		os.Exit(1)
	}

	// Reconstruct pump pattern for testing
	pump := &regret.PumpPattern{
		Pumps:       score.PumpPattern,
		Suffix:      "x", // Default non-matching suffix
		Description: "Auto-generated from complexity analysis",
	}

	if verbose {
		formatter.PrintInfo("Using auto-generated pump from complexity analysis")
	}

	// Generate test input
	input := pump.Generate(testSize)

	fmt.Printf("Generated input (n=%d): %s\n", testSize, truncate(input, 60))
	fmt.Printf("Input length: %d characters\n\n", len(input))

	// Compile regex
	re, err := regexp.Compile(pattern)
	if err != nil {
		formatter.PrintError("Failed to compile pattern: %v", err)
		os.Exit(1)
	}

	// Test matching
	fmt.Println("Testing...")
	start := time.Now()

	// Set timeout
	done := make(chan bool, 1)
	go func() {
		_ = re.MatchString(input)
		done <- true
	}()

	timeout := time.After(5 * time.Second)
	select {
	case <-done:
		elapsed := time.Since(start)
		fmt.Printf("✓ Completed in: %v\n", elapsed)

		if elapsed > 100*time.Millisecond {
			formatter.PrintWarning("Slow matching detected! (>100ms)")
			if elapsed > 1*time.Second {
				formatter.PrintWarning("Pattern exhibits ReDoS behavior")
			}
		} else {
			formatter.PrintSuccess("Fast matching (<%v)", elapsed)
		}

	case <-timeout:
		formatter.PrintError("Timeout (5s) - Pattern exhibits severe ReDoS")
		fmt.Println("\n⚠️  This pattern causes catastrophic backtracking!")
		fmt.Println("The regex engine is taking exponential time to match.")
		os.Exit(1)
	}

	// Additional analysis
	fmt.Println("\nRecommendation:")
	if score.HasEDA {
		fmt.Println("  ⛔ This pattern has exponential time complexity")
		fmt.Println("  → Avoid using this pattern with untrusted input")
	} else if score.HasIDA {
		fmt.Println("  ⚠️  This pattern has polynomial time complexity")
		fmt.Println("  → Use with caution, may be slow on large inputs")
	} else {
		fmt.Println("  ✓ Pattern appears safe for typical use")
	}
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
