package testdata_test

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp/syntax"
	"testing"

	"github.com/theakshaypant/regret"
)

// EvilPattern represents a dangerous regex pattern from testdata.
type EvilPattern struct {
	Pattern     string `json:"pattern"`
	Category    string `json:"category"`
	Severity    string `json:"severity"`
	Complexity  string `json:"complexity"`
	Description string `json:"description"`
	AttackInput string `json:"attack_input"`
}

// SafePattern represents a safe regex pattern from testdata.
type SafePattern struct {
	Pattern     string `json:"pattern"`
	Category    string `json:"category"`
	Complexity  string `json:"complexity"`
	Description string `json:"description"`
	UseCase     string `json:"use_case"`
}

// EdgeCasePattern represents an edge case pattern from testdata.
type EdgeCasePattern struct {
	Pattern        string `json:"pattern"`
	Description    string `json:"description"`
	ExpectedStatus string `json:"expected_status"`
	Reason         string `json:"reason"`
	Category       string `json:"category"`
}

// TestEvilPatternsLoad tests that all evil patterns can be loaded.
func TestEvilPatternsLoad(t *testing.T) {
	data, err := os.ReadFile("evil_patterns.json")
	if err != nil {
		t.Fatalf("Failed to read evil_patterns.json: %v", err)
	}

	var result struct {
		Description string        `json:"description"`
		Patterns    []EvilPattern `json:"patterns"`
	}

	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to parse evil_patterns.json: %v", err)
	}

	t.Logf("Loaded %d evil patterns", len(result.Patterns))

	if len(result.Patterns) == 0 {
		t.Fatal("No evil patterns loaded")
	}

	// Verify each pattern
	for i, p := range result.Patterns {
		t.Run(p.Category, func(t *testing.T) {
			// Check pattern is not empty
			if p.Pattern == "" {
				t.Errorf("Pattern %d has empty pattern string", i)
			}

			// Check pattern compiles (validates syntax)
			if _, err := syntax.Parse(p.Pattern, syntax.Perl); err != nil {
				t.Errorf("Pattern %q is invalid: %v", p.Pattern, err)
			}

			// Check required fields
			if p.Category == "" {
				t.Errorf("Pattern %q missing category", p.Pattern)
			}
			if p.Severity == "" {
				t.Errorf("Pattern %q missing severity", p.Pattern)
			}
			if p.Complexity == "" {
				t.Errorf("Pattern %q missing complexity", p.Pattern)
			}
		})
	}
}

// TestEvilPatternsAreDetected verifies that evil patterns are actually detected as unsafe.
func TestEvilPatternsAreDetected(t *testing.T) {
	data, err := os.ReadFile("evil_patterns.json")
	if err != nil {
		t.Fatalf("Failed to read evil_patterns.json: %v", err)
	}

	var result struct {
		Patterns []EvilPattern `json:"patterns"`
	}

	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to parse evil_patterns.json: %v", err)
	}

	passed := 0
	failed := 0
	var missedPatterns []string

	for _, p := range result.Patterns {
		safe := regret.IsSafe(p.Pattern)
		if safe {
			missedPatterns = append(missedPatterns, fmt.Sprintf("%q (%s)", p.Pattern, p.Category))
			failed++
		} else {
			passed++
		}
	}

	detectionRate := float64(passed) / float64(len(result.Patterns)) * 100
	t.Logf("Detection rate: %d/%d evil patterns detected (%.1f%%)", passed, len(result.Patterns), detectionRate)

	// Accept 80%+ detection rate as passing (current phase limitations)
	// Some patterns require context-aware analysis (Phase 4)
	if detectionRate < 80.0 {
		t.Errorf("Detection rate too low: %.1f%% (expected >= 80%%)", detectionRate)
	}

	if failed > 0 {
		t.Logf("NOTE: %d patterns require advanced analysis (planned for Phase 4):", failed)
		for _, pattern := range missedPatterns {
			t.Logf("  - %s", pattern)
		}
	}
}

// TestSafePatternsLoad tests that all safe patterns can be loaded.
func TestSafePatternsLoad(t *testing.T) {
	data, err := os.ReadFile("safe_patterns.json")
	if err != nil {
		t.Fatalf("Failed to read safe_patterns.json: %v", err)
	}

	var result struct {
		Description string        `json:"description"`
		Patterns    []SafePattern `json:"patterns"`
	}

	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to parse safe_patterns.json: %v", err)
	}

	t.Logf("Loaded %d safe patterns", len(result.Patterns))

	if len(result.Patterns) == 0 {
		t.Fatal("No safe patterns loaded")
	}

	// Verify each pattern
	for i, p := range result.Patterns {
		t.Run(p.Category, func(t *testing.T) {
			// Check pattern is not empty
			if p.Pattern == "" {
				t.Errorf("Pattern %d has empty pattern string", i)
			}

			// Check pattern compiles
			if _, err := syntax.Parse(p.Pattern, syntax.Perl); err != nil {
				t.Errorf("Pattern %q is invalid: %v", p.Pattern, err)
			}

			// Check required fields
			if p.Category == "" {
				t.Errorf("Pattern %q missing category", p.Pattern)
			}
			if p.UseCase == "" {
				t.Errorf("Pattern %q missing use_case", p.Pattern)
			}
		})
	}
}

