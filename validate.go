package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/errors"
	"cuelang.org/go/encoding/json"
	"cuelang.org/go/encoding/yaml"
)

// ValidationResult represents the validation result for a single file
type ValidationResult struct {
	FileName string
	IsValid  bool
	Errors   []ValidationError
}

// ValidationError represents a single validation error
type ValidationError struct {
	Line    int    // Line number in the config file
	Field   string // Field path (e.g., "spec.replicas")
	Problem string // Error message from CUE
}

// ValidateFiles validates multiple config files against a CUE schema
func ValidateFiles(schemaPath string, configPaths []string) []ValidationResult {
	ctx := cuecontext.New()

	schema, err := loadSchema(ctx, schemaPath)
	if err != nil {
		return createSchemaErrorResults(configPaths, err)
	}

	var results []ValidationResult
	for _, configPath := range configPaths {
		result := validateFile(ctx, schema, configPath)
		results = append(results, result)
	}

	return results
}

// createSchemaErrorResults creates error results for all files when schema loading fails
func createSchemaErrorResults(configPaths []string, err error) []ValidationResult {
	var results []ValidationResult
	errorMsg := fmt.Sprintf("failed to load schema: %v", err)

	for _, path := range configPaths {
		results = append(results, ValidationResult{
			FileName: path,
			IsValid:  false,
			Errors: []ValidationError{
				{Line: 0, Field: "", Problem: errorMsg},
			},
		})
	}
	return results
}

// loadSchema loads and compiles a CUE schema file
func loadSchema(ctx *cue.Context, schemaPath string) (cue.Value, error) {
	schemaData, err := os.ReadFile(schemaPath)
	if err != nil {
		return cue.Value{}, fmt.Errorf("reading schema file: %w", err)
	}

	schema := ctx.CompileBytes(schemaData, cue.Filename(schemaPath))
	if schema.Err() != nil {
		return cue.Value{}, fmt.Errorf("compiling schema: %w", schema.Err())
	}

	return schema, nil
}

// validateFile validates a single config file against the schema
func validateFile(ctx *cue.Context, schema cue.Value, configPath string) ValidationResult {
	configData, err := os.ReadFile(configPath)
	if err != nil {
		return createErrorResult(configPath, fmt.Sprintf("failed to read file: %v", err))
	}

	config, err := parseConfigFile(ctx, configPath, configData)
	if err != nil {
		return createErrorResult(configPath, err.Error())
	}

	if config.Err() != nil {
		return createValidationErrorResult(configPath, config.Err())
	}

	configDef := schema.LookupPath(cue.ParsePath("#Config"))
	if !configDef.Exists() {
		return createErrorResult(configPath, "schema does not define #Config")
	}

	unified := configDef.Unify(config)

	err = unified.Validate(cue.Concrete(true))
	if err != nil {
		return createValidationErrorResult(configPath, err)
	}

	return ValidationResult{
		FileName: configPath,
		IsValid:  true,
		Errors:   []ValidationError{},
	}
}

// parseConfigFile parses a config file based on its extension
func parseConfigFile(ctx *cue.Context, configPath string, configData []byte) (cue.Value, error) {
	ext := strings.ToLower(filepath.Ext(configPath))

	switch ext {
	case ".yaml", ".yml":
		return parseYAML(ctx, configPath, configData)
	case ".json":
		return parseJSON(ctx, configPath, configData)
	default:
		return cue.Value{}, fmt.Errorf("unsupported file format: %s (supported: .yaml, .yml, .json)", ext)
	}
}

// parseYAML parses YAML data into a CUE value
func parseYAML(ctx *cue.Context, configPath string, configData []byte) (cue.Value, error) {
	file, err := yaml.Extract(configPath, configData)
	if err != nil {
		return cue.Value{}, fmt.Errorf("failed to parse YAML: %w", err)
	}
	return ctx.BuildFile(file), nil
}

// parseJSON parses JSON data into a CUE value
func parseJSON(ctx *cue.Context, configPath string, configData []byte) (cue.Value, error) {
	expr, err := json.Extract(configPath, configData)
	if err != nil {
		return cue.Value{}, fmt.Errorf("failed to parse JSON: %w", err)
	}
	return ctx.BuildExpr(expr), nil
}

// createErrorResult creates a single error result
func createErrorResult(fileName string, problem string) ValidationResult {
	return ValidationResult{
		FileName: fileName,
		IsValid:  false,
		Errors: []ValidationError{
			{Line: 0, Field: "", Problem: problem},
		},
	}
}

// createValidationErrorResult creates a result with extracted validation errors
func createValidationErrorResult(fileName string, err error) ValidationResult {
	return ValidationResult{
		FileName: fileName,
		IsValid:  false,
		Errors:   extractValidationErrors(err),
	}
}

// extractValidationErrors extracts structured error information from CUE errors
func extractValidationErrors(err error) []ValidationError {
	cueErrors := errors.Errors(err)
	if len(cueErrors) == 0 {
		return []ValidationError{
			{Line: 0, Field: "", Problem: err.Error()},
		}
	}

	var validationErrors []ValidationError
	for _, e := range cueErrors {
		ve := extractSingleError(e)
		validationErrors = append(validationErrors, ve)
	}

	return validationErrors
}

// extractSingleError extracts information from a single CUE error
func extractSingleError(e errors.Error) ValidationError {
	return ValidationError{
		Line:    extractLineNumber(e),
		Field:   extractFieldPath(e),
		Problem: e.Error(),
	}
}

// extractLineNumber extracts the line number from error positions
func extractLineNumber(e errors.Error) int {
	positions := errors.Positions(e)
	for _, pos := range positions {
		if line := pos.Line(); line > 0 {
			return line
		}
	}
	return 0
}

// extractFieldPath extracts and formats the field path from error
func extractFieldPath(e errors.Error) string {
	path := e.Path()
	if len(path) == 0 {
		return ""
	}
	return formatPath(path)
}

// formatPath converts CUE path to a string representation
func formatPath(path []string) string {
	var parts []string

	for _, p := range path {
		if !isValidPathElement(p) {
			continue
		}
		p = strings.Trim(p, `"`)
		parts = append(parts, p)
	}

	return strings.Join(parts, ".")
}

// isValidPathElement checks if a path element should be included
func isValidPathElement(p string) bool {
	return p != "" &&
		!strings.HasPrefix(p, "[") &&
		p != "#Config"
}
