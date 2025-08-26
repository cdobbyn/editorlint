package rules

import (
  "bytes"
  "fmt"

  "github.com/dobbo-ca/editorlint/pkg/config"
)

// ValidateTrimTrailingWhitespace checks if the file has trailing whitespace when it shouldn't
func ValidateTrimTrailingWhitespace(filePath string, content []byte, cfg *config.ResolvedConfig) *ValidationError {
  // Only validate if trim_trailing_whitespace is explicitly set to true
  if cfg.TrimTrailingWhitespace == nil || !*cfg.TrimTrailingWhitespace {
    return nil
  }

  if len(content) == 0 {
    return nil // Empty files are fine
  }

  lines := bytes.Split(content, []byte("\n"))

  // Check each line for trailing whitespace
  for i, line := range lines {
    // Skip the last line if it's empty (this would be after the final newline)
    if i == len(lines)-1 && len(line) == 0 {
      continue
    }

    // Check if line has trailing whitespace
    if len(line) > 0 && isWhitespace(line[len(line)-1]) {
      lineNum := i + 1
      return &ValidationError{
        FilePath: filePath,
        Rule:     "trim_trailing_whitespace",
        Message:  fmt.Sprintf("line %d has trailing whitespace", lineNum),
      }
    }
  }

  return nil
}

// FixTrimTrailingWhitespace removes trailing whitespace from all lines
func FixTrimTrailingWhitespace(filePath string, content []byte, cfg *config.ResolvedConfig) ([]byte, bool, error) {
  // Only fix if trim_trailing_whitespace is explicitly set to true
  if cfg.TrimTrailingWhitespace == nil || !*cfg.TrimTrailingWhitespace {
    return content, false, nil
  }

  if len(content) == 0 {
    return content, false, nil // Empty files are fine
  }

  // Determine line ending to preserve
  var lineEnding []byte
  if bytes.Contains(content, []byte("\r\n")) {
    lineEnding = []byte("\r\n")
  } else if bytes.Contains(content, []byte("\r")) {
    lineEnding = []byte("\r")
  } else {
    lineEnding = []byte("\n")
  }

  lines := bytes.Split(content, []byte("\n"))
  hasChanges := false

  for i, line := range lines {
    // Skip the last line if it's empty (this would be after the final newline)
    if i == len(lines)-1 && len(line) == 0 {
      continue
    }

    // Remove trailing whitespace
    trimmed := bytes.TrimRightFunc(line, func(r rune) bool {
      return r == ' ' || r == '\t'
    })

    if !bytes.Equal(line, trimmed) {
      lines[i] = trimmed
      hasChanges = true
    }
  }

  if !hasChanges {
    return content, false, nil
  }

  // Rejoin lines with original line ending
  var result bytes.Buffer
  for i, line := range lines {
    result.Write(line)
    // Add line ending except for the last empty line
    if i < len(lines)-1 {
      if i == len(lines)-2 && len(lines[len(lines)-1]) == 0 {
        // This is the second to last line and the last line is empty
        // So we're before the final newline
        result.Write(lineEnding)
      } else if i < len(lines)-2 || len(lines[len(lines)-1]) > 0 {
        // Not the last line, or the last line is not empty
        result.Write(lineEnding)
      }
    }
  }

  return result.Bytes(), true, nil
}

// isWhitespace checks if a byte is whitespace (space or tab)
func isWhitespace(b byte) bool {
  return b == ' ' || b == '\t'
}
