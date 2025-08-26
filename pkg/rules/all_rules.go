package rules

// GetAllValidators returns all available validation functions
func GetAllValidators() []ValidatorFunc {
  return []ValidatorFunc{
    ValidateInsertFinalNewline,
    ValidateTrimTrailingWhitespace,
    ValidateEndOfLine,
    // Other validators will be added as we migrate them
  }
}

// GetAllFixers returns all available fix functions in the correct order
func GetAllFixers() []FixerFunc {
  return []FixerFunc{
    FixInsertFinalNewline,
    FixTrimTrailingWhitespace,
    FixEndOfLine,
    // Other fixers will be added as we migrate them
  }
}