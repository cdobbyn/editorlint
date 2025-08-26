package config

import (
  "os"
  "path/filepath"
  "testing"
)

func TestParseEditorConfig(t *testing.T) {
  // Create a temporary .editorconfig file
  tmpDir := t.TempDir()
  configContent := `root = true

[*]
charset = utf-8
end_of_line = lf
insert_final_newline = true
trim_trailing_whitespace = true

[*.go]
indent_style = tab
indent_size = 4

[*.{js,ts}]
indent_style = space
indent_size = 2
`

  configPath := filepath.Join(tmpDir, ".editorconfig")
  if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
    t.Fatal(err)
  }

  // Parse the config
  editorConfig, err := ParseEditorConfig(configPath)
  if err != nil {
    t.Fatal(err)
  }

  // Verify root property
  if !editorConfig.Root {
    t.Error("Expected root = true")
  }

  // Verify we have 3 sections
  if len(editorConfig.Sections) != 3 {
    t.Errorf("Expected 3 sections, got %d", len(editorConfig.Sections))
  }

  // Check first section [*]
  if editorConfig.Sections[0].Pattern != "*" {
    t.Errorf("Expected first pattern to be '*', got %s", editorConfig.Sections[0].Pattern)
  }

  if editorConfig.Sections[0].Properties["charset"] != "utf-8" {
    t.Errorf("Expected charset = utf-8")
  }

  if editorConfig.Sections[0].Properties["insert_final_newline"] != "true" {
    t.Errorf("Expected insert_final_newline = true")
  }
}

func TestResolveConfigForFile(t *testing.T) {
  // Create a temporary .editorconfig file
  tmpDir := t.TempDir()
  configContent := `root = true

[*]
insert_final_newline = true
end_of_line = lf

[*.go]
indent_style = tab
`

  configPath := filepath.Join(tmpDir, ".editorconfig")
  if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
    t.Fatal(err)
  }

  editorConfig, err := ParseEditorConfig(configPath)
  if err != nil {
    t.Fatal(err)
  }

  // Test resolution for a .go file
  goFile := filepath.Join(tmpDir, "main.go")
  resolved, err := ResolveConfigForFile(goFile, []*EditorConfig{editorConfig})
  if err != nil {
    t.Fatal(err)
  }

  // Should have insert_final_newline from [*] and indent_style from [*.go]
  if resolved.InsertFinalNewline == nil || !*resolved.InsertFinalNewline {
    t.Error("Expected insert_final_newline = true for .go file")
  }

  if resolved.IndentStyle != "tab" {
    t.Errorf("Expected indent_style = tab for .go file, got %s", resolved.IndentStyle)
  }

  if resolved.EndOfLine != "lf" {
    t.Errorf("Expected end_of_line = lf for .go file, got %s", resolved.EndOfLine)
  }
}

func TestConvertPatternToRegex(t *testing.T) {
  tests := []struct {
    pattern  string
    expected string
  }{
    {"*", "^[^/]*$"},
    {"*.go", "^[^/]*\\.go$"},
    {"**", "^.*$"},
    {"src/**", "^src/.*$"},
    {"*.{js,ts}", "^[^/]*\\.(js|ts)$"},
  }

  for _, tt := range tests {
    t.Run(tt.pattern, func(t *testing.T) {
      result, err := ConvertPatternToRegex(tt.pattern)
      if err != nil {
        t.Fatal(err)
      }

      if result != tt.expected {
        t.Errorf("Expected %s, got %s", tt.expected, result)
      }
    })
  }
}

func TestCustomConfigFile(t *testing.T) {
  // Create a temporary custom config file
  tmpDir := t.TempDir()
  customConfigPath := filepath.Join(tmpDir, "custom.editorconfig")
  customConfigContent := `[*]
insert_final_newline = true
trim_trailing_whitespace = true
indent_style = tab
`

  if err := os.WriteFile(customConfigPath, []byte(customConfigContent), 0644); err != nil {
    t.Fatal(err)
  }

  // Test FindEditorConfigsWithCustomConfig
  testFile := filepath.Join(tmpDir, "test.go")
  configs, err := FindEditorConfigsWithCustomConfig(testFile, customConfigPath)
  if err != nil {
    t.Fatal(err)
  }

  if len(configs) != 1 {
    t.Errorf("Expected 1 config, got %d", len(configs))
  }

  // Verify the config was parsed correctly
  resolved, err := ResolveConfigForFile(testFile, configs)
  if err != nil {
    t.Fatal(err)
  }

  if resolved.IndentStyle != "tab" {
    t.Errorf("Expected indent_style = tab, got %s", resolved.IndentStyle)
  }

  if resolved.InsertFinalNewline == nil || !*resolved.InsertFinalNewline {
    t.Error("Expected insert_final_newline = true")
  }

  if resolved.TrimTrailingWhitespace == nil || !*resolved.TrimTrailingWhitespace {
    t.Error("Expected trim_trailing_whitespace = true")
  }
}
