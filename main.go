package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

const version = "0.1.0"

func main() {
	var (
		schemaPath  string
		configPaths stringSlice
		showVersion bool
	)

	setupFlags(&schemaPath, &configPaths, &showVersion)
	flag.Parse()

	if showVersion {
		printVersion()
		os.Exit(0)
	}

	if err := validateArgs(schemaPath, configPaths); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n\n", err)
		flag.Usage()
		os.Exit(1)
	}

	runValidation(schemaPath, configPaths)
}

// setupFlags configures command-line flags
func setupFlags(schemaPath *string, configPaths *stringSlice, showVersion *bool) {
	flag.StringVar(schemaPath, "schema", "", "Path to CUE schema file (required)")
	flag.Var(configPaths, "config", "Path to config file to validate (can be specified multiple times)")
	flag.BoolVar(showVersion, "version", false, "Show version")
	flag.Usage = createUsageFunc()
}

// createUsageFunc creates the custom usage function
func createUsageFunc() func() {
	return func() {
		progName := filepath.Base(os.Args[0])
		fmt.Fprintf(os.Stderr, "cint - Configuration linter powered by CUE\n\n")
		fmt.Fprintf(os.Stderr, "Usage: %s --schema=<schema.cue> --config=<config.yaml> [--config=<config2.yaml>...]\n\n", progName)
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  # Validate a single file\n")
		fmt.Fprintf(os.Stderr, "  %s --schema=app.cue --config=service.yaml\n\n", progName)
		fmt.Fprintf(os.Stderr, "  # Validate multiple files\n")
		fmt.Fprintf(os.Stderr, "  %s --schema=app.cue --config=service-a.yaml --config=service-b.yaml\n\n", progName)
	}
}

// printVersion prints the version information
func printVersion() {
	fmt.Printf("cint version %s\n", version)
}

// validateArgs validates command-line arguments
func validateArgs(schemaPath string, configPaths []string) error {
	if schemaPath == "" {
		return fmt.Errorf("--schema is required")
	}
	if len(configPaths) == 0 {
		return fmt.Errorf("at least one --config is required")
	}
	return nil
}

// runValidation runs the validation and handles the results
func runValidation(schemaPath string, configPaths []string) {
	results := ValidateFiles(schemaPath, configPaths)
	output := FormatResults(results)
	fmt.Print(output)

	exitCode := determineExitCode(results)
	os.Exit(exitCode)
}

// determineExitCode determines the exit code based on validation results
func determineExitCode(results []ValidationResult) int {
	for _, result := range results {
		if !result.IsValid {
			return 1
		}
	}
	return 0
}

// stringSlice implements flag.Value for multiple string flags
type stringSlice []string

func (s *stringSlice) String() string {
	return fmt.Sprintf("%v", *s)
}

func (s *stringSlice) Set(value string) error {
	*s = append(*s, value)
	return nil
}
