package pump

import (
	"regexp/syntax"
	"strings"
	"testing"
)

func TestNewGenerator(t *testing.T) {
	tests := []struct {
		name string
		opts *Options
		want *Options
	}{
		{
			name: "with options",
			opts: &Options{
				PumpSize:       20,
				MaxPumpSize:    200,
				IncludeFailure: false,
			},
			want: &Options{
				PumpSize:       20,
				MaxPumpSize:    200,
				IncludeFailure: false,
			},
		},
		{
			name: "nil options (defaults)",
			opts: nil,
			want: &Options{
				PumpSize:       10,
				MaxPumpSize:    100,
				IncludeFailure: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewGenerator(tt.opts)
			if g == nil {
				t.Fatal("NewGenerator returned nil")
			}
			if g.opts.PumpSize != tt.want.PumpSize {
				t.Errorf("PumpSize = %v, want %v", g.opts.PumpSize, tt.want.PumpSize)
			}
			if g.opts.MaxPumpSize != tt.want.MaxPumpSize {
				t.Errorf("MaxPumpSize = %v, want %v", g.opts.MaxPumpSize, tt.want.MaxPumpSize)
			}
			if g.opts.IncludeFailure != tt.want.IncludeFailure {
				t.Errorf("IncludeFailure = %v, want %v", g.opts.IncludeFailure, tt.want.IncludeFailure)
			}
		})
	}
}

func TestGenerate(t *testing.T) {
	generator := NewGenerator(nil)

	tests := []struct {
		name               string
		pattern            string
		wantPatternsMin    int
		checkPumpComponent bool
		checkDescription   bool
	}{
		{
			name:               "nested quantifiers",
			pattern:            "(a+)+",
			wantPatternsMin:    1,
			checkPumpComponent: true,
			checkDescription:   true,
		},
		{
			name:               "double nesting",
			pattern:            "((a*)+)+",
			wantPatternsMin:    1,
			checkPumpComponent: true,
			checkDescription:   true,
		},
		{
			name:               "overlapping quantifiers",
			pattern:            "a*a*",
			wantPatternsMin:    1,
			checkPumpComponent: true,
			checkDescription:   true,
		},
		{
			name:               "overlapping alternation",
			pattern:            "(a|ab)+",
			wantPatternsMin:    1,
			checkPumpComponent: true,
			checkDescription:   true,
		},
		{
			name:               "safe pattern",
			pattern:            "^[a-z]+$",
			wantPatternsMin:    1, // Generic pump
			checkPumpComponent: true,
			checkDescription:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			re, err := syntax.Parse(tt.pattern, syntax.Perl)
			if err != nil {
				t.Fatalf("Failed to parse pattern: %v", err)
			}
			re = re.Simplify()

			patterns, err := generator.Generate(re, tt.pattern)
			if err != nil {
				t.Fatalf("Generate() error = %v", err)
			}

			if len(patterns) < tt.wantPatternsMin {
				t.Errorf("Generated %d patterns, want at least %d", len(patterns), tt.wantPatternsMin)
			}

			if tt.checkPumpComponent {
				for i, p := range patterns {
					if p.PumpComponent == "" {
						t.Errorf("Pattern %d: PumpComponent is empty", i)
					}
				}
			}

			if tt.checkDescription {
				for i, p := range patterns {
					if p.Description == "" {
						t.Errorf("Pattern %d: Description is empty", i)
					}
				}
			}
		})
	}
}

func TestPumpPattern_GenerateInput(t *testing.T) {
	tests := []struct {
		name         string
		pump         PumpPattern
		size         int
		wantContains []string
		wantSuffix   string
	}{
		{
			name: "simple pump",
			pump: PumpPattern{
				BaseString:    "",
				PumpComponent: "a",
				FailSuffix:    "x",
				Sizes:         []int{10},
			},
			size:         10,
			wantContains: []string{"a"},
			wantSuffix:   "x",
		},
		{
			name: "pump with base",
			pump: PumpPattern{
				BaseString:    "prefix_",
				PumpComponent: "ab",
				FailSuffix:    "!",
				Sizes:         []int{5},
			},
			size:         5,
			wantContains: []string{"prefix_", "ab"},
			wantSuffix:   "!",
		},
		{
			name: "zero size",
			pump: PumpPattern{
				BaseString:    "start",
				PumpComponent: "x",
				FailSuffix:    "end",
				Sizes:         []int{0},
			},
			size:         0,
			wantContains: []string{"start", "end"},
			wantSuffix:   "end",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.pump.GenerateInput(tt.size)

			for _, want := range tt.wantContains {
				if !strings.Contains(got, want) {
					t.Errorf("GenerateInput() = %q, want to contain %q", got, want)
				}
			}

			if !strings.HasSuffix(got, tt.wantSuffix) {
				t.Errorf("GenerateInput() = %q, want to end with %q", got, tt.wantSuffix)
			}

			// Check repetition count
			if tt.size > 0 {
				count := strings.Count(got, tt.pump.PumpComponent)
				if count != tt.size {
					t.Errorf("PumpComponent appears %d times, want %d", count, tt.size)
				}
			}
		})
	}
}

