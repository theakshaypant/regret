package examples

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/theakshaypant/regret"
)

// CICDReport represents a security audit report for CI/CD integration.
type CICDReport struct {
	TotalPatterns   int          `json:"total_patterns"`
	SafePatterns    int          `json:"safe_patterns"`
	UnsafePatterns  int          `json:"unsafe_patterns"`
	Files           []FileReport `json:"files"`
	Summary         string       `json:"summary"`
	ExitCode        int          `json:"exit_code"`
	FailureMessages []string     `json:"failure_messages,omitempty"`
}

// FileReport represents findings for a single file.
type FileReport struct {
	Path     string   `json:"path"`
	Patterns []string `json:"patterns"`
	Issues   []Issue  `json:"issues,omitempty"`
}

// Issue represents a detected regex issue.
type Issue struct {
	Pattern    string `json:"pattern"`
	Type       string `json:"type"`
	Severity   string `json:"severity"`
	Message    string `json:"message"`
	Suggestion string `json:"suggestion,omitempty"`
	LineNumber int    `json:"line_number,omitempty"`
}

// ScanCodebase scans a codebase for regex patterns and validates them.
// This is designed to run in CI/CD pipelines to catch ReDoS vulnerabilities.
func ScanCodebase(rootDir string, extensions []string) (*CICDReport, error) {
	report := &CICDReport{
		Files: make([]FileReport, 0),
	}

	// Walk directory tree
	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Check file extension
		ext := filepath.Ext(path)
		if !contains(extensions, ext) {
			return nil
		}

		// Scan file for regex patterns
		fileReport, err := scanFile(path)
		if err != nil {
			return fmt.Errorf("failed to scan %s: %w", path, err)
		}

		if len(fileReport.Patterns) > 0 {
			report.Files = append(report.Files, fileReport)
			report.TotalPatterns += len(fileReport.Patterns)

			// Validate each pattern
			for _, pattern := range fileReport.Patterns {
				safe := regret.IsSafe(pattern)
				if safe {
					report.SafePatterns++
				} else {
					report.UnsafePatterns++

					// Get detailed issues
					issues, _ := regret.Validate(pattern)
					for _, issue := range issues {
						fileReport.Issues = append(fileReport.Issues, Issue{
							Pattern:    pattern,
							Type:       issue.Type.String(),
							Severity:   issue.Severity.String(),
							Message:    issue.Message,
							Suggestion: issue.Suggestion,
						})
					}
				}
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Generate summary
	report.Summary = fmt.Sprintf("Scanned %d patterns: %d safe, %d unsafe",
		report.TotalPatterns, report.SafePatterns, report.UnsafePatterns)

	// Set exit code
	if report.UnsafePatterns > 0 {
		report.ExitCode = 1
		report.FailureMessages = append(report.FailureMessages,
			fmt.Sprintf("Found %d unsafe regex patterns", report.UnsafePatterns))
	} else {
		report.ExitCode = 0
	}

	return report, nil
}

// scanFile extracts regex patterns from a source file.
func scanFile(path string) (FileReport, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return FileReport{}, err
	}

	patterns := extractRegexPatterns(string(content))

	return FileReport{
		Path:     path,
		Patterns: patterns,
		Issues:   make([]Issue, 0),
	}, nil
}

// extractRegexPatterns extracts regex patterns from source code.
// This is a simplified version - a real implementation would use AST parsing.
func extractRegexPatterns(content string) []string {
	var patterns []string

	// Look for common regex compilation patterns
	regexPatterns := []string{
		`regexp\.MustCompile\("([^"]+)"\)`,
		`regexp\.Compile\("([^"]+)"\)`,
		`regexp\.MustCompile\(\x60([^\x60]+)\x60\)`, // backticks
		`regexp\.Compile\(\x60([^\x60]+)\x60\)`,     // backticks
	}

	for _, pattern := range regexPatterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindAllStringSubmatch(content, -1)
		for _, match := range matches {
			if len(match) > 1 {
				patterns = append(patterns, match[1])
			}
		}
	}

	return dedup(patterns)
}

// PreCommitHook validates regex patterns in staged files.
// Returns exit code 0 if all patterns are safe, 1 if any are unsafe.
func PreCommitHook(stagedFiles []string) int {
	hasUnsafe := false

	for _, file := range stagedFiles {
		// Skip non-code files
		if !isCodeFile(file) {
			continue
		}

		fileReport, err := scanFile(file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error scanning %s: %v\n", file, err)
			continue
		}

		for _, pattern := range fileReport.Patterns {
			safe := regret.IsSafe(pattern)
			if !safe {
				hasUnsafe = true
				fmt.Printf("❌ Unsafe regex in %s: %s\n", file, pattern)

				// Show issues
				issues, _ := regret.Validate(pattern)
				for _, issue := range issues {
					fmt.Printf("   %s: %s\n", issue.Severity, issue.Message)
					if issue.Suggestion != "" {
						fmt.Printf("   Suggestion: %s\n", issue.Suggestion)
					}
				}
			}
		}
	}

	if hasUnsafe {
		fmt.Println("\n⚠️  Unsafe regex patterns detected. Commit blocked.")
		fmt.Println("Fix the patterns or use --no-verify to bypass this check.")
		return 1
	}

	return 0
}

// GenerateJSONReport generates a JSON report suitable for CI/CD artifacts.
func GenerateJSONReport(report *CICDReport, outputPath string) error {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(outputPath, data, 0644)
}

// Helper functions

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func dedup(slice []string) []string {
	seen := make(map[string]bool)
	result := make([]string, 0)

	for _, item := range slice {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}

	return result
}

func isCodeFile(path string) bool {
	codeExtensions := []string{".go", ".js", ".ts", ".py", ".java", ".rb", ".php", ".cs"}
	ext := filepath.Ext(path)
	return contains(codeExtensions, ext)
}

// PrintReport prints a human-readable report to stdout.
func PrintReport(report *CICDReport) {
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("📋 Regex Security Audit Report")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("\nTotal Patterns: %d\n", report.TotalPatterns)
	fmt.Printf("✅ Safe: %d\n", report.SafePatterns)
	fmt.Printf("❌ Unsafe: %d\n\n", report.UnsafePatterns)

	if report.UnsafePatterns > 0 {
		fmt.Println("Unsafe Patterns Found:")
		for _, file := range report.Files {
			if len(file.Issues) > 0 {
				fmt.Printf("\n📄 %s\n", file.Path)
				for _, issue := range file.Issues {
					fmt.Printf("  ❌ Pattern: %s\n", issue.Pattern)
					fmt.Printf("     %s: %s\n", strings.ToUpper(issue.Severity), issue.Message)
					if issue.Suggestion != "" {
						fmt.Printf("     💡 %s\n", issue.Suggestion)
					}
				}
			}
		}
	}

	fmt.Println("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("Summary: %s\n", report.Summary)
	if report.ExitCode != 0 {
		fmt.Println("⚠️  Build should FAIL - unsafe patterns detected")
	} else {
		fmt.Println("✅ Build can PASS - all patterns safe")
	}
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
}
