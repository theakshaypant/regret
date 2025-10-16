# regret - Regex Threat Detector

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

**Don't regret your regex.**

regret (Regex Threat) is a Go library that detects evil regex patterns before they bite you. It can be used for validating and analyzing regular expression patterns to prevent [Regular Expression Denial of Service (ReDoS)](https://owasp.org/www-community/attacks/Regular_expression_Denial_of_Service_-_ReDoS) attacks.

## Quick Start

### Library

```go
import "github.com/theakshaypant/regret"

// Quick validation
safe := regret.IsSafe("(a+)+")  // false

// Detailed analysis with auto-generated adversarial inputs
score, _ := regret.AnalyzeComplexity("(a+)+")
// Score: 70/100, Complexity: O(2^n), HasEDA: true
// Pump patterns and worst-case inputs generated automatically
// Use score.WorstCaseInput for testing
```

### CLI Tool

```bash
# Quick validation
regret check "(a+)+"

# Detailed analysis
regret analyze "(a+)+" --mode=thorough

# Test with adversarial input
regret test "(a+)+" --size=20
```

## Installation

```bash
# Library
go get github.com/theakshaypant/regret

# CLI tool
go install github.com/theakshaypant/regret/cmd/regret@latest
```

## Features

- **Fast Heuristics** - Sub-microsecond pattern validation
- **Formal NFA Analysis** - Detect EDA (exponential) and IDA (polynomial) ambiguity
- **Complexity Scoring** - 0-100 scale with Big O notation
- **Adversarial Testing** - Generate pump patterns that expose vulnerabilities
- **Multiple Validation Modes** - Fast, Balanced, Thorough
- **CLI Tool** - Complete command-line interface for CI/CD and developer workflows

## The Problem

```go
// This innocent-looking regex...
pattern := "(a+)+b"
input := "aaaaaaaaaaaaaaaaaaaaaaaac"

// ...can hang your application for seconds or minutes!
```

**regret** detects these dangerous patterns before they cause problems.

## Common Evil Patterns

| Pattern | Issue | Complexity |
|---------|-------|------------|
| `(a+)+` | Nested quantifiers | O(2^n) - Exponential |
| `a*a*` | Overlapping quantifiers | O(n²) - Quadratic |
| `(a\|ab)+` | Overlapping alternation | Polynomial |
| `(.*)*` | Nested wildcards | Exponential |

## Documentation

Complete documentation is available in the [`docs/`](docs/) directory:

### Getting Started
- **[Getting Started Guide](docs/GETTING_STARTED.md)** - Installation, configuration, and basic usage
- **[Examples](examples/)** - Runnable code examples

### Reference
- **[API Reference](docs/API.md)** - Complete library API with all types and functions
- **[CLI Reference](docs/CLI.md)** - Command-line tool documentation

### Understanding regret
- **[When to Use regret](docs/WHEN_TO_USE.md)** - Who should use regret and in what scenarios
- **[How It Works](docs/HOW_IT_WORKS.md)** - Detection algorithms, NFA analysis, and theory
- **[Architecture](docs/ARCHITECTURE.md)** - Technical implementation and design

## Use Cases

- ✅ **Validate user input** before using in regex
- ✅ **CI/CD integration** to catch dangerous patterns
- ✅ **Security audits** to find ReDoS vulnerabilities
- ✅ **Code review** with automated pattern analysis
- ✅ **Performance testing** with adversarial inputs
- ✅ **Pre-commit hooks** to prevent unsafe patterns

## Why regret?

1. **Multiple detection layers** - Heuristics, NFA analysis, and adversarial testing
2. **Formal methods** - Based on automata theory and academic research
3. **Both library & CLI** - Use in code or as a standalone tool
4. **Production ready** - Comprehensive testing and documentation
5. **CI/CD friendly** - Exit codes, JSON output, configurable modes

## Is regret right for you?

**Not sure if you need regret?** Check out the **[When to Use regret](docs/WHEN_TO_USE.md)** guide, which includes:

- 🌳 Decision tree to evaluate if regret fits your needs
- 👥 Personas (who should use regret?)
- 📋 Scenarios (when to use regret?)
- 💡 Integration examples for different roles
- ❓ Common questions answered

**TL;DR:** Use regret if you accept regex from users, use Python/Ruby/JavaScript/PHP/Java, or want to prevent ReDoS attacks.

## Important: Go RE2 Engine

**regret validates Go-compatible regex patterns only.** Go uses the RE2 engine, which intentionally excludes features that can cause catastrophic backtracking:

### Not Supported (by design)
- ❌ **Lookaheads**: `(?=...)` and `(?!...)`
- ❌ **Lookbehinds**: `(?<=...)` and `(?<!...)`
- ❌ **Backreferences**: `\1`, `\2`, etc.
- ❌ **Conditional expressions**

### What This Means for regret

**Patterns with unsupported features are rejected:**
```go
regret.IsSafe("(?=.*[A-Z]).*")  // false - invalid syntax
regret.Validate("(?=.*[A-Z]).*") // error: "unsupported Perl syntax"
```

**Why?**
- regret uses Go's `regexp/syntax` parser
- Cannot analyze patterns that Go's parser rejects
- This is **correct** - these patterns won't work in Go anyway

### Use the Right Tool

- **For Go patterns** → Use **regret** ✓
- **For PCRE/JavaScript patterns** → Use PCRE-specific tools like [safe-regex](https://github.com/substack/safe-regex) or [rxxr2](https://github.com/superhuman/rxxr2)

See [How It Works](docs/HOW_IT_WORKS.md#limitations) for more details.

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Acknowledgments

Built with inspiration from:
- [RegexStaticAnalysis](https://github.com/NicolaasWeideman/RegexStaticAnalysis) by Nicolaas Weideman
- Academic research on NFA ambiguity detection
- [OWASP ReDoS](https://owasp.org/www-community/attacks/Regular_expression_Denial_of_Service_-_ReDoS) documentation

*This project was built using [Cursor](https://cursor.sh) and [Claude](https://anthropic.com/claude).*

---

**Built with ❤️ to make regex safer for everyone.**
