// Package regret provides validation and analysis for regex patterns
// to prevent Regular Expression Denial of Service (ReDoS) attacks.
//
// The library analyzes regex patterns for dangerous constructs using formal
// automata theory, detecting both exponential (EDA) and polynomial (IDA)
// backtracking vulnerabilities.
package regret

import "time"

// ValidationMode controls the depth of analysis performed.
type ValidationMode int

const (
	// Fast mode uses only quick heuristics (~microseconds).
	// Best for hot paths and user input validation.
	Fast ValidationMode = iota

	// Balanced mode includes syntax analysis and NFA ambiguity detection (~milliseconds).
	// Recommended for most use cases.
	Balanced

	// Thorough mode includes all checks plus adversarial input generation (~tens of milliseconds).
	// Best for configuration validation and security auditing.
	Thorough
)

// String returns the string representation of the validation mode.
func (v ValidationMode) String() string {
	switch v {
	case Fast:
		return "fast"
	case Balanced:
		return "balanced"
	case Thorough:
		return "thorough"
	default:
		return "unknown"
	}
}

// CheckFlags is a bitmask of checks to perform during validation.
type CheckFlags uint32

const (
	// CheckNestedQuantifiers detects nested quantifiers like (a+)+, (x*)*.
	CheckNestedQuantifiers CheckFlags = 1 << iota

	// CheckOverlappingAlternation detects alternations with overlapping branches like (a|ab)+.
	CheckOverlappingAlternation

	// CheckCatastrophicBacktrack detects patterns that cause catastrophic backtracking.
	CheckCatastrophicBacktrack

	// CheckUnboundedRepetition detects unbounded repetition without anchors like .*password.*.
	CheckUnboundedRepetition

	// CheckExponentialPaths detects patterns with exponential matching paths.
	CheckExponentialPaths

	// CheckComplexityScore calculates and validates complexity scores.
	CheckComplexityScore

	// CheckMemoryUsage estimates memory usage for pattern matching.
	CheckMemoryUsage

	// CheckNFAAmbiguity performs NFA analysis to detect EDA and IDA.
	CheckNFAAmbiguity

	// CheckPolynomialDegree detects and calculates polynomial backtracking degree.
	CheckPolynomialDegree

	// CheckContextAwareness analyzes pattern context and ordering for safety.
	CheckContextAwareness

	// CheckAll enables all available checks.
	CheckAll CheckFlags = ^CheckFlags(0)

	// CheckDefault includes the most important checks for typical use cases.
	CheckDefault = CheckNestedQuantifiers |
		CheckOverlappingAlternation |
		CheckCatastrophicBacktrack |
		CheckNFAAmbiguity
)

// Options configures the validation and analysis behavior.
type Options struct {
	// Mode controls the depth of analysis.
	// Default: Balanced
	Mode ValidationMode

	// Timeout sets the maximum time for analysis.
	// Analysis will return partial results if timeout is exceeded.
	// Default: 100ms for Balanced, 1s for Thorough
	Timeout time.Duration

	// Checks specifies which checks to perform (bitmask).
	// Default: CheckDefault
	Checks CheckFlags

	// MaxComplexityScore is the maximum acceptable complexity score (0-100).
	// Patterns with higher scores will be flagged.
	// Default: 70
	MaxComplexityScore int

	// MaxPatternLength is the maximum allowed pattern length.
	// Very long patterns can slow down analysis.
	// Default: 1000, set to 0 for no limit
	MaxPatternLength int

	// MaxNestingDepth is the maximum allowed quantifier nesting depth.
	// Default: 3
	MaxNestingDepth int

	// MaxQuantifiers is the maximum number of quantifiers allowed.
	// Default: 20
	MaxQuantifiers int

	// StrictMode treats warnings as errors.
	// Default: false
	StrictMode bool

	// AllowUnsafe skips validation (passthrough mode).
	// Use with caution, primarily for testing.
	// Default: false
	AllowUnsafe bool
}

// DefaultOptions returns the recommended default configuration.
func DefaultOptions() *Options {
	return &Options{
		Mode:               Balanced,
		Timeout:            100 * time.Millisecond,
		Checks:             CheckDefault,
		MaxComplexityScore: 70,
		MaxPatternLength:   1000,
		MaxNestingDepth:    3,
		MaxQuantifiers:     20,
		StrictMode:         false,
		AllowUnsafe:        false,
	}
}

