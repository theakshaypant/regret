// Package parser provides regex pattern parsing and AST manipulation.
package parser

import (
	"errors"
	"fmt"
	"regexp/syntax"
)

var (
	// ErrInvalidPattern indicates the pattern is syntactically invalid.
	ErrInvalidPattern = errors.New("invalid regex pattern")
)

// Parser wraps Go's regexp/syntax parser and provides additional utilities.
type Parser struct {
	flags syntax.Flags
}

// NewParser creates a new parser with default flags.
func NewParser() *Parser {
	return &Parser{
		flags: syntax.Perl, // Use Perl syntax (most common)
	}
}

// NewParserWithFlags creates a new parser with custom syntax flags.
func NewParserWithFlags(flags syntax.Flags) *Parser {
	return &Parser{flags: flags}
}

// Parse parses a regex pattern into an AST.
func (p *Parser) Parse(pattern string) (*syntax.Regexp, error) {
	re, err := syntax.Parse(pattern, p.flags)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidPattern, err)
	}

	// Simplify the regex AST
	re = re.Simplify()

	return re, nil
}

// MustParse is like Parse but panics on error. Useful for testing.
func (p *Parser) MustParse(pattern string) *syntax.Regexp {
	re, err := p.Parse(pattern)
	if err != nil {
		panic(err)
	}
	return re
}

// Validate checks if a pattern is syntactically valid.
func (p *Parser) Validate(pattern string) error {
	_, err := p.Parse(pattern)
	return err
}

// GetOp returns the operation type of a regex node.
func GetOp(re *syntax.Regexp) syntax.Op {
	return re.Op
}

// IsQuantifier returns true if the node is a quantifier.
func IsQuantifier(re *syntax.Regexp) bool {
	switch re.Op {
	case syntax.OpStar, syntax.OpPlus, syntax.OpQuest, syntax.OpRepeat:
		return true
	default:
		return false
	}
}

// IsAlternation returns true if the node is an alternation.
func IsAlternation(re *syntax.Regexp) bool {
	return re.Op == syntax.OpAlternate
}

// IsCapture returns true if the node is a capturing group.
func IsCapture(re *syntax.Regexp) bool {
	return re.Op == syntax.OpCapture
}

// HasQuantifier returns true if the regex contains any quantifiers.
func HasQuantifier(re *syntax.Regexp) bool {
	if IsQuantifier(re) {
		return true
	}
	for _, sub := range re.Sub {
		if HasQuantifier(sub) {
			return true
		}
	}
	return false
}

// CountQuantifiers returns the total number of quantifiers in the regex.
func CountQuantifiers(re *syntax.Regexp) int {
	count := 0
	if IsQuantifier(re) {
		count++
	}
	for _, sub := range re.Sub {
		count += CountQuantifiers(sub)
	}
	return count
}

// GetNestingDepth returns the maximum quantifier nesting depth.
func GetNestingDepth(re *syntax.Regexp) int {
	return getNestingDepthHelper(re, 0)
}

func getNestingDepthHelper(re *syntax.Regexp, currentDepth int) int {
	maxDepth := currentDepth

	if IsQuantifier(re) {
		currentDepth++
		if currentDepth > maxDepth {
			maxDepth = currentDepth
		}
	}

	for _, sub := range re.Sub {
		subDepth := getNestingDepthHelper(sub, currentDepth)
		if subDepth > maxDepth {
			maxDepth = subDepth
		}
	}

	return maxDepth
}

// CountAlternations returns the number of alternation operators.
func CountAlternations(re *syntax.Regexp) int {
	count := 0
	if IsAlternation(re) {
		count++
	}
	for _, sub := range re.Sub {
		count += CountAlternations(sub)
	}
	return count
}

// CountCaptures returns the number of capturing groups.
func CountCaptures(re *syntax.Regexp) int {
	count := 0
	if IsCapture(re) {
		count++
	}
	for _, sub := range re.Sub {
		count += CountCaptures(sub)
	}
	return count
}

// Walk traverses the regex AST and calls the visitor function for each node.
func Walk(re *syntax.Regexp, visitor func(*syntax.Regexp) bool) {
	if !visitor(re) {
		return
	}
	for _, sub := range re.Sub {
		Walk(sub, visitor)
	}
}

// FindQuantifiers finds all quantifier nodes in the regex.
func FindQuantifiers(re *syntax.Regexp) []*syntax.Regexp {
	var quantifiers []*syntax.Regexp
	Walk(re, func(node *syntax.Regexp) bool {
		if IsQuantifier(node) {
			quantifiers = append(quantifiers, node)
		}
		return true
	})
	return quantifiers
}

// FindAlternations finds all alternation nodes in the regex.
func FindAlternations(re *syntax.Regexp) []*syntax.Regexp {
	var alternations []*syntax.Regexp
	Walk(re, func(node *syntax.Regexp) bool {
		if IsAlternation(node) {
			alternations = append(alternations, node)
		}
		return true
	})
	return alternations
}

// String returns a string representation of the regex.
func String(re *syntax.Regexp) string {
	return re.String()
}
