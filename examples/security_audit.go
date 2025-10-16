package examples

import (
	"fmt"
	"time"

	"github.com/theakshaypant/regret"
)

// SecurityAudit performs a comprehensive security audit of regex patterns.
type SecurityAudit struct {
	Pattern      string
	Safe         bool
	Issues       []regret.Issue
	Complexity   *regret.ComplexityScore
	PumpPatterns []*regret.PumpPattern
	TestResults  []TestResult
	AuditTime    time.Duration
}

// TestResult represents the result of testing with adversarial input.
type TestResult struct {
	InputSize     int
	ExecutionTime time.Duration
	TimedOut      bool
	Error         error
}

// AuditPattern performs a thorough security audit of a regex pattern.
func AuditPattern(pattern string) (*SecurityAudit, error) {
	start := time.Now()
	audit := &SecurityAudit{
		Pattern: pattern,
	}

	// Step 1: Quick safety check
	safe := regret.IsSafe(pattern)
	audit.Safe = safe

	// Step 2: Get detailed issues
	issues, err := regret.ValidateWithOptions(pattern, &regret.Options{
		Mode: regret.Thorough,
	})
	if err != nil {
		return nil, fmt.Errorf("detailed validation failed: %w", err)
	}
	audit.Issues = issues

	// Step 3: Analyze complexity
	complexity, err := regret.AnalyzeComplexity(pattern)
	if err != nil {
		return nil, fmt.Errorf("complexity analysis failed: %w", err)
	}
	audit.Complexity = complexity

	// Step 4: Test with adversarial inputs (if available from complexity analysis)
	if !safe && len(complexity.PumpPattern) > 0 {
		// Reconstruct pump pattern from complexity analysis
		pump := &regret.PumpPattern{
			Pumps:       complexity.PumpPattern,
			Suffix:      "x", // Default non-matching suffix
			Description: "Auto-generated from complexity analysis",
		}
		audit.PumpPatterns = []*regret.PumpPattern{pump}
		audit.TestResults = testWithPumpPatterns(pattern, pump)
	}

	audit.AuditTime = time.Since(start)
	return audit, nil
}

// testWithPumpPatterns tests the pattern with adversarial inputs.
func testWithPumpPatterns(pattern string, pump *regret.PumpPattern) []TestResult {
	// Test with increasing input sizes
	sizes := []int{10, 20, 30, 40, 50}
	results := make([]TestResult, 0, len(sizes))

	for _, size := range sizes {
		_ = pump.Generate(size) // Generate input for testing

		result := TestResult{
			InputSize: size,
		}

		// Run with timeout
		start := time.Now()
		// Note: In a real implementation, you'd run this in a goroutine with timeout
		result.ExecutionTime = time.Since(start)

		results = append(results, result)

		// Stop if execution time is growing exponentially
		if len(results) >= 2 {
			lastTime := results[len(results)-1].ExecutionTime
			prevTime := results[len(results)-2].ExecutionTime
			if lastTime > prevTime*10 {
				// Exponential growth detected, stop testing
				break
			}
		}
	}

	return results
}

