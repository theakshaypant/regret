# regret - Architecture

This document outlines the technical architecture and implementation approach for **regret**, informed by academic research from [RegexStaticAnalysis](https://github.com/NicolaasWeideman/RegexStaticAnalysis).

## Core Concepts

### 1. Degree of Ambiguity in NFAs

The foundation of **regret** is based on analyzing the **Degree of Ambiguity** in Non-deterministic Finite Automata (NFAs):

#### Exponential Degree of Ambiguity (EDA)
- **Definition**: A pattern has EDA if the number of accepting paths in its NFA grows exponentially with input length
- **Example**: `(a|a)*` matching `aⁿ` has 2ⁿ accepting paths
- **Impact**: Matching time is O(2ⁿ) for adversarial inputs
- **Detection**: Look for patterns where multiple subexpressions can match the same input

#### Infinite Degree of Ambiguity (IDA)
- **Definition**: A pattern has IDA if the number of accepting paths grows polynomially
- **Example**: `a*a*` matching `aⁿx` has n ways to split between the two quantifiers (quadratic)
- **Example**: `a*a*a*` matching `aⁿx` has O(n³) ways to split (cubic)
- **Impact**: Matching time is O(nᵏ) where k is the polynomial degree
- **Detection**: Count overlapping quantifiers that match the same character class

### 2. Context-Aware Analysis

Order and context significantly affect pattern safety:

```
(a|a)*.*     -> Safe  (suffix .* prevents backtracking)
((a|a)*|.*)  -> Unsafe (evil pattern tries first, backtracks fully)
```

**Key Insight**: A pattern may be safe if:
- Followed by a catch-all pattern like `.*`
- In an alternation where a safer pattern is tried first
- Anchored with `^` or `$` limiting input space

## Implementation Architecture

### Package Structure

```
regret/
├── validator.go           # Main validation API
├── types.go               # Core types (Issue, Options, etc.)
├── version.go             # Version information
├── doc.go                 # Package documentation
│
├── internal/              # Internal implementation packages
│   ├── analyzer/          # Complexity analysis
│   │   ├── analyzer.go    # Complexity scoring and analysis
│   │   └── analyzer_test.go
│   │
│   ├── detector/          # Pattern detection logic
│   │   ├── detector.go    # Main detector interface & heuristics
│   │   ├── nfa_analysis.go # EDA/IDA detection via NFA analysis
│   │   ├── detector_test.go
│   │   └── nfa_analysis_test.go
│   │
│   ├── parser/            # Regex parsing & NFA construction
│   │   ├── parser.go      # Wrapper around regexp/syntax
│   │   ├── nfa.go         # NFA construction from parsed regex
│   │   ├── parser_test.go
│   │   └── nfa_test.go
│   │
│   ├── pump/              # Adversarial input generation
│   │   ├── generator.go   # Generate pump patterns
│   │   └── generator_test.go
│   │
│   ├── cli/               # CLI implementation
│   │   ├── cmd/           # CLI commands
│   │   │   ├── root.go    # Root command setup
│   │   │   ├── check.go   # Check command
│   │   │   ├── analyze.go # Analyze command
│   │   │   ├── test.go    # Test command
│   │   │   ├── scan.go    # Scan command
│   │   │   ├── benchmark.go # Benchmark command
│   │   │   └── version.go # Version command
│   │   └── output/
│   │       └── formatter.go # Output formatting (text, JSON, etc.)
│   │
│   └── util/              # Shared utilities (future)
│
├── cmd/
│   └── regret/            # CLI tool entry point
│       └── main.go
│
├── testdata/              # Test data files
│   ├── evil_patterns.json
│   ├── safe_patterns.json
│   ├── real_world_patterns.json
│   └── ...
│
└── examples/              # Example code
    ├── user_input_validation.go
    ├── security_audit.go
    └── ...
```

## Key Algorithms

### EDA Detection Algorithm

```go
func detectEDA(nfa *NFA) (hasEDA bool, examples []string) {
    // 1. Find all states reachable through epsilon transitions
    epsilonClosure := computeEpsilonClosure(nfa)
    
    // 2. For each state, count distinct paths to reach it
    for _, state := range nfa.States {
        paths := countDistinctPaths(nfa, state)
        
        // 3. If multiple paths can consume same input, check if exponential
        if paths > 1 {
            if isExponentialAmbiguity(nfa, state) {
                return true, generateEDAExamples(nfa, state)
            }
        }
    }
    
    return false, nil
}

func isExponentialAmbiguity(nfa *NFA, state *State) bool {
    // Check if ambiguity grows exponentially with input length
    // This involves detecting if there's a cycle that increases ambiguity
    
    // Look for patterns like (R|R)* where R matches same input
    // Or (R+)+ where inner and outer quantifiers overlap
    
    return hasNestedQuantifiers(state) || hasOverlappingAlternation(state)
}
```

### IDA Detection Algorithm

```go
func detectIDA(nfa *NFA) (hasIDA bool, degree int) {
    // Find sequences of overlapping quantifiers
    overlappingQuantifiers := findOverlappingQuantifiers(nfa)
    
    if len(overlappingQuantifiers) < 2 {
        return false, 0
    }
    
    // Check if quantifiers can match same character class
    for i := 0; i < len(overlappingQuantifiers)-1; i++ {
        q1 := overlappingQuantifiers[i]
        q2 := overlappingQuantifiers[i+1]
        
        if canMatchSameInput(q1, q2) {
            degree++
        }
    }
    
    if degree >= 2 {
        return true, degree
    }
    
    return false, 0
}
```

### Context-Aware Safety Check

```go
func isProtectedByContext(regex *syntax.Regexp, position int) bool {
    // Check if an evil pattern is protected by surrounding context
    
    // 1. Check for protective suffix
    if hasProtectiveSuffix(regex, position) {
        // e.g., (a|a)*.* - the .* suffix prevents backtracking
        return true
    }
    
    // 2. Check alternation ordering
    if inAlternation(regex, position) {
        // e.g., (.*|(a|a)*) - .* matches first, safe
        if safeAlternationOrder(regex, position) {
            return true
        }
    }
    
    // 3. Check for anchors that limit input
    if hasLimitingAnchors(regex, position) {
        return true
    }
    
    return false
}
```

## Testing Strategy

### Test Patterns

#### Known Evil Patterns
```go
var evilPatterns = []struct{
    pattern string
    expectedIssue IssueType
    complexity Complexity
}{
    {"(a+)+", NestedQuantifiers, Exponential},
    {"(a|a)*", OverlappingAlternation, Exponential},
    {"a*a*", RepeatedCaptureGroup, Quadratic},
    {"a*a*a*", RepeatedCaptureGroup, Cubic},
    {"(a+)+b", NestedQuantifiers, Exponential},
}
```

#### Safe Patterns
```go
var safePatterns = []string{
    "^[a-z]+$",
    "\\d{3}-\\d{3}-\\d{4}",
    "^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$",
    "(a|a)*.*",  // protected by suffix
}
```

#### Context-Dependent Patterns
```go
var contextPatterns = []struct{
    pattern string
    safe bool
    reason string
}{
    {"(a|a)*.*", true, "protective suffix"},
    {"((a|a)*|.*)", false, "evil pattern tries first"},
    {"(.*|(a|a)*)", true, "safe pattern tries first"},
}
```

## Performance Targets

| Operation | Target | Max |
|-----------|--------|-----|
| Fast mode validation | 1-10μs | 50μs |
| Balanced mode | 1-5ms | 20ms |
| Thorough mode | 10-50ms | 500ms |
| NFA construction | <1ms | 5ms |
| EDA detection | 1-10ms | 50ms |
| IDA detection | 1-10ms | 50ms |
| Pump generation | <1ms | 10ms |

## Dependencies

- **Go 1.21+**: Use latest stable Go
- **regexp/syntax**: Built-in regex parser
- **No external dependencies for core library** (keep it lightweight)
- Test dependencies: testify, stretchr for assertions

## References

### Academic Papers
1. Weideman, N., et al. "Analyzing Catastrophic Backtracking Behavior in Practical Regular Expression Matching"
2. Weideman, N., et al. "Analyzing Matching Time Behavior of Backtracking Regular Expression Matchers by Using Ambiguity of NFA"
3. Weideman, N., et al. "Turning Evil Regexes Harmless"
4. Berglund, M., et al. "Static Analysis of Regular Expressions"

### Implementations
- [RegexStaticAnalysis (Java)](https://github.com/NicolaasWeideman/RegexStaticAnalysis)
- [safe-regex (JavaScript)](https://github.com/substack/safe-regex)
- [rxxr2 (C++)](https://github.com/superhuman/rxxr2)

---
