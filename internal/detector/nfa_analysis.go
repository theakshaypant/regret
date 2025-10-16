// Package detector implements NFA-based analysis for detecting EDA and IDA.
package detector

import (
	"regexp/syntax"

	"github.com/theakshaypant/regret/internal/parser"
)

// NFAAnalyzer performs NFA-based analysis for EDA/IDA detection.
type NFAAnalyzer struct {
	nfa    *parser.NFA
	parser *parser.Parser
}

// NewNFAAnalyzer creates a new NFA analyzer.
func NewNFAAnalyzer() *NFAAnalyzer {
	return &NFAAnalyzer{
		parser: parser.NewParser(),
	}
}

// AnalyzePattern analyzes a regex pattern using NFA-based methods.
func (a *NFAAnalyzer) AnalyzePattern(re *syntax.Regexp, pattern string) ([]Issue, error) {
	// Build NFA from regex
	nfa, err := parser.BuildNFA(re)
	if err != nil {
		return nil, err
	}

	a.nfa = nfa

	var issues []Issue

	// Run EDA detection
	edaIssues := a.detectEDA(re, pattern)
	issues = append(issues, edaIssues...)

	// Run IDA detection
	idaIssues := a.detectIDA(re, pattern)
	issues = append(issues, idaIssues...)

	return issues, nil
}

// detectEDA detects Exponential Degree of Ambiguity.
// This occurs when patterns have multiple paths that can match the same input,
// and the number of paths grows exponentially with input length.
func (a *NFAAnalyzer) detectEDA(re *syntax.Regexp, pattern string) []Issue {
	var issues []Issue

	// EDA detection strategy:
	// 1. Find states with multiple epsilon paths (ambiguity sources)
	// 2. Check if ambiguity is nested (quantifiers within quantifiers)
	// 3. Check for overlapping alternations inside quantifiers

	ambiguousStates := a.findAmbiguousStates()

	for _, state := range ambiguousStates {
		// Check if this ambiguity is in a loop (quantifier)
		if a.isInQuantifierLoop(state) {
			issues = append(issues, Issue{
				Type:       "exponential_backtracking",
				Severity:   "critical",
				Position:   Position{Start: 0, End: len(pattern)},
				Pattern:    pattern,
				Message:    "Exponential ambiguity detected: multiple paths through quantifier",
				Example:    a.generateEDAExample(state),
				Suggestion: "Remove nested quantifiers or use atomic grouping",
				Complexity: 95,
			})
		}
	}

	// Additional EDA check: nested quantifiers via AST
	// This catches patterns that might be missed by pure NFA analysis
	nestedQuantifiers := a.findNestedQuantifiersInNFA(re)
	if len(nestedQuantifiers) > 0 {
		issues = append(issues, Issue{
			Type:       "exponential_backtracking",
			Severity:   "critical",
			Position:   Position{Start: 0, End: len(pattern)},
			Pattern:    pattern,
			Message:    "Nested quantifiers create exponential ambiguity",
			Example:    "aaaaaaaax",
			Suggestion: "Simplify quantifier nesting",
			Complexity: 95,
		})
	}

	return issues
}

// detectIDA detects Infinite Degree of Ambiguity (polynomial).
// This occurs when multiple quantifiers can match overlapping input,
// causing polynomial time complexity.
func (a *NFAAnalyzer) detectIDA(re *syntax.Regexp, pattern string) []Issue {
	var issues []Issue

	// IDA detection strategy:
	// 1. Find sequences of quantifiers that can match same character class
	// 2. Count the degree (number of overlapping quantifiers)
	// 3. Estimate polynomial degree

	overlappingSequences := a.findOverlappingQuantifierSequences(re)

	for _, seq := range overlappingSequences {
		degree := len(seq)
		if degree >= 2 {
			complexity := 50 + (degree * 10) // Base 50, +10 per degree
			if complexity > 90 {
				complexity = 90
			}

			complexityStr := "O(n²)"
			if degree == 3 {
				complexityStr = "O(n³)"
			} else if degree > 3 {
				complexityStr = "O(n^k)"
			}

			issues = append(issues, Issue{
				Type:       "polynomial_backtracking",
				Severity:   "high",
				Position:   Position{Start: 0, End: len(pattern)},
				Pattern:    pattern,
				Message:    "Polynomial ambiguity detected: " + complexityStr,
				Example:    "aaaaaaax",
				Suggestion: "Consolidate overlapping quantifiers or use possessive quantifiers",
				Complexity: complexity,
			})
		}
	}

	return issues
}

