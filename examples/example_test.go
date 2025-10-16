package examples_test

import (
	"fmt"
	"log"
	"time"

	"github.com/theakshaypant/regret"
)

// ExampleIsSafe demonstrates quick safety checking.
func ExampleIsSafe() {
	patterns := []string{
		"^[a-z]+$",      // safe
		"(a+)+",         // unsafe - nested quantifiers
		"\\d{3}-\\d{4}", // safe
	}

	for _, pattern := range patterns {
		if regret.IsSafe(pattern) {
			fmt.Printf("%s: safe\n", pattern)
		} else {
			fmt.Printf("%s: unsafe\n", pattern)
		}
	}
}

// ExampleValidate demonstrates basic validation.
func ExampleValidate() {
	pattern := "(a+)+"

	issues, err := regret.Validate(pattern)
	if err != nil {
		log.Fatalf("validation error: %v", err)
	}

	if len(issues) > 0 {
		fmt.Printf("Found %d issue(s):\n", len(issues))
		for _, issue := range issues {
			fmt.Printf("  - %s: %s\n", issue.Type, issue.Message)
			fmt.Printf("    Severity: %s\n", issue.Severity)
			if issue.Example != "" {
				fmt.Printf("    Example attack: %s\n", issue.Example)
			}
			if issue.Suggestion != "" {
				fmt.Printf("    Suggestion: %s\n", issue.Suggestion)
			}
		}
	}
}

// ExampleValidateWithOptions demonstrates configurable validation.
func ExampleValidateWithOptions() {
	pattern := "a*a*"

	opts := &regret.Options{
		Mode:               regret.Balanced,
		Timeout:            100 * time.Millisecond,
		MaxComplexityScore: 70,
		Checks:             regret.CheckDefault | regret.CheckPolynomialDegree,
	}

	issues, err := regret.ValidateWithOptions(pattern, opts)
	if err != nil {
		log.Fatalf("validation error: %v", err)
	}

	for _, issue := range issues {
		if issue.Severity >= regret.High {
			fmt.Printf("High severity issue: %s\n", issue.Message)
		}
	}
}

// ExampleAnalyzeComplexity demonstrates detailed complexity analysis.
func ExampleAnalyzeComplexity() {
	pattern := "(a+)+"

	score, err := regret.AnalyzeComplexity(pattern)
	if err != nil {
		log.Fatalf("analysis error: %v", err)
	}

	fmt.Printf("Pattern: %s\n", pattern)
	fmt.Printf("Complexity Score: %d/100\n", score.Overall)
	fmt.Printf("Time Complexity: %s\n", score.TimeComplexity)

	if score.HasEDA {
		fmt.Printf("⚠️  Exponential Degree of Ambiguity detected!\n")
	}
	if score.HasIDA {
		fmt.Printf("⚠️  Polynomial Degree: O(n^%d)\n", score.PolynomialDegree)
	}

	if score.WorstCaseInput != "" {
		fmt.Printf("Worst-case input: %s\n", score.WorstCaseInput)
	}

	fmt.Printf("\nExplanation: %s\n", score.Explanation)
}

// ExamplePumpPattern_Generate demonstrates using a pump pattern.
func ExamplePumpPattern_Generate() {
	pump := &regret.PumpPattern{
		Prefix:      "",
		Pumps:       []string{"a"},
		Suffix:      "x",
		Interleave:  false,
		Description: "Exposes exponential backtracking in (a+)+",
	}

	// Generate inputs
	input10 := pump.Generate(10) // "aaaaaaaaaax"
	input20 := pump.Generate(20) // "aaaaaaaaaaaaaaaaaaax"

	fmt.Printf("10 iterations: %s\n", input10)
	fmt.Printf("20 iterations: %s\n", input20)
}

// ExamplePumpPattern_GenerateSequence demonstrates generating a sequence of inputs.
func ExamplePumpPattern_GenerateSequence() {
	pump := &regret.PumpPattern{
		Prefix: "",
		Pumps:  []string{"a", "b"},
		Suffix: "x",
	}

	// Generate sequence: 5, 10, 15, 20
	sequence := pump.GenerateSequence(5, 20, 5)
	for i, input := range sequence {
		fmt.Printf("Input %d: %s\n", i+1, input)
	}
}

// ExampleOptions demonstrates creating custom options.
func ExampleOptions() {
	// Fast options for hot paths
	fast := regret.FastOptions()
	fmt.Printf("Fast mode timeout: %v\n", fast.Timeout)

	// Default balanced options
	balanced := regret.DefaultOptions()
	fmt.Printf("Balanced mode: %s\n", balanced.Mode)

	// Thorough options for security auditing
	thorough := regret.ThoroughOptions()
	fmt.Printf("Thorough checks: %v\n", thorough.Checks == regret.CheckAll)

	// Custom options
	custom := &regret.Options{
		Mode:               regret.Balanced,
		Timeout:            50 * time.Millisecond,
		Checks:             regret.CheckDefault,
		MaxComplexityScore: 60,
		StrictMode:         true,
	}
	fmt.Printf("Custom max complexity: %d\n", custom.MaxComplexityScore)
}

// ExampleComplexity_String demonstrates complexity notation.
func ExampleComplexity_String() {
	complexities := []regret.Complexity{
		regret.Constant,
		regret.Linear,
		regret.Quadratic,
		regret.Cubic,
		regret.Exponential,
	}

	for _, c := range complexities {
		fmt.Printf("%s\n", c)
	}

	// Output:
	// O(1)
	// O(n)
	// O(n²)
	// O(n³)
	// O(2^n)
}

// ExampleValidationMode demonstrates different validation modes.
func ExampleValidationMode() {
	modes := []regret.ValidationMode{
		regret.Fast,
		regret.Balanced,
		regret.Thorough,
	}

	for _, mode := range modes {
		fmt.Printf("Mode: %s\n", mode)
	}

	// Output:
	// Mode: fast
	// Mode: balanced
	// Mode: thorough
}
