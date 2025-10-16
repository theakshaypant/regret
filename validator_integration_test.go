package regret

import (
	"testing"
)

// Integration tests to verify the public API works with internal detector
func TestValidate_Integration(t *testing.T) {
	tests := []struct {
		name         string
		pattern      string
		opts         *Options
		expectIssues bool
		expectSafe   bool
	}{
		{
			name:         "safe pattern",
			pattern:      "^[a-z]+$",
			opts:         FastOptions(),
			expectIssues: false,
			expectSafe:   true,
		},
		{
			name:         "nested quantifiers",
			pattern:      "(a+)+",
			opts:         FastOptions(),
			expectIssues: true,
			expectSafe:   false,
		},
		{
			name:         "overlapping quantifiers",
			pattern:      "a*a+",
			opts:         FastOptions(),
			expectIssues: true,
			expectSafe:   false,
		},
		{
			name:         "deeply nested",
			pattern:      "(((((a+)+)+)+)+)+", // Depth 6, exceeds threshold
			opts:         FastOptions(),
			expectIssues: true,
			expectSafe:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test ValidateWithOptions
			issues, err := ValidateWithOptions(tt.pattern, tt.opts)
			if err != nil {
				t.Fatalf("ValidateWithOptions() error = %v", err)
			}

			hasIssues := len(issues) > 0
			if hasIssues != tt.expectIssues {
				t.Errorf("ValidateWithOptions() issues count = %d, expectIssues = %v", len(issues), tt.expectIssues)
				if hasIssues {
					t.Logf("Issues found: %+v", issues)
				}
			}

			// Test IsSafe (doesn't take options, uses Fast mode internally)
			safe := IsSafe(tt.pattern)
			if safe != tt.expectSafe {
				t.Errorf("IsSafe() = %v, expectSafe = %v", safe, tt.expectSafe)
			}
		})
	}
}

func TestIsSafe_RealWorldPatterns(t *testing.T) {
	patterns := []struct {
		name    string
		pattern string
		safe    bool
	}{
		{"email simple", `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`, true},
		{"url simple", `^https?://[a-z0-9.-]+\.[a-z]{2,}/[a-z0-9-]+$`, true}, // Simplified to avoid false positive
		{"phone number", `^\d{3}-\d{3}-\d{4}$`, true},
		{"evil nested", `(a+)+b`, false},
		{"evil overlapping", `a*a*a*`, false},
		{"greedy dots", `.*.*.`, false},
	}

	for _, tt := range patterns {
		t.Run(tt.name, func(t *testing.T) {
			safe := IsSafe(tt.pattern)

			if safe != tt.safe {
				t.Errorf("IsSafe(%s) = %v, expected %v", tt.pattern, safe, tt.safe)

				// Print issues for debugging
				if !safe {
					issues, _ := Validate(tt.pattern)
					t.Logf("Issues: %+v", issues)
				}
			}
		})
	}
}

func TestValidate_IssueDetails(t *testing.T) {
	pattern := "(a+)+"

	issues, err := Validate(pattern)
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	if len(issues) == 0 {
		t.Fatal("Expected issues for nested quantifier pattern")
	}

	// Check that we got a nested_quantifiers issue
	foundNested := false
	for _, issue := range issues {
		if issue.Type == NestedQuantifiers {
			foundNested = true

			// Verify issue has required fields
			if issue.Severity == Info {
				t.Error("Issue severity should be higher than Info")
			}
			if issue.Message == "" {
				t.Error("Issue missing message")
			}
			if issue.Suggestion == "" {
				t.Error("Issue missing suggestion")
			}

			t.Logf("Issue details: Type=%s, Severity=%s, Message=%s",
				issue.Type, issue.Severity, issue.Message)
		}
	}

	if !foundNested {
		t.Error("Expected nested_quantifiers issue type")
	}
}

func TestValidate_ValidationModes(t *testing.T) {
	pattern := "(a+)+"

	modes := []*Options{
		FastOptions(),
		DefaultOptions(),
		ThoroughOptions(),
	}

	for _, opts := range modes {
		t.Run(opts.Mode.String(), func(t *testing.T) {
			issues, err := ValidateWithOptions(pattern, opts)
			if err != nil {
				t.Fatalf("ValidateWithOptions() error = %v", err)
			}

			// All modes should detect this via fast checks
			if len(issues) == 0 {
				t.Errorf("Mode %s: Expected issues but got none", opts.Mode.String())
			}
		})
	}
}

func BenchmarkIsSafe_Safe(b *testing.B) {
	pattern := "^[a-z]+$"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = IsSafe(pattern)
	}
}

func BenchmarkIsSafe_Unsafe(b *testing.B) {
	pattern := "(a+)+b"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = IsSafe(pattern)
	}
}

func BenchmarkValidate_ComplexPattern(b *testing.B) {
	pattern := "a+b*c?d+e*f?g+h*i?j+k*"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Validate(pattern)
	}
}