func TestPumpPattern_GenerateSequence(t *testing.T) {
	pump := PumpPattern{
		BaseString:    "",
		PumpComponent: "a",
		FailSuffix:    "x",
		Sizes:         []int{5, 10, 15, 20},
	}

	sequence := pump.GenerateSequence()

	if len(sequence) != len(pump.Sizes) {
		t.Errorf("Generated %d inputs, want %d", len(sequence), len(pump.Sizes))
	}

	for i, input := range sequence {
		expectedCount := pump.Sizes[i]
		actualCount := strings.Count(input, pump.PumpComponent)
		if actualCount != expectedCount {
			t.Errorf("Input %d: contains %d pumps, want %d", i, actualCount, expectedCount)
		}
	}

	// Verify increasing sizes
	for i := 1; i < len(sequence); i++ {
		if len(sequence[i]) <= len(sequence[i-1]) {
			t.Errorf("Input %d is not longer than input %d", i, i-1)
		}
	}
}

func TestPumpDetection(t *testing.T) {
	generator := NewGenerator(nil)

	t.Run("nested quantifiers detection", func(t *testing.T) {
		re, _ := syntax.Parse("(a+)+", syntax.Perl)
		re = re.Simplify()

		patterns, err := generator.Generate(re, "(a+)+")
		if err != nil {
			t.Fatalf("Generate() error = %v", err)
		}

		found := false
		for _, p := range patterns {
			if strings.Contains(strings.ToLower(p.Description), "nested") ||
				strings.Contains(strings.ToLower(p.Description), "exponential") {
				found = true
				break
			}
		}

		if !found {
			t.Error("Expected to generate pump for nested quantifiers")
		}
	})

	t.Run("overlapping quantifiers detection", func(t *testing.T) {
		re, _ := syntax.Parse("a*a*", syntax.Perl)
		re = re.Simplify()

		patterns, err := generator.Generate(re, "a*a*")
		if err != nil {
			t.Fatalf("Generate() error = %v", err)
		}

		found := false
		for _, p := range patterns {
			if strings.Contains(strings.ToLower(p.Description), "overlapping") ||
				strings.Contains(strings.ToLower(p.Description), "polynomial") {
				found = true
				break
			}
		}

		if !found {
			t.Error("Expected to generate pump for overlapping quantifiers")
		}
	})

	t.Run("overlapping alternation detection", func(t *testing.T) {
		re, _ := syntax.Parse("(a|ab)+", syntax.Perl)
		re = re.Simplify()

		patterns, err := generator.Generate(re, "(a|ab)+")
		if err != nil {
			t.Fatalf("Generate() error = %v", err)
		}

		// Should generate at least one pump pattern
		if len(patterns) == 0 {
			t.Error("Expected to generate at least one pump pattern")
		}
	})
}

func TestHelperFunctions(t *testing.T) {
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

	t.Run("hasOverlappingQuantifiers", func(t *testing.T) {
		tests := []struct {
			pattern string
			want    bool
		}{
			{"a+", false},
			{"a+b+", true},
			{"a*a*", true},
			{"a+bc*", false}, // Not consecutive
		}

		for _, tt := range tests {
			re, err := syntax.Parse(tt.pattern, syntax.Perl)
			if err != nil {
				t.Fatalf("Failed to parse pattern %q: %v", tt.pattern, err)
			}
			re = re.Simplify()

			got := hasOverlappingQuantifiers(re)
			if got != tt.want {
				t.Errorf("hasOverlappingQuantifiers(%q) = %v, want %v", tt.pattern, got, tt.want)
			}
		}
	})

	t.Run("extractPumpChar", func(t *testing.T) {
		tests := []struct {
			pattern string
			want    string
		}{
			{"a+", "a"},
			{"(x*)+", "x"},
			{"[0-9]+", "a"}, // Falls back to 'a' for char classes
			{".+", "a"},     // Falls back to 'a' for any char
		}

		for _, tt := range tests {
			re, err := syntax.Parse(tt.pattern, syntax.Perl)
			if err != nil {
				t.Fatalf("Failed to parse pattern %q: %v", tt.pattern, err)
			}
			re = re.Simplify()

			got := extractPumpChar(re)
			if got != tt.want {
				t.Errorf("extractPumpChar(%q) = %v, want %v", tt.pattern, got, tt.want)
			}
		}
	})
}

func TestPumpPattern_String(t *testing.T) {
	pump := PumpPattern{
		BaseString:    "prefix",
		PumpComponent: "a",
		FailSuffix:    "x",
		Description:   "test pattern",
		Sizes:         []int{10, 20, 30},
	}

	str := pump.String()
	if str == "" {
		t.Error("String() returned empty string")
	}

	// Should contain key information
	if !strings.Contains(str, "a") {
		t.Error("String() should contain pump component")
	}
	if !strings.Contains(str, "x") {
		t.Error("String() should contain fail suffix")
	}
}

func BenchmarkGenerate(b *testing.B) {
	generator := NewGenerator(nil)
	re, _ := syntax.Parse("(a+)+", syntax.Perl)
	re = re.Simplify()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = generator.Generate(re, "(a+)+")
	}
}

func BenchmarkGenerateInput(b *testing.B) {
	pump := PumpPattern{
		BaseString:    "",
		PumpComponent: "a",
		FailSuffix:    "x",
		Sizes:         []int{100},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = pump.GenerateInput(100)
	}
}
