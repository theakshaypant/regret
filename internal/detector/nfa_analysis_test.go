package detector

import (
	"testing"

	"github.com/theakshaypant/regret/internal/parser"
)

func TestNFAAnalyzer_AnalyzePattern(t *testing.T) {
	tests := []struct {
		name         string
		pattern      string
		expectIssues bool
		expectedType string
	}{
		{
			name:         "safe pattern",
			pattern:      "^[a-z]+$",
			expectIssues: false,
		},
		{
			name:         "nested quantifiers (a+)+",
			pattern:      "(a+)+",
			expectIssues: true,
			expectedType: "exponential_backtracking",
		},
		{
			name:         "nested quantifiers (a*)*",
			pattern:      "(a*)*",
			expectIssues: true,
			expectedType: "exponential_backtracking",
		},
		{
			name:         "triple nested ((a+)+)+",
			pattern:      "((a+)+)+",
			expectIssues: true,
			expectedType: "exponential_backtracking",
		},
		{
			name:         "simple quantifier a+",
			pattern:      "a+",
			expectIssues: false,
		},
	}

	analyzer := NewNFAAnalyzer()
	p := parser.NewParser()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			re, err := p.Parse(tt.pattern)
			if err != nil {
				t.Fatalf("Parse error: %v", err)
			}

			issues, err := analyzer.AnalyzePattern(re, tt.pattern)
			if err != nil {
				t.Fatalf("AnalyzePattern() error = %v", err)
			}

			hasIssues := len(issues) > 0
			if hasIssues != tt.expectIssues {
				t.Errorf("AnalyzePattern() found %d issues, expectIssues = %v", len(issues), tt.expectIssues)
				if hasIssues {
					for _, issue := range issues {
						t.Logf("  Issue: %s - %s", issue.Type, issue.Message)
					}
				}
			}

			if tt.expectIssues && len(issues) > 0 {
				found := false
				for _, issue := range issues {
					if issue.Type == tt.expectedType {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected issue type %s not found", tt.expectedType)
				}
			}
		})
	}
}

func TestNFAAnalyzer_DetectEDA(t *testing.T) {
	tests := []struct {
		name      string
		pattern   string
		expectEDA bool
	}{
		{"nested star (a*)*", "(a*)*", true},
		{"nested plus (a+)+", "(a+)+", true},
		{"triple nested", "(((a+)+)+)", true},
		{"safe single quantifier", "a+", false},
		{"safe multiple non-nested", "a+b*c?", false},
	}

	analyzer := NewNFAAnalyzer()
	p := parser.NewParser()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			re, err := p.Parse(tt.pattern)
			if err != nil {
				t.Fatalf("Parse error: %v", err)
			}

			// Build NFA before calling detectEDA
			nfa, err := parser.BuildNFA(re)
			if err != nil {
				t.Fatalf("BuildNFA error: %v", err)
			}
			analyzer.nfa = nfa

			issues := analyzer.detectEDA(re, tt.pattern)
			hasEDA := len(issues) > 0

			if hasEDA != tt.expectEDA {
				t.Errorf("detectEDA() found %d issues, expectEDA = %v", len(issues), tt.expectEDA)
			}
		})
	}
}

func TestNFAAnalyzer_DetectIDA(t *testing.T) {
	tests := []struct {
		name      string
		pattern   string
		expectIDA bool
	}{
		{"overlapping a*a*", "a*a*", true}, // Currently detects (will refine in context analysis)
		{"overlapping \\d*\\d+", "\\d*\\d+", true},
		{"consecutive quantifiers a+b+c+", "a+b+c+", true}, // Detects consecutive quantifiers (needs refinement)
	}

	analyzer := NewNFAAnalyzer()
	p := parser.NewParser()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			re, err := p.Parse(tt.pattern)
			if err != nil {
				t.Fatalf("Parse error: %v", err)
			}

			// Build NFA before calling detectIDA
			nfa, err := parser.BuildNFA(re)
			if err != nil {
				t.Fatalf("BuildNFA error: %v", err)
			}
			analyzer.nfa = nfa

			issues := analyzer.detectIDA(re, tt.pattern)
			hasIDA := len(issues) > 0

			if hasIDA != tt.expectIDA {
				t.Logf("Pattern after parse: %s", re.String())
				t.Errorf("detectIDA() found %d issues, expectIDA = %v", len(issues), tt.expectIDA)
			}
		})
	}
}

