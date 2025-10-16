package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/fatih/color"
	"github.com/theakshaypant/regret"
)

// Formatter handles output formatting
type Formatter struct {
	writer  io.Writer
	format  string
	noColor bool
}

// NewFormatter creates a new formatter
func NewFormatter(format string, noColor bool) *Formatter {
	if noColor {
		color.NoColor = true
	}

	return &Formatter{
		writer:  os.Stdout,
		format:  format,
		noColor: noColor,
	}
}

// CheckResult represents the result of a check command
type CheckResult struct {
	Pattern    string
	Safe       bool
	Complexity string
	Score      int
	Issues     []regret.Issue
}

// AnalysisResult represents the result of an analyze command
type AnalysisResult struct {
	Pattern string
	Score   *regret.ComplexityScore
	Issues  []regret.Issue
}

// ScanResult represents the result of a scan command
type ScanResult struct {
	TotalFiles     int
	ScannedFiles   int
	TotalPatterns  int
	DangerousCount int
	Findings       []Finding
}

// Finding represents a single pattern finding in a file
type Finding struct {
	File    string
	Line    int
	Column  int
	Pattern string
	Issue   string
}

// FormatCheckResult formats a check result
func (f *Formatter) FormatCheckResult(result *CheckResult) error {
	switch f.format {
	case "json":
		return f.formatCheckJSON(result)
	case "table":
		return f.formatCheckTable(result)
	default:
		return f.formatCheckText(result)
	}
}

func (f *Formatter) formatCheckText(result *CheckResult) error {
	if result.Safe {
		fmt.Fprintf(f.writer, "%s Pattern is safe\n", f.colorize("✓", color.FgGreen))
		fmt.Fprintf(f.writer, "Complexity: %s, Score: %d/100\n", result.Complexity, result.Score)
	} else {
		fmt.Fprintf(f.writer, "%s Pattern is UNSAFE\n", f.colorize("✗", color.FgRed))
		fmt.Fprintf(f.writer, "Complexity: %s, Score: %d/100\n",
			f.colorize(result.Complexity, color.FgRed), result.Score)

		if len(result.Issues) > 0 {
			fmt.Fprintf(f.writer, "\nIssues found:\n")
			for _, issue := range result.Issues {
				severity := f.getSeveritySymbol(issue.Severity)
				fmt.Fprintf(f.writer, "  %s %s: %s\n", severity, issue.Type, issue.Message)
			}
		}
	}
	return nil
}

func (f *Formatter) formatCheckJSON(result *CheckResult) error {
	data := map[string]interface{}{
		"pattern":    result.Pattern,
		"safe":       result.Safe,
		"complexity": result.Complexity,
		"score":      result.Score,
		"issues":     result.Issues,
	}

	enc := json.NewEncoder(f.writer)
	enc.SetIndent("", "  ")
	return enc.Encode(data)
}

func (f *Formatter) formatCheckTable(result *CheckResult) error {
	// Simple table format
	fmt.Fprintln(f.writer, "┌──────────────┬────────────┬───────┐")
	fmt.Fprintln(f.writer, "│ Safe         │ Complexity │ Score │")
	fmt.Fprintln(f.writer, "├──────────────┼────────────┼───────┤")

	safeStr := f.colorize("No", color.FgRed)
	if result.Safe {
		safeStr = f.colorize("Yes", color.FgGreen)
	}

	fmt.Fprintf(f.writer, "│ %-12s │ %-10s │ %-5d │\n",
		safeStr, result.Complexity, result.Score)
	fmt.Fprintln(f.writer, "└──────────────┴────────────┴───────┘")

	return nil
}

// FormatAnalysisResult formats an analysis result
func (f *Formatter) FormatAnalysisResult(result *AnalysisResult) error {
	switch f.format {
	case "json":
		return f.formatAnalysisJSON(result)
	case "table":
		return f.formatAnalysisTable(result)
	default:
		return f.formatAnalysisText(result)
	}
}

func (f *Formatter) formatAnalysisText(result *AnalysisResult) error {
	score := result.Score

	fmt.Fprintf(f.writer, "Pattern: %s\n", f.colorize(result.Pattern, color.FgCyan))
	fmt.Fprintf(f.writer, "Status: %s\n", f.getSafetyStatus(score.Safe))
	fmt.Fprintf(f.writer, "Complexity: %s\n", f.colorize(score.TimeComplexity.String(), f.getComplexityColor(score.TimeComplexity)))
	fmt.Fprintf(f.writer, "Score: %d/100\n", score.Overall)

	if score.HasEDA {
		fmt.Fprintf(f.writer, "⚠️  EDA (Exponential Degree of Ambiguity) detected\n")
	}
	if score.HasIDA {
		fmt.Fprintf(f.writer, "⚠️  IDA (Infinite Degree of Ambiguity) detected\n")
	}

	if score.PolynomialDegree > 1 {
		fmt.Fprintf(f.writer, "Polynomial Degree: %d\n", score.PolynomialDegree)
	}

	fmt.Fprintf(f.writer, "\nMetrics:\n")
	fmt.Fprintf(f.writer, "  Nesting Depth: %d\n", score.Metrics.NestingDepth)
	fmt.Fprintf(f.writer, "  Quantifiers: %d\n", score.Metrics.QuantifierCount)
	fmt.Fprintf(f.writer, "  Alternations: %d\n", score.Metrics.AlternationCount)

	if len(result.Issues) > 0 {
		fmt.Fprintf(f.writer, "\nIssues:\n")
		for _, issue := range result.Issues {
			severity := f.getSeveritySymbol(issue.Severity)
			fmt.Fprintf(f.writer, "  %s %s: %s\n", severity, issue.Type, issue.Message)
			if issue.Suggestion != "" {
				fmt.Fprintf(f.writer, "     Suggestion: %s\n", issue.Suggestion)
			}
		}
	}

	if score.Explanation != "" {
		fmt.Fprintf(f.writer, "\nExplanation: %s\n", score.Explanation)
	}

	return nil
}

