// Package rules provides validation and fixing functions for EditorConfig rules.
package rules

import (
  "fmt"
  
  "github.com/cdobbyn/editorlint/pkg/config"
)

// ValidateInsertFinalNewline checks if the file ends with the appropriate newline character(s)
func ValidateInsertFinalNewline(filePath string, content []byte, cfg *config.ResolvedConfig) *ValidationError {
  // Only validate if insert_final_newline is explicitly set to true
  if cfg.InsertFinalNewline == nil || !*cfg.InsertFinalNewline {
    return nil
  }
  
  if len(content) == 0 {
    // Empty files should end with a newline if insert_final_newline is true
    return &ValidationError{
      FilePath: filePath,
      Rule:     "insert_final_newline",
      Message:  "empty file should end with a newline",
    }
  }
  
  // Determine what the file actually ends with
  lastChar := content[len(content)-1]
  var actualEnding string
  
  if len(content) >= 2 && content[len(content)-2] == '\r' && lastChar == '\n' {
    actualEnding = "crlf"
  } else if lastChar == '\r' {
    actualEnding = "cr"
  } else if lastChar == '\n' {
    actualEnding = "lf"
  } else {
    // File doesn't end with any recognized line ending
    return &ValidationError{
      FilePath: filePath,
      Rule:     "insert_final_newline",
      Message:  fmt.Sprintf("file should end with %s, but ends with character '%c' (0x%02x)", getEndOfLineDescription(cfg.EndOfLine), lastChar, lastChar),
    }
  }
  
  // Determine expected line ending
  expectedEnding := cfg.EndOfLine
  if expectedEnding == "" {
    expectedEnding = "lf" // Default to LF
  }
  
  // Check if actual matches expected
  if actualEnding != expectedEnding {
    return &ValidationError{
      FilePath: filePath,
      Rule:     "insert_final_newline",
      Message:  fmt.Sprintf("file should end with %s, but ends with %s", getEndOfLineDescription(expectedEnding), getEndOfLineDescription(actualEnding)),
    }
  }
  
  return nil
}

// FixInsertFinalNewline fixes the final newline in a file according to editorconfig rules
func FixInsertFinalNewline(filePath string, content []byte, cfg *config.ResolvedConfig) ([]byte, bool, error) {
  // Only fix if insert_final_newline is explicitly set to true
  if cfg.InsertFinalNewline == nil || !*cfg.InsertFinalNewline {
    return content, false, nil
  }
  
  // Determine expected line ending
  expectedEnding := cfg.EndOfLine
  if expectedEnding == "" {
    expectedEnding = "lf" // Default to LF
  }
  
  var expectedBytes []byte
  switch expectedEnding {
  case "crlf":
    expectedBytes = []byte("\r\n")
  case "cr":
    expectedBytes = []byte("\r")
  default: // "lf"
    expectedBytes = []byte("\n")
  }
  
  // Handle empty files
  if len(content) == 0 {
    return expectedBytes, true, nil
  }
  
  // Check what the file currently ends with
  lastChar := content[len(content)-1]
  var needsFix bool
  var newContent []byte
  
  if len(content) >= 2 && content[len(content)-2] == '\r' && lastChar == '\n' {
    // File ends with CRLF
    if expectedEnding != "crlf" {
      // Remove CRLF and add correct ending
      newContent = append(content[:len(content)-2], expectedBytes...)
      needsFix = true
    }
  } else if lastChar == '\r' {
    // File ends with CR
    if expectedEnding != "cr" {
      // Remove CR and add correct ending
      newContent = append(content[:len(content)-1], expectedBytes...)
      needsFix = true
    }
  } else if lastChar == '\n' {
    // File ends with LF
    if expectedEnding != "lf" {
      // Remove LF and add correct ending
      newContent = append(content[:len(content)-1], expectedBytes...)
      needsFix = true
    }
  } else {
    // File doesn't end with any line ending - add the expected one
    newContent = append(content, expectedBytes...)
    needsFix = true
  }
  
  if needsFix {
    return newContent, true, nil
  }
  
  return content, false, nil
}

// getEndOfLineDescription returns a human-readable description of line ending
func getEndOfLineDescription(endOfLine string) string {
  switch endOfLine {
  case "crlf":
    return "CRLF (\\r\\n)"
  case "cr":
    return "CR (\\r)"
  case "lf", "":
    return "LF (\\n)"
  default:
    return fmt.Sprintf("unknown line ending: %s", endOfLine)
  }
}