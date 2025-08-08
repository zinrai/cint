package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateFiles(t *testing.T) {
	tests := []struct {
		name       string
		schema     string
		configs    map[string]string // filename -> content mapping
		wantValid  bool
		wantErrors []string // Expected error message fragments
	}{
		{
			name: "valid config YAML",
			schema: `
				#Config: {
					name: string
					replicas?: int & >=1 & <=10
				}
			`,
			configs: map[string]string{
				"config.yaml": `
name: "my-service"
replicas: 3
`,
			},
			wantValid: true,
		},
		{
			name: "valid config JSON",
			schema: `
				#Config: {
					name: string
					replicas?: int & >=1 & <=10
				}
			`,
			configs: map[string]string{
				"config.json": `{
					"name": "my-service",
					"replicas": 3
				}`,
			},
			wantValid: true,
		},
		{
			name: "missing required field YAML",
			schema: `
				#Config: {
					name: string
					version: string
				}
			`,
			configs: map[string]string{
				"config.yaml": `
name: "my-service"
`,
			},
			wantValid:  false,
			wantErrors: []string{"version", "incomplete"},
		},
		{
			name: "missing required field JSON",
			schema: `
				#Config: {
					name: string
					version: string
				}
			`,
			configs: map[string]string{
				"config.json": `{"name": "my-service"}`,
			},
			wantValid:  false,
			wantErrors: []string{"version", "incomplete"},
		},
		{
			name: "value out of range YAML",
			schema: `
				#Config: {
					replicas: int & >=1 & <=10
				}
			`,
			configs: map[string]string{
				"config.yaml": `replicas: 0`,
			},
			wantValid:  false,
			wantErrors: []string{"replicas", "invalid value"},
		},
		{
			name: "value out of range JSON",
			schema: `
				#Config: {
					replicas: int & >=1 & <=10
				}
			`,
			configs: map[string]string{
				"config.json": `{"replicas": 0}`,
			},
			wantValid:  false,
			wantErrors: []string{"replicas", "invalid value"},
		},
		{
			name: "pattern mismatch YAML",
			schema: `
				#Config: {
					name: string & =~"^[a-z][a-z0-9-]*$"
				}
			`,
			configs: map[string]string{
				"config.yaml": `name: "MyService"`,
			},
			wantValid:  false,
			wantErrors: []string{"name", "out of bound"},
		},
		{
			name: "pattern mismatch JSON",
			schema: `
				#Config: {
					name: string & =~"^[a-z][a-z0-9-]*$"
				}
			`,
			configs: map[string]string{
				"config.json": `{"name": "MyService"}`,
			},
			wantValid:  false,
			wantErrors: []string{"name", "out of bound"},
		},
		{
			name: "enum constraint YAML",
			schema: `
				#Config: {
					environment: "development" | "staging" | "production"
				}
			`,
			configs: map[string]string{
				"config.yaml": `environment: "dev"`,
			},
			wantValid:  false,
			wantErrors: []string{"environment"},
		},
		{
			name: "enum constraint JSON",
			schema: `
				#Config: {
					environment: "development" | "staging" | "production"
				}
			`,
			configs: map[string]string{
				"config.json": `{"environment": "dev"}`,
			},
			wantValid:  false,
			wantErrors: []string{"environment"},
		},
		{
			name: "nested structure validation YAML",
			schema: `
				#Config: {
					metadata: {
						name: string
						labels?: {[string]: string}
					}
				}
			`,
			configs: map[string]string{
				"config.yaml": `
metadata:
  name: "test"
  labels:
    app: "web"
`,
			},
			wantValid: true,
		},
		{
			name: "nested structure validation JSON",
			schema: `
				#Config: {
					metadata: {
						name: string
						labels?: {[string]: string}
					}
				}
			`,
			configs: map[string]string{
				"config.json": `{
					"metadata": {
						"name": "test",
						"labels": {
							"app": "web"
						}
					}
				}`,
			},
			wantValid: true,
		},
		{
			name: "multiple errors YAML",
			schema: `
				#Config: {
					name: string & =~"^[a-z][a-z0-9-]*$"
					replicas: int & >=1 & <=10
					environment: "development" | "staging" | "production"
				}
			`,
			configs: map[string]string{
				"config.yaml": `
name: "BadName"
replicas: 100
environment: "testing"
`,
			},
			wantValid: false,
			wantErrors: []string{
				"name",
				"replicas",
				"environment",
			},
		},
		{
			name: "multiple errors JSON",
			schema: `
				#Config: {
					name: string & =~"^[a-z][a-z0-9-]*$"
					replicas: int & >=1 & <=10
					environment: "development" | "staging" | "production"
				}
			`,
			configs: map[string]string{
				"config.json": `{
					"name": "BadName",
					"replicas": 100,
					"environment": "testing"
				}`,
			},
			wantValid: false,
			wantErrors: []string{
				"name",
				"replicas",
				"environment",
			},
		},
		{
			name: "multiple files mixed formats",
			schema: `
				#Config: {
					name: string & =~"^[a-z][a-z0-9-]*$"
				}
			`,
			configs: map[string]string{
				"config1.yaml": `name: "valid-name"`,
				"config2.json": `{"name": "another-valid"}`,
				"config3.yml":  `name: "yet-another"`,
			},
			wantValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary files
			tmpDir := t.TempDir()

			schemaPath := filepath.Join(tmpDir, "schema.cue")
			if err := os.WriteFile(schemaPath, []byte(tt.schema), 0644); err != nil {
				t.Fatalf("failed to write schema file: %v", err)
			}

			// Write all config files and collect paths
			var configPaths []string
			for filename, content := range tt.configs {
				configPath := filepath.Join(tmpDir, filename)
				if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
					t.Fatalf("failed to write config file %s: %v", filename, err)
				}
				configPaths = append(configPaths, configPath)
			}

			// Run validation
			results := ValidateFiles(schemaPath, configPaths)

			if len(results) != len(configPaths) {
				t.Fatalf("expected %d results, got %d", len(configPaths), len(results))
			}

			// Check all results
			allValid := true
			var allErrors []string
			for _, result := range results {
				if !result.IsValid {
					allValid = false
					for _, err := range result.Errors {
						allErrors = append(allErrors, fmt.Sprintf("field=%s problem=%s", err.Field, err.Problem))
					}
				}
			}

			// Check validity
			if allValid != tt.wantValid {
				t.Errorf("IsValid = %v, want %v", allValid, tt.wantValid)
			}

			// Check errors if validation should fail
			if !tt.wantValid && len(allErrors) > 0 {
				errorStr := strings.Join(allErrors, "; ")
				for _, expectedError := range tt.wantErrors {
					if !strings.Contains(errorStr, expectedError) {
						t.Errorf("expected error containing %q in errors: %s", expectedError, errorStr)
					}
				}
			}
		})
	}
}

