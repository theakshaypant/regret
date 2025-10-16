# API Reference

Complete reference for the **regret** Go library.

## Core Functions

### IsSafe

Quick check if a regex pattern is safe.

```go
func IsSafe(pattern string) bool
```

**Parameters:**
- `pattern` - Regex pattern string to validate

**Returns:**
- `bool` - `true` if pattern is safe, `false` if dangerous or invalid

**Example:**

```go
safe := regret.IsSafe("(a+)+")

if !safe {
    fmt.Println("Pattern is dangerous!")
}
```

---

### ValidateWithOptions

Detailed validation with custom options.

```go
func ValidateWithOptions(pattern string, opts *Options) ([]Issue, error)
```

**Parameters:**
- `pattern` - Regex pattern string to validate
- `opts` - Configuration options (can be nil for defaults)

**Returns:**
- `[]Issue` - List of detected issues
- `error` - Error if pattern is invalid

**Example:**

```go
opts := &regret.Options{
    Mode: regret.Balanced,
    MaxComplexityScore: 60,
}

issues, err := regret.ValidateWithOptions("(a+)+", opts)
if err != nil {
    log.Fatal(err)
}

for _, issue := range issues {
    fmt.Printf("%s: %s\n", issue.Type, issue.Message)
}
```

---

### Validate

Simplified validation with default options.

```go
func Validate(pattern string) ([]Issue, error)
```

Equivalent to `ValidateWithOptions(pattern, nil)`.

---

### AnalyzeComplexity

Analyze pattern time complexity with automatic adversarial input generation.

```go
func AnalyzeComplexity(pattern string) (*ComplexityScore, error)
```

**Parameters:**
- `pattern` - Regex pattern to analyze

**Returns:**
- `*ComplexityScore` - Detailed complexity analysis
- `error` - Error if pattern is invalid

**Behavior:**
- For unsafe patterns (score ≥ 50), automatically generates pump patterns and worst-case inputs
- Provides concrete adversarial examples for security testing
- Pump generation failures are silently ignored (supplementary information)

**Example:**

```go
score, err := regret.AnalyzeComplexity("(a+)+")
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Score: %d/100\n", score.Overall)
fmt.Printf("Time Complexity: %s\n", score.TimeComplexity)
fmt.Printf("Has EDA: %v\n", score.HasEDA)
fmt.Printf("Has IDA: %v\n", score.HasIDA)
fmt.Printf("Explanation: %s\n", score.Explanation)

// Pump patterns are automatically generated for unsafe patterns
if len(score.PumpPattern) > 0 {
    fmt.Printf("Pump Components: %v\n", score.PumpPattern)
    fmt.Printf("Worst Case Input: %q\n", score.WorstCaseInput)
    
    // Use the worst-case input for testing
    // WARNING: This input can cause exponential backtracking!
}
```

---

## Types

### Options

Configuration for validation and analysis.

```go
type Options struct {
    Mode                ValidationMode
    Timeout             time.Duration
    Checks              CheckFlags
    MaxComplexityScore  int
    MaxPatternLength    int
    MaxNestingDepth     int
    MaxQuantifiers      int
    StrictMode          bool
    AllowUnsafe         bool
}
```

**Fields:**

- `Mode` - Validation mode (Fast, Balanced, Thorough)
- `Timeout` - Maximum analysis time (default: 100ms)
- `Checks` - Which checks to enable (bitmask)
- `MaxComplexityScore` - Maximum acceptable score (default: 100)
- `MaxPatternLength` - Maximum pattern length (default: 10000)
- `MaxNestingDepth` - Maximum quantifier nesting (default: 5)
- `MaxQuantifiers` - Maximum quantifier count (default: 20)
- `StrictMode` - Zero tolerance for issues
- `AllowUnsafe` - Allow analysis of unsafe patterns

**Example:**

```go
opts := &regret.Options{
    Mode:               regret.Balanced,
    Timeout:            50 * time.Millisecond,
    MaxComplexityScore: 70,
    StrictMode:         true,
}
```

---

### ValidationMode

Determines depth of analysis.

```go
type ValidationMode int

const (
    ModeFast      ValidationMode = iota  // Heuristics only
    ModeBalanced                          // Heuristics + NFA
    ModeThorough                          // Full analysis
)
```

**Modes:**

| Mode | Checks | Avg Time | Use Case |
|------|--------|----------|----------|
| `ModeFast` | Heuristics | < 1µs | High-throughput |
| `ModeBalanced` | + NFA analysis | < 100µs | General use |
| `ModeThorough` | + Adversarial testing | < 10ms | Security audits |

---

### Issue

Represents a detected problem.

```go
type Issue struct {
    Type       IssueType
    Severity   Severity
    Position   Position
    Pattern    string
    Message    string
    Example    string
    Suggestion string
    Complexity int
    Details    map[string]interface{}
}
```

**Fields:**

- `Type` - Category of issue
- `Severity` - How dangerous the issue is
- `Position` - Where in the pattern it occurs
- `Pattern` - The problematic sub-pattern
- `Message` - Human-readable description
- `Example` - Example adversarial input that exploits this issue
- `Suggestion` - How to fix the issue
- `Complexity` - Local complexity contribution (0-100)
- `Details` - Additional technical details about the issue