// FastOptions returns options optimized for speed.
func FastOptions() *Options {
	return &Options{
		Mode:               Fast,
		Timeout:            10 * time.Millisecond,
		Checks:             CheckNestedQuantifiers | CheckCatastrophicBacktrack,
		MaxComplexityScore: 70,
		MaxPatternLength:   1000,
		MaxNestingDepth:    3,
		MaxQuantifiers:     20,
		StrictMode:         false,
		AllowUnsafe:        false,
	}
}

// ThoroughOptions returns options for comprehensive analysis.
func ThoroughOptions() *Options {
	return &Options{
		Mode:               Thorough,
		Timeout:            1 * time.Second,
		Checks:             CheckAll,
		MaxComplexityScore: 70,
		MaxPatternLength:   2000,
		MaxNestingDepth:    5,
		MaxQuantifiers:     50,
		StrictMode:         true,
		AllowUnsafe:        false,
	}
}

// Severity represents the severity level of an issue.
type Severity int

const (
	// Critical issues will definitely cause ReDoS attacks.
	Critical Severity = iota

	// High severity issues are very likely to be exploited.
	High

	// Medium severity issues are potentially problematic.
	Medium

	// Low severity issues are minor concerns.
	Low

	// Info provides informational messages without immediate risk.
	Info
)

// String returns the string representation of the severity.
func (s Severity) String() string {
	switch s {
	case Critical:
		return "critical"
	case High:
		return "high"
	case Medium:
		return "medium"
	case Low:
		return "low"
	case Info:
		return "info"
	default:
		return "unknown"
	}
}

// IssueType represents the type of issue detected.
type IssueType int

const (
	// NestedQuantifiers indicates nested quantifiers like (a+)+.
	NestedQuantifiers IssueType = iota

	// OverlappingAlternation indicates alternations with overlapping branches.
	OverlappingAlternation

	// RepeatedCaptureGroup indicates repeated capturing groups.
	RepeatedCaptureGroup

	// ExponentialBacktracking indicates exponential backtracking (EDA).
	ExponentialBacktracking

	// PolynomialBacktracking indicates polynomial backtracking (IDA).
	PolynomialBacktracking

	// UnboundedRepetition indicates unbounded repetition without anchors.
	UnboundedRepetition

	// AmbiguousPattern indicates ambiguous matching behavior.
	AmbiguousPattern

	// ComplexityThresholdExceeded indicates complexity score is too high.
	ComplexityThresholdExceeded

	// ContextuallyDangerous indicates pattern is dangerous in current context.
	ContextuallyDangerous
)

// String returns the string representation of the issue type.
func (i IssueType) String() string {
	switch i {
	case NestedQuantifiers:
		return "nested_quantifiers"
	case OverlappingAlternation:
		return "overlapping_alternation"
	case RepeatedCaptureGroup:
		return "repeated_capture_group"
	case ExponentialBacktracking:
		return "exponential_backtracking"
	case PolynomialBacktracking:
		return "polynomial_backtracking"
	case UnboundedRepetition:
		return "unbounded_repetition"
	case AmbiguousPattern:
		return "ambiguous_pattern"
	case ComplexityThresholdExceeded:
		return "complexity_threshold_exceeded"
	case ContextuallyDangerous:
		return "contextually_dangerous"
	default:
		return "unknown"
	}
}

// Position represents a location in the regex pattern.
type Position struct {
	// Start is the starting byte offset in the pattern.
	Start int

	// End is the ending byte offset in the pattern.
	End int

	// Line is the line number for multiline patterns (1-indexed).
	Line int

	// Column is the column number (1-indexed).
	Column int
}

// Issue represents a detected problem in a regex pattern.
type Issue struct {
	// Type is the type of issue detected.
	Type IssueType

	// Severity indicates how serious the issue is.
	Severity Severity

	// Position indicates where in the pattern the issue occurs.
	Position Position

	// Pattern is the problematic sub-pattern.
	Pattern string

	// Message is a human-readable description of the issue.
	Message string

	// Example is an example adversarial input that exploits this issue.
	Example string

	// Suggestion provides guidance on how to fix the issue.
	Suggestion string

	// Complexity is the local complexity contribution (0-100).
	Complexity int

	// Details contains additional technical details about the issue.
	Details map[string]interface{}
}

// Complexity represents time or space complexity classes.
type Complexity int

