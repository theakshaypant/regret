# regret - Examples

Comprehensive examples demonstrating how to use the `regret` library in real-world scenarios.

---

## 📝 Files

### `example_test.go`
**Basic API examples** using Go's example test format.

**9 runnable examples:**
1. `ExampleIsSafe` - Quick safety checking
2. `ExampleValidate` - Basic validation
3. `ExampleValidateWithOptions` - Configurable validation
4. `ExampleAnalyzeComplexity` - Detailed complexity analysis
5. `ExamplePumpPattern_Generate` - Using pump patterns
6. `ExamplePumpPattern_GenerateSequence` - Generating test sequences
7. `ExampleOptions` - Creating custom options
8. `ExampleComplexity_String` - Complexity notation
9. `ExampleValidationMode` - Different validation modes

---

### `user_input_validation.go`
**Real-world user input validation** patterns.

**Key functions:**
- `ValidateUserRegex` - Simple validation wrapper
- `ValidateUserRegexDetailed` - Detailed feedback with warnings
- `RegexValidationMiddleware` - HTTP middleware for web applications
- `SearchFilterValidator` - Search filter validation with complexity checking

**Use cases:**
- Web applications accepting user-provided regex
- Search interfaces with pattern filters
- API parameter validation
- Multi-tenant applications

---

### `cicd_integration.go`
**CI/CD pipeline integration** for automated security scanning.

**Key functions:**
- `ScanCodebase` - Recursively scan directories for regex patterns
- `PreCommitHook` - Git pre-commit hook implementation
- `GenerateJSONReport` - Generate CI/CD artifacts
- `PrintReport` - Human-readable console reports

**Use cases:**
- Security audits in CI/CD pipelines
- Pre-commit Git hooks
- Automated code review
- Build failures on unsafe patterns

**Example:**
```go
report, _ := ScanCodebase("./src", []string{".go", ".js", ".py"})
PrintReport(report)
os.Exit(report.ExitCode) // Fail build if unsafe patterns found
```

---

### `security_audit.go`
**Comprehensive security auditing** and pattern analysis.

**Key functions:**
- `AuditPattern` - Perform thorough security audit
- `PrintAuditReport` - Generate detailed audit reports
- `ComparePatterns` - Compare multiple patterns
- `RankPatternsBySafety` - Rank patterns from safest to most dangerous

**Use cases:**
- Security research and vulnerability assessment
- Pattern comparison and evaluation
- Performance analysis with adversarial inputs
- Generating audit reports

**Example:**
```go
audit, _ := AuditPattern("(a+)+")
PrintAuditReport(audit)
// Shows: issues, complexity, adversarial testing results, recommendations
```

---

### `pump_integration.go`
**Pump pattern integration and adversarial testing** examples.

**Key functions:**
- `ExamplePumpIntegration_Basic` - Automatic pump pattern generation
- `ExamplePumpIntegration_SecurityTesting` - Security testing workflow
- `ExamplePumpIntegration_Benchmarking` - Performance testing with worst-case inputs
- `ExamplePumpIntegration_MultiPatternAnalysis` - Analyzing multiple patterns
- `ExamplePumpIntegration_ConditionalTesting` - Conditional testing based on risk
- `ExamplePumpIntegration_ComparePatterns` - Before/after comparison of patterns
- `ExamplePumpIntegration_BatchAnalysis` - Efficient batch processing

**Use cases:**
- Automated security testing with worst-case inputs
- Performance benchmarking of regex patterns
- Pattern risk assessment and reporting
- CI/CD integration for adversarial testing

**Key Features:**
- `AnalyzeComplexity()` automatically generates pump patterns for unsafe patterns (score ≥ 50)
- Provides concrete worst-case inputs in `ComplexityScore.WorstCaseInput`
- Pump components available in `ComplexityScore.PumpPattern`

**Example:**
```go
score, _ := regret.AnalyzeComplexity("(a+)+")
if score.Overall >= 50 && score.WorstCaseInput != "" {
    fmt.Printf("⚠️  Adversarial input: %q\n", score.WorstCaseInput)
    // Use for security testing with proper safeguards
}
```

---

## 🚀 Running Examples

### Run Example Tests
```bash
# All examples
go test -v ./examples

# Specific example
go test -v ./examples -run ExampleValidate

# Show example output
go test ./examples -v 2>&1 | grep -A 10 "Example"
```

