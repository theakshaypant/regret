// Package detector implements pattern detection logic for identifying dangerous regex patterns.
package detector

import (
	"fmt"
	"regexp/syntax"
	"strings"

	"github.com/theakshaypant/regret/internal/parser"
)

// ValidationMode represents the depth of analysis.
type ValidationMode int

const (
	Fast ValidationMode = iota
	Balanced
	Thorough
)

// Options contains configuration for detection.
type Options struct {
	Mode   ValidationMode
	Checks uint32
}

// Issue represents a detected problem.
type Issue struct {
	Type       string
	Severity   string
	Position   Position
	Pattern    string
	Message    string
	Example    string
	Suggestion string
	Complexity int
}

// Position represents a location in the pattern.
type Position struct {
	Start  int
	End    int
	Line   int
	Column int
}

// Detector performs pattern detection based on configured options.
type Detector struct {
	opts        *Options
	parser      *parser.Parser
	nfaAnalyzer *NFAAnalyzer
}

// NewDetector creates a new detector with the given options.
func NewDetector(opts *Options) *Detector {
	return &Detector{
		opts:        opts,
		parser:      parser.NewParser(),
		nfaAnalyzer: NewNFAAnalyzer(),
	}
}

// Detect analyzes a parsed regex and returns detected issues.
func (d *Detector) Detect(re *syntax.Regexp, pattern string) ([]Issue, error) {
	var issues []Issue

	// Run checks based on mode and flags
	switch d.opts.Mode {
	case Fast:
		issues = append(issues, d.runFastChecks(re, pattern)...)
	case Balanced:
		issues = append(issues, d.runFastChecks(re, pattern)...)
		issues = append(issues, d.runBalancedChecks(re, pattern)...)
	case Thorough:
		issues = append(issues, d.runFastChecks(re, pattern)...)
		issues = append(issues, d.runBalancedChecks(re, pattern)...)
		issues = append(issues, d.runThoroughChecks(re, pattern)...)
	}

	return issues, nil
}

func (d *Detector) runFastChecks(re *syntax.Regexp, pattern string) []Issue {
	var issues []Issue

	// 1. Pattern length validation
	if len(pattern) > 10000 {
		issues = append(issues, Issue{
			Type:       "pattern_too_long",
			Severity:   "high",
			Position:   Position{Start: 0, End: len(pattern)},
			Pattern:    pattern,
			Message:    fmt.Sprintf("Pattern exceeds maximum length (10000 characters): %d characters", len(pattern)),
			Suggestion: "Consider breaking the pattern into multiple smaller patterns",
		})
	}

	// 2. Nesting depth check
	nestingDepth := parser.GetNestingDepth(re)
	if nestingDepth > 5 {
		issues = append(issues, Issue{
			Type:       "excessive_nesting",
			Severity:   "high",
			Position:   Position{Start: 0, End: len(pattern)},
			Pattern:    pattern,
			Message:    fmt.Sprintf("Excessive quantifier nesting depth: %d (threshold: 5)", nestingDepth),
			Example:    "aaa",
			Suggestion: "Reduce nesting depth by simplifying quantifiers",
			Complexity: nestingDepth * 15, // Rough complexity estimate
		})
	}

	// 3. Quantifier count check
	quantifierCount := parser.CountQuantifiers(re)
	if quantifierCount > 20 {
		issues = append(issues, Issue{
			Type:       "too_many_quantifiers",
			Severity:   "medium",
			Position:   Position{Start: 0, End: len(pattern)},
			Pattern:    pattern,
			Message:    fmt.Sprintf("Excessive quantifiers: %d (threshold: 20)", quantifierCount),
			Suggestion: "Simplify the pattern to reduce quantifier count",
			Complexity: quantifierCount * 3,
		})
	}

	// 4. Nested quantifier detection (most dangerous)
	nestedIssues := d.detectNestedQuantifiers(re, pattern)
	issues = append(issues, nestedIssues...)

	// 5. Overlapping alternation detection
	alternationIssues := d.detectOverlappingAlternations(re, pattern)
	issues = append(issues, alternationIssues...)

	// 6. Dangerous pattern combinations
	dangerousIssues := d.detectDangerousPatterns(re, pattern)
	issues = append(issues, dangerousIssues...)

	return issues
}

func (d *Detector) runBalancedChecks(re *syntax.Regexp, pattern string) []Issue {
	// Run NFA-based EDA/IDA detection
	issues, err := d.nfaAnalyzer.AnalyzePattern(re, pattern)
	if err != nil {
		// If NFA analysis fails, return empty (fall back to fast checks)
		return []Issue{}
	}

	return issues
}

func (d *Detector) runThoroughChecks(re *syntax.Regexp, pattern string) []Issue {
	// TODO: Implement adversarial testing (Phase 3)
	return []Issue{}
}

// detectNestedQuantifiers finds patterns like (a+)+, (a*)*, (a?)+
func (d *Detector) detectNestedQuantifiers(re *syntax.Regexp, pattern string) []Issue {
	var issues []Issue

	parser.Walk(re, func(node *syntax.Regexp) bool {
		if !parser.IsQuantifier(node) {
			return true
		}

		// Check if this quantifier has a child quantifier
		if len(node.Sub) > 0 {
			for _, sub := range node.Sub {
				if parser.HasQuantifier(sub) {
					issues = append(issues, Issue{
						Type:       "nested_quantifiers",
						Severity:   "critical",
						Position:   Position{Start: 0, End: len(pattern)},
						Pattern:    node.String(),
						Message:    fmt.Sprintf("Nested quantifiers detected: %s", node.String()),
						Example:    generateNestedQuantifierExample(node),
						Suggestion: "Remove nesting: simplify to a single quantifier",
						Complexity: 90, // Very high complexity
					})
				}
			}
		}

		return true
	})

	return issues
}

