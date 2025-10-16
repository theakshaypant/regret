package testdata

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/theakshaypant/regret"
)

type PumpPattern struct {
	ExpectedComponents []string `json:"expected_components"`
	MinInputLength     int      `json:"min_input_length"`
}

type PumpTestPattern struct {
	Pattern           string       `json:"pattern"`
	Description       string       `json:"description"`
	ExpectedScoreMin  int          `json:"expected_score_min"`
	ExpectedScoreMax  int          `json:"expected_score_max"`
	ExpectedPump      bool         `json:"expected_pump"`
	ExpectedWorstCase bool         `json:"expected_worst_case"`
	PumpDetails       *PumpPattern `json:"pump_details"`
	Note              string       `json:"note"`
}

type PumpCategory struct {
	Description string            `json:"description"`
	Patterns    []PumpTestPattern `json:"patterns"`
}

type PumpTestData struct {
	Description string                  `json:"description"`
	Version     string                  `json:"version"`
	Categories  map[string]PumpCategory `json:"categories"`
}

func loadPumpPatterns(t *testing.T) *PumpTestData {
	path := filepath.Join(".", "pump_patterns.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read pump_patterns.json: %v", err)
	}

	var testData PumpTestData
	if err := json.Unmarshal(data, &testData); err != nil {
		t.Fatalf("failed to parse pump_patterns.json: %v", err)
	}

	return &testData
}

func TestPumpPatterns_GeneratesPump(t *testing.T) {
	testData := loadPumpPatterns(t)
	category := testData.Categories["generates_pump"]

	t.Logf("Testing %d patterns that should generate pump patterns", len(category.Patterns))

	for _, tc := range category.Patterns {
		t.Run(tc.Description, func(t *testing.T) {
			score, err := regret.AnalyzeComplexity(tc.Pattern)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Check score
			if score.Overall < tc.ExpectedScoreMin {
				t.Errorf("score = %d, want >= %d", score.Overall, tc.ExpectedScoreMin)
			}

			// Check pump generation
			hasPump := len(score.PumpPattern) > 0
			if hasPump != tc.ExpectedPump {
				t.Errorf("pump generated = %v, want %v", hasPump, tc.ExpectedPump)
			}

			// Check worst case generation
			hasWorstCase := score.WorstCaseInput != ""
			if hasWorstCase != tc.ExpectedWorstCase {
				t.Errorf("worst case generated = %v, want %v", hasWorstCase, tc.ExpectedWorstCase)
			}

			// Validate pump details if provided
			if tc.PumpDetails != nil && hasPump {
				if len(score.PumpPattern) == 0 {
					t.Error("expected pump components but none found")
				}

				if len(score.WorstCaseInput) < tc.PumpDetails.MinInputLength {
					t.Errorf("worst case input length = %d, want >= %d",
						len(score.WorstCaseInput), tc.PumpDetails.MinInputLength)
				}

				// Check expected components if specified
				if len(tc.PumpDetails.ExpectedComponents) > 0 {
					found := false
					for _, expected := range tc.PumpDetails.ExpectedComponents {
						for _, actual := range score.PumpPattern {
							if expected == actual {
								found = true
								break
							}
						}
					}
					if !found {
						t.Errorf("expected pump components %v, got %v",
							tc.PumpDetails.ExpectedComponents, score.PumpPattern)
					}
				}
			}

			t.Logf("✓ Pattern: %s", tc.Pattern)
			t.Logf("  Score: %d", score.Overall)
			t.Logf("  Pump: %v", score.PumpPattern)
			t.Logf("  Worst Case: %q", score.WorstCaseInput)
		})
	}
}

