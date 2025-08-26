// Package validator provides the main validation and fixing functionality for editorlint.
//
// This package orchestrates the entire validation and fixing process by:
// - Discovering and parsing EditorConfig files
// - Resolving configuration for specific files
// - Applying validation rules and fix functions
// - Handling both single files and directory trees
//
// The Validator type is the main entry point that coordinates between the config
// resolution and rules application.
package validator

import (
  "fmt"
  "os"
  "path/filepath"
  "regexp"
  "runtime"
  "strings"
  "sync"

  "github.com/dobbo-ca/editorlint/pkg/config"
  "github.com/dobbo-ca/editorlint/pkg/output"
  "github.com/dobbo-ca/editorlint/pkg/rules"
)

// Config holds configuration options for the validator.
type Config struct {
  // CustomConfigPath specifies an alternative .editorconfig file to use instead
  // of searching hierarchically. If empty, uses standard EditorConfig discovery.
  CustomConfigPath string

  // Recursive indicates whether to process files in subdirectories when
  // validating a directory target.
  Recursive        bool

  // Fix indicates whether to automatically fix validation errors instead
  // of just reporting them.
  Fix              bool

  // OutputFormat specifies how to format output: default, tabular, json, quiet
  OutputFormat     string

  // Workers specifies the number of parallel workers for file processing.
  // If 0, uses runtime.NumCPU().
  Workers          int

  // Quiet enables minimal output mode
  Quiet            bool

  // ExcludePatterns specifies glob patterns for files/directories to exclude
  ExcludePatterns  []string
}

// Validator handles file validation and fixing according to EditorConfig rules.
// It coordinates between configuration resolution and rule application.
type Validator struct {
  config    Config
  formatter *output.Formatter
  workers   int
}

// New creates a new validator with the given configuration.
func New(cfg Config) *Validator {
  // Set default workers if not specified
  workers := cfg.Workers
  if workers <= 0 {
    workers = runtime.NumCPU()
  }

  formatter := output.NewFormatter(cfg.OutputFormat, cfg.Quiet)

  return &Validator{
    config:    cfg,
    formatter: formatter,
    workers:   workers,
  }
}

// ValidateTarget validates a target file or directory according to EditorConfig rules.
//
// If target is a file, validates that single file. If target is a directory,
// validates all eligible files in the directory (recursively if Recursive is true).
// When Fix is true, automatically fixes validation errors instead of reporting them.
//
// Returns an error if validation fails or if any validation errors are found
// (in non-fix mode).
func (v *Validator) ValidateTarget(target string) error {
  // Check if target is a file or directory
  info, err := os.Stat(target)
  if err != nil {
    return fmt.Errorf("cannot access target: %w", err)
  }

  if info.IsDir() {
    return v.validateDirectory(target)
  } else {
    return v.validateSingleFile(target)
  }
}

func (v *Validator) validateDirectory(directory string) error {
  // Check if .editorconfig exists (unless using custom config)
  if v.config.CustomConfigPath == "" {
    if err := v.checkForEditorConfig(directory); err != nil {
      return err
    }
  }

  // Print progress unless in quiet mode
  if !v.config.Quiet {
    mode := "Validating"
    if v.config.Fix {
      mode = "Fixing"
    }
    fmt.Printf("%s directory: %s (recursive: %v)\n", mode, directory, v.config.Recursive)
  }

  if v.config.Fix {
    // Fix mode: fix all validation errors
    fixed, totalFiles, err := v.fixFilesParallel(directory)
    if err != nil {
      return err
    }

    result := &output.Result{
      FixedFiles: fixed,
      TotalFiles: totalFiles,
      Success:    len(fixed) == 0, // Success if no fixes were needed
      Mode:       "fix",
    }

    v.formatter.FormatResults(result)
    return nil
  } else {
    // Validate mode: report validation errors
    errors, totalFiles, err := v.validateFilesParallel(directory)
    if err != nil {
      return err
    }

    result := &output.Result{
      Errors:     errors,
      TotalFiles: totalFiles,
      Success:    len(errors) == 0,
      Mode:       "validate",
    }

    v.formatter.FormatResults(result)

    if len(errors) > 0 {
      return fmt.Errorf("validation failed with %d errors", len(errors))
    }

    return nil
  }
}