const (
	// Constant represents O(1) complexity.
	Constant Complexity = iota

	// Linear represents O(n) complexity.
	Linear

	// Quadratic represents O(n²) complexity.
	Quadratic

	// Cubic represents O(n³) complexity.
	Cubic

	// Polynomial represents O(n^k) complexity for k > 3.
	Polynomial

	// Exponential represents O(2^n) complexity.
	Exponential

	// Unknown represents unknown or indeterminate complexity.
	Unknown
)

// String returns the string representation of the complexity.
func (c Complexity) String() string {
	switch c {
	case Constant:
		return "O(1)"
	case Linear:
		return "O(n)"
	case Quadratic:
		return "O(n²)"
	case Cubic:
		return "O(n³)"
	case Polynomial:
		return "O(n^k)"
	case Exponential:
		return "O(2^n)"
	case Unknown:
		return "O(?)"
	default:
		return "unknown"
	}
}

// BigO returns the mathematical Big-O notation.
func (c Complexity) BigO() string {
	return c.String()
}

// ComplexityScore contains detailed complexity analysis results.
type ComplexityScore struct {
	// Overall is the overall complexity score (0-100).
	// Lower is better. Scores above 70 indicate problematic patterns.
	Overall int

	// TimeComplexity is the estimated worst-case time complexity.
	TimeComplexity Complexity

	// SpaceComplexity is the estimated space complexity.
	SpaceComplexity Complexity

	// HasEDA indicates if Exponential Degree of Ambiguity was detected.
	// This means the pattern has exponentially many ways to match input.
	HasEDA bool

	// HasIDA indicates if Infinite Degree of Ambiguity was detected.
	// This means the pattern has polynomially many ways to match input.
	HasIDA bool

	// PolynomialDegree is the degree of polynomial backtracking.
	// 2 = quadratic, 3 = cubic, etc. Only set if HasIDA is true.
	PolynomialDegree int

	// Metrics contains detailed metrics about the pattern.
	Metrics Metrics

	// WorstCaseInput is an example input that triggers worst-case behavior.
	WorstCaseInput string

	// PumpPattern contains the pump components for generating adversarial inputs.
	PumpPattern []string

	// Explanation is a human-readable explanation of the complexity analysis.
	Explanation string

	// Safe indicates whether the pattern is considered safe based on the analysis.
	Safe bool
}

// Metrics contains detailed metrics about a regex pattern.
type Metrics struct {
	// NestingDepth is the maximum quantifier nesting depth.
	NestingDepth int

	// QuantifierCount is the total number of quantifiers in the pattern.
	QuantifierCount int

	// AlternationCount is the number of alternation operators (|).
	AlternationCount int
}

// PumpPattern represents a pattern for generating adversarial inputs.
// It uses the "pumping" technique to create progressively longer inputs
// that expose exponential or polynomial backtracking.
type PumpPattern struct {
	// Prefix is the initial string before the pumped section.
	Prefix string

	// Pumps contains the repeating components.
	// Multiple pumps can be interleaved or concatenated.
	Pumps []string

	// Suffix is the final string after the pumped section.
	// Often a character that doesn't match, forcing backtracking.
	Suffix string

	// Interleave indicates whether pumps should be interleaved.
	// If true: pump[0], pump[1], pump[0], pump[1], ...
	// If false: pump[0] * n, pump[1] * m, ...
	Interleave bool

	// Description explains what this pump pattern tests.
	Description string
}

// Generate creates an adversarial input of the specified size.
// The size parameter controls how many times the pump components are repeated.
func (p *PumpPattern) Generate(size int) string {
	if size <= 0 {
		return p.Prefix + p.Suffix
	}

	result := p.Prefix

	if p.Interleave {
		// Interleave pumps: p[0]p[1]p[0]p[1]...
		for i := 0; i < size; i++ {
			for _, pump := range p.Pumps {
				result += pump
			}
		}
	} else {
		// Concatenate pumps: p[0]*n p[1]*m ...
		for _, pump := range p.Pumps {
			for i := 0; i < size; i++ {
				result += pump
			}
		}
	}

	result += p.Suffix
	return result
}

// GenerateSequence creates a sequence of adversarial inputs with increasing sizes.
func (p *PumpPattern) GenerateSequence(start, end, step int) []string {
	if start <= 0 {
		start = 1
	}
	if step <= 0 {
		step = 1
	}

	var sequence []string
	for size := start; size <= end; size += step {
		sequence = append(sequence, p.Generate(size))
	}
	return sequence
}
