# Regret CLI Tool Documentation

The `regret` command-line tool provides a comprehensive interface for detecting and analyzing ReDoS vulnerabilities in regex patterns.

## Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/theakshaypant/regret
cd regret

# Build
go build -o regret ./cmd/regret

# Install to $GOPATH/bin
go install ./cmd/regret
```

### From GitHub (Future)

```bash
go install github.com/theakshaypant/regret/cmd/regret@latest
```

## Commands

### `check` - Quick Validation

Quickly validates a regex pattern for ReDoS vulnerabilities. Perfect for CI/CD pipelines.

**Usage:**
```bash
regret check <pattern> [flags]
```

**Examples:**
```bash
# Basic check
regret check "(a+)+"

# Different validation modes
regret check "(a+)+" --mode=fast
regret check "(a+)+" --mode=thorough

# JSON output for scripting
regret check "(a+)+" --output=json

# Use in scripts (exit code 0=safe, 1=unsafe)
regret check "$USER_PATTERN" || echo "Unsafe pattern detected!"
```

**Output:**
```
✗ Pattern is UNSAFE
Complexity: O(2^n), Score: 70/100

Issues found:
  ⛔ nested_quantifiers: Nested quantifiers detected: (a+)+
  ⛔ exponential_backtracking: Nested quantifiers create exponential ambiguity
```

### `analyze` - Detailed Analysis

Performs comprehensive complexity analysis on a regex pattern.

**Usage:**
```bash
regret analyze <pattern> [flags]
```

**Examples:**
```bash
# Detailed analysis
regret analyze "(a+)+"

# Thorough mode
regret analyze "(a+)+" --mode=thorough

# Table format
regret analyze "(a+)+" --output=table

# Verbose output
regret analyze "(a+)+" --verbose
```

**Output:**
```
Pattern: (a+)+
Status: ✗ UNSAFE
Complexity: O(2^n)
Score: 70/100
⚠️  EDA (Exponential Degree of Ambiguity) detected

Metrics:
  Nesting Depth: 2
  Quantifiers: 2
  Alternations: 0

Issues:
  ⛔ nested_quantifiers: Nested quantifiers detected: (a+)+
     Suggestion: Remove nesting: simplify to a single quantifier

Explanation: Exponential time complexity - catastrophic backtracking risk
```

### `test` - Adversarial Testing

Tests a pattern with adversarial inputs to detect actual ReDoS behavior.

**Usage:**
```bash
regret test <pattern> [flags]
```

**Flags:**
- `-s, --size int` - Pump size (number of repetitions) (default: 20)

**Examples:**
```bash
# Test with default size
regret test "(a+)+"

# Custom size
regret test "(a+)+" --size=30

# Verbose mode
regret test "(a+)+" --size=20 --verbose
```

**Output:**
```
Generated input (n=20): aaaaaaaaaaaaaaaaaaax
Input length: 21 characters

Testing...
✓ Completed in: 245ms
⚠️  Slow matching detected! (>100ms)
⚠️  Pattern exhibits ReDoS behavior

Recommendation:
  ⛔ This pattern has exponential time complexity
  → Avoid using this pattern with untrusted input
```

### `version` - Version Information

Display version information.

**Usage:**
```bash
regret version
```

**Output:**
```
regret version 0.1.0
Regex threat detection and analysis tool

Features:
  • Fast heuristics detection
  • NFA-based formal analysis (EDA/IDA)
  • Complexity scoring (0-100)
  • Adversarial input generation
