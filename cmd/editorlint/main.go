package main

import (
  "fmt"
  "os"

  "github.com/cdobbyn/editorlint/pkg/validator"
  "github.com/spf13/cobra"
)

var (
  recurseFlag    bool
  fixFlag        bool
  configFlag     string
  outputFlag     string
  workersFlag    int
  quietFlag      bool
)

var rootCmd = &cobra.Command{
  Use:   "editorlint [directory|file]",
  Short: "A tool to validate files against .editorconfig rules",
  Long:  "editorlint reads .editorconfig files and validates that all files in a repository follow the specified configuration rules.",
  Args:  cobra.ExactArgs(1),
  Run: func(cmd *cobra.Command, args []string) {
    target := args[0]

    // Create validator with config
    v := validator.New(validator.Config{
      CustomConfigPath: configFlag,
      Recursive:        recurseFlag,
      Fix:              fixFlag,
      OutputFormat:     outputFlag,
      Workers:          workersFlag,
      Quiet:            quietFlag,
    })

    err := v.ValidateTarget(target)
    if err != nil {
      fmt.Fprintf(os.Stderr, "Error: %v\n", err)
      os.Exit(1)
    }
  },
}

func init() {
  rootCmd.Flags().BoolVarP(&recurseFlag, "recurse", "r", false, "Scan directories recursively")
  rootCmd.Flags().BoolVarP(&fixFlag, "fix", "f", false, "Automatically fix validation errors")
  rootCmd.Flags().StringVarP(&configFlag, "config", "c", "", "Use specific .editorconfig file instead of searching hierarchy")
  rootCmd.Flags().StringVarP(&outputFlag, "output", "o", "default", "Output format: default, tabular, json, quiet")
  rootCmd.Flags().IntVarP(&workersFlag, "workers", "w", 0, "Number of parallel workers (0 = auto-detect)")
  rootCmd.Flags().BoolVarP(&quietFlag, "quiet", "q", false, "Quiet mode - minimal output")
}

func main() {
  if err := rootCmd.Execute(); err != nil {
    fmt.Fprintf(os.Stderr, "Error: %v\n", err)
    os.Exit(1)
  }
}
