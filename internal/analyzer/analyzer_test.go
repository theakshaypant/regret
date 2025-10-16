package analyzer

import (
	"regexp/syntax"
	"testing"
	"time"
)

func TestNewAnalyzer(t *testing.T) {
	tests := []struct {
		name string
		opts *Options
		want *Options
	}{
		{
			name: "with options",
			opts: &Options{
				Timeout:            10 * time.Second,
				MaxComplexityScore: 80,
			},
			want: &Options{
				Timeout:            10 * time.Second,
				MaxComplexityScore: 80,
			},
		},
		{
			name: "nil options (defaults)",
			opts: nil,
			want: &Options{
				Timeout:            5 * time.Second,
				MaxComplexityScore: 100,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := NewAnalyzer(tt.opts)
			if a == nil {
				t.Fatal("NewAnalyzer returned nil")
			}
			if a.opts.Timeout != tt.want.Timeout {
				t.Errorf("Timeout = %v, want %v", a.opts.Timeout, tt.want.Timeout)
			}
			if a.opts.MaxComplexityScore != tt.want.MaxComplexityScore {
				t.Errorf("MaxComplexityScore = %v, want %v", a.opts.MaxComplexityScore, tt.want.MaxComplexityScore)
			}
		})
	}
}

func TestAnalyze(t *testing.T) {
	analyzer := NewAnalyzer(nil)

	tests := []struct {
		name           string
		pattern        string
		wantClass      string
		wantScoreMin   int
		wantScoreMax   int
		wantComplexity string
	}{
		{
			name:           "simple literal",
			pattern:        "hello",
			wantClass:      "linear", // Simplify may optimize differently
			wantScoreMin:   0,
			wantScoreMax:   20,
			wantComplexity: "O(n)",
		},
		{
			name:           "linear pattern",
			pattern:        "ab+cd*",
			wantClass:      "linear",
			wantScoreMin:   0,
			wantScoreMax:   30,
			wantComplexity: "O(n)",
		},
		{
			name:           "nested quantifiers (exponential)",
			pattern:        "(a+)+",
			wantClass:      "exponential",
			wantScoreMin:   50,
			wantScoreMax:   100,
			wantComplexity: "O(2^n)",
		},
		{
			name:           "double nesting",
			pattern:        "((a*)+)+",
			wantClass:      "exponential",
			wantScoreMin:   60,
			wantScoreMax:   100,
			wantComplexity: "O(2^n)",
		},
		{
			name:           "overlapping quantifiers (polynomial)",
			pattern:        "\\d*\\d+",
			wantClass:      "polynomial",
			wantScoreMin:   25,
			wantScoreMax:   70,
			wantComplexity: "O(n²)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			re, err := syntax.Parse(tt.pattern, syntax.Perl)
			if err != nil {
				t.Fatalf("Failed to parse pattern: %v", err)
			}
			re = re.Simplify()

			result, err := analyzer.Analyze(re, tt.pattern)
			if err != nil {
				t.Fatalf("Analyze() error = %v", err)
			}

			if result.TimeClass != tt.wantClass {
				t.Errorf("TimeClass = %v, want %v", result.TimeClass, tt.wantClass)
			}

			if result.Score < tt.wantScoreMin || result.Score > tt.wantScoreMax {
				t.Errorf("Score = %v, want between %v and %v", result.Score, tt.wantScoreMin, tt.wantScoreMax)
			}

			if result.Complexity != tt.wantComplexity {
				t.Errorf("Complexity = %v, want %v", result.Complexity, tt.wantComplexity)
			}

			// Verify description exists
			if result.Description == "" {
				t.Error("Description is empty")
			}

			// Verify metrics
			if result.Metrics == nil {
				t.Error("Metrics is nil")
			}
		})
	}
}

