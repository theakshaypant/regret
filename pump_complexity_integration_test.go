package regret

import (
	"testing"
)

func TestAnalyzeComplexity_PumpIntegration(t *testing.T) {
	tests := []struct {
		name                 string
		pattern              string
		expectPumpComponents bool
		expectWorstCaseInput bool
		minScore             int // Minimum expected score for unsafe patterns
	}{
		{
			name:                 "nested quantifiers should generate pump",
			pattern:              "(a+)+",
			expectPumpComponents: true,
			expectWorstCaseInput: true,
			minScore:             50,
		},
		{
			name:                 "overlapping quantifiers (below threshold)",
			pattern:              "a*a+",
			expectPumpComponents: false, // Score 45 - below threshold
			expectWorstCaseInput: false,
			minScore:             0,
		},
		{
			name:                 "overlapping alternation (safe in Go)",
			pattern:              "(a|ab)+",
			expectPumpComponents: false, // Go's regex engine handles this efficiently
			expectWorstCaseInput: false,
			minScore:             0,
		},
		{
			name:                 "double nested quantifiers should generate pump",
			pattern:              "((a+)+)+",
			expectPumpComponents: true,
			expectWorstCaseInput: true,
			minScore:             50,
		},
		{
			name:                 "safe pattern should not generate pump",
			pattern:              "^[a-z]+$",
			expectPumpComponents: false,
			expectWorstCaseInput: false,
			minScore:             0,
		},
		{
			name:                 "simple literal should not generate pump",
			pattern:              "hello",
			expectPumpComponents: false,
			expectWorstCaseInput: false,
			minScore:             0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score, err := AnalyzeComplexity(tt.pattern)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Check if score is as expected
			if score.Overall < tt.minScore {
				t.Errorf("score.Overall = %d, want >= %d", score.Overall, tt.minScore)
			}

			// Check PumpPattern field
			hasPumpComponents := len(score.PumpPattern) > 0
			if hasPumpComponents != tt.expectPumpComponents {
				t.Errorf("PumpPattern populated = %v, want %v (components: %v)",
					hasPumpComponents, tt.expectPumpComponents, score.PumpPattern)
			}

			// Check WorstCaseInput field
			hasWorstCase := score.WorstCaseInput != ""
			if hasWorstCase != tt.expectWorstCaseInput {
				t.Errorf("WorstCaseInput populated = %v, want %v (input: %q)",
					hasWorstCase, tt.expectWorstCaseInput, score.WorstCaseInput)
			}

			// Additional validation for patterns that should have pump data
			if tt.expectPumpComponents {
				if len(score.PumpPattern) == 0 {
					t.Error("expected PumpPattern to be populated but it was empty")
				}
				if score.WorstCaseInput == "" {
					t.Error("expected WorstCaseInput to be populated but it was empty")
				}

				// Verify pump components are reasonable
				for i, component := range score.PumpPattern {
					if component == "" {
						t.Errorf("PumpPattern[%d] is empty", i)
					}
				}

				t.Logf("Pattern: %s", tt.pattern)
				t.Logf("Score: %d", score.Overall)
				t.Logf("Pump Components: %v", score.PumpPattern)
				t.Logf("Worst Case Input (len=%d): %q", len(score.WorstCaseInput), score.WorstCaseInput)
			}
		})
	}
}

func TestAnalyzeComplexity_PumpInputGeneration(t *testing.T) {
	// Test with a known dangerous pattern
	score, err := AnalyzeComplexity("(a+)+b")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if score.Overall < 50 {
		t.Errorf("expected high complexity score, got %d", score.Overall)
	}

	if len(score.PumpPattern) == 0 {
		t.Fatal("expected PumpPattern to be populated")
	}

	if score.WorstCaseInput == "" {
		t.Fatal("expected WorstCaseInput to be populated")
	}

	// Verify the worst case input contains the pump component
	pumpComponent := score.PumpPattern[0]
	if pumpComponent == "" {
		t.Fatal("pump component is empty")
	}

	// The worst case input should contain multiple instances of the pump component
	// or at least be long enough to demonstrate the issue
	if len(score.WorstCaseInput) < 10 {
		t.Errorf("worst case input seems too short: %q", score.WorstCaseInput)
	}

	t.Logf("Pump component: %q", pumpComponent)
	t.Logf("Worst case input: %q", score.WorstCaseInput)
	t.Logf("Input length: %d", len(score.WorstCaseInput))
}

func TestAnalyzeComplexity_PumpConsistency(t *testing.T) {
	// Run analysis multiple times to ensure consistency
	pattern := "(a+)+"

	var prevScore *ComplexityScore
	for i := 0; i < 3; i++ {
		score, err := AnalyzeComplexity(pattern)
		if err != nil {
			t.Fatalf("iteration %d: unexpected error: %v", i, err)
		}

		if i > 0 {
			// Check consistency with previous run
			if score.Overall != prevScore.Overall {
				t.Errorf("iteration %d: score changed from %d to %d", i, prevScore.Overall, score.Overall)
			}

			if len(score.PumpPattern) != len(prevScore.PumpPattern) {
				t.Errorf("iteration %d: pump pattern count changed", i)
			}

			// WorstCaseInput should be the same each time
			if score.WorstCaseInput != prevScore.WorstCaseInput {
				t.Errorf("iteration %d: worst case input changed", i)
			}
		}

		prevScore = score
	}
}