```

## Global Flags

These flags apply to all commands:

- `-m, --mode string` - Validation mode: `fast`, `balanced`, `thorough` (default: "balanced")
- `-o, --output string` - Output format: `text`, `json`, `table` (default: "text")
- `-v, --verbose` - Verbose output
- `-q, --quiet` - Quiet mode (errors only)
- `--no-color` - Disable color output
- `-c, --config string` - Config file path
- `-h, --help` - Help for any command

## Output Formats

### Text (Default)

Human-readable format with colors and emoji indicators.

```bash
regret check "(a+)+" --output=text
```

### JSON

Machine-readable format for scripting and tooling.

```bash
regret check "(a+)+" --output=json
```

```json
{
  "pattern": "(a+)+",
  "safe": false,
  "complexity": "O(2^n)",
  "score": 70,
  "issues": [...]
}
```

### Table

Structured table format.

```bash
regret check "(a+)+" --output=table
```

```
┌──────────────┬────────────┬───────┐
│ Safe         │ Complexity │ Score │
├──────────────┼────────────┼───────┤
│ No           │ O(2^n)     │ 70    │
└──────────────┴────────────┴───────┘
```

## Exit Codes

- `0` - Pattern is safe / No issues found
- `1` - Pattern is unsafe / Issues found / Error occurred

This makes the CLI perfect for CI/CD integration:

```bash
regret check "$USER_INPUT" || exit 1
```

## CI/CD Integration

### GitHub Actions

```yaml
name: Regex Security Check

on: [push, pull_request]

jobs:
  regret:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Setup Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      
      - name: Install regret
        run: go install github.com/theakshaypant/regret/cmd/regret@latest
      
      - name: Validate patterns
        run: |
          # Check patterns in Go files with custom scanning logic
          go run examples/cicd_integration.go
      
      - name: Upload report
        uses: actions/upload-artifact@v3
        with:
          name: regret-report
          path: report.json
```

### Pre-commit Hook

```bash
#!/bin/sh
# .git/hooks/pre-commit

echo "Checking regex patterns..."
# Check patterns using custom logic (see examples/cicd_integration.go)
if ! go run examples/cicd_integration.go; then
    echo "❌ Dangerous regex patterns detected!"
    exit 1
fi

echo "✓ All regex patterns are safe"
```

### GitLab CI

```yaml
regex_check:
  stage: test
  script:
    - go install github.com/theakshaypant/regret/cmd/regret@latest
    # Validate patterns using custom logic
    - go run examples/cicd_integration.go
```

## Common Workflows

### 1. Quick Pattern Validation

```bash
# Validate user input before using in application
regret check "$USER_PATTERN" && use_pattern "$USER_PATTERN"
```

### 2. Code Review

```bash
# Analyze pattern during code review
regret analyze "(a+)+" --mode=thorough --verbose
```

### 3. Security Audit

```bash
# Audit patterns using custom logic (see examples/security_audit.go)
go run examples/security_audit.go "$PATTERN"
```

### 4. Pattern Development

```bash
# Test pattern while developing
regret test "$PATTERN" --size=50 --verbose
```

## Tips and Best Practices

### 1. Use Appropriate Validation Modes

- `fast` - Quick checks, minimal overhead (CI/CD)
- `balanced` - Good balance of speed and accuracy (default)
- `thorough` - Comprehensive analysis (security audits)

### 2. Leverage Exit Codes

Always check exit codes in scripts:

```bash
if regret check "$PATTERN"; then
    echo "Pattern is safe"
else
    echo "Pattern is unsafe"
    exit 1
fi
```

### 3. Use JSON for Automation

Parse JSON output for integration with other tools:

```bash
regret analyze "$PATTERN" --output=json | jq '.Overall'
```

### 4. Combine Commands

```bash
# Check, then analyze if unsafe
regret check "$PATTERN" || regret analyze "$PATTERN" --mode=thorough
```

## Troubleshooting

### Pattern Escaping

When using patterns with special shell characters, use quotes:

```bash
# Good
regret check "(a+)+"
regret check '(a|b)*'

# Bad (shell interprets parentheses)
regret check (a+)+
```

### Large Codebases

For very large codebases, validate patterns using custom scanning logic:

```bash
# Use the provided examples for codebase scanning
go run examples/cicd_integration.go
```

### Timeout Issues

If analysis times out, use fast mode:

```bash
# Use fast mode for quick checks
regret check "$PATTERN" --mode=fast

# Or analyze specific patterns
regret analyze "$PATTERN" --mode=balanced
```

## Examples

See the [examples directory](../examples/) for more usage examples and integration patterns.

## Support

- GitHub Issues: https://github.com/theakshaypant/regret/issues
- Documentation: https://github.com/theakshaypant/regret/docs
- Library API: See [API Documentation](../README.md)

