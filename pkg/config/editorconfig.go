// Package config provides EditorConfig file parsing and configuration resolution.
package config

import (
  "bufio"
  "fmt"
  "os"
  "path/filepath"
  "regexp"
  "strconv"
  "strings"
)

// EditorConfig represents a parsed .editorconfig file
type EditorConfig struct {
  Root     bool
  Sections []Section
  FilePath string
}

// Section represents a section in .editorconfig with pattern and properties
type Section struct {
  Pattern    string
  Properties map[string]string
}

// ResolvedConfig represents the final configuration for a specific file
type ResolvedConfig struct {
  IndentStyle              string
  IndentSize               *int
  TabWidth                 *int
  EndOfLine                string
  Charset                  string
  TrimTrailingWhitespace   *bool
  InsertFinalNewline       *bool
  MaxLineLength            *int
}

// FindEditorConfigs walks up the directory tree to find all applicable .editorconfig files
func FindEditorConfigs(targetPath string) ([]*EditorConfig, error) {
  return FindEditorConfigsWithCustomConfig(targetPath, "")
}

// FindEditorConfigsWithCustomConfig allows specifying a custom config file path
func FindEditorConfigsWithCustomConfig(targetPath, customConfigPath string) ([]*EditorConfig, error) {
  // If custom config file is specified, use it exclusively
  if customConfigPath != "" {
    absConfigPath, err := filepath.Abs(customConfigPath)
    if err != nil {
      return nil, fmt.Errorf("failed to get absolute path for config file: %w", err)
    }

    config, err := ParseEditorConfig(absConfigPath)
    if err != nil {
      return nil, fmt.Errorf("failed to parse custom config %s: %w", absConfigPath, err)
    }

    return []*EditorConfig{config}, nil
  }

  // Original hierarchical search logic
  var configs []*EditorConfig

  dir := filepath.Dir(targetPath)
  if !filepath.IsAbs(dir) {
    absDir, err := filepath.Abs(dir)
    if err != nil {
      return nil, fmt.Errorf("failed to get absolute path: %w", err)
    }
    dir = absDir
  }

  for {
    configPath := filepath.Join(dir, ".editorconfig")

    if _, err := os.Stat(configPath); err == nil {
      config, err := ParseEditorConfig(configPath)
      if err != nil {
        return nil, fmt.Errorf("failed to parse %s: %w", configPath, err)
      }

      configs = append([]*EditorConfig{config}, configs...) // Prepend to maintain parent->child order

      if config.Root {
        break
      }
    }

    parent := filepath.Dir(dir)
    if parent == dir {
      // Reached filesystem root
      break
    }
    dir = parent
  }

  return configs, nil
}

// ParseEditorConfig parses a single .editorconfig file
func ParseEditorConfig(filePath string) (*EditorConfig, error) {
  file, err := os.Open(filePath)
  if err != nil {
    return nil, err
  }
  defer file.Close()

  config := &EditorConfig{FilePath: filePath}
  scanner := bufio.NewScanner(file)

  var currentSection *Section
  inHeaderSection := true

  for scanner.Scan() {
    line := strings.TrimSpace(scanner.Text())

    // Skip empty lines and comments
    if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
      continue
    }

    // Check for section headers [pattern]
    if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
      pattern := line[1 : len(line)-1]
      currentSection = &Section{
        Pattern:    pattern,
        Properties: make(map[string]string),
      }
      config.Sections = append(config.Sections, *currentSection)
      inHeaderSection = false
      continue
    }

    // Parse key = value pairs
    if strings.Contains(line, "=") {
      parts := strings.SplitN(line, "=", 2)
      if len(parts) != 2 {
        continue
      }

      key := strings.TrimSpace(parts[0])
      value := strings.TrimSpace(parts[1])

      // Handle root property in header section
      if inHeaderSection && key == "root" {
        config.Root = strings.ToLower(value) == "true"
        continue
      }

      // Add property to current section
      if currentSection != nil {
        currentSection.Properties[key] = value
      }
    }
  }

  return config, scanner.Err()
}

