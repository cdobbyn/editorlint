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
```

### Command Line Options

| Flag | Short | Description |
|------|-------|-------------|
| `--recurse` | `-r` | Scan directories recursively |
| `--fix` | `-f` | Automatically fix validation errors |
| `--config` | `-c` | Use specific .editorconfig file instead of searching hierarchy |

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

### GitHub Action

Use editorlint in your GitHub workflows to automatically validate and fix EditorConfig violations:

```yaml
name: EditorConfig Check

on: [push, pull_request]

jobs:
  editorconfig:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: dobbo-ca/editorlint@v1
        with:
          path: '.'
          fix: false
          fail-on-violations: true
```

**Auto-fix with commits:**

```yaml
name: Auto-fix EditorConfig

on: [push]

jobs:
  editorconfig-fix:
    runs-on: ubuntu-latest
    permissions:
      contents: write
    steps:
      - uses: actions/checkout@v4
        with:
          token: ${{ secrets.GITHUB_TOKEN }}
      - uses: dobbo-ca/editorlint@v1
        with:
          fix: true
          auto-commit: true
          token: ${{ secrets.GITHUB_TOKEN }}
```

**Action Inputs:**

| Input | Description | Required | Default |
|-------|-------------|----------|---------|
| `path` | Path to file or directory to validate | No | `.` |
| `config-file` | Path to custom .editorconfig file | No | `` |
| `recurse` | Process files recursively in directories | No | `true` |
| `fix` | Automatically fix violations instead of just reporting | No | `false` |
| `output-format` | Output format: default, tabular, json, quiet | No | `default` |
| `fail-on-violations` | Fail the action if violations are found | No | `true` |
| `auto-commit` | Automatically commit fixes when fix=true | No | `false` |
| `commit-message` | Commit message for auto-committed fixes | No | `fix: auto-fix editorconfig violations` |
| `git-user-name` | Git user name for auto-commits | No | `github-actions[bot]` |
| `git-user-email` | Git user email for auto-commits | No | `github-actions[bot]@users.noreply.github.com` |
| `token` | GitHub token for auto-commits (defaults to github.token) | No | `` |
| `ignore-patterns` | Comma-separated list of glob patterns to ignore | No | `` |

**Action Outputs:**

| Output | Description |
|--------|-------------|
| `violations-found` | Whether any violations were found (true/false) |
| `files-processed` | Number of files processed |
| `files-fixed` | Number of files fixed (when fix=true) |

**Example Workflows:**

```yaml
# Basic validation - fail on violations
- uses: dobbo-ca/editorlint@v1

# Auto-fix violations and commit changes
- uses: dobbo-ca/editorlint@v1
  with:
    fix: true
    auto-commit: true
    commit-message: 'style: fix editorconfig violations'
    token: ${{ secrets.GITHUB_TOKEN }}

# Validate specific directory with custom config
- uses: dobbo-ca/editorlint@v1
  with:
    path: 'src/'
    config-file: '.editorconfig.strict'
    ignore-patterns: '*.tmp,node_modules/**'

# Use outputs in subsequent steps
- uses: dobbo-ca/editorlint@v1
  id: lint
- run: echo "Fixed ${{ steps.lint.outputs.files-fixed }} files"
```

### Homebrew (macOS and Linux)

```bash
brew tap dobbo-ca/taps
brew install editorlint
```

Or install directly:

```bash
brew install dobbo-ca/taps/editorlint
```

### Chocolatey (Windows)

```powershell
choco install editorlint -s https://github.com/dobbo-ca/chocolatey-packages
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

## GitHub Actions Marketplace

### Publishing to Marketplace

To publish this action to the GitHub Actions Marketplace:

1. **Create a Release**: The action must be tagged with a semantic version
   ```bash
   git tag v1.0.0
   git push origin v1.0.0
   ```

2. **Use Major Version Tags**: Create and maintain major version tags for easier consumption
   ```bash
   git tag v1
   git push origin v1
   ```

3. **Marketplace Requirements Met**:
   - ✅ `action.yml` with proper metadata and branding
   - ✅ Comprehensive README with usage examples
   - ✅ MIT License
   - ✅ Semantic versioning
   - ✅ Proper input/output documentation

### Version Tags

For marketplace compatibility, this action follows semantic versioning:

- **v1.x.x**: Latest stable version with major version tag `v1`
- **v1.0.0+**: Initial marketplace releases
- Users can reference `@v1` for automatic minor/patch updates
- Users can reference `@v1.0.0` for exact version pinning

### Marketplace Benefits

Publishing to the marketplace provides:
- **Discoverability**: Users can find the action in GitHub's marketplace
- **Trust**: Verified actions with clear documentation
- **Version Management**: Automatic handling of major version aliases
- **Usage Analytics**: Insights into action adoption

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

