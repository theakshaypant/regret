package regret

import (
	"testing"
	"time"
)

func TestValidationMode_String(t *testing.T) {
	tests := []struct {
		mode ValidationMode
		want string
	}{
		{Fast, "fast"},
		{Balanced, "balanced"},
		{Thorough, "thorough"},
		{ValidationMode(999), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.mode.String(); got != tt.want {
				t.Errorf("ValidationMode.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSeverity_String(t *testing.T) {
	tests := []struct {
		severity Severity
		want     string
	}{
		{Critical, "critical"},
		{High, "high"},
		{Medium, "medium"},
		{Low, "low"},
		{Info, "info"},
		{Severity(999), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.severity.String(); got != tt.want {
				t.Errorf("Severity.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIssueType_String(t *testing.T) {
	tests := []struct {
		issueType IssueType
		want      string
	}{
		{NestedQuantifiers, "nested_quantifiers"},
		{OverlappingAlternation, "overlapping_alternation"},
		{ExponentialBacktracking, "exponential_backtracking"},
		{PolynomialBacktracking, "polynomial_backtracking"},
		{IssueType(999), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.issueType.String(); got != tt.want {
				t.Errorf("IssueType.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestComplexity_String(t *testing.T) {
	tests := []struct {
		complexity Complexity
		want       string
	}{
		{Constant, "O(1)"},
		{Linear, "O(n)"},
		{Quadratic, "O(n²)"},
		{Cubic, "O(n³)"},
		{Polynomial, "O(n^k)"},
		{Exponential, "O(2^n)"},
		{Unknown, "O(?)"},
		{Complexity(999), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.complexity.String(); got != tt.want {
				t.Errorf("Complexity.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDefaultOptions(t *testing.T) {
	opts := DefaultOptions()

	if opts.Mode != Balanced {
		t.Errorf("DefaultOptions().Mode = %v, want %v", opts.Mode, Balanced)
	}
	if opts.Timeout != 100*time.Millisecond {
		t.Errorf("DefaultOptions().Timeout = %v, want %v", opts.Timeout, 100*time.Millisecond)
	}
	if opts.Checks != CheckDefault {
		t.Errorf("DefaultOptions().Checks = %v, want %v", opts.Checks, CheckDefault)
	}
	if opts.MaxComplexityScore != 70 {
		t.Errorf("DefaultOptions().MaxComplexityScore = %v, want 70", opts.MaxComplexityScore)
	}
}

func TestFastOptions(t *testing.T) {
	opts := FastOptions()

	if opts.Mode != Fast {
		t.Errorf("FastOptions().Mode = %v, want %v", opts.Mode, Fast)
	}
	if opts.Timeout != 10*time.Millisecond {
		t.Errorf("FastOptions().Timeout = %v, want %v", opts.Timeout, 10*time.Millisecond)
	}
}

func TestThoroughOptions(t *testing.T) {
	opts := ThoroughOptions()

	if opts.Mode != Thorough {
		t.Errorf("ThoroughOptions().Mode = %v, want %v", opts.Mode, Thorough)
	}
	if opts.Checks != CheckAll {
		t.Errorf("ThoroughOptions().Checks = %v, want %v", opts.Checks, CheckAll)
	}
	if !opts.StrictMode {
		t.Error("ThoroughOptions().StrictMode = false, want true")
	}
}

func TestPumpPattern_Generate(t *testing.T) {
	tests := []struct {
		name string
		pump PumpPattern
		size int
		want string
	}{
		{
			name: "simple pump",
			pump: PumpPattern{
				Prefix: "test",
				Pumps:  []string{"a"},
				Suffix: "x",
			},
			size: 3,
			want: "testaaax",
		},
		{
			name: "zero size",
			pump: PumpPattern{
				Prefix: "pre",
				Pumps:  []string{"a"},
				Suffix: "post",
			},
			size: 0,
			want: "prepost",
		},
		{
			name: "multiple pumps concatenated",
			pump: PumpPattern{
				Prefix:     "",
				Pumps:      []string{"a", "b"},
				Suffix:     "x",
				Interleave: false,
			},
			size: 2,
			want: "aabbx",
		},
		{
			name: "multiple pumps interleaved",
			pump: PumpPattern{
				Prefix:     "",
				Pumps:      []string{"a", "b"},
				Suffix:     "x",
				Interleave: true,
			},
			size: 2,
			want: "ababx",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.pump.Generate(tt.size)
			if got != tt.want {
				t.Errorf("PumpPattern.Generate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPumpPattern_GenerateSequence(t *testing.T) {
	pump := PumpPattern{
		Prefix: "",
		Pumps:  []string{"a"},
		Suffix: "x",
	}

	sequence := pump.GenerateSequence(1, 3, 1)

	want := []string{"ax", "aax", "aaax"}
	if len(sequence) != len(want) {
		t.Fatalf("GenerateSequence() returned %d items, want %d", len(sequence), len(want))
	}

	for i, got := range sequence {
		if got != want[i] {
			t.Errorf("GenerateSequence()[%d] = %v, want %v", i, got, want[i])
		}
	}
}

func TestCheckFlags(t *testing.T) {
	// Test that CheckAll includes all flags
	if CheckAll&CheckNestedQuantifiers == 0 {
		t.Error("CheckAll should include CheckNestedQuantifiers")
	}
	if CheckAll&CheckNFAAmbiguity == 0 {
		t.Error("CheckAll should include CheckNFAAmbiguity")
	}

	// Test CheckDefault includes expected flags
	if CheckDefault&CheckNestedQuantifiers == 0 {
		t.Error("CheckDefault should include CheckNestedQuantifiers")
	}
	if CheckDefault&CheckNFAAmbiguity == 0 {
		t.Error("CheckDefault should include CheckNFAAmbiguity")
	}

	// Test flag combinations
	combined := CheckNestedQuantifiers | CheckOverlappingAlternation
	if combined&CheckNestedQuantifiers == 0 {
		t.Error("Combined flags should include CheckNestedQuantifiers")
	}
	if combined&CheckOverlappingAlternation == 0 {
		t.Error("Combined flags should include CheckOverlappingAlternation")
	}
}

func TestComplexity_BigO(t *testing.T) {
	tests := []struct {
		complexity Complexity
		want       string
	}{
		{Linear, "O(n)"},
		{Quadratic, "O(n²)"},
		{Exponential, "O(2^n)"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.complexity.BigO(); got != tt.want {
				t.Errorf("Complexity.BigO() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFullVersion(t *testing.T) {
	version := FullVersion()
	if version == "" {
		t.Error("FullVersion() returned empty string")
	}
	// Should include prerelease suffix
	if VersionPrerelease != "" && version == Version {
		t.Errorf("FullVersion() = %v, expected to include prerelease suffix", version)
	}
}