func (v *Validator) validateSingleFile(filePath string) error {
  // Print progress unless in quiet mode
  if !v.config.Quiet {
    mode := "Validating"
    if v.config.Fix {
      mode = "Fixing"
    }
    fmt.Printf("%s file: %s\n", mode, filePath)
  }

  if v.config.Fix {
    // Fix mode: fix validation errors in single file
    fixed, err := v.fixSingleFile(filePath)
    if err != nil {
      return err
    }

    var fixedFiles []string
    if fixed {
      fixedFiles = []string{filePath}
    }

    result := &output.Result{
      FixedFiles: fixedFiles,
      TotalFiles: 1,
      Success:    !fixed, // Success if no fixes were needed
      Mode:       "fix",
    }

    v.formatter.FormatResults(result)
    return nil
  } else {
    // Validate mode: report validation errors
    errors, err := v.validateSingleFileErrors(filePath)
    if err != nil {
      return err
    }

    result := &output.Result{
      Errors:     errors,
      TotalFiles: 1,
      Success:    len(errors) == 0,
      Mode:       "validate",
    }

    v.formatter.FormatResults(result)

    if len(errors) > 0 {
      return fmt.Errorf("validation failed with %d errors", len(errors))
    }

    return nil
  }
}

// validateFiles validates all files in the given directory against editorconfig rules
func (v *Validator) validateFiles(directory string) ([]rules.ValidationError, error) {
  var errors []rules.ValidationError

  walkErr := filepath.Walk(directory, func(path string, info os.FileInfo, walkErr error) error {
    if walkErr != nil {
      return walkErr
    }

    // Skip directories
    if info.IsDir() {
      // Check if directory should be ignored
      if v.shouldIgnore(path) {
        return filepath.SkipDir
      }
      // If not recursive, skip subdirectories
      if !v.config.Recursive && path != directory {
        return filepath.SkipDir
      }
      return nil
    }

    // Skip .editorconfig files themselves
    if info.Name() == ".editorconfig" {
      return nil
    }

    // Skip hidden files and directories (optional - could be configurable)
    if strings.HasPrefix(info.Name(), ".") {
      return nil
    }

    // Check if file should be ignored
    if v.shouldIgnore(path) {
      return nil
    }

    // Skip binary files and executable files
    if isBinaryFile(path, info) {
      return nil
    }

    // Convert to absolute path for config resolution
    absPath, err := filepath.Abs(path)
    if err != nil {
      return fmt.Errorf("failed to get absolute path for %s: %w", path, err)
    }

    // Find applicable editorconfig files for this file
    var configs []*config.EditorConfig
    var configErr error

    if v.config.CustomConfigPath != "" {
      configs, configErr = config.FindEditorConfigsWithCustomConfig(absPath, v.config.CustomConfigPath)
    } else {
      configs, configErr = config.FindEditorConfigs(absPath)
    }

    if configErr != nil {
      return fmt.Errorf("failed to find editorconfig for %s: %w", path, configErr)
    }

    if len(configs) == 0 {
      // No .editorconfig found
      return fmt.Errorf(".editorconfig file not found in directory hierarchy for %s", path)
    }

    // Resolve configuration for this specific file
    resolvedConfig, err := config.ResolveConfigForFile(absPath, configs)
    if err != nil {
      return fmt.Errorf("failed to resolve config for %s: %w", path, err)
    }

    // Validate the file against the resolved configuration (use original path for error messages)
    fileErrors := v.validateFile(path, resolvedConfig)
    errors = append(errors, fileErrors...)

    return nil
  })

  return errors, walkErr
}

// validateFile validates a single file against its resolved editorconfig
func (v *Validator) validateFile(filePath string, cfg *config.ResolvedConfig) []rules.ValidationError {
  var errors []rules.ValidationError

  // Read the file
  content, err := os.ReadFile(filePath)
  if err != nil {
    errors = append(errors, rules.ValidationError{
      FilePath: filePath,
      Rule:     "file_access",
      Message:  fmt.Sprintf("could not read file: %v", err),
    })
    return errors
  }

  // Run all validation checks
  validators := rules.GetAllValidators()

  for _, validator := range validators {
    if err := validator(filePath, content, cfg); err != nil {
      errors = append(errors, *err)
    }
  }

  return errors
}