// ResolveConfigForFile resolves the final configuration for a specific file
func ResolveConfigForFile(filePath string, configs []*EditorConfig) (*ResolvedConfig, error) {
  resolved := &ResolvedConfig{}

  for _, config := range configs {
    for _, section := range config.Sections {
      matches, err := matchesPattern(filePath, section.Pattern, config.FilePath)
      if err != nil {
        return nil, err
      }

      if matches {
        applyProperties(resolved, section.Properties)
      }
    }
  }

  return resolved, nil
}

// matchesPattern checks if a file path matches an editorconfig pattern
func matchesPattern(filePath, pattern, configPath string) (bool, error) {
  // Convert file path to relative path from config location
  configDir := filepath.Dir(configPath)
  relPath, err := filepath.Rel(configDir, filePath)
  if err != nil {
    return false, err
  }

  // Normalize path separators to forward slashes
  relPath = filepath.ToSlash(relPath)

  // Convert editorconfig pattern to regex
  regexPattern, err := ConvertPatternToRegex(pattern)
  if err != nil {
    return false, err
  }

  matched, err := regexp.MatchString(regexPattern, relPath)
  return matched, err
}

// ConvertPatternToRegex converts an editorconfig glob pattern to a regex
func ConvertPatternToRegex(pattern string) (string, error) {
  // Escape regex special characters except our glob characters
  pattern = regexp.QuoteMeta(pattern)

  // Convert escaped glob patterns back to regex equivalents
  pattern = strings.ReplaceAll(pattern, "\\*\\*", ".*")     // ** matches anything including path separators

  // Special case: if pattern is just "*", treat it like "**" for compatibility
  // This matches common EditorConfig usage where [*] is expected to match all files
  if pattern == "\\*" {
    pattern = ".*"
  } else {
    pattern = strings.ReplaceAll(pattern, "\\*", "[^/]*")    // * matches anything except path separators
  }
  pattern = strings.ReplaceAll(pattern, "\\?", ".")        // ? matches any single character

  // Handle character classes [abc] and [!abc]
  pattern = strings.ReplaceAll(pattern, "\\[!", "[^")
  pattern = strings.ReplaceAll(pattern, "\\[", "[")
  pattern = strings.ReplaceAll(pattern, "\\]", "]")

  // Handle brace expansion {js,ts,jsx}
  braceRegex := regexp.MustCompile(`\\{([^}]+)\\}`)
  pattern = braceRegex.ReplaceAllStringFunc(pattern, func(match string) string {
    // Remove escaped braces
    content := match[2 : len(match)-2]
    // Split by comma and create alternation
    parts := strings.Split(content, ",")
    for i, part := range parts {
      parts[i] = regexp.QuoteMeta(part)
    }
    return "(" + strings.Join(parts, "|") + ")"
  })

  // Anchor the pattern to match the full path
  return "^" + pattern + "$", nil
}

// applyProperties applies properties to a ResolvedConfig
func applyProperties(config *ResolvedConfig, properties map[string]string) {
  for key, value := range properties {
    switch key {
    case "indent_style":
      if value == "tab" || value == "space" {
        config.IndentStyle = value
      }
    case "indent_size":
      if value == "tab" {
        // Use tab_width value if available
        config.IndentSize = nil
      } else if size, err := strconv.Atoi(value); err == nil && size > 0 {
        config.IndentSize = &size
      }
    case "tab_width":
      if width, err := strconv.Atoi(value); err == nil && width > 0 {
        config.TabWidth = &width
      }
    case "end_of_line":
      if value == "lf" || value == "crlf" || value == "cr" {
        config.EndOfLine = value
      }
    case "charset":
      config.Charset = value
    case "trim_trailing_whitespace":
      if b, err := strconv.ParseBool(value); err == nil {
        config.TrimTrailingWhitespace = &b
      }
    case "insert_final_newline":
      if b, err := strconv.ParseBool(value); err == nil {
        config.InsertFinalNewline = &b
      }
    case "max_line_length":
      if value == "off" {
        // max_line_length = off means no limit
        config.MaxLineLength = nil
      } else if length, err := strconv.Atoi(value); err == nil && length > 0 {
        config.MaxLineLength = &length
      }
    }
  }
}
