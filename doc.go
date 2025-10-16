/*
Package regret provides validation and analysis for regex patterns to prevent
Regular Expression Denial of Service (ReDoS) attacks.

# Overview

When your application accepts regex patterns from untrusted sources (user input,
configuration files, APIs), you need to ensure these patterns won't cause
catastrophic backtracking. regret analyzes patterns using formal automata theory
to detect both exponential (EDA) and polynomial (IDA) backtracking vulnerabilities.

# Quick Start

	import "github.com/theakshaypant/regret"

	// Quick safety check
	if !regret.IsSafe("(a+)+") {
	    return errors.New("unsafe regex pattern")
	}

	// Detailed validation
	issues, err := regret.Validate(pattern)
	if err != nil {
	    return err
	}
	for _, issue := range issues {
	    fmt.Printf("%s: %s\n", issue.Type, issue.Message)
	}

# Validation Modes

regret offers three validation modes with different performance characteristics:

  - Fast: Quick heuristics only (~microseconds)
    Best for hot paths and real-time validation

  - Balanced: Heuristics + NFA analysis (~milliseconds)
    Recommended for most use cases

  - Thorough: Full analysis + adversarial testing (~tens of milliseconds)
    Best for configuration validation and security auditing

# Issue Detection

The library detects multiple types of dangerous patterns:

  - Exponential Backtracking (EDA): Patterns with exponentially many matching paths
    Example: (a+)+, (a|a)*
    Complexity: O(2^n)

  - Polynomial Backtracking (IDA): Patterns with polynomial ambiguity
    Example: a*a* (quadratic), a*a*a* (cubic)
    Complexity: O(n^k)

  - Nested Quantifiers: Quantifiers inside quantified groups
    Example: (x*)*,  (\w+)+

  - Overlapping Alternation: Alternations with overlapping branches
    Example: (a|ab)+, (foo|foobar)*

  - Context-Dependent Issues: Patterns that are safe/unsafe based on context
    Example: ((a|a)*|.*) is unsafe, but (a|a)*.* is safe

# Configuration

Customize validation behavior with Options:

	opts := &regret.Options{
	    Mode:               regret.Balanced,
	    Timeout:            100 * time.Millisecond,
	    MaxComplexityScore: 70,
	    Checks:             regret.CheckDefault,
	    StrictMode:         true,
	}
	issues, err := regret.ValidateWithOptions(pattern, opts)

# Complexity Analysis

Get detailed complexity metrics:

	score, err := regret.AnalyzeComplexity(pattern)
	if err != nil {
	    return err
	}

	fmt.Printf("Complexity: %d/100\n", score.Overall)
	fmt.Printf("Time Complexity: %s\n", score.TimeComplexity)

	if score.HasEDA {
	    fmt.Println("Exponential backtracking detected!")
	}
	if score.HasIDA {
	    fmt.Printf("Polynomial degree: O(n^%d)\n", score.PolynomialDegree)
	}

# Adversarial Input Generation

Adversarial inputs are automatically generated during complexity analysis:

	score, err := regret.AnalyzeComplexity("(a+)+")
	if err != nil {
	    return err
	}

	// Use auto-generated worst-case input
	if score.WorstCaseInput != "" {
	    fmt.Printf("Worst-case input: %s\n", score.WorstCaseInput)
	}

	// Or create custom inputs using pump components
	if len(score.PumpPattern) > 0 {
	    pump := &regret.PumpPattern{
	        Pumps:  score.PumpPattern,
	        Suffix: "x",
	    }
	    for n := 10; n <= 100; n += 10 {
	        input := pump.Generate(n)
	        start := time.Now()
	        re.MatchString(input)
	        fmt.Printf("n=%d, time=%v\n", n, time.Since(start))
	    }
	}

# Use Cases

User-Facing Applications:

	func handleUserRegex(pattern string) error {
	    if !regret.IsSafe(pattern) {
	        return errors.New("unsafe pattern")
	    }
	    re := regexp.MustCompile(pattern)
	    // Use safely...
	    return nil
	}

Configuration Validation:

	for _, pattern := range config.Patterns {
	    issues, _ := regret.Validate(pattern)
	    for _, issue := range issues {
	        if issue.Severity >= regret.High {
	            return fmt.Errorf("unsafe pattern: %s", issue.Message)
	        }
	    }
	}

API Endpoints:

	opts := &regret.Options{
	    Mode:    regret.Fast,
	    Timeout: 50 * time.Millisecond,
	}
	issues, _ := regret.ValidateWithOptions(userPattern, opts)
	if len(issues) > 0 {
	    http.Error(w, "Invalid pattern", http.StatusBadRequest)
	    return
	}

# Theory Background

regret uses formal automata theory to analyze patterns:

1. Parse regex into Abstract Syntax Tree (AST)
2. Construct Non-deterministic Finite Automaton (NFA)
3. Analyze NFA for ambiguity:
  - EDA (Exponential Degree of Ambiguity)
  - IDA (Infinite Degree of Ambiguity)

4. Calculate polynomial degree for IDA patterns
5. Perform context-aware analysis
6. Generate adversarial inputs (pumping)

This approach is based on academic research:
  - "Analyzing Catastrophic Backtracking Behavior in Practical Regular Expression Matching"
  - "Analyzing Matching Time Behavior of Backtracking Regular Expression Matchers by Using Ambiguity of NFA"

# Performance

Typical performance characteristics:

	Operation           Fast Mode    Balanced Mode    Thorough Mode
	----------------------------------------------------------------
	Validation          1-10μs       1-5ms            10-50ms
	Complexity Analysis N/A          5-20ms           20-100ms
	Pump Generation     N/A          <1ms             1-10ms

# Thread Safety

All public functions are safe for concurrent use. Options and results are
immutable after creation.

# Error Handling

The library returns meaningful errors:

  - ErrInvalidPattern: Syntactically invalid regex
  - ErrPatternTooLong: Pattern exceeds MaxPatternLength
  - ErrTimeout: Analysis exceeded configured timeout
  - ErrUnsupportedFeature: Pattern uses unsupported features

# Version Information

	fmt.Println(regret.FullVersion())

# More Information

See README.md for comprehensive documentation and examples.
GitHub: https://github.com/theakshaypant/regret
*/
package regret
