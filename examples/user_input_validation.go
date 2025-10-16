package examples

import (
	"fmt"
	"net/http"

	"github.com/theakshaypant/regret"
)

// ValidateUserRegex validates a user-provided regex pattern before compilation.
// This is critical for preventing ReDoS attacks in applications that accept
// user input for filtering, searching, or validation.
func ValidateUserRegex(pattern string) error {
	// Quick safety check
	safe := regret.IsSafe(pattern)
	if !safe {
		return fmt.Errorf("regex pattern is unsafe: potential ReDoS vulnerability")
	}

	return nil
}

// ValidateUserRegexDetailed provides detailed validation feedback.
func ValidateUserRegexDetailed(pattern string) (bool, []string, error) {
	issues, err := regret.ValidateWithOptions(pattern, &regret.Options{
		Mode:               regret.Balanced,
		MaxComplexityScore: 60,
		StrictMode:         false,
	})
	if err != nil {
		return false, nil, err
	}

	var warnings []string
	hasErrors := false

	for _, issue := range issues {
		if issue.Severity == regret.Critical || issue.Severity == regret.High {
			hasErrors = true
			warnings = append(warnings, fmt.Sprintf("[%s] %s", issue.Severity, issue.Message))
		} else if issue.Severity == regret.Medium {
			warnings = append(warnings, fmt.Sprintf("[WARNING] %s", issue.Message))
		}
	}

	return !hasErrors, warnings, nil
}

// RegexValidationMiddleware is an HTTP middleware that validates regex patterns
// in query parameters before allowing requests to proceed.
func RegexValidationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check for regex pattern in query params
		pattern := r.URL.Query().Get("pattern")
		if pattern == "" {
			next.ServeHTTP(w, r)
			return
		}

		// Validate the pattern
		safe, warnings, err := ValidateUserRegexDetailed(pattern)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid pattern: %v", err), http.StatusBadRequest)
			return
		}

		if !safe {
			http.Error(w, fmt.Sprintf("Unsafe regex pattern:\n%s", warnings), http.StatusBadRequest)
			return
		}

		// Add warnings to response header if any
		if len(warnings) > 0 {
			for _, warning := range warnings {
				w.Header().Add("X-Regex-Warning", warning)
			}
		}

		next.ServeHTTP(w, r)
	})
}

// SearchFilterValidator validates user-provided search filters before executing searches.
type SearchFilterValidator struct {
	opts *regret.Options
}

// NewSearchFilterValidator creates a new validator with custom options.
func NewSearchFilterValidator(strict bool) *SearchFilterValidator {
	opts := regret.DefaultOptions()
	opts.StrictMode = strict
	if strict {
		opts.MaxComplexityScore = 40 // Very conservative
	}

	return &SearchFilterValidator{opts: opts}
}

// Validate checks if a search pattern is safe to use.
func (v *SearchFilterValidator) Validate(pattern string) error {
	issues, err := regret.ValidateWithOptions(pattern, v.opts)
	if err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	if len(issues) == 0 {
		return nil
	}

	// Collect all issues
	var criticalIssues []string
	for _, issue := range issues {
		if issue.Severity == regret.Critical || issue.Severity == regret.High {
			criticalIssues = append(criticalIssues, issue.Message)
		}
	}

	if len(criticalIssues) > 0 {
		return fmt.Errorf("unsafe pattern detected: %v", criticalIssues)
	}

	return nil
}

// ValidateWithComplexity validates and returns complexity information.
func (v *SearchFilterValidator) ValidateWithComplexity(pattern string) (bool, *regret.ComplexityScore, error) {
	// First check if it's safe
	issues, err := regret.ValidateWithOptions(pattern, v.opts)
	if err != nil {
		return false, nil, err
	}

	// Analyze complexity
	score, err := regret.AnalyzeComplexity(pattern)
	if err != nil {
		return false, nil, err
	}

	// Determine safety
	safe := len(issues) == 0 && score.Overall <= v.opts.MaxComplexityScore

	return safe, score, nil
}
