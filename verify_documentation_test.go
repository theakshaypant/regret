package regret

import (
	"reflect"
	"testing"
)

// TestComplexityScore_FieldsMatchDocumentation verifies that ComplexityScore has all documented fields
func TestComplexityScore_FieldsMatchDocumentation(t *testing.T) {
	score := ComplexityScore{
		Overall:          70,
		TimeComplexity:   Exponential,
		SpaceComplexity:  Linear,
		HasEDA:           true,
		HasIDA:           false,
		PolynomialDegree: 0,
		Metrics: Metrics{
			NestingDepth:     2,
			QuantifierCount:  3,
			AlternationCount: 0,
		},
		WorstCaseInput: "test",
		PumpPattern:    []string{"a"},
		Explanation:    "test",
		Safe:           false,
	}

	// Verify all fields are accessible (compilation check)
	if score.Overall != 70 {
		t.Error("Overall field issue")
	}
	if score.TimeComplexity != Exponential {
		t.Error("TimeComplexity field issue")
	}
	if score.SpaceComplexity != Linear {
		t.Error("SpaceComplexity field issue")
	}
	if !score.HasEDA {
		t.Error("HasEDA field issue")
	}
	if score.HasIDA {
		t.Error("HasIDA field issue")
	}
	if score.PolynomialDegree != 0 {
		t.Error("PolynomialDegree field issue")
	}
	if score.WorstCaseInput != "test" {
		t.Error("WorstCaseInput field issue")
	}
	if len(score.PumpPattern) != 1 || score.PumpPattern[0] != "a" {
		t.Error("PumpPattern field issue")
	}
	if score.Explanation != "test" {
		t.Error("Explanation field issue")
	}
	if score.Safe {
		t.Error("Safe field issue")
	}

	t.Logf("✓ All ComplexityScore fields accessible and correct type")
}

// TestPumpIntegration_AutomaticGeneration verifies automatic pump generation for unsafe patterns
func TestPumpIntegration_AutomaticGeneration(t *testing.T) {
	tests := []struct {
		name            string
		pattern         string
		minScore        int
		expectPump      bool
		expectWorstCase bool
	}{
		{
			name:            "nested quantifiers should generate pump",
			pattern:         "(a+)+",
			minScore:        50,
			expectPump:      true,
			expectWorstCase: true,
		},
		{
			name:            "double nested should generate pump",
			pattern:         "((a+)+)+",
			minScore:        50,
			expectPump:      true,
			expectWorstCase: true,
		},
		{
			name:            "safe pattern should not generate pump",
			pattern:         "^[a-z]+$",
			minScore:        0,
			expectPump:      false,
			expectWorstCase: false,
		},
		{
			name:            "below threshold should not generate pump",
			pattern:         "a*a+",
			minScore:        0,
			expectPump:      false,
			expectWorstCase: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score, err := AnalyzeComplexity(tt.pattern)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Verify score
			if score.Overall < tt.minScore {
				t.Errorf("score = %d, want >= %d", score.Overall, tt.minScore)
			}

			// Verify pump generation
			hasPump := len(score.PumpPattern) > 0
			if hasPump != tt.expectPump {
				t.Errorf("pump generated = %v, want %v (components: %v)",
					hasPump, tt.expectPump, score.PumpPattern)
			}

			// Verify worst case generation
			hasWorstCase := score.WorstCaseInput != ""
			if hasWorstCase != tt.expectWorstCase {
				t.Errorf("worst case generated = %v, want %v (input: %q)",
					hasWorstCase, tt.expectWorstCase, score.WorstCaseInput)
			}

			// For patterns that should have pumps, verify quality
			if tt.expectPump {
				if len(score.PumpPattern) == 0 {
					t.Error("expected pump components but got none")
				}
				if score.WorstCaseInput == "" {
					t.Error("expected worst case input but got empty string")
				}
				if len(score.WorstCaseInput) < 5 {
					t.Errorf("worst case input too short: %q", score.WorstCaseInput)
				}

				// Verify pump components are not empty
				for i, component := range score.PumpPattern {
					if component == "" {
						t.Errorf("PumpPattern[%d] is empty", i)
					}
				}
			}

			t.Logf("✓ Pattern: %s | Score: %d | Pump: %v | WorstCase: %v",
				tt.pattern, score.Overall, hasPump, hasWorstCase)
		})
	}
}

