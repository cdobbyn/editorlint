package rules

import (
  "bytes"
  "fmt"

  "github.com/dobbo-ca/editorlint/pkg/config"
)

// ValidateEndOfLine checks if all line endings in the file match the configured style
func ValidateEndOfLine(filePath string, content []byte, config *config.ResolvedConfig) *ValidationError {
  // Only validate if end_of_line is set
  if config.EndOfLine == "" {
    return nil
  }

  if len(content) == 0 {
    return nil // Empty files are fine
  }

  // Define expected line ending
  var expectedEnding []byte
  var expectedName string

  switch config.EndOfLine {
  case "lf":
    expectedEnding = []byte("\n")
    expectedName = "LF (\\n)"
  case "crlf":
    expectedEnding = []byte("\r\n")
    expectedName = "CRLF (\\r\\n)"
  case "cr":
    expectedEnding = []byte("\r")
    expectedName = "CR (\\r)"
  default:
    return nil // Unknown line ending style
  }

  // Find all line endings in the file
  lineEndings := findLineEndings(content)

  for _, ending := range lineEndings {
    if !bytes.Equal(ending.bytes, expectedEnding) {
      return &ValidationError{
        FilePath: filePath,
        Rule:     "end_of_line",
        Message:  fmt.Sprintf("line %d uses %s but should use %s", ending.line, ending.name, expectedName),
      }
    }
  }

  return nil
}

// FixEndOfLine converts all line endings to the configured style
func FixEndOfLine(filePath string, content []byte, config *config.ResolvedConfig) ([]byte, bool, error) {
  // Only fix if end_of_line is set
  if config.EndOfLine == "" {
    return content, false, nil
  }

  if len(content) == 0 {
    return content, false, nil
  }

  // Define target line ending
  var targetEnding []byte

  switch config.EndOfLine {
  case "lf":
    targetEnding = []byte("\n")
  case "crlf":
    targetEnding = []byte("\r\n")
  case "cr":
    targetEnding = []byte("\r")
  default:
    return content, false, nil // Unknown line ending style
  }

  // Convert all line endings to the target style
  result := normalizeLineEndings(content, targetEnding)

  // Check if any changes were made
  changed := !bytes.Equal(content, result)

  return result, changed, nil
}

// LineEnding represents a line ending found in the file
type LineEnding struct {
  bytes []byte
  name  string
  line  int
}

// findLineEndings finds all line endings in the content and their positions
func findLineEndings(content []byte) []LineEnding {
  var endings []LineEnding
  lineNum := 1

  for i := 0; i < len(content); i++ {
    if content[i] == '\r' {
      if i+1 < len(content) && content[i+1] == '\n' {
        // CRLF
        endings = append(endings, LineEnding{
          bytes: []byte("\r\n"),
          name:  "CRLF (\\r\\n)",
          line:  lineNum,
        })
        i++ // Skip the \n
      } else {
        // CR only
        endings = append(endings, LineEnding{
          bytes: []byte("\r"),
          name:  "CR (\\r)",
          line:  lineNum,
        })
      }
      lineNum++
    } else if content[i] == '\n' {
      // LF only
      endings = append(endings, LineEnding{
        bytes: []byte("\n"),
        name:  "LF (\\n)",
        line:  lineNum,
      })
      lineNum++
    }
  }

  return endings
}

// normalizeLineEndings converts all line endings in content to the target ending
func normalizeLineEndings(content []byte, targetEnding []byte) []byte {
  // First, normalize all line endings to LF
  // Replace CRLF with LF
  normalized := bytes.ReplaceAll(content, []byte("\r\n"), []byte("\n"))
  // Replace remaining CR with LF
  normalized = bytes.ReplaceAll(normalized, []byte("\r"), []byte("\n"))

  // If target is LF, we're done
  if bytes.Equal(targetEnding, []byte("\n")) {
    return normalized
  }

  // Convert LF to target ending
  result := bytes.ReplaceAll(normalized, []byte("\n"), targetEnding)
  return result
}