### Use Example Functions
```go
package main

import (
    "os"
    "github.com/theakshaypant/regret/examples"
)

func main() {
    // User input validation
    if err := examples.ValidateUserRegex("(a+)+"); err != nil {
        panic(err) // Pattern rejected
    }
    
    // Security audit
    audit, _ := examples.AuditPattern("(a+)+")
    examples.PrintAuditReport(audit)
    
    // CI/CD scanning
    report, _ := examples.ScanCodebase("./src", []string{".go"})
    examples.PrintReport(report)
    os.Exit(report.ExitCode)
}
```

---

## 📚 Example Categories

### Basic Usage
- `ExampleIsSafe` - Quick boolean check
- `ExampleValidate` - Get detailed issues
- `ExampleValidateWithOptions` - Custom configuration

### Advanced Features
- `ExampleAnalyzeComplexity` - Full complexity analysis with auto-generated pump patterns

### Configuration
- `ExampleOptions` - Creating Options
- `ExampleValidationMode` - Different modes (Fast, Balanced, Thorough)

### Real-World Integration
- `user_input_validation.go` - HTTP middleware, validation wrappers
- `cicd_integration.go` - Pipeline integration, pre-commit hooks
- `security_audit.go` - Comprehensive auditing, reporting

---

## 🎯 Use Cases

| Use Case | File | Key Functions |
|----------|------|---------------|
| **Validate user input** | `user_input_validation.go` | `ValidateUserRegex`, `RegexValidationMiddleware` |
| **HTTP API validation** | `user_input_validation.go` | `RegexValidationMiddleware` |
| **Search filters** | `user_input_validation.go` | `SearchFilterValidator` |
| **CI/CD security** | `cicd_integration.go` | `ScanCodebase`, `GenerateJSONReport` |
| **Pre-commit hooks** | `cicd_integration.go` | `PreCommitHook` |
| **Security audits** | `security_audit.go` | `AuditPattern`, `PrintAuditReport` |
| **Pattern comparison** | `security_audit.go` | `ComparePatterns`, `RankPatternsBySafety` |
| **Learn the API** | `example_test.go` | All Example* functions |

---

## 🔧 Integration Patterns

### HTTP Middleware
```go
import "github.com/theakshaypant/regret/examples"

http.Handle("/search", examples.RegexValidationMiddleware(searchHandler))
```

### Pre-commit Hook
```bash
#!/bin/bash
# .git/hooks/pre-commit
FILES=$(git diff --cached --name-only --diff-filter=ACM | grep '\.go$')
if [ -n "$FILES" ]; then
    go run examples/check_patterns.go $FILES
    exit $?
fi
```

### CI/CD Pipeline (GitHub Actions)
```yaml
name: Regex Security Scan
on: [push, pull_request]
jobs:
  scan:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
      - name: Validate patterns
        run: |
          go install github.com/theakshaypant/regret/cmd/regret@latest
          # Check patterns in Go files with custom scanning logic
          go run examples/cicd_integration.go
      - name: Upload report
        uses: actions/upload-artifact@v3
        with:
          name: regex-security-report
          path: report.json
```

### Docker Integration
```dockerfile
FROM golang:1.21
RUN go install github.com/theakshaypant/regret/cmd/regret@latest
COPY . /app
WORKDIR /app
RUN go run examples/cicd_integration.go || exit 1
```

---

## 💡 Adding New Examples

### Example Test Format
```go
func ExampleNewFeature() {
    pattern := "(a+)+"
    safe, _ := regret.IsSafe(pattern)
    fmt.Printf("Safe: %v\n", safe)
    
    // Output:
    // Safe: false
}
```

### Guidelines
1. **Keep examples simple** - Focus on one feature
2. **Show realistic usage** - Real-world scenarios
3. **Include error handling** - Demonstrate proper usage
4. **Add output comments** - For example tests
5. **Update this README** - Document new examples

---

## 📖 See Also

- [Getting Started Guide](../docs/GETTING_STARTED.md) - Installation and basic usage
- [API Reference](../docs/API.md) - Complete API documentation
- [Test Data](../testdata/README.md) - Curated pattern collections
- [How It Works](../docs/HOW_IT_WORKS.md) - Detection algorithms explained

---

## 📊 Statistics

- **Example tests**: 11 functions
- **Integration examples**: 4 files (user_input_validation, cicd_integration, security_audit, pump_integration)
- **Total functions**: 38+
- **Use cases covered**: 15+
- **Lines of example code**: ~1,750+

---

**Note**: Example tests use the `examples_test` package to demonstrate external usage of the library, simulating how users would import and use `regret` in their own projects.
