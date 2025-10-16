# When to Use regret

A practical guide to help you decide if regret is right for your situation.

## Quick Decision Tree

```
Do you work with regular expressions?
├─ No → You probably don't need regret
└─ Yes → Continue...
    │
    ├─ Are you using Go exclusively?
    │   ├─ Yes → Do you accept regex from users/external sources?
    │   │   ├─ Yes → ✅ Use regret for validation
    │   │   └─ No → ⚠️ Optional, useful for code review
    │   └─ No → Continue...
    │
    ├─ Do you use Python, Ruby, JavaScript, PHP, Perl, or Java?
    │   └─ Yes → ✅ Use regret! Your language is vulnerable to ReDoS
    │
    ├─ Do you accept regex patterns from untrusted sources?
    │   └─ Yes → ✅ Use regret! Critical for security
    │
    └─ Want to prevent ReDoS attacks?
        └─ Yes → ✅ Use regret for protection
```

---

## Who Should Use regret?

### ✅ Definitely Use regret

#### 1. **Application Developers Accepting User Input**

**You should use regret if:**
- Your app allows users to provide regex patterns
- You have search features with regex support
- You accept patterns via APIs or configuration files
- You build tools where users write their own patterns

**Why:**
User-provided regex is the #1 ReDoS attack vector. Even if you use Go (which is safe), validating patterns protects your entire ecosystem.

**Example scenarios:**
```go
// Search API accepting regex
func searchHandler(w http.ResponseWriter, r *http.Request) {
    pattern := r.URL.Query().Get("pattern")
    
    // ✅ Validate first
    if !regret.IsSafe(pattern) {
        http.Error(w, "Invalid pattern", http.StatusBadRequest)
        return
    }
    
    // Now safe to use
    re := regexp.MustCompile(pattern)
    // ... search logic
}
```

```javascript
// JavaScript app - validate patterns server-side
// POST /validate-pattern
{
  "pattern": "(a+)+"
}

// Go backend validates before allowing use
issues, _ := regret.Validate(pattern)
if len(issues) > 0 {
    return errors.New("dangerous pattern")
}
```

---

#### 2. **Security Engineers & Auditors**

**You should use regret if:**
- You perform security audits
- You review code for vulnerabilities
- You need to find ReDoS risks in codebases
- You write security policies

**Why:**
ReDoS is often overlooked in security audits. regret helps you find these vulnerabilities quickly.

**Example scenarios:**
```bash
# Audit entire codebase
regret check --recursive --format=json > security-audit.json

# Find high-severity issues
regret check --fail-on=high

# Check specific files
regret check
```

**Real-world finding:**
```bash
$ regret check

Found 3 dangerous patterns:
  
  ./api/search.py:45
    Pattern: (.*)*
    Severity: HIGH
    Issue: Nested wildcards cause exponential backtracking
    Recommendation: Use .* instead
    
  ./middleware/validator.js:112
    Pattern: (\w+)*\w+
    Severity: HIGH
    Issue: Overlapping quantifiers
    Recommendation: Use \w+ instead
```

---

#### 3. **DevOps/Platform Engineers**

**You should use regret if:**
- You manage CI/CD pipelines
- You enforce code quality gates
- You operate multi-language environments
- You manage infrastructure with regex configs

**Why:**
Prevent dangerous patterns from reaching production. One check in CI saves hours of incident response.

**Example scenarios:**
```yaml
# .github/workflows/security.yml
name: Security Checks
on: [push, pull_request]

jobs:
  regex-security:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      
      - name: Install regret
        run: go install github.com/theakshaypant/regret/cmd/regret@latest
      
      - name: Scan for dangerous regex patterns
        run: regret check --recursive --fail-on=high
```

```bash
# Pre-commit hook
#!/bin/bash
# .git/hooks/pre-commit

regret check --fail-on=high
if [ $? -ne 0 ]; then
    echo "❌ Dangerous regex patterns detected!"
    echo "Run: regret check --recursive"
    exit 1
fi
```

---

#### 4. **Backend Developers (Python, Ruby, JavaScript, PHP, Java)**

**You should use regret if:**
- You write backend services in languages with backtracking regex engines
- You process user input with regex
- You use regex for validation or parsing
- You maintain legacy codebases

