# Test Data

This directory contains curated collections of regex patterns for testing and validation.

## Files

### `evil_patterns.json`
**25 known dangerous patterns** that cause catastrophic backtracking (ReDoS).

Categories:
- Nested quantifiers (e.g., `(a+)+`)
- Overlapping alternation (e.g., `(a|ab)+`)
- Nested wildcards (e.g., `(.*)* `)
- Overlapping quantifiers (e.g., `a*a*`)
- Multiple nested groups

Each pattern includes:
- Attack input examples
- Complexity analysis
- Severity rating
- Source/reference

**Use for:** Testing detection accuracy, validating fixes, security research.

---

### `safe_patterns.json`
**20 verified safe patterns** that don't cause catastrophic backtracking.

Categories:
- Email validation
- URL validation
- Phone numbers
- UUIDs
- Dates and timestamps
- Fixed formats
- Character classes

Each pattern includes:
- Complexity analysis
- Common use cases
- Category classification

**Use for:** Avoiding false positives, benchmarking performance, pattern templates.

---

### `real_world_patterns.json`
**20 patterns from popular projects** and standards.

Sources:
- RFC specifications (email, DNS, etc.)
- Cloud platforms (AWS, GCP, Kubernetes)
- Programming standards (semantic versioning, identifiers)
- Security tools (password validation)

Each pattern includes:
- Source attribution
- Status (safe/unsafe)
- Practical use case
- Category

**Use for:** Real-world validation, integration testing, documentation examples.

---

### `edge_cases.json`
**25 tricky patterns** that test detection accuracy.

Categories:
- Boundary cases (anchors, word boundaries)
- Non-capturing groups
- Bounded quantifiers
- Lookaheads
- Optional quantifiers
- Adjacent vs nested
- Greedy vs non-greedy

Each pattern includes:
- Expected status (safe/unsafe/warning)
- Reasoning
- Test category

**Use for:** Testing detection edge cases, improving accuracy, finding false positives/negatives.

---

### `performance_patterns.json`
**15 patterns for benchmarking** with different complexity classes.

Complexity classes:
- Constant time: O(1)
- Linear time: O(n)
- Quadratic time: O(n²)
- Cubic time: O(n³)
- Exponential time: O(2^n)

Each pattern includes:
- Input size recommendations
- Expected growth rate
- Attack suffix
- Benchmark configuration

**Use for:** Performance testing, complexity validation, benchmarking.

---

### `pump_patterns.json`
**18 patterns for testing pump pattern integration** with `AnalyzeComplexity()`.

Categories:
- Generates pump (score >= 50, unsafe patterns with adversarial inputs)
- No pump - safe (safe patterns, no pump needed)
- No pump - below threshold (score < 50, no pump despite issues)
- Edge cases (boundary conditions for pump generation)

Each pattern includes:
- Expected score range (min/max)
- Expected pump generation (yes/no)
- Expected worst-case input (yes/no)
- Pump details (expected components, minimum input length)
- Notes explaining the behavior

**Use for:** Testing automatic pump generation, validating adversarial input creation, threshold behavior.

---

## Usage

### Loading Patterns

```go
import (
    "encoding/json"
    "os"
)

type Pattern struct {
    Pattern     string `json:"pattern"`
    Category    string `json:"category"`
    Description string `json:"description"`
    // ... other fields
}

type PatternSet struct {
    Description string    `json:"description"`
    Patterns    []Pattern `json:"patterns"`
}

func LoadPatterns(filename string) (*PatternSet, error) {
    data, err := os.ReadFile(filename)
    if err != nil {
        return nil, err
    }
    
    var set PatternSet
    err = json.Unmarshal(data, &set)
    return &set, err
}
```

### Testing All Evil Patterns

```go
func TestAllEvilPatterns(t *testing.T) {
    set, err := LoadPatterns("testdata/evil_patterns.json")
    if err != nil {
        t.Fatal(err)
    }
    
    for _, p := range set.Patterns {
        t.Run(p.Category, func(t *testing.T) {
            safe, _ := regret.IsSafe(p.Pattern)
            if safe {
                t.Errorf("Pattern %s should be unsafe but was marked safe", p.Pattern)
            }
        })
    }
}
```

### Benchmarking

```go
func BenchmarkPatterns(b *testing.B) {
    set, _ := LoadPatterns("testdata/performance_patterns.json")
    
    for _, p := range set.Patterns {
        b.Run(p.Description, func(b *testing.B) {
            for i := 0; i < b.N; i++ {
                regret.IsSafe(p.Pattern)
            }
        })
    }
}
```

---

## Statistics

| File | Patterns | Categories | Status |
|------|----------|------------|--------|
| `evil_patterns.json` | 25 | 7 | All unsafe |
| `safe_patterns.json` | 20 | 10 | All safe |
| `real_world_patterns.json` | 20 | 11 | Mixed |
| `edge_cases.json` | 25 | 15 | Mixed |
| `performance_patterns.json` | 15 | 5 | Test data |
| `pump_patterns.json` | 18 | 4 | Pump integration test |
| **Total** | **123** | **52** | **Comprehensive** |

---

## Contributing Patterns

When adding new patterns:

1. Choose the appropriate file
2. Follow the existing JSON structure
3. Include all required fields
4. Add meaningful descriptions
5. Verify the pattern behavior
6. Document the source/reference

### Required Fields

**Evil patterns:**
- `pattern`, `category`, `severity`, `complexity`, `description`, `attack_input`

**Safe patterns:**
- `pattern`, `category`, `complexity`, `description`, `use_case`

**Real-world patterns:**
- `pattern`, `source`, `category`, `status`, `description`

**Edge cases:**
- `pattern`, `description`, `expected_status`, `reason`, `category`

**Performance patterns:**
- `pattern`, `input_sizes`, `expected_growth`, `attack_suffix`, `description`

---

## References

- [OWASP ReDoS](https://owasp.org/www-community/attacks/Regular_expression_Denial_of_Service_-_ReDoS)
- [RegexStaticAnalysis](https://github.com/NicolaasWeideman/RegexStaticAnalysis)
- [How It Works](../docs/HOW_IT_WORKS.md) - Detailed explanation of detection algorithms

---

## License

These patterns are collected from public sources and research papers. Attribution is provided where applicable.

