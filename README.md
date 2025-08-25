# editorlint

A Go tool to validate and fix files according to .editorconfig specifications.

## Overview

editorlint reads .editorconfig files and validates that all files in a repository follow the specified configuration rules.

## Features

- Validates files against .editorconfig rules
- Supports recursive directory scanning with -r/--recurse flag
- Currently supports insert_final_newline validation
- More validations coming soon

## Usage

```bash
# Check files in current directory
editorlint .

# Check files recursively
editorlint -r .
```

## Installation

```bash
go install github.com/cdobbyn/editorlint@latest
```