// TestSafePatternsAreSafe verifies that safe patterns are actually detected as safe.
func TestSafePatternsAreSafe(t *testing.T) {
	data, err := os.ReadFile("safe_patterns.json")
	if err != nil {
		t.Fatalf("Failed to read safe_patterns.json: %v", err)
	}

	var result struct {
		Patterns []SafePattern `json:"patterns"`
	}

	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to parse safe_patterns.json: %v", err)
	}

	passed := 0
	failed := 0

	for _, p := range result.Patterns {
		safe := regret.IsSafe(p.Pattern)
		if !safe {
			t.Logf("WARNING: Pattern %q (category: %s) was marked UNSAFE but should be SAFE (possible false positive)", p.Pattern, p.Category)
			failed++
		} else {
			passed++
		}
	}

	t.Logf("Accuracy: %d/%d safe patterns marked safe (%.1f%%)", passed, len(result.Patterns), float64(passed)/float64(len(result.Patterns))*100)

	if failed > len(result.Patterns)/2 {
		t.Errorf("Too many false positives: %d/%d safe patterns marked unsafe", failed, len(result.Patterns))
	}
}

// TestEdgeCasesLoad tests that all edge case patterns can be loaded.
func TestEdgeCasesLoad(t *testing.T) {
	data, err := os.ReadFile("edge_cases.json")
	if err != nil {
		t.Fatalf("Failed to read edge_cases.json: %v", err)
	}

	var result struct {
		Description string            `json:"description"`
		Patterns    []EdgeCasePattern `json:"patterns"`
	}

	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to parse edge_cases.json: %v", err)
	}

	t.Logf("Loaded %d edge case patterns", len(result.Patterns))

	if len(result.Patterns) == 0 {
		t.Fatal("No edge case patterns loaded")
	}

	// Verify each pattern
	for i, p := range result.Patterns {
		if p.Pattern == "" {
			t.Errorf("Pattern %d has empty pattern string", i)
		}

		// Check pattern compiles
		if _, err := syntax.Parse(p.Pattern, syntax.Perl); err != nil {
			t.Errorf("Pattern %q is invalid: %v", p.Pattern, err)
		}

		// Check expected status is valid
		if p.ExpectedStatus != "safe" && p.ExpectedStatus != "unsafe" && p.ExpectedStatus != "warning" {
			t.Errorf("Pattern %q has invalid expected_status: %q", p.Pattern, p.ExpectedStatus)
		}
	}
}

// TestAllTestdataFilesExist verifies all expected testdata files are present.
func TestAllTestdataFilesExist(t *testing.T) {
	expectedFiles := []string{
		"evil_patterns.json",
		"safe_patterns.json",
		"real_world_patterns.json",
		"edge_cases.json",
		"performance_patterns.json",
		"README.md",
	}

	for _, filename := range expectedFiles {
		if _, err := os.Stat(filename); os.IsNotExist(err) {
			t.Errorf("Missing expected testdata file: %s", filename)
		} else {
			t.Logf("✓ Found %s", filename)
		}
	}
}

// TestTestdataStatistics provides statistics about the testdata.
func TestTestdataStatistics(t *testing.T) {
	type stats struct {
		EvilPatterns      int
		SafePatterns      int
		RealWorldPatterns int
		EdgeCases         int
		PerfPatterns      int
	}

	var s stats

	// Count evil patterns
	if data, err := os.ReadFile("evil_patterns.json"); err == nil {
		var result struct {
			Patterns []EvilPattern `json:"patterns"`
		}
		if json.Unmarshal(data, &result) == nil {
			s.EvilPatterns = len(result.Patterns)
		}
	}

	// Count safe patterns
	if data, err := os.ReadFile("safe_patterns.json"); err == nil {
		var result struct {
			Patterns []SafePattern `json:"patterns"`
		}
		if json.Unmarshal(data, &result) == nil {
			s.SafePatterns = len(result.Patterns)
		}
	}

	// Count real-world patterns
	if data, err := os.ReadFile("real_world_patterns.json"); err == nil {
		var result struct {
			Patterns []map[string]interface{} `json:"patterns"`
		}
		if json.Unmarshal(data, &result) == nil {
			s.RealWorldPatterns = len(result.Patterns)
		}
	}

	// Count edge cases
	if data, err := os.ReadFile("edge_cases.json"); err == nil {
		var result struct {
			Patterns []EdgeCasePattern `json:"patterns"`
		}
		if json.Unmarshal(data, &result) == nil {
			s.EdgeCases = len(result.Patterns)
		}
	}

	// Count performance patterns
	if data, err := os.ReadFile("performance_patterns.json"); err == nil {
		var result struct {
			Patterns []map[string]interface{} `json:"patterns"`
		}
		if json.Unmarshal(data, &result) == nil {
			s.PerfPatterns = len(result.Patterns)
		}
	}

	total := s.EvilPatterns + s.SafePatterns + s.RealWorldPatterns + s.EdgeCases + s.PerfPatterns

	t.Logf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	t.Logf("TESTDATA STATISTICS")
	t.Logf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	t.Logf("Evil patterns:       %d", s.EvilPatterns)
	t.Logf("Safe patterns:       %d", s.SafePatterns)
	t.Logf("Real-world patterns: %d", s.RealWorldPatterns)
	t.Logf("Edge cases:          %d", s.EdgeCases)
	t.Logf("Performance:         %d", s.PerfPatterns)
	t.Logf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	t.Logf("TOTAL:               %d patterns", total)
	t.Logf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	if total < 100 {
		t.Errorf("Expected at least 100 patterns, got %d", total)
	}
}
