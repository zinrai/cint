package main

import (
	"fmt"
	"strings"
)

// FormatResults formats validation results into a human-readable string
func FormatResults(results []ValidationResult) string {
	var output strings.Builder

	for _, result := range results {
		formatSingleResult(&output, result)
	}

	return output.String()
}

// formatSingleResult formats a single validation result
func formatSingleResult(output *strings.Builder, result ValidationResult) {
	if result.IsValid {
		fmt.Fprintf(output, "%s: ok\n", result.FileName)
		return
	}

	fmt.Fprintf(output, "FAIL: %s\n", result.FileName)
	for _, err := range result.Errors {
		formatError(output, err)
	}
}

// formatError formats a single validation error
func formatError(output *strings.Builder, err ValidationError) {
	switch {
	case err.Line > 0 && err.Field != "":
		fmt.Fprintf(output, "  line %d, field \"%s\": %s\n",
			err.Line, err.Field, err.Problem)
	case err.Line > 0:
		fmt.Fprintf(output, "  line %d: %s\n",
			err.Line, err.Problem)
	case err.Field != "":
		fmt.Fprintf(output, "  field \"%s\": %s\n",
			err.Field, err.Problem)
	default:
		fmt.Fprintf(output, "  %s\n", err.Problem)
	}
}