// Additional methods for fixing, single file validation, etc.
func (v *Validator) validateSingleFileErrors(filePath string) ([]rules.ValidationError, error) {
  // Convert to absolute path for config resolution
  absPath, err := filepath.Abs(filePath)
  if err != nil {
    return nil, fmt.Errorf("failed to get absolute path for %s: %w", filePath, err)
  }

  // Find applicable editorconfig files for this file
  var configs []*config.EditorConfig

  if v.config.CustomConfigPath != "" {
    configs, err = config.FindEditorConfigsWithCustomConfig(absPath, v.config.CustomConfigPath)
  } else {
    configs, err = config.FindEditorConfigs(absPath)
  }

  if err != nil {
    return nil, fmt.Errorf("failed to find editorconfig for %s: %w", filePath, err)
  }

  if len(configs) == 0 {
    return nil, fmt.Errorf(".editorconfig file not found in directory hierarchy for %s", filePath)
  }

  // Resolve configuration for this specific file
  resolvedConfig, err := config.ResolveConfigForFile(absPath, configs)
  if err != nil {
    return nil, fmt.Errorf("failed to resolve config for %s: %w", filePath, err)
  }

  // Validate the file against the resolved configuration
  errors := v.validateFile(filePath, resolvedConfig)
  return errors, nil
}

func (v *Validator) fixSingleFile(filePath string) (bool, error) {
  // Convert to absolute path for config resolution
  absPath, err := filepath.Abs(filePath)
  if err != nil {
    return false, fmt.Errorf("failed to get absolute path for %s: %w", filePath, err)
  }

  // Find applicable editorconfig files for this file
  var configs []*config.EditorConfig

  if v.config.CustomConfigPath != "" {
    configs, err = config.FindEditorConfigsWithCustomConfig(absPath, v.config.CustomConfigPath)
  } else {
    configs, err = config.FindEditorConfigs(absPath)
  }

  if err != nil {
    return false, fmt.Errorf("failed to find editorconfig for %s: %w", filePath, err)
  }

  if len(configs) == 0 {
    return false, fmt.Errorf(".editorconfig file not found in directory hierarchy for %s", filePath)
  }

  // Resolve configuration for this specific file
  resolvedConfig, err := config.ResolveConfigForFile(absPath, configs)
  if err != nil {
    return false, fmt.Errorf("failed to resolve config for %s: %w", filePath, err)
  }

  // Read the file
  content, err := os.ReadFile(filePath)
  if err != nil {
    return false, fmt.Errorf("could not read file %s: %w", filePath, err)
  }

  // Apply all fixers
  fixers := rules.GetAllFixers()
  modified := false

  for _, fixer := range fixers {
    newContent, changed, err := fixer(filePath, content, resolvedConfig)
    if err != nil {
      return false, fmt.Errorf("failed to apply fixer to %s: %w", filePath, err)
    }
    if changed {
      content = newContent
      modified = true
    }
  }

  // Write back to file if modified
  if modified {
    err = os.WriteFile(filePath, content, 0644)
    if err != nil {
      return false, fmt.Errorf("failed to write fixed file %s: %w", filePath, err)
    }
  }

  return modified, nil
}

func (v *Validator) fixFiles(directory string) ([]string, error) {
  var fixedFiles []string

  walkErr := filepath.Walk(directory, func(path string, info os.FileInfo, walkErr error) error {
    if walkErr != nil {
      return walkErr
    }

    // Skip directories
    if info.IsDir() {
      // Check if directory should be ignored
      if v.shouldIgnore(path) {
        return filepath.SkipDir
      }
      // If not recursive, skip subdirectories
      if !v.config.Recursive && path != directory {
        return filepath.SkipDir
      }
      return nil
    }

    // Skip .editorconfig files themselves
    if info.Name() == ".editorconfig" {
      return nil
    }

    // Skip hidden files and directories (optional - could be configurable)
    if strings.HasPrefix(info.Name(), ".") {
      return nil
    }

    // Check if file should be ignored
    if v.shouldIgnore(path) {
      return nil
    }

    // Skip binary files and executable files
    if isBinaryFile(path, info) {
      return nil
    }

    // Try to fix the file
    fixed, err := v.fixSingleFile(path)
    if err != nil {
      return fmt.Errorf("failed to fix %s: %w", path, err)
    }

    if fixed {
      fixedFiles = append(fixedFiles, path)
    }

    return nil
  })

  return fixedFiles, walkErr
}

