// Package rules provides validation and fixing functions for EditorConfig rules.
//
// This package implements various EditorConfig properties validation and their
// corresponding fix functions. Each rule is implemented as a ValidatorFunc and FixerFunc
// that can be applied to file content based on resolved configuration.
package rules

import (
  "fmt"

  "github.com/dobbo-ca/editorlint/pkg/config"
)

// ValidationError represents a validation failure
type ValidationError struct {
  FilePath string
  Rule     string
  Message  string
}

func (e ValidationError) Error() string {
  return fmt.Sprintf("%s: %s violation - %s", e.FilePath, e.Rule, e.Message)
}

// ValidatorFunc is a function that validates a file against a specific rule
type ValidatorFunc func(string, []byte, *config.ResolvedConfig) *ValidationError

// FixerFunc is a function that fixes violations of a specific rule
type FixerFunc func(string, []byte, *config.ResolvedConfig) ([]byte, bool, error)
