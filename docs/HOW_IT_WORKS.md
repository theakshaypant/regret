# How regret Works

Understanding the detection algorithms and theory behind **regret**.

## Overview

**regret** uses a multi-layered approach to detect dangerous regex patterns:

1. **Fast Heuristics** - Pattern-matching rules (< 1µs)
2. **NFA Analysis** - Formal automata theory (< 100µs)
3. **Adversarial Testing** - Generate and test inputs (< 10ms)

## The Problem: Catastrophic Backtracking

### What is ReDoS?

Regular Expression Denial of Service (ReDoS) occurs when a regex engine's backtracking algorithm enters exponential time complexity.

**Example:**

```go
pattern := "(a+)+"
input := "aaaaaaaaaaaaaaaaaaaaaaaax"
```

The regex engine tries:
- `a+` matches "a"
- `(a+)+` matches "a" (entire string)
- Suffix doesn't match (no "b")
- **Backtrack**: Try `a+` with 2 a's, then 1 a, then 0...
- Repeat for every possible split: `a|aaa...`, `aa|aa...`, `aaa|a...`
- **Result**: O(2^n) attempts for input length n

### Why It Happens

Three main causes:

1. **Nested Quantifiers**: `(a+)+`, `(a*)*`
2. **Overlapping Quantifiers**: `a*a+`, `\d*\d+`
3. **Overlapping Alternation**: `(a|ab)+`, `(.*|.+)`

## Layer 1: Fast Heuristics

### Pattern Length Check

**Purpose:** Catch excessively long patterns

**Algorithm:**
```go
if len(pattern) > MAX_PATTERN_LENGTH {
    return Issue{Type: PatternTooLong}
}
```

**Threshold:** 10,000 characters

---

### Nesting Depth Check

**Purpose:** Detect deeply nested quantifiers

**Algorithm:**
```
function getNestingDepth(node):
    if node is not quantifier:
        return max depth of children
    else:
        return 1 + max depth of children
```

**Example:**
- `a+` → depth 1
- `(a+)+` → depth 2
- `((a+)+)+` → depth 3 ✗

**Threshold:** 5 levels

---

### Quantifier Count Check

**Purpose:** Detect patterns with too many quantifiers

**Algorithm:**
```
function countQuantifiers(node):
    count = 0
    if node is quantifier:
        count = 1
    for child in node.children:
        count += countQuantifiers(child)
    return count
```

**Threshold:** 20 quantifiers

---

### Nested Quantifier Detection

**Purpose:** Find quantifiers directly nested within other quantifiers

**Algorithm:**
```
function hasNestedQuantifier(node):
    if node is quantifier:
        for child in node.children:
            if child is quantifier:
                return true  // Direct nesting!
    for child in node.children:
        if hasNestedQuantifier(child):
            return true
    return false
```

**Examples:**
- `(a+)+` ✗ Nested
- `(a+)(b+)` ✓ Adjacent, not nested
- `(a|b)+` ✓ Quantifier over alternation is OK

---

### Overlapping Alternation Detection

**Purpose:** Find alternation branches that can match the same input

**Algorithm:**
```
function branchesOverlap(branch1, branch2):
    prefix1 = getLiteralPrefix(branch1)
    prefix2 = getLiteralPrefix(branch2)
    
    if prefix1 and prefix2:
        return prefix1.startsWith(prefix2) or prefix2.startsWith(prefix1)
    
    charset1 = getCharClass(branch1)
    charset2 = getCharClass(branch2)
    
    return charset1.intersects(charset2)
```

**Examples:**
- `a|ab` ✗ Overlaps (prefix "a")
- `a|b` ✓ Disjoint
- `\d|\w` ✗ Overlaps (digits are word chars)

---

### Dangerous Pattern Detection

**Purpose:** Catch known problematic patterns

**Patterns:**
- `.*.*` - Nested wildcards
- `.*+.*` - Overlapping wildcards
- `.+.+` - Overlapping any-char
- `\d*\d+` - Overlapping digit quantifiers
- `\w*\w+` - Overlapping word quantifiers

---

## Layer 2: NFA Analysis

### What is an NFA?

A **Non-deterministic Finite Automaton** is a formal model of computation that can be in multiple states simultaneously.

**Example NFA for `(a|b)*c`:**

