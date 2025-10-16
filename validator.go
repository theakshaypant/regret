package regret

import (
	"errors"
	"fmt"

	"github.com/theakshaypant/regret/internal/analyzer"
	"github.com/theakshaypant/regret/internal/detector"
	"github.com/theakshaypant/regret/internal/parser"
	"github.com/theakshaypant/regret/internal/pump"
)

var (
	// ErrInvalidPattern indicates the pattern is syntactically invalid.
	ErrInvalidPattern = errors.New("invalid regex pattern")

	// ErrPatternTooLong indicates the pattern exceeds the maximum allowed length.
	ErrPatternTooLong = errors.New("pattern too long")

	// ErrTimeout indicates the analysis exceeded the configured timeout.
	ErrTimeout = errors.New("analysis timeout exceeded")

	// ErrUnsupportedFeature indicates the pattern uses unsupported regex features.
	ErrUnsupportedFeature = errors.New("unsupported regex feature")
)

// IsSafe performs a quick safety check on a regex pattern using strict default settings.
// Returns true if the pattern is safe to use, false otherwise.
//
// This function uses Fast mode with CheckDefault flags and is optimized for performance.
// For detailed information about issues, use Validate() instead.
//
// Example:
//
//	if !regret.IsSafe("(a+)+") {
//	    return errors.New("unsafe regex pattern")
//	}
func IsSafe(pattern string) bool {
	opts := FastOptions()
	opts.StrictMode = true
	issues, err := ValidateWithOptions(pattern, opts)
	if err != nil {
		return false
	}
	return len(issues) == 0
}

// Validate analyzes a regex pattern and returns all detected issues.
// Uses default options (Balanced mode with CheckDefault flags).
//
// Returns a slice of issues (which may be empty) and an error if the pattern
// cannot be analyzed (e.g., syntax errors, timeout).
//
// Example:
//
//	issues, err := regret.Validate("(a+)+")
//	if err != nil {
//	    return fmt.Errorf("validation failed: %w", err)
//	}
//	for _, issue := range issues {
//	    fmt.Printf("Issue: %s at position %d\n", issue.Message, issue.Position.Start)
//	}
func Validate(pattern string) ([]Issue, error) {
	return ValidateWithOptions(pattern, DefaultOptions())
}

// ValidateWithOptions analyzes a regex pattern with custom configuration options.
//
// The validation process runs in multiple layers depending on the Mode:
//   - Fast: Quick heuristics only (~microseconds)
//   - Balanced: Heuristics + NFA analysis (~milliseconds)
//   - Thorough: Full analysis + adversarial testing (~tens of milliseconds)
//
// Returns a slice of issues (which may be empty) and an error if the pattern
// cannot be analyzed.
//
// Example:
//
//	opts := &regret.Options{
//	    Mode: regret.Balanced,
//	    Timeout: 100 * time.Millisecond,
//	    MaxComplexityScore: 70,
//	}
//	issues, err := regret.ValidateWithOptions(pattern, opts)
func ValidateWithOptions(pattern string, opts *Options) ([]Issue, error) {
	if opts == nil {
		opts = DefaultOptions()
	}

	// Handle passthrough mode
	if opts.AllowUnsafe {
		return []Issue{}, nil
	}

	// Check pattern length
	if opts.MaxPatternLength > 0 && len(pattern) > opts.MaxPatternLength {
		return nil, fmt.Errorf("%w: %d > %d", ErrPatternTooLong, len(pattern), opts.MaxPatternLength)
	}

	// Create validator
	v := newValidator(opts)
	return v.validate(pattern)
}

// AnalyzeComplexity performs detailed complexity analysis on a regex pattern.
//
// This function provides comprehensive information including:
//   - Complexity score (0-100)
//   - Time and space complexity estimates
//   - EDA/IDA detection
//   - Polynomial degree (if applicable)
//   - Detailed metrics
//   - Adversarial input examples
//
// Uses Thorough mode for complete analysis.
//
// Example:
//
//	score, err := regret.AnalyzeComplexity("(a+)+")
//	if err != nil {
//	    return err
//	}
//	fmt.Printf("Complexity: %d/100\n", score.Overall)
//	if score.HasEDA {
//	    fmt.Printf("Exponential backtracking detected!\n")
//	    fmt.Printf("Worst-case input: %s\n", score.WorstCaseInput)
//	}
func AnalyzeComplexity(pattern string) (*ComplexityScore, error) {
	opts := ThoroughOptions()
	opts.Checks = CheckAll

	// Create analyzer
	a := newAnalyzer(opts)
	return a.analyze(pattern)
}

// validator is the internal validator implementation.
type validator struct {
	opts   *Options
	parser *parser.Parser
	detect *detector.Detector
}

func newValidator(opts *Options) *validator {
	// Convert public options to internal detector options
	detectorOpts := &detector.Options{
		Mode:   detector.ValidationMode(opts.Mode),
		Checks: uint32(opts.Checks),
	}

	return &validator{
		opts:   opts,
		parser: parser.NewParser(),
		detect: detector.NewDetector(detectorOpts),
	}
}

func (v *validator) validate(pattern string) ([]Issue, error) {
	// Parse the pattern
	re, err := v.parser.Parse(pattern)
	if err != nil {
		return nil, err
	}

	// Run detection based on mode
	internalIssues, err := v.detect.Detect(re, pattern)
	if err != nil {
		return nil, err
	}

	// Convert internal issues to public issues
	return convertIssues(internalIssues), nil
}

