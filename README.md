# editorlint

A comprehensive Go tool to validate and fix files according to .editorconfig specifications.

## Overview

editorlint reads .editorconfig files and validates that all files in a repository follow the specified configuration rules. It supports hierarchical configuration files and can automatically fix many types of violations.

## Features

### Core Functionality
- **Hierarchical Configuration**: Properly handles multiple `.editorconfig` files with correct precedence rules
- **Pattern Matching**: Full support for EditorConfig glob patterns (`*`, `**`, `[seq]`, `{alt1,alt2}`)
- **Recursive Scanning**: Scan directories recursively with `-r/--recurse` flag
- **Automatic Fixing**: Fix validation errors automatically with `-f/--fix` flag

### Supported EditorConfig Rules

- ✅ **insert_final_newline**: Ensures files end with appropriate newline characters
- ✅ **trim_trailing_whitespace**: Removes trailing spaces and tabs from lines
- ✅ **indent_style**: Validates consistent use of spaces or tabs for indentation
- ✅ **indent_size**: Ensures proper indentation size (works with spaces)
- ✅ **tab_width**: Handles tab width for display calculations
- ✅ **end_of_line**: Validates and converts line ending styles (LF, CRLF, CR)
- ✅ **max_line_length**: Validates maximum line length (with tab expansion)
- ✅ **charset**: Validates and converts character encodings (UTF-8, UTF-16, Latin-1, with/without BOM)

### Validation Features
- **Multiple Rule Validation**: Runs all applicable rules on each file
- **Detailed Error Messages**: Clear descriptions of violations with line numbers
- **Pattern-Specific Rules**: Different rules can apply to different file patterns
- **UTF-8 Support**: Proper handling of Unicode characters in length calculations

### Fix Features
- **Safe Fixing**: Only modifies files when actual violations are found
- **Ordered Processing**: Fixes are applied in logical order (line endings first, final newline last)
- **Preserve File Permissions**: Maintains original file permissions after fixing
- **Comprehensive Fixing**: Can fix multiple types of violations in a single pass

## Usage

```bash
# Validate files in current directory
editorlint .

# Validate files recursively
editorlint -r .

# Validate a single file
editorlint src/main.go

# Fix violations automatically
editorlint -f .

# Fix violations recursively
editorlint -r -f .

# Fix a single file
editorlint -f src/main.go

# Use custom .editorconfig file
editorlint -c custom.editorconfig .

# Use custom config with single file
editorlint -c custom.editorconfig src/main.go

# Exclude specific files or patterns
editorlint --exclude "*.tmp" --exclude "node_modules/**" .

# Exclude with short flag
editorlint -e "*.log" -e "build/" .
```

### Command Line Options

| Flag | Short | Description |
|------|-------|-------------|
| `--recurse` | `-r` | Scan directories recursively |
| `--fix` | `-f` | Automatically fix validation errors |
| `--config` | `-c` | Use specific .editorconfig file instead of searching hierarchy |
| `--exclude` | `-e` | Exclude files matching glob patterns (can be specified multiple times) |
| `--output` | `-o` | Output format: default, tabular, json, quiet |
| `--workers` | `-w` | Number of parallel workers (0 = auto-detect) |
| `--quiet` | `-q` | Quiet mode - minimal output |

### Target Types

editorlint can work with both **directories** and **individual files**:

- **Directory mode**: Validates all files in the directory (optionally recursive)
- **File mode**: Validates a single specific file

When targeting a single file, editorlint will:
1. Look for `.editorconfig` in the file's directory hierarchy (unless `-c` is used)
2. Apply the appropriate rules based on file patterns
3. Report or fix violations for just that file

## EditorConfig Support

### File Hierarchy
editorlint properly handles the EditorConfig hierarchy:
- Searches up the directory tree for `.editorconfig` files
- Stops at files marked with `root = true`
- Child configurations override parent configurations
- More specific patterns override less specific patterns

### Custom Configuration Files
Use the `-c/--config` flag to specify a custom `.editorconfig` file:
- Bypasses the hierarchical search
- Uses only the specified configuration file
- Useful for testing different configurations
- Can be used with both directory and file targets

### Supported Properties