**Why:**
Your language is vulnerable to ReDoS attacks. Unlike Go, Python/Ruby/JavaScript/etc. use backtracking engines that can hang on malicious input.

**Example scenarios:**

**Python:**
```python
# Django API - validate patterns before use
import subprocess
import json

def validate_pattern(pattern):
    """Validate regex pattern using regret"""
    result = subprocess.run(
        ['regret', 'check', pattern, '--format=json'],
        capture_output=True,
        text=True
    )
    
    if result.returncode != 0:
        data = json.loads(result.stdout)
        raise ValueError(f"Unsafe pattern: {data['issues']}")
    
    return True

# In your view
def search_view(request):
    pattern = request.GET.get('pattern')
    
    # Validate with regret
    validate_pattern(pattern)
    
    # Now safe to use
    import re
    regex = re.compile(pattern)
    results = regex.findall(request.data)
    return JsonResponse({'results': results})
```

**JavaScript:**
```javascript
// Express.js - validate patterns
const { exec } = require('child_process');
const util = require('util');
const execPromise = util.promisify(exec);

async function validatePattern(pattern) {
    try {
        await execPromise(`regret check "${pattern}"`);
        return true;
    } catch (error) {
        throw new Error('Dangerous regex pattern');
    }
}

app.post('/search', async (req, res) => {
    const { pattern } = req.body;
    
    // Validate first
    try {
        await validatePattern(pattern);
    } catch (error) {
        return res.status(400).json({ error: error.message });
    }
    
    // Use pattern
    const regex = new RegExp(pattern);
    const results = data.filter(item => regex.test(item));
    res.json({ results });
});
```

---

#### 5. **API Gateway/Service Mesh Operators**

**You should use regret if:**
- You operate API gateways
- You manage service meshes
- You handle routing/filtering with regex
- You expose regex configuration to teams

**Why:**
Gateways are high-value targets. One bad regex can DoS your entire infrastructure.

**Example scenarios:**
```go
// API Gateway - validate route patterns
type RouteConfig struct {
    Path    string
    Pattern string  // User-defined regex pattern
    Service string
}

func validateRouteConfig(config RouteConfig) error {
    // Validate regex pattern
    issues, err := regret.Validate(config.Pattern)
    if err != nil {
        return fmt.Errorf("invalid pattern: %w", err)
    }
    
    for _, issue := range issues {
        if issue.Severity >= regret.High {
            return fmt.Errorf("dangerous pattern: %s", issue.Message)
        }
    }
    
    return nil
}

// Before applying new routes
func applyRoutes(configs []RouteConfig) error {
    for _, config := range configs {
        if err := validateRouteConfig(config); err != nil {
            return err
        }
    }
    
    // Safe to apply
    gateway.UpdateRoutes(configs)
    return nil
}
```

---

### ⚠️ Consider Using regret

#### 6. **Library/Framework Authors**

**Consider regret if:**
- Your library accepts regex patterns
- You build validation frameworks
- You create testing tools
- You provide search/filter functionality

**Why:**
Protect your users from shooting themselves in the foot.

**Example scenarios:**
```go
// Validation library
package validator

import "github.com/theakshaypant/regret"

func (v *Validator) MatchesRegex(pattern string) *Rule {
    // Validate pattern safety
    if !regret.IsSafe(pattern) {
        return &Rule{
            Valid: false,
            Error: "unsafe regex pattern provided",
        }
    }
    
    re := regexp.MustCompile(pattern)
    return &Rule{
        Valid: true,
        Regex: re,
    }
}
```

---

#### 7. **Frontend Developers**

**Consider regret if:**
- You implement client-side search
- You validate input with regex
- You build dev tools or IDEs
- You test regex patterns

**Why:**
JavaScript's regex engine is vulnerable. Even though attacks are client-side, they can freeze browsers.

**Example scenarios:**
```javascript
// Browser IDE - validate patterns before execution
async function testRegexPattern(pattern, testString) {
    // Call backend validation API
    const response = await fetch('/api/validate-pattern', {
        method: 'POST',
        body: JSON.stringify({ pattern }),
        headers: { 'Content-Type': 'application/json' }
    });
    
    if (!response.ok) {
        throw new Error('Dangerous pattern detected');
    }
    
    // Safe to test in browser
    const regex = new RegExp(pattern);
    return regex.test(testString);
}
```