func (v *Validator) checkForEditorConfig(directory string) error {
  // If using custom config file, check that it exists
  if v.config.CustomConfigPath != "" {
    if _, err := os.Stat(v.config.CustomConfigPath); err != nil {
      return fmt.Errorf("custom config file not found: %s", v.config.CustomConfigPath)
    }
    return nil
  }

  // Original logic for hierarchical search
  absDir, err := filepath.Abs(directory)
  if err != nil {
    return fmt.Errorf("failed to get absolute path: %w", err)
  }

  dir := absDir
  for {
    configPath := filepath.Join(dir, ".editorconfig")

    if _, err := os.Stat(configPath); err == nil {
      return nil // Found .editorconfig
    }

    parent := filepath.Dir(dir)
    if parent == dir {
      // Reached filesystem root
      break
    }
    dir = parent
  }

  return fmt.Errorf(".editorconfig file not found in directory hierarchy starting from %s", directory)
}

// isBinaryFile checks if a file should be skipped (binary or executable)
func isBinaryFile(filePath string, info os.FileInfo) bool {
  ext := strings.ToLower(filepath.Ext(filePath))

  // Skip executable files with no extension
  if ext == "" && info.Mode()&0111 != 0 {
    return true
  }

  // Only process known text file extensions
  textExtensions := map[string]bool{
    ".go":   true, ".py":   true, ".js":   true, ".ts":   true,
    ".html": true, ".css":  true, ".scss": true, ".sass": true,
    ".json": true, ".xml":  true, ".yaml": true, ".yml":  true,
    ".md":   true, ".txt":  true, ".csv":  true, ".sql":  true,
    ".sh":   true, ".bash": true, ".zsh":  true, ".fish": true,
    ".c":    true, ".cpp":  true, ".h":    true, ".hpp":  true,
    ".java": true, ".kt":   true, ".rs":   true, ".rb":   true,
    ".php":  true, ".swift": true, ".dart": true, ".r":   true,
    ".tex":  true, ".lua":  true, ".vim":  true, ".ini":  true,
    ".conf": true, ".cfg":  true, ".toml": true, ".lock": true,
  }

  // If it has a known text extension, it's not binary
  if textExtensions[ext] {
    return false
  }

  // If it has no extension or unknown extension, check for null bytes
  file, err := os.Open(filePath)
  if err != nil {
    return true // If we can't read it, skip it
  }
  defer file.Close()

  buffer := make([]byte, 512)
  n, _ := file.Read(buffer)

  // Check for null bytes (indicates binary content)
  for i := 0; i < n; i++ {
    if buffer[i] == 0 {
      return true
    }
  }

  return false
}

// FileJob represents a file processing job
type FileJob struct {
  Path string
  Info os.FileInfo
}

// validateFilesParallel validates files in parallel using worker goroutines
func (v *Validator) validateFilesParallel(directory string) ([]rules.ValidationError, int, error) {
  // Collect all files to process
  files, err := v.collectFiles(directory)
  if err != nil {
    return nil, 0, err
  }

  if len(files) == 0 {
    return []rules.ValidationError{}, 0, nil
  }

  // Create channels for job distribution and result collection
  jobs := make(chan FileJob, len(files))
  results := make(chan []rules.ValidationError, len(files))

  // Start worker goroutines
  var wg sync.WaitGroup
  for i := 0; i < v.workers; i++ {
    wg.Add(1)
    go func() {
      defer wg.Done()
      for job := range jobs {
        errors := v.validateSingleFileSync(job.Path)
        results <- errors
      }
    }()
  }

  // Send jobs to workers
  go func() {
    defer close(jobs)
    for _, file := range files {
      jobs <- file
    }
  }()

  // Wait for all workers to complete
  go func() {
    wg.Wait()
    close(results)
  }()

  // Collect results
  var allErrors []rules.ValidationError
  for errors := range results {
    allErrors = append(allErrors, errors...)
  }

  return allErrors, len(files), nil
}