func TestNFAAnalyzer_ComputeAmbiguityDegree(t *testing.T) {
	tests := []struct {
		name              string
		pattern           string
		expectDegree      int
		expectExponential bool
	}{
		{
			name:              "nested (a+)+",
			pattern:           "(a+)+",
			expectDegree:      1,
			expectExponential: true,
		},
		{
			name:              "triple nested",
			pattern:           "((a+)+)+",
			expectDegree:      2,
			expectExponential: true,
		},
		{
			name:              "safe pattern",
			pattern:           "a+b+",
			expectDegree:      1,
			expectExponential: false,
		},
	}

	analyzer := NewNFAAnalyzer()
	p := parser.NewParser()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			re, err := p.Parse(tt.pattern)
			if err != nil {
				t.Fatalf("Parse error: %v", err)
			}

			degree, isExp := analyzer.ComputeAmbiguityDegree(re)

			if degree < tt.expectDegree {
				t.Errorf("ComputeAmbiguityDegree() degree = %d, want >= %d", degree, tt.expectDegree)
			}

			if isExp != tt.expectExponential {
				t.Errorf("ComputeAmbiguityDegree() exponential = %v, want %v", isExp, tt.expectExponential)
			}

			t.Logf("Pattern '%s': degree=%d, exponential=%v", tt.pattern, degree, isExp)
		})
	}
}

func TestDetector_BalancedMode(t *testing.T) {
	tests := []struct {
		name         string
		pattern      string
		expectIssues bool
		mode         ValidationMode
	}{
		{
			name:         "nested quantifiers - balanced mode",
			pattern:      "(a+)+",
			expectIssues: true,
			mode:         Balanced,
		},
		{
			name:         "nested quantifiers - fast mode",
			pattern:      "(a+)+",
			expectIssues: true,
			mode:         Fast,
		},
		{
			name:         "safe pattern - balanced mode",
			pattern:      "^[a-z]+$",
			expectIssues: false,
			mode:         Balanced,
		},
	}

	p := parser.NewParser()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			re, err := p.Parse(tt.pattern)
			if err != nil {
				t.Fatalf("Parse error: %v", err)
			}

			opts := &Options{Mode: tt.mode}
			detector := NewDetector(opts)

			issues, err := detector.Detect(re, tt.pattern)
			if err != nil {
				t.Fatalf("Detect() error = %v", err)
			}

			hasIssues := len(issues) > 0
			if hasIssues != tt.expectIssues {
				t.Errorf("Detect() in %v mode found %d issues, expectIssues = %v",
					tt.mode, len(issues), tt.expectIssues)
				if hasIssues {
					for _, issue := range issues {
						t.Logf("  Issue: %s - %s", issue.Type, issue.Message)
					}
				}
			}
		})
	}
}

func TestNFAAnalyzer_FindNestedQuantifiers(t *testing.T) {
	tests := []struct {
		name     string
		pattern  string
		expected int // minimum expected nested quantifiers
	}{
		{"single nested", "(a+)+", 1},
		{"double nested", "((a+)+)+", 2},
		{"no nesting", "a+b*", 0},
	}

	analyzer := NewNFAAnalyzer()
	p := parser.NewParser()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			re, err := p.Parse(tt.pattern)
			if err != nil {
				t.Fatalf("Parse error: %v", err)
			}

			nested := analyzer.findNestedQuantifiersInNFA(re)
			if len(nested) < tt.expected {
				t.Errorf("findNestedQuantifiersInNFA() found %d, want >= %d", len(nested), tt.expected)
			}

			if len(nested) > 0 {
				t.Logf("Found nested quantifiers: %v", nested)
			}
		})
	}
}