---

### IssueType

Categories of detected issues.

```go
type IssueType int

const (
    NestedQuantifiers IssueType = iota
    OverlappingAlternation
    RepeatedCaptureGroup
    ExponentialBacktracking  // EDA
    PolynomialBacktracking   // IDA
    UnboundedRepetition
    AmbiguousPattern
    ComplexityThresholdExceeded
    ContextuallyDangerous
)
```

---

### Severity

How dangerous an issue is.

```go
type Severity string

const (
    Critical Severity = "critical"  // Always fails in production
    High     Severity = "high"      // Very likely to cause issues
    Medium   Severity = "medium"    // Potential performance impact
    Low      Severity = "low"       // Minor concern
    Info     Severity = "info"      // Informational only
)
```

---

### ComplexityScore

Detailed complexity analysis result.

```go
type ComplexityScore struct {
    Overall          int
    TimeComplexity   Complexity
    SpaceComplexity  Complexity
    HasEDA           bool
    HasIDA           bool
    PolynomialDegree int
    Metrics          Metrics
    WorstCaseInput   string
    PumpPattern      []string
    Explanation      string
    Safe             bool
}
```

**Fields:**

- `Overall` - Overall complexity score (0-100, lower is better)
- `TimeComplexity` - Estimated worst-case time complexity (Complexity enum)
- `SpaceComplexity` - Estimated space complexity (Complexity enum)
- `HasEDA` - Exponential Degree of Ambiguity detected
- `HasIDA` - Infinite Degree of Ambiguity detected (polynomial)
- `PolynomialDegree` - Polynomial degree (2=quadratic, 3=cubic, etc.)
- `Metrics` - Detailed metrics about the pattern
- `WorstCaseInput` - Example input that triggers worst-case behavior (automatically generated for score ≥ 50)
- `PumpPattern` - Pump components for generating adversarial inputs (automatically populated for score ≥ 50)
- `Explanation` - Human-readable explanation of the complexity
- `Safe` - Whether the pattern is considered safe

**Note:** When `AnalyzeComplexity()` detects an unsafe pattern (score ≥ 50), it automatically populates `WorstCaseInput` and `PumpPattern` with adversarial test inputs. For safe patterns, these fields will be empty/nil.

---

### PumpPattern

Adversarial input generator.

```go
type PumpPattern struct {
    Prefix      string
    Pumps       []string
    Suffix      string
    Interleave  bool
    Description string
}
```

**Fields:**

- `Prefix` - Initial string before the pumped section
- `Pumps` - Repeating components (can be multiple)
- `Suffix` - Final string after pumped section (often non-matching char)
- `Interleave` - Whether to interleave pumps or concatenate them
- `Description` - Explanation of what this pump pattern tests

**Methods:**

```go
// Generate single adversarial input of specified size
func (p *PumpPattern) Generate(n int) string

// Generate sequence of inputs with sizes from start to end by step
func (p *PumpPattern) GenerateSequence(start, end, step int) []string
```

**Example:**

```go
pump := &regret.PumpPattern{
    Prefix:      "",
    Pumps:       []string{"a"},
    Suffix:      "x",
    Interleave:  false,
    Description: "Tests exponential backtracking in (a+)+",
}

input := pump.Generate(20)  // "aaaaaaaaaaaaaaaaaaax"

// Generate sequence from size 10 to 30 with step 10
sequence := pump.GenerateSequence(10, 30, 10)
// ["aaaaaaaaax", "aaaaaaaaaaaaaaaaaaax", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaax"]
```

---

### CheckFlags

Bitmask for enabling specific checks.

```go
type CheckFlags uint32

const (
    CheckNestedQuantifiers CheckFlags = 1 << iota
    CheckOverlappingAlternation
    CheckCatastrophicBacktrack
    CheckUnboundedRepetition
    CheckExponentialPaths
    CheckComplexityScore
    CheckMemoryUsage
    CheckNFAAmbiguity
    CheckPolynomialDegree
    CheckContextAwareness
    
    // CheckAll enables all available checks
    CheckAll CheckFlags = ^CheckFlags(0)
    
    // CheckDefault includes the most important checks for typical use cases
    CheckDefault = CheckNestedQuantifiers | 
                   CheckOverlappingAlternation | 
                   CheckCatastrophicBacktrack | 
                   CheckNFAAmbiguity
)
```

**Example:**

```go
opts := &regret.Options{
    Checks: regret.CheckNestedQuantifiers | regret.CheckOverlappingAlternation,
}
```

---

## Performance Characteristics

| Function | Typical Time | Use Case |
|----------|-------------|----------|
| `IsSafe()` (Fast) | < 1µs | Real-time validation |
| `IsSafe()` (Balanced) | < 100µs | General use |
| `AnalyzeComplexity()` | < 1ms | Detailed analysis (includes pump generation) |

---

## Related Documentation

- [Getting Started](GETTING_STARTED.md) - Installation and basic usage
- [How It Works](HOW_IT_WORKS.md) - Understanding the detection algorithms
- [CLI Reference](CLI.md) - Command-line tool documentation
- [Examples](../examples/) - Runnable code examples