---

#### 8. **Go Developers**

**Consider regret if:**
- You accept regex from users
- You work in multi-language environments
- You want better code review insights
- You build tools for other languages

**Why:**
Even though Go is safe (uses RE2/Thompson NFA), regret helps in these scenarios:

1. **Validation for other languages:** Patterns from your Go API might be used by Python/JS clients
2. **Better error messages:** Explain WHY patterns are rejected
3. **Team education:** Help developers understand complexity
4. **Cross-service compatibility:** Ensure patterns work safely everywhere

**Example scenarios:**
```go
// Shared pattern service
type PatternService struct {
    // ... 
}

func (s *PatternService) ValidatePattern(pattern string) (*ValidationResult, error) {
    // Validate syntax (Go specific)
    _, err := regexp.Compile(pattern)
    if err != nil {
        return nil, fmt.Errorf("invalid syntax: %w", err)
    }
    
    // Validate safety (cross-language)
    score, err := regret.AnalyzeComplexity(pattern)
    if err != nil {
        return nil, err
    }
    
    return &ValidationResult{
        Pattern:         pattern,
        SafeInGo:        true,  // Always true if compile succeeded
        SafeInPython:    score.Overall < 70,
        SafeInJS:        score.Overall < 70,
        ComplexityScore: score.Overall,
        Warnings:        score.Issues,
    }, nil
}
```

---

### ❌ Don't Need regret

#### Who Doesn't Need regret?

**You probably don't need regret if:**

1. **You never use regex**
   - Your application doesn't use regular expressions at all
   - You use other parsing methods (parser generators, etc.)

2. **You only use hard-coded, reviewed patterns**
   - All regex patterns are in source code
   - Patterns are reviewed and tested
   - No dynamic or user-provided patterns
   - Small codebase where manual review is sufficient

3. **You only use simple patterns**
   - Literal string matching only
   - No quantifiers or alternation
   - No complex patterns

**Even then, regret can be useful for:**
- Code review insights
- Developer education
- Future-proofing as complexity grows

---

## When to Use regret (By Scenario)

### Scenario 1: New Project Setup

**When:** Starting a new project that will use regex

**Use regret for:**
- [ ] Set up pre-commit hooks
- [ ] Add CI/CD validation
- [ ] Document regex best practices
- [ ] Establish complexity thresholds

**Example:**
```bash
# Project initialization
go install github.com/theakshaypant/regret/cmd/regret@latest

# Add to pre-commit
cat > .git/hooks/pre-commit << 'EOF'
#!/bin/bash
regret check --fail-on=high
EOF
chmod +x .git/hooks/pre-commit

# Add to CI
# (see DevOps examples above)
```

---

### Scenario 2: Existing Project Security Audit

**When:** Auditing an existing codebase for security issues

**Use regret for:**
- [ ] Full codebase scan
- [ ] Generate security report
- [ ] Prioritize findings by severity
- [ ] Fix high-risk patterns first

**Example:**
```bash
# Initial audit
regret check --recursive --format=json > audit-$(date +%Y%m%d).json

# Review findings
regret check --recursive --fail-on=high

# Focus on critical paths
regret check ./src/auth --fail-on=medium
```

---

### Scenario 3: Before Production Deployment

**When:** Deploying new code to production

**Use regret to:**
- [ ] Validate all new patterns
- [ ] Check configuration files
- [ ] Verify user-facing features
- [ ] Test edge cases

**Example:**
```bash
# Pre-deployment checklist
regret check --fail-on=medium
regret check --fail-on=high

# Test specific patterns
regret check "pattern-from-config-file"
```

---

### Scenario 4: Incident Response

**When:** Investigating a performance incident or outage

**Use regret to:**
- [ ] Check if ReDoS is the cause
- [ ] Identify problematic patterns
- [ ] Generate test cases
- [ ] Validate fixes

**Example:**
```bash
# Found suspicious pattern in logs
regret analyze "(a+)+" --mode=thorough

# Generate test input
regret test "(a+)+" --size=20

# Test the actual input that caused issues
echo "aaaaaaaaaaaaaaaaaax" | your-app --pattern "(a+)+"
```

---

### Scenario 5: Code Review

**When:** Reviewing pull requests or code changes