// detectOverlappingAlternations finds patterns like (a|ab)+, (a|a)*
func (d *Detector) detectOverlappingAlternations(re *syntax.Regexp, pattern string) []Issue {
	var issues []Issue

	parser.Walk(re, func(node *syntax.Regexp) bool {
		if !parser.IsAlternation(node) {
			return true
		}

		// Check if alternation branches can match the same prefix
		if len(node.Sub) >= 2 {
			for i := 0; i < len(node.Sub); i++ {
				for j := i + 1; j < len(node.Sub); j++ {
					if branchesOverlap(node.Sub[i], node.Sub[j]) {
						issues = append(issues, Issue{
							Type:       "overlapping_alternation",
							Severity:   "high",
							Position:   Position{Start: 0, End: len(pattern)},
							Pattern:    node.String(),
							Message:    fmt.Sprintf("Overlapping alternation branches: %s", node.String()),
							Example:    "ababababx",
							Suggestion: "Reorder branches or use atomic grouping",
							Complexity: 70,
						})
						break
					}
				}
			}
		}

		return true
	})

	return issues
}

// detectDangerousPatterns finds other dangerous combinations
func (d *Detector) detectDangerousPatterns(re *syntax.Regexp, pattern string) []Issue {
	var issues []Issue

	// Pattern 1: Multiple overlapping quantifiers like a*a*
	if strings.Contains(pattern, "*.*") || strings.Contains(pattern, "+.+") {
		issues = append(issues, Issue{
			Type:       "polynomial_backtracking",
			Severity:   "high",
			Position:   Position{Start: 0, End: len(pattern)},
			Pattern:    pattern,
			Message:    "Overlapping unbounded quantifiers detected",
			Example:    "aaaaaaaax",
			Suggestion: "Use possessive quantifiers or atomic grouping",
			Complexity: 60,
		})
	}

	// Pattern 2: Greedy quantifier followed by similar pattern
	// Look for a*a+, a+a*, etc.
	dangerousPatterns := []string{
		"a*a+", "a+a*", "a*a*",
		"\\d*\\d+", "\\d+\\d*", "\\d*\\d*",
		"\\w*\\w+", "\\w+\\w*", "\\w*\\w*",
		".*.", ".+.", ".*.*",
	}

	for _, dp := range dangerousPatterns {
		if strings.Contains(pattern, dp) {
			issues = append(issues, Issue{
				Type:       "polynomial_backtracking",
				Severity:   "high",
				Position:   Position{Start: 0, End: len(pattern)},
				Pattern:    dp,
				Message:    fmt.Sprintf("Potentially dangerous pattern detected: %s", dp),
				Example:    "aaaaaaax",
				Suggestion: "Consolidate or reorder quantifiers",
				Complexity: 65,
			})
		}
	}

	return issues
}

// Helper function to generate example input for nested quantifiers
func generateNestedQuantifierExample(node *syntax.Regexp) string {
	// For patterns like (a+)+, generate aaaaaaa
	// The actual adversarial input would be aaaa...x (where x doesn't match)
	switch node.Op {
	case syntax.OpStar, syntax.OpPlus:
		return "aaaaaa"
	case syntax.OpQuest:
		return "a"
	default:
		return "test"
	}
}

// Helper function to check if two alternation branches can overlap
func branchesOverlap(a, b *syntax.Regexp) bool {
	// Unwrap captures to get to the actual content
	for a.Op == syntax.OpCapture && len(a.Sub) > 0 {
		a = a.Sub[0]
	}
	for b.Op == syntax.OpCapture && len(b.Sub) > 0 {
		b = b.Sub[0]
	}

	// Simple heuristic: check if one branch is a prefix of another
	aStr := a.String()
	bStr := b.String()

	// Check string prefixes
	if len(aStr) > 0 && len(bStr) > 0 {
		if strings.HasPrefix(aStr, bStr) || strings.HasPrefix(bStr, aStr) {
			return true
		}
	}

	// Check if both branches start with the same literal character
	if a.Op == syntax.OpLiteral && b.Op == syntax.OpLiteral {
		if len(a.Rune) > 0 && len(b.Rune) > 0 && a.Rune[0] == b.Rune[0] {
			return true
		}
	}

	// Check if both start with concat and their first elements overlap
	if a.Op == syntax.OpConcat && b.Op == syntax.OpConcat {
		if len(a.Sub) > 0 && len(b.Sub) > 0 {
			return branchesOverlap(a.Sub[0], b.Sub[0])
		}
	}

	// Check if one is concat and other is literal - compare first element
	if a.Op == syntax.OpConcat && b.Op == syntax.OpLiteral {
		if len(a.Sub) > 0 {
			return branchesOverlap(a.Sub[0], b)
		}
	}
	if a.Op == syntax.OpLiteral && b.Op == syntax.OpConcat {
		if len(b.Sub) > 0 {
			return branchesOverlap(a, b.Sub[0])
		}
	}

	// Check if both use wildcards or character classes
	if (a.Op == syntax.OpAnyChar || a.Op == syntax.OpCharClass) &&
		(b.Op == syntax.OpAnyChar || b.Op == syntax.OpCharClass) {
		return true
	}

	return false
}