func TestValidateFilesWithInvalidSchema(t *testing.T) {
	tmpDir := t.TempDir()

	// Invalid schema file (syntax error)
	schemaPath := filepath.Join(tmpDir, "invalid.cue")
	if err := os.WriteFile(schemaPath, []byte(`#Config: {invalid syntax`), 0644); err != nil {
		t.Fatalf("failed to write schema file: %v", err)
	}

	// Test with both YAML and JSON
	testCases := []struct {
		filename string
		content  string
	}{
		{"config.yaml", `name: test`},
		{"config.json", `{"name": "test"}`},
	}

	for _, tc := range testCases {
		t.Run(tc.filename, func(t *testing.T) {
			configPath := filepath.Join(tmpDir, tc.filename)
			if err := os.WriteFile(configPath, []byte(tc.content), 0644); err != nil {
				t.Fatalf("failed to write config file: %v", err)
			}

			results := ValidateFiles(schemaPath, []string{configPath})

			if len(results) != 1 {
				t.Fatalf("expected 1 result, got %d", len(results))
			}

			if results[0].IsValid {
				t.Error("expected validation to fail with invalid schema")
			}

			if !strings.Contains(results[0].Errors[0].Problem, "failed to load schema") {
				t.Errorf("expected schema loading error, got: %s", results[0].Errors[0].Problem)
			}
		})
	}
}

func TestValidateFilesWithNonExistentFile(t *testing.T) {
	tmpDir := t.TempDir()

	schemaPath := filepath.Join(tmpDir, "schema.cue")
	if err := os.WriteFile(schemaPath, []byte(`#Config: {name: string}`), 0644); err != nil {
		t.Fatalf("failed to write schema file: %v", err)
	}

	// Test with both YAML and JSON extensions
	testCases := []string{
		"nonexistent.yaml",
		"nonexistent.json",
	}

	for _, filename := range testCases {
		t.Run(filename, func(t *testing.T) {
			nonExistentPath := filepath.Join(tmpDir, filename)

			results := ValidateFiles(schemaPath, []string{nonExistentPath})

			if len(results) != 1 {
				t.Fatalf("expected 1 result, got %d", len(results))
			}

			if results[0].IsValid {
				t.Error("expected validation to fail for non-existent file")
			}

			if !strings.Contains(results[0].Errors[0].Problem, "failed to read file") {
				t.Errorf("expected file reading error, got: %s", results[0].Errors[0].Problem)
			}
		})
	}
}

func TestValidateFilesWithUnsupportedFormat(t *testing.T) {
	tmpDir := t.TempDir()

	schemaPath := filepath.Join(tmpDir, "schema.cue")
	if err := os.WriteFile(schemaPath, []byte(`#Config: {name: string}`), 0644); err != nil {
		t.Fatalf("failed to write schema file: %v", err)
	}

	// Test various unsupported formats
	testCases := []struct {
		filename string
		content  string
	}{
		{"config.toml", `name = "test"`},
		{"config.ini", `name=test`},
		{"config.xml", `<name>test</name>`},
	}

	for _, tc := range testCases {
		t.Run(tc.filename, func(t *testing.T) {
			configPath := filepath.Join(tmpDir, tc.filename)
			if err := os.WriteFile(configPath, []byte(tc.content), 0644); err != nil {
				t.Fatalf("failed to write config file: %v", err)
			}

			results := ValidateFiles(schemaPath, []string{configPath})

			if len(results) != 1 {
				t.Fatalf("expected 1 result, got %d", len(results))
			}

			if results[0].IsValid {
				t.Error("expected validation to fail for unsupported format")
			}

			if !strings.Contains(results[0].Errors[0].Problem, "unsupported file format") {
				t.Errorf("expected unsupported format error, got: %s", results[0].Errors[0].Problem)
			}
		})
	}
}