**Use regret to:**
- [ ] Validate new patterns
- [ ] Check pattern complexity
- [ ] Suggest improvements
- [ ] Educate team members

**Example:**
```bash
# Check patterns in PR
git diff main... | grep -E "regexp\.|re\." | regret check --stdin

# Or scan changed files
regret check diff --name-only main...)

# Detailed analysis for discussion
regret analyze "proposed-pattern" --mode=thorough
```

---

### Scenario 6: Multi-Language Microservices

**When:** Operating a system with multiple programming languages

**Use regret to:**
- [ ] Centralized pattern validation
- [ ] Ensure cross-language compatibility
- [ ] Protect vulnerable services
- [ ] Maintain pattern registry

**Example:**
```go
// Pattern validation service
type PatternRegistry struct {
    store map[string]*ValidatedPattern
}

type ValidatedPattern struct {
    Pattern         string
    ValidatedAt     time.Time
    SafeInGo        bool
    SafeInPython    bool
    SafeInJS        bool
    ComplexityScore int
}

func (r *PatternRegistry) Register(pattern string) error {
    // Validate with regret
    score, err := regret.AnalyzeComplexity(pattern)
    if err != nil {
        return err
    }
    
    if score.Overall >= 70 {
        return errors.New("pattern too complex for multi-language use")
    }
    
    r.store[pattern] = &ValidatedPattern{
        Pattern:         pattern,
        ValidatedAt:     time.Now(),
        SafeInGo:        true,  // Go always safe with RE2
        SafeInPython:    score.Overall < 70,
        SafeInJS:        score.Overall < 70,
        ComplexityScore: score.Overall,
    }
    
    return nil
}
```

---

### Scenario 7: Building Internal Tools

**When:** Creating tools for other developers

**Use regret to:**
- [ ] Validate patterns in tool inputs
- [ ] Provide helpful error messages
- [ ] Teach best practices
- [ ] Prevent tool abuse

**Example:**
```go
// Log analyzer tool
func main() {
    pattern := flag.String("pattern", "", "Regex pattern to search logs")
    flag.Parse()
    
    // Validate pattern
    issues, err := regret.Validate(*pattern)
    if err != nil {
        log.Fatalf("Invalid pattern: %v", err)
    }
    
    for _, issue := range issues {
        if issue.Severity >= regret.High {
            log.Printf("⚠️  Warning: %s", issue.Message)
            log.Printf("   This pattern may be slow on large log files")
            log.Printf("   Consider using: %s", issue.Suggestion)
            
            fmt.Print("Continue anyway? [y/N]: ")
            var response string
            fmt.Scanln(&response)
            if response != "y" && response != "Y" {
                os.Exit(1)
            }
        }
    }
    
    // Use pattern
    re := regexp.MustCompile(*pattern)
    // ... analyze logs
}
```

---

### Scenario 8: Configuration Management

**When:** Managing application configuration with regex patterns

**Use regret to:**
- [ ] Validate config files before deployment
- [ ] Check patterns at startup
- [ ] Prevent misconfigurations
- [ ] Provide clear feedback

**Example:**
```go
// Config validation
type Config struct {
    Routes []RouteConfig `json:"routes"`
}

type RouteConfig struct {
    Path    string `json:"path"`
    Pattern string `json:"pattern"`
}

func validateConfig(configFile string) error {
    var config Config
    data, _ := os.ReadFile(configFile)
    json.Unmarshal(data, &config)
    
    for i, route := range config.Routes {
        issues, err := regret.Validate(route.Pattern)
        if err != nil {
            return fmt.Errorf("route %d: invalid pattern: %w", i, err)
        }
        
        for _, issue := range issues {
            if issue.Severity >= regret.High {
                return fmt.Errorf("route %d: unsafe pattern: %s", i, issue.Message)
            }
        }
    }
    
    return nil
}

func main() {
    if err := validateConfig("config.json"); err != nil {
        log.Fatalf("Configuration error: %v", err)
    }
    
    // Start application
}
```

---

## Decision Matrix

Use this matrix to decide if regret is right for you:

