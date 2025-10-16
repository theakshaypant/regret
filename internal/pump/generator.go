// Package pump provides adversarial input generation for regex testing.
package pump

import (
	"fmt"
	"regexp/syntax"
	"strings"
)

// Options contains configuration for pump pattern generation.
type Options struct {
	PumpSize       int  // Size of pumped component (default: 10)
	MaxPumpSize    int  // Maximum pump size (default: 100)
	IncludeFailure bool // Include failing suffix (default: true)
}

// PumpPattern represents an adversarial input pattern.
type PumpPattern struct {
	BaseString    string // The base pattern to repeat
	PumpComponent string // The component to pump (repeat)
	FailSuffix    string // Suffix that causes failure
	Description   string // Description of why this triggers backtracking
	Sizes         []int  // Suggested pump sizes to test
}

// Generator generates adversarial inputs for regex patterns.
type Generator struct {
	opts *Options
}

// NewGenerator creates a new pump pattern generator.
func NewGenerator(opts *Options) *Generator {
	if opts == nil {
		opts = &Options{
			PumpSize:       10,
			MaxPumpSize:    100,
			IncludeFailure: true,
		}
	}
	return &Generator{opts: opts}
}

// Generate creates pump patterns for a given regex.
func (g *Generator) Generate(re *syntax.Regexp, pattern string) ([]PumpPattern, error) {
	var patterns []PumpPattern

	// Detect nested quantifiers
	if hasNestedQuantifiers(re) {
		patterns = append(patterns, g.generateNestedQuantifierPump(re))
	}

	// Detect overlapping quantifiers
	if hasOverlappingQuantifiers(re) {
		patterns = append(patterns, g.generateOverlappingQuantifierPump(re))
	}

	// Detect overlapping alternations
	if hasOverlappingAlternation(re) {
		patterns = append(patterns, g.generateAlternationPump(re))
	}

	// If no specific patterns found, generate generic pump
	if len(patterns) == 0 {
		patterns = append(patterns, g.generateGenericPump(re))
	}

	return patterns, nil
}

// generateNestedQuantifierPump generates pump for patterns like (a+)+.
func (g *Generator) generateNestedQuantifierPump(re *syntax.Regexp) PumpPattern {
	// For (a+)+, generate aaaaaa...x where x doesn't match
	// The pump component is 'a', which gets repeated

	baseChar := extractPumpChar(re)

	return PumpPattern{
		BaseString:    "",
		PumpComponent: baseChar,
		FailSuffix:    "x",
		Description:   "Nested quantifiers cause exponential backtracking. Each 'a' doubles the number of ways to match.",
		Sizes:         []int{5, 10, 15, 20, 25},
	}
}

// generateOverlappingQuantifierPump generates pump for patterns like a*a*.
func (g *Generator) generateOverlappingQuantifierPump(re *syntax.Regexp) PumpPattern {
	// For a*a*, generate aaaaaa...x
	// The pump component is 'a'

	baseChar := extractPumpChar(re)

	return PumpPattern{
		BaseString:    "",
		PumpComponent: baseChar,
		FailSuffix:    "x",
		Description:   "Overlapping quantifiers cause polynomial backtracking. Regex tries all ways to split input between quantifiers.",
		Sizes:         []int{10, 20, 30, 40, 50},
	}
}

// generateAlternationPump generates pump for patterns like (a|ab)+.
func (g *Generator) generateAlternationPump(re *syntax.Regexp) PumpPattern {
	// For (a|ab)+, generate ababab...x
	// This forces backtracking between the alternation branches

	return PumpPattern{
		BaseString:    "",
		PumpComponent: "ab",
		FailSuffix:    "x",
		Description:   "Overlapping alternation branches cause backtracking. Regex tries each branch at each position.",
		Sizes:         []int{5, 10, 15, 20},
	}
}

// generateGenericPump generates a generic pump pattern.
func (g *Generator) generateGenericPump(re *syntax.Regexp) PumpPattern {
	baseChar := extractPumpChar(re)
	if baseChar == "" {
		baseChar = "a"
	}

	return PumpPattern{
		BaseString:    "",
		PumpComponent: baseChar,
		FailSuffix:    "x",
		Description:   "Generic pump pattern to test regex performance",
		Sizes:         []int{10, 50, 100},
	}
}

