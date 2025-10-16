package examples

import (
	"fmt"
	"regexp"
	"time"

	"github.com/theakshaypant/regret"
)

// ExamplePumpIntegration_Basic demonstrates automatic pump pattern generation.
func ExamplePumpIntegration_Basic() {
	// Analyze a dangerous pattern
	score, err := regret.AnalyzeComplexity("(a+)+")
	if err != nil {
		panic(err)
	}

	fmt.Printf("Pattern: (a+)+\n")
	fmt.Printf("Score: %d/100\n", score.Overall)
	fmt.Printf("Safe: %v\n", score.Safe)

	// For unsafe patterns (score >= 50), pump patterns are automatically generated
	if len(score.PumpPattern) > 0 {
		fmt.Printf("Pump Components: %v\n", score.PumpPattern)
		fmt.Printf("Worst Case Input: %q\n", score.WorstCaseInput)
		fmt.Printf("Input Length: %d\n", len(score.WorstCaseInput))
	}

	// Output (pattern-dependent):
	// Pattern: (a+)+
	// Score: 70/100
	// Safe: false
	// Pump Components: [a]
	// Worst Case Input: "aaaaaaaaaaaaaaaaaaaax"
	// Input Length: 21
}

// ExamplePumpIntegration_SecurityTesting shows using worst-case inputs for security testing.
func ExamplePumpIntegration_SecurityTesting() {
	patterns := []string{
		"(a+)+",
		"(a|ab)+",
		"a*a+",
		"^[a-z]+$", // Safe pattern
	}

	for _, pattern := range patterns {
		score, err := regret.AnalyzeComplexity(pattern)
		if err != nil {
			fmt.Printf("Pattern %q: ERROR - %v\n", pattern, err)
			continue
		}

		fmt.Printf("Pattern: %s\n", pattern)
		fmt.Printf("  Score: %d/100\n", score.Overall)

		if score.Overall >= 50 {
			if score.WorstCaseInput != "" {
				fmt.Printf("  ⚠️  UNSAFE - Adversarial input available\n")
				fmt.Printf("  Test Input: %q\n", score.WorstCaseInput)
			} else {
				fmt.Printf("  ⚠️  UNSAFE - No pump pattern generated\n")
			}
		} else {
			fmt.Printf("  ✓ SAFE\n")
		}
		fmt.Println()
	}
}

// ExamplePumpIntegration_Benchmarking demonstrates using pump patterns for performance testing.
func ExamplePumpIntegration_Benchmarking() {
	pattern := "(a+)+"

	score, err := regret.AnalyzeComplexity(pattern)
	if err != nil {
		panic(err)
	}

	if score.WorstCaseInput == "" {
		fmt.Println("No worst-case input available")
		return
	}

	// WARNING: This pattern is unsafe - only use for demonstration!
	// Do NOT use in production without proper safeguards
	re := regexp.MustCompile(pattern)

	// Test with progressively larger inputs based on the pump pattern
	testInput := score.WorstCaseInput

	fmt.Printf("Testing pattern: %s\n", pattern)
	fmt.Printf("Base input length: %d\n", len(testInput))
	fmt.Println()

	// Test with the generated worst-case input
	start := time.Now()
	// Note: In a real test, you'd want timeout protection
	matched := re.MatchString(testInput)
	duration := time.Since(start)

	fmt.Printf("Matched: %v\n", matched)
	fmt.Printf("Duration: %v\n", duration)
	fmt.Printf("Input: %q\n", testInput)
}

// ExamplePumpIntegration_MultiPatternAnalysis shows analyzing multiple patterns efficiently.
func ExamplePumpIntegration_MultiPatternAnalysis() {
	type PatternTest struct {
		Name    string
		Pattern string
	}

	tests := []PatternTest{
		{"Nested Quantifiers", "(a+)+"},
		{"Double Nested", "((a+)+)+"},
		{"Overlapping Quant", "a*a+"},
		{"Safe Email", `^[a-z0-9._%+-]+@[a-z0-9.-]+\.[a-z]{2,}$`},
		{"Safe URL", `^https?://[a-z0-9.-]+\.[a-z]{2,}(/.*)?$`},
	}

	fmt.Println("=== Pattern Analysis Report ===")

	for _, test := range tests {
		score, err := regret.AnalyzeComplexity(test.Pattern)
		if err != nil {
			fmt.Printf("%s: ERROR - %v\n\n", test.Name, err)
			continue
		}

		fmt.Printf("%s\n", test.Name)
		fmt.Printf("  Pattern: %s\n", test.Pattern)
		fmt.Printf("  Score: %d/100\n", score.Overall)
		fmt.Printf("  Complexity: %s\n", score.TimeComplexity)

		if len(score.PumpPattern) > 0 {
			fmt.Printf("  Adversarial: YES\n")
			fmt.Printf("    Components: %v\n", score.PumpPattern)
			fmt.Printf("    Sample: %q\n", score.WorstCaseInput)
		} else {
			fmt.Printf("  Adversarial: NO\n")
		}
		fmt.Println()
	}
}

// ExamplePumpIntegration_ConditionalTesting shows conditional testing based on pump availability.
func ExamplePumpIntegration_ConditionalTesting() {
	pattern := "(a+)+"

	score, err := regret.AnalyzeComplexity(pattern)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Analyzing: %s\n", pattern)
	fmt.Printf("Score: %d/100\n", score.Overall)

	// Only perform expensive testing if we have adversarial inputs
	if score.WorstCaseInput != "" {
		fmt.Println("\n⚠️  Unsafe pattern detected!")
		fmt.Println("Running adversarial testing...")
		fmt.Printf("Test input: %q\n", score.WorstCaseInput)

		// In production, you would:
		// 1. Log the finding
		// 2. Run controlled tests with timeouts
		// 3. Consider alternative patterns
		// 4. Add monitoring

		fmt.Println("\n✓ Testing complete")
	} else {
		fmt.Println("\n✓ Pattern appears safe - no adversarial inputs needed")
	}
}

// ExamplePumpIntegration_BatchAnalysis shows efficient batch processing.
func ExamplePumpIntegration_BatchAnalysis() {
	patterns := []string{
		"(a+)+",
		"((b+)+)+",
		"a*a+",
		".*password.*",
		"^[a-z]+$",
	}

	type Result struct {
		Pattern        string
		Score          int
		HasAdversarial bool
		WorstCase      string
	}

	results := make([]Result, 0, len(patterns))

	// Analyze all patterns
	for _, pattern := range patterns {
		score, err := regret.AnalyzeComplexity(pattern)
		if err != nil {
			continue
		}

		results = append(results, Result{
			Pattern:        pattern,
			Score:          score.Overall,
			HasAdversarial: len(score.PumpPattern) > 0,
			WorstCase:      score.WorstCaseInput,
		})
	}

	// Report findings
	fmt.Println("=== Batch Analysis Results ===")
	fmt.Printf("Total patterns analyzed: %d\n\n", len(results))

	unsafeCount := 0
	for _, r := range results {
		if r.Score >= 50 {
			unsafeCount++
			fmt.Printf("⚠️  UNSAFE: %s (score: %d)\n", r.Pattern, r.Score)
			if r.HasAdversarial {
				fmt.Printf("    Adversarial input available: %q\n", r.WorstCase)
			}
		}
	}

	fmt.Printf("\nUnsafe patterns: %d/%d\n", unsafeCount, len(results))
}
