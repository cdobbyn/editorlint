package rules

import (
  "testing"
  
  "github.com/cdobbyn/editorlint/pkg/config"
)

func TestValidateInsertFinalNewline(t *testing.T) {
  tests := []struct {
    name      string
    content   string
    endOfLine string
    insertFinalNewline bool
    wantError bool
  }{
    {
      name:      "file with LF ending",
      content:   "package main\n",
      endOfLine: "lf",
      insertFinalNewline: true,
      wantError: false,
    },
    {
      name:      "file without newline",
      content:   "package main",
      endOfLine: "lf",
      insertFinalNewline: true,
      wantError: true,
    },
    {
      name:      "file with CRLF ending when expecting LF",
      content:   "package main\r\n",
      endOfLine: "lf",
      insertFinalNewline: true,
      wantError: true,
    },
    {
      name:      "file with CRLF ending when expecting CRLF",
      content:   "package main\r\n",
      endOfLine: "crlf",
      insertFinalNewline: true,
      wantError: false,
    },
    {
      name:      "empty file",
      content:   "",
      endOfLine: "lf",
      insertFinalNewline: true,
      wantError: true,
    },
    {
      name:      "file without newline but insert_final_newline is false",
      content:   "package main",
      endOfLine: "lf",
      insertFinalNewline: false,
      wantError: false,
    },
  }
  
  for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
      resolvedConfig := &config.ResolvedConfig{
        EndOfLine: tt.endOfLine,
        InsertFinalNewline: &tt.insertFinalNewline,
      }
      
      err := ValidateInsertFinalNewline("test.go", []byte(tt.content), resolvedConfig)
      
      if tt.wantError && err == nil {
        t.Error("Expected validation error, but got none")
      }
      
      if !tt.wantError && err != nil {
        t.Errorf("Expected no validation error, but got: %v", err)
      }
    })
  }
}

func TestFixInsertFinalNewline(t *testing.T) {
  tests := []struct {
    name      string
    content   string
    endOfLine string
    insertFinalNewline bool
    expectedContent string
    expectFixed bool
  }{
    {
      name:      "file without newline - should add LF",
      content:   "package main",
      endOfLine: "lf",
      insertFinalNewline: true,
      expectedContent: "package main\n",
      expectFixed: true,
    },
    {
      name:      "file with CRLF when expecting LF",
      content:   "package main\r\n",
      endOfLine: "lf",
      insertFinalNewline: true,
      expectedContent: "package main\n",
      expectFixed: true,
    },
    {
      name:      "file already correct",
      content:   "package main\n",
      endOfLine: "lf",
      insertFinalNewline: true,
      expectedContent: "package main\n",
      expectFixed: false,
    },
    {
      name:      "insert_final_newline is false - no fix",
      content:   "package main",
      endOfLine: "lf",
      insertFinalNewline: false,
      expectedContent: "package main",
      expectFixed: false,
    },
  }
  
  for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
      resolvedConfig := &config.ResolvedConfig{
        EndOfLine: tt.endOfLine,
        InsertFinalNewline: &tt.insertFinalNewline,
      }
      
      newContent, fixed, err := FixInsertFinalNewline("test.go", []byte(tt.content), resolvedConfig)
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