| Question | Yes | No |
|----------|-----|-----|
| Do you accept regex from users? | ✅ **Use regret** | Continue... |
| Do you use Python/Ruby/JS/PHP/Java? | ✅ **Use regret** | Continue... |
| Do you have regex in production? | ✅ **Consider regret** | Continue... |
| Do you want ReDoS protection? | ✅ **Use regret** | Continue... |
| Are you performing security audit? | ✅ **Use regret** | Continue... |
| Do you work with regex at all? | ⚠️ **Maybe useful** | ❌ **Not needed** |

---

## Common Questions

### "I only use Go. Do I need regret?"

**Short answer:** Probably not for Go itself, but yes if:
- You accept patterns from users (security)
- You work in multi-language environments (compatibility)
- You want better developer experience (education)

**Why:** Go uses RE2 (Thompson NFA), so it's inherently safe. But regret helps validate patterns that might be used by other services or languages.

---

### "How do I integrate regret into my workflow?"

**Start small:**

1. **Week 1:** Install CLI, try `regret check` on your patterns
2. **Week 2:** Add to pre-commit hooks for new changes only
3. **Week 3:** Add to CI/CD pipeline
4. **Week 4:** Fix high-severity issues found in audit

**For library integration:**
```go
go get github.com/theakshaypant/regret
```

Then use in your code (see examples above).

---

### "What if regret flags a safe pattern?"

regret is conservative and may have false positives. If you're certain a pattern is safe:

1. **Review the warning:** Understand why it was flagged
2. **Test it:** Use `regret test` to generate adversarial inputs
3. **Document it:** Add comments explaining why it's safe
4. **Configure thresholds:** Adjust `MaxComplexityScore` if needed
5. **Override if necessary:** Use bypass mechanisms for specific patterns

**Example:**
```go
// Pattern is flagged but safe in our case because:
// 1. Input is bounded to max 10 characters
// 2. Only used internally, never from users
// 3. Tested with adversarial inputs up to n=100
pattern := "(a{2,5})+"  // regret score: 45 (medium)

// Explicitly acknowledge and document
if os.Getenv("SKIP_PATTERN_VALIDATION") != "true" {
    // Still validate in production
    issues, _ := regret.Validate(pattern)
    log.Printf("Pattern warnings: %v", issues)
}
```

---

### "Is regret required for Go projects?"

**No, but recommended if:**

Your project meets ANY of these criteria:
- [ ] Accepts regex from external sources
- [ ] Part of multi-language system
- [ ] Security-critical application
- [ ] Team is learning regex best practices
- [ ] You want defense-in-depth

Go's RE2 engine protects the Go service, but regret protects your entire ecosystem.

---

## Getting Started

Ready to use regret? Here's how to start:

### 1. Installation

```bash
# CLI tool
go install github.com/theakshaypant/regret/cmd/regret@latest

# Library
go get github.com/theakshaypant/regret
```

### 2. Quick Test

```bash
# Test a pattern
regret check "(a+)+"

# Scan your code
regret check
```

### 3. Integration

Choose based on your role:

- **Developers:** Add to pre-commit hooks
- **DevOps:** Add to CI/CD pipeline
- **Security:** Run full audit
- **Architects:** Add to pattern validation service

See [GETTING_STARTED.md](GETTING_STARTED.md) for detailed instructions.

---

## Summary

### Use regret if you:

✅ Accept regex patterns from users or external sources  
✅ Use Python, Ruby, JavaScript, PHP, Java, or other backtracking engines  
✅ Want to prevent ReDoS attacks  
✅ Perform security audits  
✅ Operate in multi-language environments  
✅ Care about application security and performance  

### Don't need regret if you:

❌ Never use regular expressions  
❌ Only use hard-coded, simple, well-tested patterns  
❌ Have no external inputs or user-provided patterns  

### Still not sure?

**Try it!** It takes 5 minutes:

```bash
# Install
go install github.com/theakshaypant/regret/cmd/regret@latest

# Scan your code
regret check

# Review findings
```

If it finds issues, you needed it. If it doesn't, you have peace of mind.

---

## Further Reading

- [Getting Started](GETTING_STARTED.md) - Installation and setup
- [API Reference](API.md) - Library integration details
- [CLI Reference](CLI.md) - Command-line tool usage
- [How It Works](HOW_IT_WORKS.md) - Technical details
- [Examples](../examples/) - Code examples

---

**Questions?** [Open an issue](https://github.com/theakshaypant/regret/issues) or check the [documentation](README.md).

