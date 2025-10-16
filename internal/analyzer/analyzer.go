// Package analyzer implements complexity analysis for regex patterns.
package analyzer

import (
	"regexp/syntax"
	"time"
)

// Options contains configuration for analysis.
type Options struct {
	Timeout            time.Duration
	MaxComplexityScore int
}

// ComplexityScore contains complexity analysis results (internal format).
type ComplexityScore struct {
	Score       int                    // 0-100 scale
	Complexity  string                 // O(1), O(n), O(n²), O(2^n)
	Description string                 // Human-readable description
	TimeClass   string                 // constant, linear, polynomial, exponential
	Degree      int                    // For polynomial: degree (2=quadratic, 3=cubic)
	Issues      []string               // List of contributing issues
	Metrics     map[string]interface{} // Detailed metrics
}

// Analyzer performs complexity analysis on regex patterns.
type Analyzer struct {
	opts *Options
}

// NewAnalyzer creates a new analyzer with the given options.
func NewAnalyzer(opts *Options) *Analyzer {
	if opts == nil {
		opts = &Options{
			Timeout:            5 * time.Second,
			MaxComplexityScore: 100,
		}
	}
	return &Analyzer{opts: opts}
}

// Analyze performs comprehensive complexity analysis on a regex pattern.
func (a *Analyzer) Analyze(re *syntax.Regexp, pattern string) (*ComplexityScore, error) {
	score := &ComplexityScore{
		Score:       0,
		Complexity:  "O(n)",
		Description: "Linear time complexity",
		TimeClass:   "linear",
		Degree:      1,
		Issues:      make([]string, 0),
		Metrics:     make(map[string]interface{}),
	}

	// Analyze different aspects
	a.analyzeNesting(re, score)
	a.analyzeQuantifiers(re, score)
	a.analyzeAlternations(re, score)
	a.analyzePattern(re, score)

	// Determine final complexity class
	a.determineComplexity(score)

	// Cap score at max
	if score.Score > a.opts.MaxComplexityScore {
		score.Score = a.opts.MaxComplexityScore
	}

	return score, nil
}

// EstimateComplexity provides a quick complexity estimate.
func (a *Analyzer) EstimateComplexity(re *syntax.Regexp) string {
	// Quick checks
	if hasNestedQuantifiers(re) {
		return "O(2^n)"
	}

	overlapping := findOverlappingQuantifiers(re)
	if len(overlapping) > 0 {
		degree := len(overlapping) + 1
		if degree == 2 {
			return "O(n²)"
		} else if degree == 3 {
			return "O(n³)"
		}
		return "O(n^k)"
	}

	quantifiers := countQuantifiers(re)
	if quantifiers > 0 {
		return "O(n)"
	}

	return "O(1)"
}

// Analysis methods

func (a *Analyzer) analyzeNesting(re *syntax.Regexp, score *ComplexityScore) {
	maxDepth := 0
	nestedCount := 0

	walkRegexp(re, func(node *syntax.Regexp) bool {
		if isQuantifier(node) {
			depth := getQuantifierDepth(node)
			if depth > maxDepth {
				maxDepth = depth
			}

			// Only count true nesting (quantifier directly inside quantifier)
			// Don't count Concat/Capture that happen to contain quantifiers
			if isTrulyNested(node) {
				nestedCount++
			}
		}
		return true
	})

	score.Metrics["nesting_depth"] = maxDepth
	score.Metrics["nested_quantifiers"] = nestedCount

	if nestedCount > 0 {
		score.Score += 40 + (nestedCount * 10)
		score.Issues = append(score.Issues, "nested quantifiers (exponential risk)")
		score.TimeClass = "exponential"
		score.Degree = nestedCount + 1
	} else if maxDepth > 3 {
		score.Score += 15 + (maxDepth * 5)
		score.Issues = append(score.Issues, "deep nesting")
	}
}

func (a *Analyzer) analyzeQuantifiers(re *syntax.Regexp, score *ComplexityScore) {
	quantifierCount := countQuantifiers(re)
	overlappingSeqs := findOverlappingQuantifiers(re)

	score.Metrics["quantifier_count"] = quantifierCount
	score.Metrics["overlapping_sequences"] = len(overlappingSeqs)

	if len(overlappingSeqs) > 0 {
		degree := len(overlappingSeqs) + 1
		baseScore := 25 + (degree * 10)
		score.Score += baseScore

		if degree == 2 {
			score.Issues = append(score.Issues, "overlapping quantifiers (quadratic)")
		} else if degree == 3 {
			score.Issues = append(score.Issues, "overlapping quantifiers (cubic)")
		} else {
			score.Issues = append(score.Issues, "overlapping quantifiers (high polynomial)")
		}

		if score.TimeClass == "linear" {
			score.TimeClass = "polynomial"
			score.Degree = degree
		}
	}

	if quantifierCount > 15 {
		score.Score += 10 + (quantifierCount - 15)
		score.Issues = append(score.Issues, "excessive quantifiers")
	}
}

func (a *Analyzer) analyzeAlternations(re *syntax.Regexp, score *ComplexityScore) {
	alternationCount := 0
	overlappingAlts := 0

	walkRegexp(re, func(node *syntax.Regexp) bool {
		if node.Op == syntax.OpAlternate {
			alternationCount++
			if hasOverlappingBranches(node) {
				overlappingAlts++
			}
		}
		return true
	})

	score.Metrics["alternations"] = alternationCount
	score.Metrics["overlapping_alternations"] = overlappingAlts

	if overlappingAlts > 0 {
		score.Score += 20 + (overlappingAlts * 5)
		score.Issues = append(score.Issues, "overlapping alternation branches")
	}
}