| Property | Description | Fix Support |
|----------|-------------|-------------|
| `insert_final_newline` | Ensure file ends with newline | ✅ Yes |
| `trim_trailing_whitespace` | Remove trailing whitespace | ✅ Yes |
| `indent_style` | Use tabs or spaces | ✅ Yes |
| `indent_size` | Number of spaces per indent | ✅ Yes |
| `tab_width` | Width of tab character | ✅ Used in calculations |
| `end_of_line` | Line ending style | ✅ Yes |
| `max_line_length` | Maximum line length | ⚠️ Validation only |
| `charset` | File character encoding | ✅ Yes |

### Example .editorconfig

```ini
root = true

# Default settings for all files
[*]
charset = utf-8
end_of_line = lf
insert_final_newline = true
trim_trailing_whitespace = true

# 4 space indentation for Python
[*.py]
indent_style = space
indent_size = 4
max_line_length = 88

# Tab indentation for Go
[*.go]
indent_style = tab
indent_size = 4
max_line_length = 100

# 2 space indentation for web files
[*.{js,ts,html,css,json}]
indent_style = space
indent_size = 2

# No trailing whitespace trimming for Markdown
[*.md]
trim_trailing_whitespace = false
max_line_length = 80

# Windows line endings for batch files
[*.bat]
end_of_line = crlf
```

## Installation

### Homebrew (macOS and Linux)

```bash
brew tap cdobbyn/taps
brew install editorlint
```

Or install directly:

```bash
brew install cdobbyn/taps/editorlint
```

### Go Install

```bash
go install github.com/dobbo-ca/editorlint@latest
```

### Pre-built Binaries

Download pre-built binaries from the [releases page](https://github.com/dobbo-ca/editorlint/releases).

## Architecture

editorlint is built with a modular architecture where each EditorConfig rule is implemented in its own file:

- `insert_final_newline.go` - Final newline validation and fixing
- `trim_trailing_whitespace.go` - Trailing whitespace handling
- `indent_style.go` - Indentation validation and conversion
- `max_line_length.go` - Line length validation
- `charset.go` - Character encoding validation and conversion
- `end_of_line.go` - Line ending validation and conversion

This makes it easy to add new validation rules or modify existing ones.

## Examples

### Basic Usage

```bash
# Check all files in current directory
$ editorlint .
Validating directory: . (recursive: false)
Found 2 validation errors:
  ./main.go: trim_trailing_whitespace violation - line 15 has trailing whitespace
  ./README.md: insert_final_newline violation - file should end with LF (\n), but ends with character 'o' (0x6f)

To fix these errors automatically, run with --fix flag

# Fix the errors
$ editorlint -f .
Fixing files in directory: . (recursive: false)
Fixed 2 files:
  ./main.go
  ./README.md
```

### Single File Usage

```bash
# Check a specific file
$ editorlint src/utils.go
Validating file: src/utils.go
✓ File passes editorconfig validation: src/utils.go

# Fix a specific file
$ editorlint -f src/main.go
Fixing file: src/main.go
✓ Fixed: src/main.go
```

### Custom Configuration

```bash
# Use a custom .editorconfig file
$ editorlint -c strict.editorconfig src/
Validating directory: src/ (recursive: false)
✓ All files pass editorconfig validation

# Test different config on single file
$ editorlint -c experimental.editorconfig main.go
Validating file: main.go
Found 1 validation errors:
  main.go: indent_style violation - line 5 uses spaces but should use tabs
```

### Error Types

Common validation errors you might see:

```bash
# Trailing whitespace
./file.txt: trim_trailing_whitespace violation - line 10 has trailing whitespace

# Wrong indentation
./script.py: indent_style violation - line 5 uses tabs but should use spaces
./code.js: indent_size violation - line 8 has 3 spaces, should be multiple of 2

# Line endings
./readme.md: end_of_line violation - line 2 uses CRLF (\r\n) but should use LF (\n)

# Missing final newline
./config.json: insert_final_newline violation - file should end with LF (\n), but ends with character '}' (0x7d)

# Line length
./long-line.txt: max_line_length violation - line 1 is 120 characters long, exceeds maximum of 100

# Character encoding
./file.txt: charset violation - file has UTF-8 BOM but charset is set to utf-8 (no BOM)
```

