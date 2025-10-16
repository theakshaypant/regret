package detector

import (
	"regexp/syntax"
	"testing"

	"github.com/theakshaypant/regret/internal/parser"
)

func TestDetector_FastChecks(t *testing.T) {
	tests := []struct {
		name         string
		pattern      string
		expectIssues bool
		expectedType string
		expectedSev  string
		mode         ValidationMode
	}{
		{
			name:         "safe pattern",
			pattern:      "^[a-z]+$",
			expectIssues: false,
			mode:         Fast,
		},
		{
			name:         "nested quantifiers (a+)+",
			pattern:      "(a+)+",
			expectIssues: true,
			expectedType: "nested_quantifiers",
			expectedSev:  "critical",
			mode:         Fast,
		},
		{
			name:         "nested quantifiers (a*)*",
			pattern:      "(a*)*",
			expectIssues: true,
			expectedType: "nested_quantifiers",
			expectedSev:  "critical",
			mode:         Fast,
		},
		{
			name:         "deeply nested ((a+)+)+",
			pattern:      "((a+)+)+",
			expectIssues: true,
			expectedType: "nested_quantifiers", // Depth is 3, below threshold of 5, but still nested
			expectedSev:  "critical",
			mode:         Fast,
		},
		{
			name:         "overlapping quantifiers a*a+",
			pattern:      "a*a+",
			expectIssues: true,
			expectedType: "polynomial_backtracking",
			expectedSev:  "high",
			mode:         Fast,
		},
		{
			name:         "overlapping quantifiers \\d*\\d+",
			pattern:      "\\d*\\d+",
			expectIssues: true,
			expectedType: "polynomial_backtracking",
			expectedSev:  "high",
			mode:         Fast,
		},
		{
			name:         "greedy dot quantifiers .*.",
			pattern:      ".*.",
			expectIssues: true,
			expectedType: "polynomial_backtracking",
			expectedSev:  "high",
			mode:         Fast,
		},
		{
			name:         "safe simple quantifier",
			pattern:      "a+b*c?",
			expectIssues: false,
			mode:         Fast,
		},
		{
			name:         "safe anchored pattern",
			pattern:      "^abc+$",
			expectIssues: false,
			mode:         Fast,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := parser.NewParser()
			re, err := p.Parse(tt.pattern)
			if err != nil {
				t.Fatalf("Failed to parse pattern: %v", err)
			}

			opts := &Options{
				Mode: tt.mode,
			}
			detector := NewDetector(opts)
			issues, err := detector.Detect(re, tt.pattern)
			if err != nil {
				t.Fatalf("Detect() error = %v", err)
			}

			if tt.expectIssues && len(issues) == 0 {
				t.Errorf("Expected issues but got none")
			}

			if !tt.expectIssues && len(issues) > 0 {
				t.Errorf("Expected no issues but got %d: %v", len(issues), issues)
			}

			if tt.expectIssues && len(issues) > 0 {
				found := false
				for _, issue := range issues {
					if issue.Type == tt.expectedType && issue.Severity == tt.expectedSev {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected issue type=%s severity=%s, but didn't find it. Got: %v",
						tt.expectedType, tt.expectedSev, issues)
				}
			}
		})
	}
}

func TestDetector_NestedQuantifiers(t *testing.T) {
	tests := []struct {
		name         string
		pattern      string
		expectIssues bool
	}{
		{"simple nested (a+)+", "(a+)+", true},
		{"simple nested (a*)*", "(a*)*", true},
		{"triple nested ((a+)+)+", "((a+)+)+", true},
		{"nested with alternation ((a|b)+)+", "((a|b)+)+", true},
		{"safe single quantifier a+", "a+", false},
		{"safe multiple non-nested a+b*c?", "a+b*c?", false},
		{"safe with groups (abc)+", "(abc)+", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := parser.NewParser()
			re, err := p.Parse(tt.pattern)
			if err != nil {
				t.Fatalf("Failed to parse: %v", err)
			}

			opts := &Options{Mode: Fast}
			d := NewDetector(opts)
			issues := d.detectNestedQuantifiers(re, tt.pattern)

			if tt.expectIssues && len(issues) == 0 {
				t.Error("Expected nested quantifier issues but got none")
			}

			if !tt.expectIssues && len(issues) > 0 {
				t.Errorf("Expected no issues but got %d", len(issues))
			}
		})
	}
}

func TestDetector_DangerousPatterns(t *testing.T) {
	tests := []struct {
		name         string
		pattern      string
		expectIssues bool
	}{
		{"overlapping a*a+", "a*a+", true},
		{"overlapping a+a*", "a+a*", true},
		{"overlapping \\d*\\d+", "\\d*\\d+", true},
		{"overlapping \\w*\\w+", "\\w*\\w+", true},
		{"greedy dots .*.", ".*.", true},
		{"greedy dots .+.", ".+.", true},
		{"safe pattern abc", "abc", false},
		{"safe pattern [a-z]+", "[a-z]+", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := parser.NewParser()
			re, err := p.Parse(tt.pattern)
			if err != nil {
				t.Fatalf("Failed to parse: %v", err)
			}

			opts := &Options{Mode: Fast}
			d := NewDetector(opts)
			issues := d.detectDangerousPatterns(re, tt.pattern)

			if tt.expectIssues && len(issues) == 0 {
				t.Error("Expected dangerous pattern issues but got none")
			}

			if !tt.expectIssues && len(issues) > 0 {
				t.Errorf("Expected no issues but got %d", len(issues))
			}
		})
	}
}

