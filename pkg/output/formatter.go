// Package output provides various formatting options for validation results.
package output

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/cdobbyn/editorlint/pkg/rules"
)

// OutputFormat represents different output formatting options
type OutputFormat string

const (
	FormatDefault  OutputFormat = "default"
	FormatTabular  OutputFormat = "tabular"
	FormatJSON     OutputFormat = "json"
	FormatQuiet    OutputFormat = "quiet"
)

// Result represents the validation results for output formatting
type Result struct {
	Errors      []rules.ValidationError
	FixedFiles  []string
	TotalFiles  int
	Success     bool
	Mode        string // "validate" or "fix"
}

// Formatter handles different output formats
type Formatter struct {
	format OutputFormat
	quiet  bool
}

// NewFormatter creates a new output formatter
func NewFormatter(format string, quiet bool) *Formatter {
	f := &Formatter{
		format: OutputFormat(format),
		quiet:  quiet,
	}

	// Override format if quiet mode is enabled
	if quiet {
		f.format = FormatQuiet
	}

	return f
}

// FormatResults outputs the validation results in the specified format
func (f *Formatter) FormatResults(result *Result) {
	switch f.format {
	case FormatJSON:
		f.formatJSON(result)
	case FormatTabular:
		f.formatTabular(result)
	case FormatQuiet:
		f.formatQuiet(result)
	default:
		f.formatDefault(result)
	}
}

// formatDefault outputs in the current default format
func (f *Formatter) formatDefault(result *Result) {
	if result.Mode == "fix" {
		f.formatFixResults(result)
	} else {
		f.formatValidationResults(result)
	}
}

func (f *Formatter) formatValidationResults(result *Result) {
	if result.Success {
		fmt.Printf("✓ All files pass editorconfig validation\n")
		return
	}

	// Group errors by rule
	errorsByRule := make(map[string][]rules.ValidationError)
	for _, err := range result.Errors {
		errorsByRule[err.Rule] = append(errorsByRule[err.Rule], err)
	}

	fmt.Printf("Found %d validation errors:\n\n", len(result.Errors))

	for rule, errors := range errorsByRule {
		fmt.Printf("📋 %s (%d files):\n", rule, len(errors))
		for _, err := range errors {
			fmt.Printf("  • %s - %s\n", err.FilePath, err.Message)
		}
		fmt.Println()
	}

	fmt.Printf("To fix these errors automatically, run with --fix flag\n")
}

func (f *Formatter) formatFixResults(result *Result) {
	if len(result.FixedFiles) > 0 {
		fmt.Printf("✅ Fixed %d files:\n", len(result.FixedFiles))
		for _, file := range result.FixedFiles {
			fmt.Printf("  • %s\n", file)
		}
	} else {
		fmt.Printf("✓ No fixes needed - all files already pass editorconfig validation\n")
	}
}

// formatTabular outputs results in a table format
func (f *Formatter) formatTabular(result *Result) {
	if result.Success {
		fmt.Printf("✓ All files pass editorconfig validation\n")
		return
	}

	if result.Mode == "fix" {
		f.formatFixResults(result)
		return
	}

	// Group errors by file
	errorsByFile := make(map[string][]rules.ValidationError)
	allRules := make(map[string]bool)

	for _, err := range result.Errors {
		errorsByFile[err.FilePath] = append(errorsByFile[err.FilePath], err)
		allRules[err.Rule] = true
	}

	// Sort rules for consistent output
	var ruleList []string
	for rule := range allRules {
		ruleList = append(ruleList, rule)
	}
	sort.Strings(ruleList)

	// Create table
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	// Header
	header := []string{"File"}
	header = append(header, ruleList...)
	fmt.Fprintln(w, strings.Join(header, "\t"))

	// Separator line
	separator := make([]string, len(header))
	for i := range separator {
		separator[i] = "---"
	}
	fmt.Fprintln(w, strings.Join(separator, "\t"))

	// Data rows - format paths for better display
	var files []string
	for file := range errorsByFile {
		files = append(files, file)
	}
	sort.Strings(files)

	// Format paths for tabular display
	displayPaths := f.formatPathsForTable(files)

	for i, file := range files {
		displayPath := displayPaths[i]
		row := []string{displayPath}
		fileErrors := errorsByFile[file]

		// Create map of rules for this file
		fileRules := make(map[string]string)
		for _, err := range fileErrors {
			fileRules[err.Rule] = "❌"
		}

		// Add status for each rule
		for _, rule := range ruleList {
			if status, exists := fileRules[rule]; exists {
				row = append(row, status)
			} else {
				row = append(row, "✅")
			}
		}

		fmt.Fprintln(w, strings.Join(row, "\t"))
	}

	w.Flush()
	fmt.Printf("\nFound %d validation errors in %d files\n", len(result.Errors), len(errorsByFile))
}

