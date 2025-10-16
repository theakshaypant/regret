package parser

import (
	"regexp/syntax"
	"testing"
)

func TestParser_Parse(t *testing.T) {
	p := NewParser()

	tests := []struct {
		name    string
		pattern string
		wantErr bool
	}{
		{"simple", "abc", false},
		{"quantifier", "a+", false},
		{"alternation", "a|b", false},
		{"capture", "(abc)", false},
		{"nested", "(a+)+", false},
		{"invalid", "[", true},
		{"invalid_escape", "\\", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			re, err := p.Parse(tt.pattern)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && re == nil {
				t.Error("Parse() returned nil regex without error")
			}
		})
	}
}

func TestParser_Validate(t *testing.T) {
	p := NewParser()

	tests := []struct {
		pattern string
		valid   bool
	}{
		{"abc", true},
		{"(a+)+", true},
		{"[a-z]+", true},
		{"[", false},
		{"(?P<)", false},
	}

	for _, tt := range tests {
		t.Run(tt.pattern, func(t *testing.T) {
			err := p.Validate(tt.pattern)
			if (err == nil) != tt.valid {
				t.Errorf("Validate() error = %v, want valid = %v", err, tt.valid)
			}
		})
	}
}

func TestIsQuantifier(t *testing.T) {
	p := NewParser()

	tests := []struct {
		pattern string
		want    bool
	}{
		{"a+", true},
		{"a*", true},
		{"a?", true},
		{"a{2,5}", false}, // After Simplify(), root is Concat, not a quantifier
		{"a", false},
		{"ab", false},
	}

	for _, tt := range tests {
		t.Run(tt.pattern, func(t *testing.T) {
			re := p.MustParse(tt.pattern)
			got := IsQuantifier(re)
			if got != tt.want {
				t.Errorf("IsQuantifier() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCountQuantifiers(t *testing.T) {
	p := NewParser()

	tests := []struct {
		pattern string
		want    int
	}{
		{"a", 0},
		{"a+", 1},
		{"a+b*", 2},
		{"(a+)+", 2},
		{"a+b*c?", 3},
		{"(a{2,5})+", 4}, // After Simplify(), a{2,5} becomes aa(a(a(a)?)?)? with 3 Quest + 1 Plus
	}

	for _, tt := range tests {
		t.Run(tt.pattern, func(t *testing.T) {
			re := p.MustParse(tt.pattern)
			got := CountQuantifiers(re)
			if got != tt.want {
				t.Errorf("CountQuantifiers() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetNestingDepth(t *testing.T) {
	p := NewParser()

	tests := []struct {
		pattern string
		want    int
	}{
		{"a", 0},
		{"a+", 1},
		{"(a+)+", 2},
		{"((a+)+)+", 3},
		{"a+b+", 1},
		{"(a+b+)+", 2},
	}

	for _, tt := range tests {
		t.Run(tt.pattern, func(t *testing.T) {
			re := p.MustParse(tt.pattern)
			got := GetNestingDepth(re)
			if got != tt.want {
				t.Errorf("GetNestingDepth() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCountAlternations(t *testing.T) {
	p := NewParser()

	tests := []struct {
		pattern string
		want    int
	}{
		{"a", 0},
		{"a|b", 0},         // After Simplify(), becomes CharClass [ab], no alternation
		{"a|b|c", 0},       // After Simplify(), becomes CharClass [abc], no alternation
		{"(a|b)c", 0},      // After Simplify(), becomes [ab]c, no alternation
		{"(a|b)|(c|d)", 1}, // Capture groups prevent simplification, creates 1 Alternate node
		{"ab|cd", 1},       // Multi-char alternations can't be simplified to CharClass
		{"(ab)|(cd)", 1},   // With captures
	}

	for _, tt := range tests {
		t.Run(tt.pattern, func(t *testing.T) {
			re := p.MustParse(tt.pattern)
			got := CountAlternations(re)
			if got != tt.want {
				t.Errorf("CountAlternations() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFindQuantifiers(t *testing.T) {
	p := NewParser()

	pattern := "(a+)+b*"
	re := p.MustParse(pattern)

	quantifiers := FindQuantifiers(re)
	if len(quantifiers) != 3 {
		t.Errorf("FindQuantifiers() found %d quantifiers, want 3", len(quantifiers))
	}
}

func TestWalk(t *testing.T) {
	p := NewParser()

	pattern := "(a+)+b*"
	re := p.MustParse(pattern)

	count := 0
	Walk(re, func(node *syntax.Regexp) bool {
		count++
		return true
	})

	if count == 0 {
		t.Error("Walk() visited 0 nodes, expected more")
	}
}