func (a *Analyzer) analyzePattern(re *syntax.Regexp, score *ComplexityScore) {
	patternLen := len(re.String())
	score.Metrics["pattern_length"] = patternLen

	if patternLen > 500 {
		score.Score += 10
		score.Issues = append(score.Issues, "very long pattern")
	}

	if hasDotStar(re) {
		score.Score += 5
		score.Metrics["has_dotstar"] = true
	}
}

func (a *Analyzer) determineComplexity(score *ComplexityScore) {
	switch score.TimeClass {
	case "exponential":
		score.Complexity = "O(2^n)"
		score.Description = "Exponential time complexity - catastrophic backtracking risk"
		if score.Score < 70 {
			score.Score = 70
		}

	case "polynomial":
		if score.Degree == 2 {
			score.Complexity = "O(n²)"
			score.Description = "Quadratic time complexity - moderate backtracking risk"
		} else if score.Degree == 3 {
			score.Complexity = "O(n³)"
			score.Description = "Cubic time complexity - high backtracking risk"
		} else {
			score.Complexity = "O(n^k)"
			score.Description = "Polynomial time complexity - backtracking risk"
		}
		if score.Score < 40 {
			score.Score = 40
		}

	case "linear":
		if score.Score < 20 {
			score.Complexity = "O(n)"
			score.Description = "Linear time complexity - good performance"
		} else {
			score.Complexity = "O(n)"
			score.Description = "Linear time complexity with some inefficiencies"
		}

	default:
		if score.Score < 10 {
			score.Complexity = "O(1)"
			score.Description = "Constant time complexity - excellent performance"
			score.TimeClass = "constant"
		}
	}
}

// Helper functions

func walkRegexp(re *syntax.Regexp, visitor func(*syntax.Regexp) bool) {
	if !visitor(re) {
		return
	}
	for _, sub := range re.Sub {
		walkRegexp(sub, visitor)
	}
}

func isQuantifier(re *syntax.Regexp) bool {
	switch re.Op {
	case syntax.OpStar, syntax.OpPlus, syntax.OpQuest, syntax.OpRepeat:
		return true
	}
	return false
}

func hasQuantifier(re *syntax.Regexp) bool {
	if isQuantifier(re) {
		return true
	}
	for _, sub := range re.Sub {
		if hasQuantifier(sub) {
			return true
		}
	}
	return false
}

func hasNestedQuantifiers(re *syntax.Regexp) bool {
	result := false
	walkRegexp(re, func(node *syntax.Regexp) bool {
		if isQuantifier(node) && isTrulyNested(node) {
			result = true
			return false
		}
		return true
	})
	return result
}

// isTrulyNested checks if a quantifier contains another quantifier.
// It recursively checks through Concat/Capture to find nested quantifiers.
func isTrulyNested(re *syntax.Regexp) bool {
	if !isQuantifier(re) || len(re.Sub) == 0 {
		return false
	}

	// Recursively check for quantifiers in children
	return containsQuantifierRecursive(re.Sub[0])
}

func containsQuantifierRecursive(re *syntax.Regexp) bool {
	if isQuantifier(re) {
		return true
	}

	// Recurse through Concat and Capture
	if re.Op == syntax.OpConcat || re.Op == syntax.OpCapture {
		for _, sub := range re.Sub {
			if containsQuantifierRecursive(sub) {
				return true
			}
		}
	}

	return false
}

func getQuantifierDepth(re *syntax.Regexp) int {
	if !isQuantifier(re) {
		return 0
	}

	maxSubDepth := 0
	for _, sub := range re.Sub {
		depth := getQuantifierDepthHelper(sub, 0)
		if depth > maxSubDepth {
			maxSubDepth = depth
		}
	}

	return maxSubDepth + 1
}

func getQuantifierDepthHelper(re *syntax.Regexp, current int) int {
	if isQuantifier(re) {
		current++
	}

	maxDepth := current
	for _, sub := range re.Sub {
		depth := getQuantifierDepthHelper(sub, current)
		if depth > maxDepth {
			maxDepth = depth
		}
	}

	return maxDepth
}

func countQuantifiers(re *syntax.Regexp) int {
	count := 0
	walkRegexp(re, func(node *syntax.Regexp) bool {
		if isQuantifier(node) {
			count++
		}
		return true
	})
	return count
}

func findOverlappingQuantifiers(re *syntax.Regexp) []string {
	var sequences []string

	walkRegexp(re, func(node *syntax.Regexp) bool {
		if node.Op == syntax.OpConcat {
			consecutive := 0
			for _, sub := range node.Sub {
				if isQuantifier(sub) {
					consecutive++
					if consecutive >= 2 {
						sequences = append(sequences, node.String())
						break
					}
				} else {
					consecutive = 0
				}
			}
		}
		return true
	})

	return sequences
}

func hasOverlappingBranches(re *syntax.Regexp) bool {
	if re.Op != syntax.OpAlternate || len(re.Sub) < 2 {
		return false
	}

	firstOp := re.Sub[0].Op
	for i := 1; i < len(re.Sub); i++ {
		if re.Sub[i].Op == firstOp {
			return true
		}
	}

	return false
}

func hasDotStar(re *syntax.Regexp) bool {
	result := false
	walkRegexp(re, func(node *syntax.Regexp) bool {
		if node.Op == syntax.OpStar && len(node.Sub) > 0 {
			if node.Sub[0].Op == syntax.OpAnyChar || node.Sub[0].Op == syntax.OpAnyCharNotNL {
				result = true
				return false
			}
		}
		return true
	})
	return result
}