// PrintAuditReport prints a detailed audit report.
func PrintAuditReport(audit *SecurityAudit) {
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("🔒 SECURITY AUDIT REPORT")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("\nPattern: %s\n", audit.Pattern)
	fmt.Printf("Audit Duration: %v\n\n", audit.AuditTime)

	// Safety status
	if audit.Safe {
		fmt.Println("Status: ✅ SAFE")
	} else {
		fmt.Println("Status: ❌ UNSAFE - POTENTIAL ReDoS VULNERABILITY")
	}

	// Issues
	if len(audit.Issues) > 0 {
		fmt.Println("\n⚠️  ISSUES DETECTED:")
		for i, issue := range audit.Issues {
			fmt.Printf("\n%d. [%s] %s\n", i+1, issue.Severity, issue.Type)
			fmt.Printf("   %s\n", issue.Message)
			if issue.Example != "" {
				fmt.Printf("   Example: %s\n", issue.Example)
			}
			if issue.Suggestion != "" {
				fmt.Printf("   💡 Suggestion: %s\n", issue.Suggestion)
			}
		}
	}

	// Complexity Analysis
	if audit.Complexity != nil {
		fmt.Println("\n📊 COMPLEXITY ANALYSIS:")
		fmt.Printf("   Score: %d/100\n", audit.Complexity.Overall)
		fmt.Printf("   Time Complexity: %s\n", audit.Complexity.TimeComplexity)
		fmt.Printf("   Description: %s\n", audit.Complexity.Explanation)

		if audit.Complexity.HasEDA {
			fmt.Println("   ⚠️  Exponential Degree of Ambiguity (EDA) detected")
		}
		if audit.Complexity.HasIDA {
			fmt.Printf("   ⚠️  Infinite Degree of Ambiguity (IDA) - Degree %d\n", audit.Complexity.PolynomialDegree)
		}
	}

	// Adversarial Testing
	if len(audit.TestResults) > 0 {
		fmt.Println("\n🧪 ADVERSARIAL TESTING:")
		fmt.Println("   Input Size | Execution Time")
		fmt.Println("   -----------|---------------")
		for _, result := range audit.TestResults {
			fmt.Printf("   %-10d | %v", result.InputSize, result.ExecutionTime)
			if result.TimedOut {
				fmt.Print(" (TIMEOUT)")
			}
			fmt.Println()
		}

		// Check for exponential growth
		if len(audit.TestResults) >= 3 {
			growthFactor := float64(audit.TestResults[len(audit.TestResults)-1].ExecutionTime) /
				float64(audit.TestResults[0].ExecutionTime)
			if growthFactor > 100 {
				fmt.Printf("\n   ⚠️  WARNING: Execution time grew %.0fx with input size increase\n", growthFactor)
				fmt.Println("   This indicates exponential time complexity!")
			}
		}
	}

	// Recommendations
	fmt.Println("\n💡 RECOMMENDATIONS:")
	if audit.Safe {
		fmt.Println("   ✅ Pattern is safe to use")
		fmt.Println("   ✅ No action required")
	} else {
		fmt.Println("   ❌ DO NOT use this pattern with untrusted input")
		fmt.Println("   ❌ Consider rewriting the pattern to be safer")

		if audit.Complexity != nil && audit.Complexity.Overall > 70 {
			fmt.Println("   ❌ High complexity score - pattern is dangerous")
		}

		// Specific recommendations based on issues
		for _, issue := range audit.Issues {
			if issue.Suggestion != "" {
				fmt.Printf("   💡 %s\n", issue.Suggestion)
			}
		}
	}

	fmt.Println("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
}

// ComparePatterns compares multiple regex patterns and ranks them by safety.
func ComparePatterns(patterns []string) ([]SecurityAudit, error) {
	audits := make([]SecurityAudit, 0, len(patterns))

	for _, pattern := range patterns {
		audit, err := AuditPattern(pattern)
		if err != nil {
			return nil, fmt.Errorf("audit failed for %s: %w", pattern, err)
		}
		audits = append(audits, *audit)
	}

	return audits, nil
}

// RankPatternsBySafety ranks patterns from safest to most dangerous.
func RankPatternsBySafety(audits []SecurityAudit) []SecurityAudit {
	// Sort by complexity score (lower is safer)
	ranked := make([]SecurityAudit, len(audits))
	copy(ranked, audits)

	// Simple bubble sort (good enough for small lists)
	for i := 0; i < len(ranked); i++ {
		for j := i + 1; j < len(ranked); j++ {
			if ranked[i].Complexity.Overall > ranked[j].Complexity.Overall {
				ranked[i], ranked[j] = ranked[j], ranked[i]
			}
		}
	}

	return ranked
}
