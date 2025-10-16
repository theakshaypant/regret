# Getting Started with regret

This guide will help you get started with **regret** for both library and CLI usage.

## Installation

### Library

Add regret to your Go project:

```bash
go get github.com/theakshaypant/regret
```

### CLI Tool

Install the command-line tool:

```bash
go install github.com/theakshaypant/regret/cmd/regret@latest
```

Verify installation:

```bash
regret version
```

## Basic Usage (Library)

### 1. Quick Validation

The simplest way to check if a pattern is safe:

```go
package main

import (
    "fmt"
    "github.com/theakshaypant/regret"
)

func main() {
    safe := regret.IsSafe("(a+)+")
    
    if !safe {
        fmt.Println("⚠️  Dangerous pattern detected!")
    }
}
```

### 2. Detailed Validation

Get detailed information about detected issues:

```go
issues, err := regret.ValidateWithOptions("(a+)+", &regret.Options{
    Mode: regret.Balanced,
})

for _, issue := range issues {
    fmt.Printf("Issue: %s\n", issue.Message)
    fmt.Printf("Severity: %s\n", issue.Severity)
    fmt.Printf("Suggestion: %s\n", issue.Suggestion)
}
```

### 3. Complexity Analysis

Analyze pattern complexity with automatic adversarial input generation:

```go
score, err := regret.AnalyzeComplexity("(a+)+")
if err != nil {
    panic(err)
}

fmt.Printf("Complexity Score: %d/100\n", score.Overall)
fmt.Printf("Time Complexity: %s\n", score.TimeComplexity)
fmt.Printf("Explanation: %s\n", score.Explanation)

// Check specific attributes
if score.HasEDA {
    fmt.Println("⚠️  Pattern has Exponential Degree of Ambiguity")
}

// For unsafe patterns (score >= 50), pump patterns are automatically generated
if len(score.PumpPattern) > 0 {
    fmt.Printf("Pump Components: %v\n", score.PumpPattern)
    fmt.Printf("Worst Case Input: %q\n", score.WorstCaseInput)
    
    // Use this input for security testing
    // WARNING: This can cause exponential backtracking in vulnerable engines!
}
```

### 4. Adversarial Testing

**Option A: Automatic (via AnalyzeComplexity)**

For unsafe patterns, `AnalyzeComplexity()` automatically provides adversarial inputs:

```go
score, err := regret.AnalyzeComplexity("(a+)+")
if err != nil {
    panic(err)
}

if score.Overall >= 50 && score.WorstCaseInput != "" {
    fmt.Printf("Testing with worst-case input: %q\n", score.WorstCaseInput)
    // Use score.WorstCaseInput for benchmarking
}
```

## Basic Usage (CLI)

### 1. Check Pattern

Quick validation:

```bash
regret check "(a+)+"
```

Output:
```
❌ Pattern is UNSAFE
  ⚠️  nested_quantifiers: Nested quantifiers detected at position 0-5
      Pattern: (a+)+
      Example: "aaaaax" could cause exponential backtracking
```

### 2. Analyze Pattern

Detailed complexity analysis:

```bash
regret analyze "(a+)+" --mode=thorough
```

### 3. Test Pattern

Generate and test with adversarial input:

```bash
regret test "(a+)+" --size=20 --benchmark
```

### 4. Scan Codebase

Validate patterns in your code using custom logic:

```bash
# Use the provided examples for codebase scanning
go run examples/cicd_integration.go
```

## Configuration

### Validation Modes

Choose the right mode for your use case:

```go
// Fast mode - basic heuristics only (< 1µs)
opts := &regret.Options{Mode: regret.Fast}

// Balanced mode - heuristics + NFA analysis (< 100µs)
opts := &regret.Options{Mode: regret.Balanced}

// Thorough mode - full analysis including adversarial testing
opts := &regret.Options{Mode: regret.Thorough}
```

### Custom Thresholds

Adjust sensitivity:

```go
opts := &regret.Options{
    Mode:                regret.Balanced,
    MaxComplexityScore:  60,      // Fail if score > 60
    MaxNestingDepth:     3,       // Fail if nesting > 3
    MaxQuantifiers:      15,      // Fail if quantifiers > 15
    StrictMode:          true,    // Zero tolerance for issues
}
```

### Timeout Configuration

Set analysis timeout:

```go
opts := &regret.Options{
    Mode:    regret.Thorough,
    Timeout: 100 * time.Millisecond,  // Max analysis time
}
```

## Common Patterns

### Validating User Input

```go
func validateUserRegex(pattern string) error {
    safe := regret.IsSafe(pattern)
    
    if !safe {
        return fmt.Errorf("pattern is unsafe for use")
    }
    
    return nil
}
```

### CI/CD Integration

```bash
#!/bin/bash
# pre-commit hook

# Validate patterns using custom logic
if ! go run examples/cicd_integration.go; then
    echo "❌ Dangerous regex patterns detected!"
    exit 1
fi
```

### Performance Testing

```go
func benchmarkPattern(pattern string, size int) time.Duration {
    // Analyze pattern to get auto-generated worst-case input
    score, err := regret.AnalyzeComplexity(pattern)
    if err != nil || score.WorstCaseInput == "" {
        return 0
    }
    
    // Use the auto-generated worst-case input or create custom from pump components
    input := score.WorstCaseInput
    if len(score.PumpPattern) > 0 && size > 20 {
        // Create custom input with specified size
        pump := &regret.PumpPattern{
            Pumps:  score.PumpPattern,
            Suffix: "x",
        }
        input = pump.Generate(size)
    }
    
    start := time.Now()
    re := regexp.MustCompile(pattern)
    re.MatchString(input)
    return time.Since(start)
}
```

## Next Steps

- **[API Reference](API.md)** - Complete API documentation
- **[CLI Reference](CLI.md)** - All CLI commands and flags
- **[How It Works](HOW_IT_WORKS.md)** - Understanding the detection algorithms
- **[Examples](../examples/)** - More code examples

## Next Steps

- Read the [API Reference](API.md) for complete function documentation
- Learn [How It Works](HOW_IT_WORKS.md) to understand the detection algorithms
- Check out the [examples](../examples/) for more code samples