// formatPathsForTable optimizes file paths for tabular display
func (f *Formatter) formatPathsForTable(files []string) []string {
	if len(files) == 0 {
		return files
	}

	// Get working directory for relative path calculation
	wd, err := os.Getwd()
	if err != nil {
		wd = ""
	}

	// Convert to relative paths where possible
	relativePaths := make([]string, len(files))
	for i, file := range files {
		if wd != "" {
			if rel, err := filepath.Rel(wd, file); err == nil && !strings.HasPrefix(rel, "../") {
				relativePaths[i] = rel
			} else {
				relativePaths[i] = file
			}
		} else {
			relativePaths[i] = file
		}
	}

	// Calculate max reasonable column width (40% of terminal width, min 20, max 60)
	maxWidth := f.getMaxPathWidth()

	// Check if any paths exceed the max width
	needsTruncation := false
	for _, path := range relativePaths {
		if len(path) > maxWidth {
			needsTruncation = true
			break
		}
	}

	// If truncation is needed, apply smart shortening
	if needsTruncation {
		return f.shortenPaths(relativePaths, maxWidth)
	}

	return relativePaths
}

// getMaxPathWidth determines reasonable maximum width for file paths
func (f *Formatter) getMaxPathWidth() int {
	// Try to get terminal width, fallback to reasonable default
	// For now, use a conservative default of 50 characters
	// In a real implementation, you could use a library like:
	// - golang.org/x/crypto/ssh/terminal
	// - github.com/mattn/go-isatty with syscalls
	return 50
}

// shortenPaths intelligently shortens file paths to fit in the specified width
func (f *Formatter) shortenPaths(paths []string, maxWidth int) []string {
	shortened := make([]string, len(paths))

	for i, path := range paths {
		if len(path) <= maxWidth {
			shortened[i] = path
			continue
		}

		// Strategy 1: Try showing just the filename if directory structure is deep
		filename := filepath.Base(path)
		if len(filename) <= maxWidth-3 { // -3 for ".../"
			shortened[i] = ".../" + filename
			continue
		}

		// Strategy 2: Truncate in the middle, preserving start and end
		if maxWidth > 10 {
			start := path[:maxWidth/2-2]
			end := path[len(path)-(maxWidth/2-2):]
			shortened[i] = start + "..." + end
		} else {
			// Last resort: just truncate with ellipsis
			shortened[i] = path[:maxWidth-3] + "..."
		}
	}

	return shortened
}

// formatJSON outputs results in JSON format
func (f *Formatter) formatJSON(result *Result) {
	type jsonError struct {
		FilePath string `json:"file_path"`
		Rule     string `json:"rule"`
		Message  string `json:"message"`
	}

	type jsonResult struct {
		Success    bool        `json:"success"`
		Mode       string      `json:"mode"`
		TotalFiles int         `json:"total_files"`
		Errors     []jsonError `json:"errors,omitempty"`
		FixedFiles []string    `json:"fixed_files,omitempty"`
	}

	jsonErrors := make([]jsonError, len(result.Errors))
	for i, err := range result.Errors {
		jsonErrors[i] = jsonError{
			FilePath: err.FilePath,
			Rule:     err.Rule,
			Message:  err.Message,
		}
	}

	output := jsonResult{
		Success:    result.Success,
		Mode:       result.Mode,
		TotalFiles: result.TotalFiles,
		Errors:     jsonErrors,
		FixedFiles: result.FixedFiles,
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	encoder.Encode(output)
}

// formatQuiet outputs minimal results
func (f *Formatter) formatQuiet(result *Result) {
	if result.Mode == "fix" {
		if len(result.FixedFiles) > 0 {
			fmt.Printf("Fixed %d files\n", len(result.FixedFiles))
		} else {
			fmt.Printf("No fixes needed\n")
		}
	} else {
		if result.Success {
			fmt.Printf("✓ All files valid\n")
		} else {
			fmt.Printf("❌ %d errors found\n", len(result.Errors))
		}
	}
}