func TestEstimateComplexity(t *testing.T) {
	analyzer := NewAnalyzer(nil)

	tests := []struct {
		name    string
		pattern string
		want    string
	}{
		{
			name:    "constant time",
			pattern: "hello",
			want:    "O(1)",
		},
		{
			name:    "linear time",
			pattern: "a+",
			want:    "O(n)",
		},
		{
			name:    "quadratic time",
			pattern: "a*a*",
			want:    "O(n²)",
		},
		{
			name:    "quadratic or polynomial time",
			pattern: "\\d*\\d+\\w*",
			want:    "O(n²)", // Or O(n³), depends on Simplify
		},
		{
			name:    "exponential time",
			pattern: "(a+)+",
			want:    "O(2^n)",
		},
		{
			name:    "exponential time (double nesting)",
			pattern: "((a*)+)+",
			want:    "O(2^n)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			re, err := syntax.Parse(tt.pattern, syntax.Perl)
			if err != nil {
				t.Fatalf("Failed to parse pattern: %v", err)
			}
			re = re.Simplify()

			got := analyzer.EstimateComplexity(re)
			if got != tt.want {
				t.Errorf("EstimateComplexity() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestComplexityScoring(t *testing.T) {
	analyzer := NewAnalyzer(nil)

	tests := []struct {
		name            string
		pattern         string
		checkIssues     bool
		expectedIssues  []string
		checkMetrics    bool
		expectedMetrics map[string]int
	}{
		{
			name:        "nested quantifiers issue",
			pattern:     "(a+)+",
			checkIssues: true,
			expectedIssues: []string{
				"nested quantifiers (exponential risk)",
			},
			checkMetrics: true,
			expectedMetrics: map[string]int{
				"nested_quantifiers": 1, // One outer quantifier with nested inner
				"nesting_depth":      2,
			},
		},
		{
			name:        "overlapping quantifiers issue",
			pattern:     "a*a*",
			checkIssues: true,
			expectedIssues: []string{
				"overlapping quantifiers (quadratic)",
			},
			checkMetrics: true,
			expectedMetrics: map[string]int{
				"overlapping_sequences": 1,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			re, err := syntax.Parse(tt.pattern, syntax.Perl)
			if err != nil {
				t.Fatalf("Failed to parse pattern: %v", err)
			}
			re = re.Simplify()

			result, err := analyzer.Analyze(re, tt.pattern)
			if err != nil {
				t.Fatalf("Analyze() error = %v", err)
			}

			if tt.checkIssues {
				if len(result.Issues) == 0 {
					t.Error("Expected issues, got none")
				}
				for _, expected := range tt.expectedIssues {
					found := false
					for _, issue := range result.Issues {
						if issue == expected {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("Expected issue %q not found in %v", expected, result.Issues)
					}
				}
			}

			if tt.checkMetrics {
				for key, expected := range tt.expectedMetrics {
					if val, ok := result.Metrics[key]; ok {
						if intVal, ok := val.(int); ok {
							if intVal != expected {
								t.Errorf("Metric %s = %v, want %v", key, intVal, expected)
							}
						} else {
							t.Errorf("Metric %s is not an int", key)
						}
					} else {
						t.Errorf("Metric %s not found", key)
					}
				}
			}
		})
	}
}

func TestMaxComplexityScore(t *testing.T) {
	// Create analyzer with max score of 50
	analyzer := NewAnalyzer(&Options{
		Timeout:            5 * time.Second,
		MaxComplexityScore: 50,
	})

	// Pattern that would normally score > 50
	re, err := syntax.Parse("((((a+)+)+)+)+", syntax.Perl)
	if err != nil {
		t.Fatalf("Failed to parse pattern: %v", err)
	}
	re = re.Simplify()

	result, err := analyzer.Analyze(re, "((((a+)+)+)+)+")
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}

	if result.Score > 50 {
		t.Errorf("Score = %v, want <= 50 (capped by MaxComplexityScore)", result.Score)
	}
}

func TestHelperFunctions(t *testing.T) {
	t.Run("isQuantifier", func(t *testing.T) {
		tests := []struct {
			pattern string
			want    bool
		}{
			{"a+", true},
			{"a*", true},
			{"a?", true},
			{"a", false},
			{"ab", false},
		}

		for _, tt := range tests {
			re, err := syntax.Parse(tt.pattern, syntax.Perl)
			if err != nil {
				t.Fatalf("Failed to parse pattern %q: %v", tt.pattern, err)
			}
			re = re.Simplify()

			got := isQuantifier(re)
			if got != tt.want {
				t.Errorf("isQuantifier(%q) = %v, want %v", tt.pattern, got, tt.want)
			}
		}
	})

	t.Run("countQuantifiers", func(t *testing.T) {
		tests := []struct {
			pattern string
			want    int
		}{
			{"abc", 0},
			{"a+", 1},
			{"a+b*", 2},
			{"(a+b*)+", 3}, // outer quantifier + inner 2
		}

		for _, tt := range tests {
			re, err := syntax.Parse(tt.pattern, syntax.Perl)
			if err != nil {
				t.Fatalf("Failed to parse pattern %q: %v", tt.pattern, err)
			}
			re = re.Simplify()

			got := countQuantifiers(re)
			if got != tt.want {
				t.Errorf("countQuantifiers(%q) = %v, want %v", tt.pattern, got, tt.want)
			}
		}
	})

	t.Run("hasNestedQuantifiers", func(t *testing.T) {
		tests := []struct {
			pattern string
			want    bool
		}{
			{"a+", false},
			{"a+b*", false},
			{"(a+)+", true},
			{"(a*)*", true},
			{"((a+)+)+", true},
		}

		for _, tt := range tests {
			re, err := syntax.Parse(tt.pattern, syntax.Perl)
			if err != nil {
				t.Fatalf("Failed to parse pattern %q: %v", tt.pattern, err)
			}
			re = re.Simplify()

			got := hasNestedQuantifiers(re)
			if got != tt.want {
				t.Errorf("hasNestedQuantifiers(%q) = %v, want %v", tt.pattern, got, tt.want)
			}
		}
	})
}

func BenchmarkAnalyze(b *testing.B) {
	analyzer := NewAnalyzer(nil)
	pattern := "(a+)+"
	re, _ := syntax.Parse(pattern, syntax.Perl)
	re = re.Simplify()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = analyzer.Analyze(re, pattern)
	}
}

func BenchmarkEstimateComplexity(b *testing.B) {
	analyzer := NewAnalyzer(nil)
	pattern := "(a+)+"
	re, _ := syntax.Parse(pattern, syntax.Perl)
	re = re.Simplify()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = analyzer.EstimateComplexity(re)
	}
}
