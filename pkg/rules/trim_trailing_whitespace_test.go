package rules

import (
  "testing"

  "github.com/cdobbyn/editorlint/pkg/config"
)

func TestValidateTrimTrailingWhitespace(t *testing.T) {
  tests := []struct {
    name                   string
    content                string
    trimTrailingWhitespace bool
    wantError              bool
  }{
    {
      name:                   "no trailing whitespace",
      content:                "package main\nfunc main() {\n}\n",
      trimTrailingWhitespace: true,
      wantError:              false,
    },
    {
      name:                   "trailing spaces",
      content:                "package main \nfunc main() {\n}\n",
      trimTrailingWhitespace: true,
      wantError:              true,
    },
    {
      name:                   "trailing tabs",
      content:                "package main\t\nfunc main() {\n}\n",
      trimTrailingWhitespace: true,
      wantError:              true,
    },
    {
      name:                   "trailing whitespace but rule disabled",
      content:                "package main \nfunc main() {\n}\n",
      trimTrailingWhitespace: false,
      wantError:              false,
    },
  }

  for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
      cfg := &config.ResolvedConfig{
        TrimTrailingWhitespace: &tt.trimTrailingWhitespace,
      }

      err := ValidateTrimTrailingWhitespace("test.go", []byte(tt.content), cfg)

      if tt.wantError && err == nil {
        t.Error("Expected validation error, but got none")
      }

      if !tt.wantError && err != nil {
        t.Errorf("Expected no validation error, but got: %v", err)
      }
    })
  }
}

func TestFixTrimTrailingWhitespace(t *testing.T) {
  tests := []struct {
    name                   string
    content                string
    trimTrailingWhitespace bool
    expectedContent        string
    expectFixed            bool
  }{
    {
      name:                   "remove trailing spaces",
      content:                "package main \nfunc main() {\n}\n",
      trimTrailingWhitespace: true,
      expectedContent:        "package main\nfunc main() {\n}\n",
      expectFixed:            true,
    },
    {
      name:                   "remove trailing tabs",
      content:                "package main\t\t\nfunc main() {\n}\n",
      trimTrailingWhitespace: true,
      expectedContent:        "package main\nfunc main() {\n}\n",
      expectFixed:            true,
    },
    {
      name:                   "no trailing whitespace",
      content:                "package main\nfunc main() {\n}\n",
      trimTrailingWhitespace: true,
      expectedContent:        "package main\nfunc main() {\n}\n",
      expectFixed:            false,
    },
    {
      name:                   "rule disabled",
      content:                "package main \nfunc main() {\n}\n",
      trimTrailingWhitespace: false,
      expectedContent:        "package main \nfunc main() {\n}\n",
      expectFixed:            false,
    },
  }

  for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
      cfg := &config.ResolvedConfig{
        TrimTrailingWhitespace: &tt.trimTrailingWhitespace,
      }

      newContent, fixed, err := FixTrimTrailingWhitespace("test.go", []byte(tt.content), cfg)
      if err != nil {
        t.Fatal(err)
      }

      if fixed != tt.expectFixed {
        t.Errorf("Expected fixed=%v, got %v", tt.expectFixed, fixed)
      }

      if string(newContent) != tt.expectedContent {
        t.Errorf("Expected content %q, got %q", tt.expectedContent, string(newContent))
      }
    })
  }
}