// GenerateInput generates an actual test input from a pump pattern.
func (p *PumpPattern) GenerateInput(size int) string {
	var builder strings.Builder

	builder.WriteString(p.BaseString)

	for i := 0; i < size; i++ {
		builder.WriteString(p.PumpComponent)
	}

	builder.WriteString(p.FailSuffix)

	return builder.String()
}

// GenerateSequence generates a sequence of test inputs with increasing sizes.
func (p *PumpPattern) GenerateSequence() []string {
	inputs := make([]string, len(p.Sizes))
	for i, size := range p.Sizes {
		inputs[i] = p.GenerateInput(size)
	}
	return inputs
}

// Helper functions

func hasNestedQuantifiers(re *syntax.Regexp) bool {
	result := false
	walk(re, func(node *syntax.Regexp) bool {
		if isQuantifier(node) {
			for _, sub := range node.Sub {
				if containsQuantifier(sub) {
					result = true
					return false
				}
			}
		}
		return true
	})
	return result
}

func hasOverlappingQuantifiers(re *syntax.Regexp) bool {
	result := false
	walk(re, func(node *syntax.Regexp) bool {
		if node.Op == syntax.OpConcat {
			consecutive := 0
			for _, sub := range node.Sub {
				if isQuantifier(sub) {
					consecutive++
					if consecutive >= 2 {
						result = true
						return false
					}
				} else {
					consecutive = 0
				}
			}
		}
		return true
	})
	return result
}

func hasOverlappingAlternation(re *syntax.Regexp) bool {
	result := false
	walk(re, func(node *syntax.Regexp) bool {
		if node.Op == syntax.OpAlternate && len(node.Sub) >= 2 {
			// Simple check: if any branches share a prefix
			for i := 0; i < len(node.Sub); i++ {
				for j := i + 1; j < len(node.Sub); j++ {
					if branchesOverlap(node.Sub[i], node.Sub[j]) {
						result = true
						return false
					}
				}
			}
		}
		return true
	})
	return result
}

func branchesOverlap(a, b *syntax.Regexp) bool {
	// Simple heuristic
	aStr := a.String()
	bStr := b.String()

	if len(aStr) > 0 && len(bStr) > 0 {
		return strings.HasPrefix(aStr, bStr) || strings.HasPrefix(bStr, aStr)
	}

	return false
}

func extractPumpChar(re *syntax.Regexp) string {
	// Try to extract a character that can be pumped
	var result string

	walk(re, func(node *syntax.Regexp) bool {
		if node.Op == syntax.OpLiteral && len(node.Rune) > 0 {
			result = string(node.Rune[0])
			return false
		}
		if node.Op == syntax.OpCharClass {
			// Use 'a' for character classes
			result = "a"
			return false
		}
		if node.Op == syntax.OpAnyChar || node.Op == syntax.OpAnyCharNotNL {
			result = "a"
			return false
		}
		return true
	})

	if result == "" {
		result = "a"
	}

	return result
}

func walk(re *syntax.Regexp, visitor func(*syntax.Regexp) bool) {
	if !visitor(re) {
		return
	}
	for _, sub := range re.Sub {
		walk(sub, visitor)
	}
}

func isQuantifier(re *syntax.Regexp) bool {
	switch re.Op {
	case syntax.OpStar, syntax.OpPlus, syntax.OpQuest, syntax.OpRepeat:
		return true
	}
	return false
}

func containsQuantifier(re *syntax.Regexp) bool {
	if isQuantifier(re) {
		return true
	}
	for _, sub := range re.Sub {
		if containsQuantifier(sub) {
			return true
		}
	}
	return false
}

// String returns a string representation of the pump pattern.
func (p *PumpPattern) String() string {
	return fmt.Sprintf("Pump{component=%q, fail=%q, sizes=%v}",
		p.PumpComponent, p.FailSuffix, p.Sizes)
}