func (f *Formatter) formatAnalysisJSON(result *AnalysisResult) error {
	enc := json.NewEncoder(f.writer)
	enc.SetIndent("", "  ")
	return enc.Encode(result)
}

func (f *Formatter) formatAnalysisTable(result *AnalysisResult) error {
	score := result.Score

	fmt.Fprintln(f.writer, "┌────────────────┬─────────────────────────┐")
	fmt.Fprintln(f.writer, "│ Metric         │ Value                   │")
	fmt.Fprintln(f.writer, "├────────────────┼─────────────────────────┤")

	safeStr := "No"
	if score.Safe {
		safeStr = "Yes"
	}
	fmt.Fprintf(f.writer, "│ Safe           │ %-23s │\n", safeStr)
	fmt.Fprintf(f.writer, "│ Complexity     │ %-23s │\n", score.TimeComplexity.String())
	fmt.Fprintf(f.writer, "│ Score          │ %-23d │\n", score.Overall)
	fmt.Fprintf(f.writer, "│ Has EDA        │ %-23v │\n", score.HasEDA)
	fmt.Fprintf(f.writer, "│ Has IDA        │ %-23v │\n", score.HasIDA)

	fmt.Fprintln(f.writer, "└────────────────┴─────────────────────────┘")

	return nil
}

// FormatScanResult formats a scan result
func (f *Formatter) FormatScanResult(result *ScanResult) error {
	switch f.format {
	case "json":
		return f.formatScanJSON(result)
	default:
		return f.formatScanText(result)
	}
}

func (f *Formatter) formatScanText(result *ScanResult) error {
	fmt.Fprintf(f.writer, "Scanned %d files\n", result.ScannedFiles)
	fmt.Fprintf(f.writer, "Found %d regex patterns\n", result.TotalPatterns)

	if result.DangerousCount == 0 {
		fmt.Fprintf(f.writer, "%s No dangerous patterns found\n", f.colorize("✓", color.FgGreen))
	} else {
		fmt.Fprintf(f.writer, "%s Found %d dangerous pattern(s)\n",
			f.colorize("⚠", color.FgYellow), result.DangerousCount)

		fmt.Fprintln(f.writer, "\nFindings:")
		for _, finding := range result.Findings {
			fmt.Fprintf(f.writer, "  %s:%d:%d: %s\n",
				finding.File, finding.Line, finding.Column, finding.Pattern)
			if finding.Issue != "" {
				fmt.Fprintf(f.writer, "    Issue: %s\n", finding.Issue)
			}
		}
	}

	return nil
}

func (f *Formatter) formatScanJSON(result *ScanResult) error {
	enc := json.NewEncoder(f.writer)
	enc.SetIndent("", "  ")
	return enc.Encode(result)
}

// Helper functions

func (f *Formatter) colorize(text string, attr color.Attribute) string {
	if f.noColor {
		return text
	}
	return color.New(attr).Sprint(text)
}

func (f *Formatter) getSeveritySymbol(severity regret.Severity) string {
	switch severity {
	case regret.Critical:
		return f.colorize("⛔", color.FgRed)
	case regret.High:
		return f.colorize("🔴", color.FgRed)
	case regret.Medium:
		return f.colorize("🟡", color.FgYellow)
	case regret.Low:
		return f.colorize("🔵", color.FgBlue)
	default:
		return f.colorize("ℹ️", color.FgCyan)
	}
}

func (f *Formatter) getSafetyStatus(safe bool) string {
	if safe {
		return f.colorize("✓ SAFE", color.FgGreen)
	}
	return f.colorize("✗ UNSAFE", color.FgRed)
}

func (f *Formatter) getComplexityColor(complexity regret.Complexity) color.Attribute {
	switch complexity {
	case regret.Exponential:
		return color.FgRed
	case regret.Polynomial, regret.Cubic, regret.Quadratic:
		return color.FgYellow
	case regret.Linear:
		return color.FgGreen
	default:
		return color.FgWhite
	}
}

// PrintError prints an error message
func (f *Formatter) PrintError(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintf(os.Stderr, "%s %s\n", f.colorize("Error:", color.FgRed), msg)
}

// PrintWarning prints a warning message
func (f *Formatter) PrintWarning(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintf(f.writer, "%s %s\n", f.colorize("Warning:", color.FgYellow), msg)
}

// PrintInfo prints an info message
func (f *Formatter) PrintInfo(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintf(f.writer, "%s %s\n", f.colorize("Info:", color.FgCyan), msg)
}

// PrintSuccess prints a success message
func (f *Formatter) PrintSuccess(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintf(f.writer, "%s %s\n", f.colorize("✓", color.FgGreen), msg)
}
