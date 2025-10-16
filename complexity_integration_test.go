package regret

import (
	"testing"
)

// TestComplexityAnalysis tests the full complexity analysis API.
func TestComplexityAnalysis(t *testing.T) {
	tests := []struct {
		name                string
		pattern             string
		wantComplexityClass Complexity
		wantScoreMin        int
		wantScoreMax        int
		wantHasEDA          bool
		wantHasIDA          bool
	}{
		{
			name:                "simple safe pattern",
			pattern:             "^[a-z]+$",
			wantComplexityClass: Linear,
			wantScoreMin:        0,
			wantScoreMax:        30,
			wantHasEDA:          false,
			wantHasIDA:          false,
		},
		{
			name:                "nested quantifiers (EDA)",
			pattern:             "(a+)+",
			wantComplexityClass: Exponential,
			wantScoreMin:        50,
			wantScoreMax:        100,
			wantHasEDA:          true,
			wantHasIDA:          false,
		},
		{
			name:                "overlapping quantifiers (IDA)",
			pattern:             "\\d*\\d+",
			wantComplexityClass: Quadratic,
			wantScoreMin:        25,
			wantScoreMax:        70,
			wantHasEDA:          false,
			wantHasIDA:          true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score, err := AnalyzeComplexity(tt.pattern)
			if err != nil {
				t.Fatalf("AnalyzeComplexity() error = %v", err)
			}

			if score.TimeComplexity != tt.wantComplexityClass {
				t.Errorf("TimeComplexity = %v, want %v", score.TimeComplexity, tt.wantComplexityClass)
			}

			if score.Overall < tt.wantScoreMin || score.Overall > tt.wantScoreMax {
				t.Errorf("Overall score = %v, want between %v and %v", score.Overall, tt.wantScoreMin, tt.wantScoreMax)
			}

			if score.HasEDA != tt.wantHasEDA {
				t.Errorf("HasEDA = %v, want %v", score.HasEDA, tt.wantHasEDA)
			}

			if score.HasIDA != tt.wantHasIDA {
				t.Errorf("HasIDA = %v, want %v", score.HasIDA, tt.wantHasIDA)
			}

			// Verify explanation exists
			if score.Explanation == "" {
				t.Error("Explanation is empty")
			}

			// Verify metrics
			if score.Metrics.NestingDepth < 0 {
				t.Error("NestingDepth is negative")
			}
			if score.Metrics.QuantifierCount < 0 {
				t.Error("QuantifierCount is negative")
			}
		})
	}
}

// TestEndToEnd tests the complete workflow combining validation,
// complexity analysis, and pump generation.
func TestEndToEnd(t *testing.T) {
	pattern := "(a+)+"

	// 1. Validate pattern
	issues, err := Validate(pattern)
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	if len(issues) == 0 {
		t.Error("Expected issues for dangerous pattern, got none")
	}

	// 2. Analyze complexity
	score, err := AnalyzeComplexity(pattern)
	if err != nil {
		t.Fatalf("AnalyzeComplexity() error = %v", err)
	}

	if score.TimeComplexity != Exponential {
		t.Errorf("TimeComplexity = %v, want Exponential", score.TimeComplexity)
	}

	if !score.HasEDA {
		t.Error("Expected HasEDA = true")
	}

	if score.Overall < 50 {
		t.Errorf("Overall score = %v, want >= 50 for exponential pattern", score.Overall)
	}

	// 3. Check auto-generated pump patterns
	if len(score.PumpPattern) == 0 {
		t.Error("Expected auto-generated pump patterns in complexity score, got none")
	}

	// 4. Check worst-case input
	if score.WorstCaseInput == "" {
		t.Error("Expected worst-case input to be generated")
	}

	t.Logf("Pattern: %s", pattern)
	t.Logf("Complexity: %v (score: %d/100)", score.TimeComplexity, score.Overall)
	t.Logf("Issues: %d", len(issues))
	t.Logf("Pump components: %v", score.PumpPattern)
	t.Logf("Worst-case input: %s", score.WorstCaseInput)
}

// BenchmarkAnalyzeComplexity benchmarks complexity analysis.
func BenchmarkAnalyzeComplexity(b *testing.B) {
	pattern := "(a+)+"
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = AnalyzeComplexity(pattern)
	}
}
