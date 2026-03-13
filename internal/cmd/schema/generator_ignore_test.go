package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerator_GenerateWithIgnoreErrors(t *testing.T) {
	testDir := "testdata_ignore"
	schemaFile := filepath.Join(testDir, "problem.json")

	defer os.RemoveAll(testDir)

	// Schema with valid types including an empty struct (which is now handled correctly)
	os.MkdirAll(testDir, 0755)
	os.WriteFile(schemaFile, []byte(`{
		"$defs": {
			"ValidType": {
				"type": "string",
				"description": "A valid type"
			},
			"EmptyStruct": {
				"type": "object",
				"description": "A struct with no properties"
			},
			"AnotherValidType": {
				"type": "integer",
				"description": "Another valid type"
			}
		}
	}`), 0644)

	// With the new generator, empty structs are handled correctly
	config := &Config{
		InputFile:    schemaFile,
		OutputFile:   "output.go",
		PackageName:  "test",
		IgnoreErrors: false,
	}

	generator := NewGenerator(config)

	err := generator.LoadSchema()
	if err != nil {
		t.Fatalf("Failed to load schema: %v", err)
	}

	err = generator.Generate()
	if err != nil {
		t.Errorf("Generator.Generate() should succeed with empty structs: %v", err)
	}
}

func TestGenerator_GenerateWithIgnoreTypes(t *testing.T) {
	testDir := "testdata_ignore_types"
	schemaFile := filepath.Join(testDir, "test.json")
	outputFile := filepath.Join(testDir, "output.go")

	defer os.RemoveAll(testDir)

	os.MkdirAll(testDir, 0755)
	os.WriteFile(schemaFile, []byte(`{
		"$defs": {
			"IncludedType": {
				"type": "string",
				"description": "This type should be included"
			},
			"ExcludedType": {
				"type": "string",
				"description": "This type should be excluded"
			}
		}
	}`), 0644)

	config := &Config{
		InputFile:   schemaFile,
		OutputFile:  outputFile,
		PackageName: "test",
		IgnoreTypes: []string{"ExcludedType"},
	}

	err := config.Validate()
	if err != nil {
		t.Fatalf("Failed to validate config: %v", err)
	}

	generator := NewGenerator(config)

	err = generator.LoadSchema()
	if err != nil {
		t.Fatalf("Failed to load schema: %v", err)
	}

	err = generator.Generate()
	if err != nil {
		t.Fatalf("Failed to generate: %v", err)
	}

	err = generator.SaveToFile()
	if err != nil {
		t.Fatalf("Failed to save: %v", err)
	}

	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	contentStr := string(content)

	if !strings.Contains(contentStr, "IncludedType") {
		t.Error("Output should contain IncludedType")
	}

	if strings.Contains(contentStr, "ExcludedType") {
		t.Error("Output should not contain ExcludedType")
	}

	if generator.GetSkippedCount() != 1 {
		t.Errorf("Expected 1 skipped item, got %d", generator.GetSkippedCount())
	}
}

func TestGenerator_IgnoreErrorsOutput(t *testing.T) {
	testDir := "testdata_ignore2"
	schemaFile := filepath.Join(testDir, "mixed.json")
	outputFile := filepath.Join(testDir, "output.go")

	defer os.RemoveAll(testDir)

	os.MkdirAll(testDir, 0755)
	os.WriteFile(schemaFile, []byte(`{
		"$defs": {
			"ValidEnum": {
				"type": "string",
				"oneOf": [
					{"const": "valid", "description": "Valid value"}
				]
			},
			"EmptyStruct": {
				"type": "object",
				"description": "Empty struct - handled correctly now"
			},
			"ValidType": {
				"type": "string",
				"description": "Another valid type"
			}
		}
	}`), 0644)

	config := &Config{
		InputFile:    schemaFile,
		OutputFile:   outputFile,
		PackageName:  "test",
		IgnoreErrors: true,
	}

	err := config.Validate()
	if err != nil {
		t.Fatalf("Failed to validate config: %v", err)
	}

	generator := NewGenerator(config)

	err = generator.LoadSchema()
	if err != nil {
		t.Fatalf("Failed to load schema: %v", err)
	}

	err = generator.Generate()
	if err != nil {
		t.Errorf("Generator.Generate() with ignore errors should not fail: %v", err)
		return
	}

	err = generator.SaveToFile()
	if err != nil {
		t.Errorf("Generator.SaveToFile() error = %v", err)
		return
	}

	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Errorf("Failed to read output file: %v", err)
		return
	}

	contentStr := string(content)

	if !strings.Contains(contentStr, "ValidEnum") {
		t.Error("Output should contain ValidEnum type")
	}

	if !strings.Contains(contentStr, "ValidType") {
		t.Error("Output should contain ValidType")
	}

	// EmptyStruct is now generated correctly as empty struct
	if !strings.Contains(contentStr, "EmptyStruct") {
		t.Error("Output should contain EmptyStruct (empty structs are now handled)")
	}
}