// fixFilesParallel fixes files in parallel using worker goroutines
func (v *Validator) fixFilesParallel(directory string) ([]string, int, error) {
  // Collect all files to process
  files, err := v.collectFiles(directory)
  if err != nil {
    return nil, 0, err
  }

  if len(files) == 0 {
    return []string{}, 0, nil
  }

  // Create channels for job distribution and result collection
  jobs := make(chan FileJob, len(files))
  results := make(chan string, len(files)) // Empty string means no fix needed

  // Start worker goroutines
  var wg sync.WaitGroup
  for i := 0; i < v.workers; i++ {
    wg.Add(1)
    go func() {
      defer wg.Done()
      for job := range jobs {
        fixed, err := v.fixSingleFile(job.Path)
        if err != nil {
          // Log error but continue processing other files
          continue
        }
        if fixed {
          results <- job.Path
        } else {
          results <- "" // No fix needed
        }
      }
    }()
  }

  // Send jobs to workers
  go func() {
    defer close(jobs)
    for _, file := range files {
      jobs <- file
    }
  }()

  // Wait for all workers to complete
  go func() {
    wg.Wait()
    close(results)
  }()

  // Collect results
  var fixedFiles []string
  for result := range results {
    if result != "" { // Non-empty means file was fixed
      fixedFiles = append(fixedFiles, result)
    }
  }

  return fixedFiles, len(files), nil
}

// collectFiles gathers all files that should be processed
func (v *Validator) collectFiles(directory string) ([]FileJob, error) {
  var files []FileJob

  err := filepath.Walk(directory, func(path string, info os.FileInfo, walkErr error) error {
    if walkErr != nil {
      return walkErr
    }

    // Skip directories
    if info.IsDir() {
      // Check if directory should be ignored
      if v.shouldIgnore(path) {
        return filepath.SkipDir
      }
      // If not recursive, skip subdirectories
      if !v.config.Recursive && path != directory {
        return filepath.SkipDir
      }
      return nil
    }

    // Skip .editorconfig files themselves
    if info.Name() == ".editorconfig" {
      return nil
    }

    // Skip hidden files and directories
    if strings.HasPrefix(info.Name(), ".") {
      return nil
    }

    // Check if file should be ignored
    if v.shouldIgnore(path) {
      return nil
    }

    // Skip binary files and executable files
    if isBinaryFile(path, info) {
      return nil
    }

    files = append(files, FileJob{Path: path, Info: info})
    return nil
  })

  return files, err
}

// validateSingleFileSync performs synchronous validation of a single file
func (v *Validator) validateSingleFileSync(filePath string) []rules.ValidationError {
  errors, err := v.validateSingleFileErrors(filePath)
  if err != nil {
    // Convert error to ValidationError for consistent handling
    return []rules.ValidationError{{
      FilePath: filePath,
      Rule:     "file_access",
      Message:  err.Error(),
    }}
  }
  return errors
}

// shouldIgnore checks if a file path should be ignored based on ignore patterns
func (v *Validator) shouldIgnore(filePath string) bool {
  if len(v.config.ExcludePatterns) == 0 {
    return false
  }

  // Convert to forward slashes for consistent matching across platforms
  normalizedPath := filepath.ToSlash(filePath)

  for _, pattern := range v.config.ExcludePatterns {
    // Convert glob pattern to regex for more powerful matching
    regexPattern, err := config.ConvertPatternToRegex(pattern)
    if err != nil {
      continue // Skip invalid patterns
    }

    // Check if the path matches the regex pattern
    matched, err := regexp.MatchString(regexPattern, normalizedPath)
    if err == nil && matched {
      return true
    }

    // Also check relative paths (remove leading directories)
    pathParts := strings.Split(normalizedPath, "/")
    for i := 0; i < len(pathParts); i++ {
      relativePath := strings.Join(pathParts[i:], "/")
      matched, err := regexp.MatchString(regexPattern, relativePath)
      if err == nil && matched {
        return true
      }
    }
  }

  return false
}