// convertIssues converts internal detector issues to public API issues.
func convertIssues(internal []detector.Issue) []Issue {
	issues := make([]Issue, len(internal))
	for i, iss := range internal {
		issues[i] = convertIssue(iss)
	}
	return issues
}

// convertIssue converts a single internal detector issue to public API issue.
func convertIssue(iss detector.Issue) Issue {
	return Issue{
		Type:       issueTypeFromString(iss.Type),
		Severity:   severityFromString(iss.Severity),
		Position:   Position{Start: iss.Position.Start, End: iss.Position.End, Line: iss.Position.Line, Column: iss.Position.Column},
		Pattern:    iss.Pattern,
		Message:    iss.Message,
		Example:    iss.Example,
		Suggestion: iss.Suggestion,
		Complexity: iss.Complexity,
		Details:    make(map[string]interface{}),
	}
}

func issueTypeFromString(s string) IssueType {
	switch s {
	case "nested_quantifiers":
		return NestedQuantifiers
	case "overlapping_alternation":
		return OverlappingAlternation
	case "exponential_backtracking":
		return ExponentialBacktracking
	case "polynomial_backtracking":
		return PolynomialBacktracking
	default:
		return AmbiguousPattern
	}
}

func severityFromString(s string) Severity {
	switch s {
	case "critical":
		return Critical
	case "high":
		return High
	case "medium":
		return Medium
	case "low":
		return Low
	default:
		return Info
	}
}

// anlz wraps the internal analyzer.
type anlz struct {
	opts   *Options
	impl   *analyzer.Analyzer
	parser *parser.Parser
}

func newAnalyzer(opts *Options) *anlz {
	analyzerOpts := &analyzer.Options{
		Timeout:            opts.Timeout,
		MaxComplexityScore: opts.MaxComplexityScore,
	}

	return &anlz{
		opts:   opts,
		impl:   analyzer.NewAnalyzer(analyzerOpts),
		parser: parser.NewParser(),
	}
}

func (a *anlz) analyze(pattern string) (*ComplexityScore, error) {
	// Parse pattern
	re, err := a.parser.Parse(pattern)
	if err != nil {
		return nil, err
	}

	// Analyze complexity
	result, err := a.impl.Analyze(re, pattern)
	if err != nil {
		return nil, err
	}

	// Convert internal result to public result
	complexity := complexityFromString(result.Complexity)

	// Generate pump pattern for adversarial testing
	var pumpComponents []string
	var worstCaseInput string

	// Only generate pump pattern if the pattern is potentially unsafe
	if result.Score >= 50 {
		pumpGen := newPumpGenerator(a.opts)
		pump, err := pumpGen.generate(pattern)
		if err == nil && pump != nil {
			pumpComponents = pump.Pumps
			// Generate a worst-case input with moderate pump size
			// Use first pump size if available, otherwise default to 20
			pumpSize := 20
			if len(pump.Pumps) > 0 {
				worstCaseInput = pump.Generate(pumpSize)
			}
		}
		// Silently ignore pump generation errors - it's supplementary information
	}

	return &ComplexityScore{
		Overall:          result.Score,
		TimeComplexity:   complexity,
		SpaceComplexity:  Linear, // TODO: implement space complexity analysis
		HasEDA:           result.TimeClass == "exponential",
		HasIDA:           result.TimeClass == "polynomial",
		PolynomialDegree: result.Degree,
		Metrics: Metrics{
			NestingDepth:     getMetricInt(result.Metrics, "nesting_depth"),
			QuantifierCount:  getMetricInt(result.Metrics, "quantifier_count"),
			AlternationCount: getMetricInt(result.Metrics, "alternations"),
		},
		WorstCaseInput: worstCaseInput,
		PumpPattern:    pumpComponents,
		Explanation:    result.Description,
		Safe:           result.Score < 50,
	}, nil
}

// pumpGen wraps the internal pump generator.
type pumpGen struct {
	opts   *Options
	impl   *pump.Generator
	parser *parser.Parser
}

func newPumpGenerator(opts *Options) *pumpGen {
	pumpOpts := &pump.Options{
		PumpSize:       10,
		MaxPumpSize:    100,
		IncludeFailure: true,
	}

	return &pumpGen{
		opts:   opts,
		impl:   pump.NewGenerator(pumpOpts),
		parser: parser.NewParser(),
	}
}

func (g *pumpGen) generate(pattern string) (*PumpPattern, error) {
	// Parse pattern
	re, err := g.parser.Parse(pattern)
	if err != nil {
		return nil, err
	}

	// Generate pump patterns
	results, err := g.impl.Generate(re, pattern)
	if err != nil {
		return nil, err
	}

	// Return first pump pattern (most relevant)
	if len(results) == 0 {
		return nil, fmt.Errorf("no pump pattern generated")
	}

	result := results[0]

	// Convert internal result to public result
	pumps := []string{result.PumpComponent}

	return &PumpPattern{
		Prefix:      result.BaseString,
		Pumps:       pumps,
		Suffix:      result.FailSuffix,
		Interleave:  false,
		Description: result.Description,
	}, nil
}

// Helper functions

func complexityFromString(s string) Complexity {
	switch s {
	case "O(1)":
		return Constant
	case "O(n)":
		return Linear
	case "O(n²)":
		return Quadratic
	case "O(n³)":
		return Cubic
	case "O(n^k)":
		return Polynomial
	case "O(2^n)":
		return Exponential
	default:
		return Unknown
	}
}

func getMetricInt(metrics map[string]interface{}, key string) int {
	if val, ok := metrics[key]; ok {
		if intVal, ok := val.(int); ok {
			return intVal
		}
	}
	return 0
}