// findAmbiguousStates finds states that can be reached via multiple paths.
func (a *NFAAnalyzer) findAmbiguousStates() []*parser.State {
	var ambiguous []*parser.State

	for _, state := range a.nfa.States {
		// Count distinct paths to this state
		pathCount := a.countPathsToState(state)
		if pathCount > 1 {
			ambiguous = append(ambiguous, state)
		}
	}

	return ambiguous
}

// countPathsToState counts distinct paths from start to given state.
func (a *NFAAnalyzer) countPathsToState(target *parser.State) int {
	// Simplified path counting (exact counting is expensive)
	// We use epsilon closure size as a proxy for ambiguity
	closure := parser.ComputeEpsilonClosure(target)

	// If state is reachable via many epsilon transitions, it's likely ambiguous
	if len(closure) > 3 {
		return len(closure)
	}

	return 1
}

// isInQuantifierLoop checks if a state is part of a quantifier loop.
func (a *NFAAnalyzer) isInQuantifierLoop(state *parser.State) bool {
	// Check if state has epsilon transitions that form a cycle
	visited := make(map[*parser.State]bool)
	return a.hasCycle(state, visited)
}

// hasCycle checks if there's a cycle reachable from the given state.
func (a *NFAAnalyzer) hasCycle(state *parser.State, visited map[*parser.State]bool) bool {
	if visited[state] {
		return true // Found a cycle
	}

	visited[state] = true

	for _, next := range state.EpsilonTo {
		if a.hasCycle(next, visited) {
			return true
		}
	}

	delete(visited, state) // Backtrack
	return false
}

// findNestedQuantifiersInNFA finds nested quantifiers using AST traversal.
func (a *NFAAnalyzer) findNestedQuantifiersInNFA(re *syntax.Regexp) []string {
	var nested []string

	parser.Walk(re, func(node *syntax.Regexp) bool {
		if !parser.IsQuantifier(node) {
			return true
		}

		// Check if this quantifier contains another quantifier
		for _, sub := range node.Sub {
			if parser.HasQuantifier(sub) {
				nested = append(nested, node.String())
				break
			}
		}

		return true
	})

	return nested
}

// findOverlappingQuantifierSequences finds sequences of quantifiers that can match overlapping input.
func (a *NFAAnalyzer) findOverlappingQuantifierSequences(re *syntax.Regexp) [][]string {
	var sequences [][]string
	var currentSeq []string

	// Walk the AST looking for consecutive quantifiers
	parser.Walk(re, func(node *syntax.Regexp) bool {
		if node.Op == syntax.OpConcat {
			// Check children for quantifier sequences
			currentSeq = []string{}
			for _, sub := range node.Sub {
				if parser.IsQuantifier(sub) {
					// Check if this quantifier overlaps with previous
					if len(currentSeq) > 0 && a.quantifiersCanOverlap(sub, sub) {
						currentSeq = append(currentSeq, sub.String())
					} else if len(currentSeq) == 0 {
						currentSeq = append(currentSeq, sub.String())
					} else {
						// End of sequence
						if len(currentSeq) >= 2 {
							sequences = append(sequences, currentSeq)
						}
						currentSeq = []string{sub.String()}
					}
				} else {
					// Non-quantifier breaks the sequence
					if len(currentSeq) >= 2 {
						sequences = append(sequences, currentSeq)
					}
					currentSeq = []string{}
				}
			}

			// Add final sequence if valid
			if len(currentSeq) >= 2 {
				sequences = append(sequences, currentSeq)
			}
		}

		return true
	})

	return sequences
}

// quantifiersCanOverlap checks if two quantifiers can match overlapping character sets.
func (a *NFAAnalyzer) quantifiersCanOverlap(q1, q2 *syntax.Regexp) bool {
	// Simplified check: assume quantifiers can overlap if they're both present
	// A more sophisticated check would compare character classes
	return true
}

// generateEDAExample generates an example input that triggers EDA.
func (a *NFAAnalyzer) generateEDAExample(state *parser.State) string {
	// Generate string that would cause exponential backtracking
	return "aaaaaaaaaaaax"
}

// ComputeAmbiguityDegree estimates the degree of ambiguity for a pattern.
// Returns (degree, isExponential).
func (a *NFAAnalyzer) ComputeAmbiguityDegree(re *syntax.Regexp) (int, bool) {
	// Check for exponential ambiguity (nested quantifiers)
	nested := a.findNestedQuantifiersInNFA(re)
	if len(nested) > 0 {
		return len(nested), true
	}

	// Check for polynomial ambiguity (overlapping quantifiers)
	sequences := a.findOverlappingQuantifierSequences(re)
	maxDegree := 0
	for _, seq := range sequences {
		if len(seq) > maxDegree {
			maxDegree = len(seq)
		}
	}

	if maxDegree >= 2 {
		return maxDegree, false
	}

	return 1, false
}
