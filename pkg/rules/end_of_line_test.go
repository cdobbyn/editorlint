package rules

import (
  "testing"

  "github.com/cdobbyn/editorlint/pkg/config"
)

func TestValidateEndOfLine(t *testing.T) {
  tests := []struct {
    name      string
    content   string
    endOfLine string
    wantError bool
  }{
    {
      name:      "LF correct",
      content:   "line 1\nline 2\n",
      endOfLine: "lf",
      wantError: false,
    },
    {
      name:      "CRLF correct",
      content:   "line 1\r\nline 2\r\n",
      endOfLine: "crlf",
      wantError: false,
    },
    {
      name:      "CR correct",
      content:   "line 1\rline 2\r",
      endOfLine: "cr",
      wantError: false,
    },
    {
      name:      "mixed line endings",
      content:   "line 1\nline 2\r\nline 3\r",
      endOfLine: "lf",
      wantError: true,
    },
    {
      name:      "CRLF when expecting LF",
      content:   "line 1\r\nline 2\r\n",
      endOfLine: "lf",
      wantError: true,
    },
  }

  for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
      cfg := &config.ResolvedConfig{
        EndOfLine: tt.endOfLine,
      }

      err := ValidateEndOfLine("test.go", []byte(tt.content), cfg)

      if tt.wantError && err == nil {
        t.Error("Expected validation error, but got none")
      }

      if !tt.wantError && err != nil {
        t.Errorf("Expected no validation error, but got: %v", err)
      }
    })
  }
}

func TestFixEndOfLine(t *testing.T) {
  tests := []struct {
    name            string
    content         string
    endOfLine       string
    expectedContent string
    expectFixed     bool
  }{
    {
      name:            "convert CRLF to LF",
      content:         "line 1\r\nline 2\r\n",
      endOfLine:       "lf",
      expectedContent: "line 1\nline 2\n",
      expectFixed:     true,
    },
    {
      name:            "convert LF to CRLF",
      content:         "line 1\nline 2\n",
      endOfLine:       "crlf",
      expectedContent: "line 1\r\nline 2\r\n",
      expectFixed:     true,
    },
    {
      name:            "mixed to LF",
      content:         "line 1\r\nline 2\rline 3\n",
      endOfLine:       "lf",
      expectedContent: "line 1\nline 2\nline 3\n",
      expectFixed:     true,
    },
    {
      name:            "already correct",
      content:         "line 1\nline 2\n",
      endOfLine:       "lf",
      expectedContent: "line 1\nline 2\n",
      expectFixed:     false,
    },
  }

  for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
      cfg := &config.ResolvedConfig{
        EndOfLine: tt.endOfLine,
      }

      newContent, fixed, err := FixEndOfLine("test.go", []byte(tt.content), cfg)
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