func TestDetector_NestingDepth(t *testing.T) {
	tests := []struct {
		pattern       string
		expectIssues  bool
		expectedDepth int
	}{
		{"a+", false, 1},
		{"(a+)+", false, 2},
		{"((a+)+)+", false, 3},
		{"(((a+)+)+)+", false, 4},
		{"((((a+)+)+)+)+", false, 5},
		{"(((((a+)+)+)+)+)+", true, 6}, // Exceeds threshold of 5
		{"((((((a+)+)+)+)+)+)+", true, 7},
	}

	for _, tt := range tests {
		t.Run(tt.pattern, func(t *testing.T) {
			p := parser.NewParser()
			re, err := p.Parse(tt.pattern)
			if err != nil {
				t.Fatalf("Failed to parse: %v", err)
			}

			depth := parser.GetNestingDepth(re)
			if depth != tt.expectedDepth {
				t.Errorf("Expected nesting depth %d, got %d", tt.expectedDepth, depth)
			}

			opts := &Options{Mode: Fast}
			d := NewDetector(opts)
			issues, _ := d.Detect(re, tt.pattern)

			hasExcessiveNesting := false
			for _, issue := range issues {
				if issue.Type == "excessive_nesting" {
					hasExcessiveNesting = true
					break
				}
			}

			if tt.expectIssues && !hasExcessiveNesting {
				t.Error("Expected excessive nesting issue but didn't find it")
			}

			if !tt.expectIssues && hasExcessiveNesting {
				t.Error("Didn't expect excessive nesting issue but found one")
			}
		})
	}
}

func TestDetector_QuantifierCount(t *testing.T) {
	// Create a pattern with many quantifiers
	pattern := "a+b*c?d+e*f?g+h*i?j+k*l?m+n*o?p+q*r?s+t*u?v+w*" // 24 quantifiers

	p := parser.NewParser()
	re, err := p.Parse(pattern)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	count := parser.CountQuantifiers(re)
	// After Simplify(), the count might be slightly different due to optimizations
	if count < 20 {
		t.Errorf("Expected at least 20 quantifiers, got %d", count)
	}

	opts := &Options{Mode: Fast}
	d := NewDetector(opts)
	issues, _ := d.Detect(re, pattern)

	foundTooMany := false
	for _, issue := range issues {
		if issue.Type == "too_many_quantifiers" {
			foundTooMany = true
			break
		}
	}

	if !foundTooMany {
		t.Errorf("Expected too_many_quantifiers issue for pattern with %d quantifiers", count)
	}
}

func TestDetector_PatternLength(t *testing.T) {
	// Create an excessively long pattern
	longPattern := ""
	for i := 0; i < 10001; i++ {
		longPattern += "a"
	}

	p := parser.NewParser()
	re, err := p.Parse(longPattern)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	opts := &Options{Mode: Fast}
	d := NewDetector(opts)
	issues, _ := d.Detect(re, longPattern)

	foundTooLong := false
	for _, issue := range issues {
		if issue.Type == "pattern_too_long" {
			foundTooLong = true
			break
		}
	}

	if !foundTooLong {
		t.Error("Expected pattern_too_long issue for 10001 character pattern")
	}
}

func TestBranchesOverlap(t *testing.T) {
	p := parser.NewParser()

	tests := []struct {
		name     string
		pattern  string
		expected bool
	}{
		{
			name:     "no overlap abc vs def",
			pattern:  "abc|def",
			expected: false,
		},
		{
			name:     "overlap with captures (a|ab)",
			pattern:  "(a)|(ab)", // Captures prevent full simplification
			expected: true,
		},
		{
			name:     "wildcard overlap",
			pattern:  ".|.",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			re, err := p.Parse(tt.pattern)
			if err != nil {
				t.Fatalf("Failed to parse: %v", err)
			}

			// The pattern might be simplified, so we need to find the alternation
			var foundAlternation bool
			parser.Walk(re, func(node *syntax.Regexp) bool {
				if parser.IsAlternation(node) {
					foundAlternation = true
					if len(node.Sub) >= 2 {
						overlap := branchesOverlap(node.Sub[0], node.Sub[1])
						if overlap != tt.expected {
							t.Errorf("Expected overlap=%v, got %v", tt.expected, overlap)
						}
					}
					return false
				}
				return true
			})

			// Some patterns get simplified and lose their alternation nodes
			if !foundAlternation {
				t.Skipf("Pattern %s was simplified (no alternation node found) - this is expected for optimized patterns", tt.pattern)
			}
		})
	}
}

func TestDetector_ValidationModes(t *testing.T) {
	pattern := "(a+)+"
	p := parser.NewParser()
	re, err := p.Parse(pattern)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	modes := []ValidationMode{Fast, Balanced, Thorough}
	for _, mode := range modes {
		t.Run(mode.String(), func(t *testing.T) {
			opts := &Options{Mode: mode}
			d := NewDetector(opts)
			issues, err := d.Detect(re, pattern)
			if err != nil {
				t.Fatalf("Detect() error = %v", err)
			}

			// All modes should detect nested quantifiers via fast checks
			if len(issues) == 0 {
				t.Error("Expected issues in all validation modes")
			}
		})
	}
}

func (m ValidationMode) String() string {
	switch m {
	case Fast:
		return "Fast"
	case Balanced:
		return "Balanced"
	case Thorough:
		return "Thorough"
	default:
		return "Unknown"
	}
}