// TestPumpIntegration_ConsistencyWithDocumentation tests that behavior matches documented threshold
func TestPumpIntegration_ConsistencyWithDocumentation(t *testing.T) {
	// Documentation states: "For unsafe patterns (score ≥ 50), automatically generates pump patterns"

	patterns := []string{
		"(a+)+",    // Should have pump (≥ 50)
		"((a+)+)+", // Should have pump (≥ 50)
		"a*a+",     // Should NOT have pump (< 50)
		"^[a-z]+$", // Should NOT have pump (< 50)
	}

	for _, pattern := range patterns {
		score, err := AnalyzeComplexity(pattern)
		if err != nil {
			t.Fatalf("pattern %q: unexpected error: %v", pattern, err)
		}

		hasPump := len(score.PumpPattern) > 0
		shouldHavePump := score.Overall >= 50

		if hasPump != shouldHavePump {
			t.Errorf("pattern %q: pump generated = %v, but score = %d (expected pump if ≥ 50)",
				pattern, hasPump, score.Overall)
		}

		// Also verify worst case consistency
		hasWorstCase := score.WorstCaseInput != ""
		if hasPump != hasWorstCase {
			t.Errorf("pattern %q: inconsistent pump state - pump=%v but worstCase=%v",
				pattern, hasPump, hasWorstCase)
		}

		t.Logf("✓ Pattern: %s | Score: %d | Pump: %v | Consistent: ✓",
			pattern, score.Overall, hasPump)
	}
}

// TestPumpPattern_TypeIsCorrect verifies PumpPattern field is []string as documented
func TestPumpPattern_TypeIsCorrect(t *testing.T) {
	score, err := AnalyzeComplexity("(a+)+")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify type using reflection
	pumpType := reflect.TypeOf(score.PumpPattern)
	if pumpType.Kind() != reflect.Slice {
		t.Fatalf("PumpPattern is not a slice: %v", pumpType)
	}

	if pumpType.Elem().Kind() != reflect.String {
		t.Fatalf("PumpPattern is not []string: %v", pumpType)
	}

	t.Logf("✓ PumpPattern type is correct: []string")
}

// TestWorstCaseInput_TypeIsCorrect verifies WorstCaseInput field is string as documented
func TestWorstCaseInput_TypeIsCorrect(t *testing.T) {
	score, err := AnalyzeComplexity("(a+)+")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify type using reflection
	worstCaseType := reflect.TypeOf(score.WorstCaseInput)
	if worstCaseType.Kind() != reflect.String {
		t.Fatalf("WorstCaseInput is not a string: %v", worstCaseType)
	}

	t.Logf("✓ WorstCaseInput type is correct: string")
}

// TestExamples_FromDocumentation tests code examples from documentation
func TestExamples_FromDocumentation(t *testing.T) {
	// Example from docs/API.md
	t.Run("API.md example", func(t *testing.T) {
		score, err := AnalyzeComplexity("(a+)+")
		if err != nil {
			t.Fatal(err)
		}

		// These should be accessible as per documentation
		_ = score.Overall
		_ = score.TimeComplexity
		_ = score.HasEDA
		_ = score.HasIDA
		_ = score.Explanation

		// Pump patterns are automatically generated for unsafe patterns
		if len(score.PumpPattern) > 0 {
			t.Logf("Pump Components: %v", score.PumpPattern)
			t.Logf("Worst Case Input: %q", score.WorstCaseInput)
		}

		t.Logf("✓ API.md example works correctly")
	})

	// Example from docs/GETTING_STARTED.md
	t.Run("GETTING_STARTED.md example", func(t *testing.T) {
		score, err := AnalyzeComplexity("(a+)+")
		if err != nil {
			panic(err)
		}

		_ = score.Overall
		_ = score.TimeComplexity
		_ = score.Explanation

		if score.HasEDA {
			t.Logf("Pattern has Exponential Degree of Ambiguity")
		}

		// For unsafe patterns (score >= 50), pump patterns are automatically generated
		if len(score.PumpPattern) > 0 {
			_ = score.PumpPattern
			_ = score.WorstCaseInput
		}

		t.Logf("✓ GETTING_STARTED.md example works correctly")
	})
}