```
    ┌─────────┐
    │  Start  │
    └────┬────┘
         │ ε
    ┌────▼────┐
    │  State  │◄─────┐
    │    1    │      │
    └─┬────┬──┘      │
      │    │         │
    a │    │ b       │ ε
      │    │         │
    ┌─▼────▼──┐      │
    │  State  ├──────┘
    │    2    │
    └────┬────┘
         │ c
    ┌────▼────┐
    │ Accept  │
    └─────────┘
```

### Building the NFA

**Algorithm:**

```go
function buildNFA(regex):
    nfa = new NFA()
    start = nfa.addState()
    accept = nfa.addState()
    
    buildFromRegexp(nfa, regex, start, accept)
    
    return nfa
```

**Construction rules:**

| Regex Op | NFA Construction |
|----------|------------------|
| Literal `a` | `start --a--> accept` |
| Concat `ab` | `start --a--> mid --b--> accept` |
| Alternate `a\|b` | Branch on ε-transitions |
| Star `a*` | ε-loop back to start |
| Plus `a+` | One required, then loop |

### EDA Detection

**EDA** (Exponential Degree of Ambiguity) = O(2^n) complexity

**Algorithm:**

```
function detectEDA(nfa, pattern):
    // 1. Find nested quantifiers in AST
    nested = findNestedQuantifiers(pattern)
    if nested:
        return Issue{Type: ExponentialAmbiguity}
    
    // 2. Find overlapping quantifiers
    overlapping = findOverlappingQuantifiers(pattern)
    if overlapping:
        return Issue{Type: ExponentialAmbiguity}
    
    return nil
```

**Why nested quantifiers cause EDA:**

For pattern `(a+)+` and input "aaa...a":
- Outer `+` can match: 1, 2, 3, ..., n times
- For each, inner `a+` can split: n, n-1, n-2, ..., 1 ways
- Total: 2^n combinations

---

### IDA Detection

**IDA** (Infinite Degree of Ambiguity) = polynomial complexity (O(n^k))

**Algorithm:**

```
function detectIDA(nfa, pattern):
    // Find adjacent quantifiers over same input
    quantifiers = findAllQuantifiers(pattern)
    
    for each pair (q1, q2) in quantifiers:
        if q1 and q2 are adjacent:
            if canMatchSameInput(q1, q2):
                degree = estimatePolynomialDegree(q1, q2)
                return Issue{Type: PolynomialAmbiguity, Degree: degree}
    
    return nil
```

**Example:**

Pattern: `a*a+`
- `a*` can match 0, 1, 2, ..., n a's
- `a+` must match remaining a's
- For input "aaa...a" (n chars), there are n ways to split
- Complexity: O(n)

---

### Ambiguity Degree Calculation

```
function computeAmbiguityDegree(pattern):
    nestingDepth = getNestingDepth(pattern)
    
    if nestingDepth >= 2:
        return EXPONENTIAL  // O(2^n)
    
    quantifierCount = countQuantifiers(pattern)
    
    if quantifierCount >= 3:
        return POLYNOMIAL  // O(n^k)
    
    return LINEAR  // O(n)
```

---

## Layer 3: Adversarial Testing

### Pump Pattern Generation

**Purpose:** Generate inputs that expose worst-case behavior

**Algorithm:**

```
function generatePumpPattern(pattern):
    pump = {prefix: "", pumps: [], suffix: ""}
    
    if hasNestedQuantifiers(pattern):
        pump = generateNestedPump(pattern)
    else if hasOverlappingQuantifiers(pattern):
        pump = generateOverlappingPump(pattern)
    else if hasAlternation(pattern):
        pump = generateAlternationPump(pattern)
    else:
        pump = generateGenericPump(pattern)
    
    return pump
```

### Nested Quantifier Pumps

**Pattern:** `(a+)+`

**Pump:**
- Prefix: `""`
- Pump: `"a"`
- Suffix: `"x"` (non-matching)

**Generated inputs:**
- n=5: `"aaaaax"`
- n=10: `"aaaaaaaaaax"`
- n=20: `"aaaaaaaaaaaaaaaaaaax"`

**Why it works:** The non-matching suffix forces the engine to try all possible ways to split the prefix.

---

### Overlapping Quantifier Pumps

**Pattern:** `a*a+`

**Pump:**
- Prefix: `""`
- Pump: `"a"`
- Suffix: `"x"`

Similar to nested quantifiers, but with linear complexity.

---

### Alternation Pumps

**Pattern:** `(a|ab)+`

**Pump:**
- Prefix: `""`
- Pump: `"ab"`
- Suffix: `"x"`