func TestPumpPatterns_NoPumpSafe(t *testing.T) {
	testData := loadPumpPatterns(t)
	category := testData.Categories["no_pump_safe"]

	t.Logf("Testing %d safe patterns that should not generate pump patterns", len(category.Patterns))

	for _, tc := range category.Patterns {
		t.Run(tc.Description, func(t *testing.T) {
			score, err := regret.AnalyzeComplexity(tc.Pattern)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Check score
			if tc.ExpectedScoreMax > 0 && score.Overall > tc.ExpectedScoreMax {
				t.Errorf("score = %d, want <= %d", score.Overall, tc.ExpectedScoreMax)
			}

			// Check no pump generation
			hasPump := len(score.PumpPattern) > 0
			if hasPump != tc.ExpectedPump {
				t.Errorf("pump generated = %v, want %v (components: %v)",
					hasPump, tc.ExpectedPump, score.PumpPattern)
			}

			// Check no worst case generation
			hasWorstCase := score.WorstCaseInput != ""
			if hasWorstCase != tc.ExpectedWorstCase {
				t.Errorf("worst case generated = %v, want %v (input: %q)",
					hasWorstCase, tc.ExpectedWorstCase, score.WorstCaseInput)
			}

			t.Logf("✓ Pattern: %s", tc.Pattern)
			t.Logf("  Score: %d (safe)", score.Overall)
		})
	}
}

func TestPumpPatterns_BelowThreshold(t *testing.T) {
	testData := loadPumpPatterns(t)
	category := testData.Categories["no_pump_below_threshold"]

	t.Logf("Testing %d patterns below pump threshold", len(category.Patterns))

	for _, tc := range category.Patterns {
		t.Run(tc.Description, func(t *testing.T) {
			score, err := regret.AnalyzeComplexity(tc.Pattern)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Check score is below threshold
			if score.Overall >= 50 {
				t.Errorf("score = %d, expected < 50 (below pump threshold)", score.Overall)
			}

			// Check no pump generation (because below threshold)
			hasPump := len(score.PumpPattern) > 0
			if hasPump != tc.ExpectedPump {
				t.Errorf("pump generated = %v, want %v", hasPump, tc.ExpectedPump)
			}

			t.Logf("✓ Pattern: %s", tc.Pattern)
			t.Logf("  Score: %d (below threshold)", score.Overall)
			if tc.Note != "" {
				t.Logf("  Note: %s", tc.Note)
			}
		})
	}
}

func TestPumpPatterns_EdgeCases(t *testing.T) {
	testData := loadPumpPatterns(t)
	category := testData.Categories["edge_cases"]

	t.Logf("Testing %d edge case patterns", len(category.Patterns))

	for _, tc := range category.Patterns {
		t.Run(tc.Description, func(t *testing.T) {
			score, err := regret.AnalyzeComplexity(tc.Pattern)
			if err != nil {
				t.Logf("Pattern %q resulted in error: %v (may be expected)", tc.Pattern, err)
				return
			}

			hasPump := len(score.PumpPattern) > 0
			hasWorstCase := score.WorstCaseInput != ""

			t.Logf("✓ Pattern: %s", tc.Pattern)
			t.Logf("  Score: %d", score.Overall)
			t.Logf("  Pump: %v", hasPump)
			t.Logf("  Worst Case: %v", hasWorstCase)
			if tc.Note != "" {
				t.Logf("  Note: %s", tc.Note)
			}

			// Edge cases should at least not crash
			if hasPump && !hasWorstCase {
				t.Error("inconsistent state: pump pattern generated but no worst case input")
			}
		})
	}
}

func TestPumpPatterns_Comprehensive(t *testing.T) {
	testData := loadPumpPatterns(t)

	totalPatterns := 0
	patternsWithPump := 0
	patternsWithoutPump := 0
	errors := 0

	for categoryName, category := range testData.Categories {
		t.Logf("\nCategory: %s (%d patterns)", categoryName, len(category.Patterns))

		for _, tc := range category.Patterns {
			totalPatterns++

			score, err := regret.AnalyzeComplexity(tc.Pattern)
			if err != nil {
				errors++
				continue
			}

			if len(score.PumpPattern) > 0 {
				patternsWithPump++
			} else {
				patternsWithoutPump++
			}
		}
	}

	t.Logf("\n=== Pump Pattern Test Summary ===")
	t.Logf("Total patterns tested: %d", totalPatterns)
	t.Logf("Patterns with pump: %d", patternsWithPump)
	t.Logf("Patterns without pump: %d", patternsWithoutPump)
	t.Logf("Errors: %d", errors)

	if totalPatterns == 0 {
		t.Error("no patterns were tested")
	}
}