**Generated:** `"abababab...x"`

Forces engine to try matching `"a"` vs `"ab"` at each position.

---

## Complexity Scoring

### Score Calculation

```go
score = 0

// Base score from pattern characteristics
score += nestingDepth * 15      // 15 points per level
score += quantifierCount * 3    // 3 points per quantifier
score += alternationCount * 5   // 5 points per alternation

// Penalties for specific issues
if hasNestedQuantifiers:
    score += 25
if hasOverlappingQuantifiers:
    score += 15
if hasOverlappingAlternation:
    score += 10

// Cap at maximum
score = min(score, MAX_SCORE)

return score
```

### Time Complexity Classification

| Score | Complexity | Time Class | Safety |
|-------|------------|------------|--------|
| 0-20 | O(1) - O(n) | Linear | ✅ Safe |
| 21-40 | O(n²) | Quadratic | ⚠️ Caution |
| 41-60 | O(n³) | Polynomial | ⚠️ Risky |
| 61-100 | O(2^n) | Exponential | ❌ Dangerous |

---

## Context-Aware Analysis

### Why Context Matters

The **order** and **structure** of regex components affects safety:

**Example 1:**
```
Pattern: (a|a)*.*
Analysis: EDA from nested ambiguity (outer *, inner alternation and .*)
```

**Example 2:**
```
Pattern: ((a|a)*|.*)
Analysis: IDA from alternation (branches don't multiply)
```

### Parenthesis Impact

Parentheses control **scope of ambiguity**:

- `(a+)+` - Nested scope → Exponential
- `(a+)+(b+)` - Separate scopes → Linear

---

## Academic Foundation

**regret** is based on formal research:

1. **"Static Analysis for Regular Expression Denial-of-Service Attacks"**
   - Weideman, van der Merwe, Berglund, Watson (2017)
   - Defines EDA and IDA

2. **"Taxonomy of ReDoS Vulnerabilities"**
   - Davis, Michael, Coghlan, Servant (2018)
   - Classifies attack patterns

3. **NFA Ambiguity Detection**
   - Theoretical foundations from automata theory

---

## Validation Modes Explained

### Fast Mode (< 1µs)

**Checks:**
- Pattern length
- Nesting depth
- Quantifier count
- Nested quantifiers (AST scan)
- Dangerous patterns

**Skips:**
- NFA construction
- Formal ambiguity analysis
- Adversarial testing

**Use when:** High throughput required (e.g., API gateway)

---

### Balanced Mode (< 100µs)

**Checks:**
- All fast checks
- NFA construction
- EDA detection
- IDA detection
- Formal ambiguity analysis

**Skips:**
- Adversarial input generation
- Actual performance testing

**Use when:** General development (default)

---

### Thorough Mode (< 10ms)

**Checks:**
- All balanced checks
- Pump pattern generation
- Adversarial input testing
- Actual regex benchmark (optional)

**Use when:** Security audits, CI/CD, production deployment

---

## Performance Optimization

### Why So Fast?

1. **Lazy Evaluation**: Stop at first issue in fast mode
2. **AST Caching**: Parse once, analyze multiple times
3. **Minimal Allocations**: Reuse data structures
4. **Early Exit**: Return as soon as dangerous pattern detected

### Benchmarks

```
BenchmarkIsSafe_Safe/email-8              500000    2.1 µs/op
BenchmarkIsSafe_Safe/url-8                300000    3.8 µs/op
BenchmarkIsSafe_Evil/nested_quant-8       200000    6.2 µs/op
BenchmarkAnalyzeComplexity-8              100000   12.5 µs/op
BenchmarkGeneratePump-8                    20000   78.3 µs/op
```

---

## Limitations

### False Positives

Some safe patterns may be flagged:

```go
// Flagged but actually safe (due to RE2 limitations)
`(/.*)?`         // Optional group
`(a{2,5})+`      // Bounded quantifier
```

**Why:** Conservative analysis prefers safety over precision.

---

### False Negatives

Some dangerous patterns may be missed:

```go
// Context-dependent issues
`a+|a+`          // Identical branches (will be optimized)
`[a-z]{10,100}`  // Very wide bounded repeat
```

**Why:** Complexity vs. accuracy tradeoff.

---

## See Also

- [Getting Started](GETTING_STARTED.md)
- [API Reference](API.md)
- [Architecture](ARCHITECTURE.md)

