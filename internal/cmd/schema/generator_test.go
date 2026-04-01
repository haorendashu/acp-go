package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerator_LoadSchema(t *testing.T) {
	testDir := "testdata_gen"
	validSchemaFile := filepath.Join(testDir, "valid.json")
	invalidSchemaFile := filepath.Join(testDir, "invalid.json")

	defer os.RemoveAll(testDir)

	os.MkdirAll(testDir, 0755)
	os.WriteFile(validSchemaFile, []byte(`{
		"type": "object",
		"properties": {
			"name": {"type": "string"}
		}
	}`), 0644)
	os.WriteFile(invalidSchemaFile, []byte(`{invalid json`), 0644)

	tests := []struct {
		name       string
		configFile string
		wantErr    bool
		errMsg     string
	}{
		{
			name:       "valid schema",
			configFile: validSchemaFile,
			wantErr:    false,
		},
		{
			name:       "invalid schema",
			configFile: invalidSchemaFile,
			wantErr:    true,
			errMsg:     "failed to parse schema file",
		},
		{
			name:       "non-existent file",
			configFile: "non/existent/file.json",
			wantErr:    true,
			errMsg:     "failed to read schema file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{
				InputFile:   tt.configFile,
				OutputFile:  "output.go",
				PackageName: "test",
			}

			generator := NewGenerator(config)
			err := generator.LoadSchema()

			if tt.wantErr {
				if err == nil {
					t.Errorf("Generator.LoadSchema() error = nil, wantErr %v", tt.wantErr)
					return
				}
				if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Generator.LoadSchema() error = %v, want error containing %v", err, tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("Generator.LoadSchema() error = %v, wantErr %v", err, tt.wantErr)
					return
				}

				if generator.schema == nil {
					t.Error("Generator.LoadSchema() schema is nil after successful load")
				}
			}
		})
	}
}

func TestGenerator_Generate(t *testing.T) {
	testDir := "testdata_gen2"
	schemaFile := filepath.Join(testDir, "enum.json")

	defer os.RemoveAll(testDir)

	os.MkdirAll(testDir, 0755)
	os.WriteFile(schemaFile, []byte(`{
		"$defs": {
			"Status": {
				"type": "string",
				"oneOf": [
					{
						"const": "active",
						"description": "User is active"
					},
					{
						"const": "inactive",
						"description": "User is inactive"
					}
				]
			}
		}
	}`), 0644)

	config := &Config{
		InputFile:   schemaFile,
		OutputFile:  "output.go",
		PackageName: "test",
	}

	generator := NewGenerator(config)

	err := generator.LoadSchema()
	if err != nil {
		t.Fatalf("Failed to load schema: %v", err)
	}

	err = generator.Generate()
	if err != nil {
		t.Errorf("Generator.Generate() error = %v", err)
	}

	// Test generation without loaded schema
	generator2 := NewGenerator(config)
	err = generator2.Generate()
	if err == nil {
		t.Error("Generator.Generate() should fail when schema not loaded")
	}
	if !strings.Contains(err.Error(), "schema not loaded") {
		t.Errorf("Generator.Generate() error = %v, want error about schema not loaded", err)
	}
}

func TestGenerator_SaveToFile(t *testing.T) {
	testDir := "testdata_gen3"
	outputDir := "output_gen"
	schemaFile := filepath.Join(testDir, "simple.json")
	outputFile := filepath.Join(outputDir, "types.go")

	defer func() {
		os.RemoveAll(testDir)
		os.RemoveAll(outputDir)
	}()

	os.MkdirAll(testDir, 0755)
	os.WriteFile(schemaFile, []byte(`{
		"$defs": {
			"SimpleType": {
				"type": "string"
			}
		}
	}`), 0644)

	config := &Config{
		InputFile:   schemaFile,
		OutputFile:  outputFile,
		PackageName: "test",
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
		t.Errorf("Generator.SaveToFile() error = %v", err)
		return
	}

	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		t.Error("Output file was not created")
	}

	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Errorf("Failed to read output file: %v", err)
		return
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "package test") {
		t.Error("Output file should contain correct package declaration")
	}

	if !strings.Contains(contentStr, "SimpleType") {
		t.Error("Output file should contain generated type")
	}
}

func TestGenerator_StructFieldsAlphabeticalOrder(t *testing.T) {
	testDir := "testdata_order"
	schemaFile := filepath.Join(testDir, "test.json")
	outputFile := filepath.Join(testDir, "output.go")

	defer os.RemoveAll(testDir)

	os.MkdirAll(testDir, 0755)
	os.WriteFile(schemaFile, []byte(`{
		"$defs": {
			"TestStruct": {
				"type": "object",
				"properties": {
					"zField": {"type": "string", "description": "Z field"},
					"aField": {"type": "string", "description": "A field"},
					"mField": {"type": "string", "description": "M field"},
					"bField": {"type": "string", "description": "B field"}
				}
			}
		}
	}`), 0644)

	config := &Config{
		InputFile:   schemaFile,
		OutputFile:  outputFile,
		PackageName: "test",
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
		t.Errorf("Generator.SaveToFile() error = %v", err)
		return
	}

	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Errorf("Failed to read output file: %v", err)
		return
	}

	contentStr := string(content)

	aFieldPos := strings.Index(contentStr, "AField")
	bFieldPos := strings.Index(contentStr, "BField")
	mFieldPos := strings.Index(contentStr, "MField")
	zFieldPos := strings.Index(contentStr, "ZField")

	if aFieldPos == -1 || bFieldPos == -1 || mFieldPos == -1 || zFieldPos == -1 {
		t.Logf("Generated content:\n%s", contentStr)
		t.Error("Not all fields found in generated code")
		return
	}

	if !(aFieldPos < bFieldPos && bFieldPos < mFieldPos && mFieldPos < zFieldPos) {
		t.Errorf("Fields are not in alphabetical order. Positions: A=%d, B=%d, M=%d, Z=%d",
			aFieldPos, bFieldPos, mFieldPos, zFieldPos)
	}
}

func TestGenerator_OneOfWithDiscriminator(t *testing.T) {
	testDir := "testdata_oneof"
	schemaFile := filepath.Join(testDir, "test.json")
	outputFile := filepath.Join(testDir, "output.go")

	defer os.RemoveAll(testDir)

	os.MkdirAll(testDir, 0755)
	os.WriteFile(schemaFile, []byte(`{
		"$defs": {
			"TestOneOf": {
				"description": "Test OneOf with type discriminator",
				"oneOf": [
					{
						"description": "Text variant",
						"properties": {
							"text": {"type": "string"},
							"type": {"const": "text", "type": "string"}
						},
						"required": ["type", "text"],
						"type": "object"
					},
					{
						"description": "Number variant",
						"properties": {
							"number": {"type": "integer"},
							"type": {"const": "number", "type": "string"}
						},
						"required": ["type", "number"],
						"type": "object"
					}
				]
			}
		}
	}`), 0644)

	config := &Config{
		InputFile:   schemaFile,
		OutputFile:  outputFile,
		PackageName: "test",
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
		t.Errorf("Generator.SaveToFile() error = %v", err)
		return
	}

	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Errorf("Failed to read output file: %v", err)
		return
	}

	contentStr := string(content)

	// Verify variant field using marker interface pattern
	if !strings.Contains(contentStr, "variant testOneOfVariant") {
		t.Error("Expected variant field with marker interface not found")
		t.Logf("Generated content:\n%s", contentStr)
	}

	// Verify marker interface
	if !strings.Contains(contentStr, "isTestOneOfVariant()") {
		t.Error("Expected marker interface method not found")
	}

	// Verify variant structs were generated
	if !strings.Contains(contentStr, "type TestOneOfText struct") {
		t.Error("Expected TestOneOfText struct not found")
	}
	if !strings.Contains(contentStr, "type TestOneOfNumber struct") {
		t.Error("Expected TestOneOfNumber struct not found")
	}

	// Verify As* accessor methods
	if !strings.Contains(contentStr, "AsText()") {
		t.Error("Expected AsText accessor method not found")
	}
	if !strings.Contains(contentStr, "AsNumber()") {
		t.Error("Expected AsNumber accessor method not found")
	}
}

func TestGenerator_OneOfWithJSONMethods(t *testing.T) {
	testDir := "testdata_json"
	schemaFile := filepath.Join(testDir, "test.json")
	outputFile := filepath.Join(testDir, "output.go")

	defer os.RemoveAll(testDir)

	os.MkdirAll(testDir, 0755)
	os.WriteFile(schemaFile, []byte(`{
		"$defs": {
			"TestOneOf": {
				"description": "Test OneOf with JSON methods",
				"oneOf": [
					{
						"description": "Text variant",
						"properties": {
							"text": {"type": "string"},
							"type": {"const": "text", "type": "string"}
						},
						"required": ["type", "text"],
						"type": "object"
					},
					{
						"description": "Number variant",
						"properties": {
							"number": {"type": "integer"},
							"type": {"const": "number", "type": "string"}
						},
						"required": ["type", "number"],
						"type": "object"
					}
				]
			}
		}
	}`), 0644)

	config := &Config{
		InputFile:   schemaFile,
		OutputFile:  outputFile,
		PackageName: "test",
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
		t.Errorf("Generator.SaveToFile() error = %v", err)
		return
	}

	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Errorf("Failed to read output file: %v", err)
		return
	}

	contentStr := string(content)

	// Verify union struct with variant field
	if !strings.Contains(contentStr, "variant testOneOfVariant") {
		t.Error("Expected variant field not found")
		t.Logf("Generated content:\n%s", contentStr)
	}

	// Verify JSON methods
	if !strings.Contains(contentStr, "MarshalJSON") {
		t.Error("Expected MarshalJSON method not found")
	}
	if !strings.Contains(contentStr, "UnmarshalJSON") {
		t.Error("Expected UnmarshalJSON method not found")
	}

	// Verify imports
	if !strings.Contains(contentStr, `"encoding/json"`) {
		t.Error("Expected encoding/json import not found")
	}
	if !strings.Contains(contentStr, `"fmt"`) {
		t.Error("Expected fmt import not found")
	}
}

func TestNewGenerator(t *testing.T) {
	config := &Config{
		InputFile:   "input.json",
		OutputFile:  "output.go",
		PackageName: "test",
	}

	generator := NewGenerator(config)

	if generator == nil {
		t.Fatal("NewGenerator() returned nil")
	}

	if generator.config != config {
		t.Error("NewGenerator() config not set correctly")
	}

	if generator.builder == nil {
		t.Error("NewGenerator() builder not initialized")
	}

	if generator.schema != nil {
		t.Error("NewGenerator() schema should be nil initially")
	}
}

func TestGenerator_StructFieldUnionFallsBackToRawMessage(t *testing.T) {
	testDir := "testdata_union_field"
	schemaFile := filepath.Join(testDir, "test.json")
	outputFile := filepath.Join(testDir, "output.go")

	defer os.RemoveAll(testDir)

	os.MkdirAll(testDir, 0755)
	os.WriteFile(schemaFile, []byte(`{
		"$defs": {
			"ExampleRequest": {
				"type": "object",
				"properties": {
					"id": {"type": "string"},
					"method": {"type": "string"},
					"params": {
						"anyOf": [
							{"$ref": "#/$defs/FooParams"},
							{"$ref": "#/$defs/BarParams"},
							{"type": "null"}
						]
					}
				},
				"required": ["id", "method" ]
			},
			"FooParams": {
				"type": "object",
				"properties": {"foo": {"type": "string"}}
			},
			"BarParams": {
				"type": "object",
				"properties": {"bar": {"type": "integer"}}
			}
		}
	}`), 0644)

	config := &Config{
		InputFile:   schemaFile,
		OutputFile:  outputFile,
		PackageName: "test",
	}

	generator := NewGenerator(config)
	if err := generator.LoadSchema(); err != nil {
		t.Fatalf("Failed to load schema: %v", err)
	}
	if err := generator.Generate(); err != nil {
		t.Fatalf("Failed to generate: %v", err)
	}
	if err := generator.SaveToFile(); err != nil {
		t.Fatalf("Failed to save: %v", err)
	}

	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "type ExampleRequest struct") {
		t.Fatalf("expected ExampleRequest struct to be generated, got:\n%s", contentStr)
	}
	if !strings.Contains(contentStr, "Params json.RawMessage `json:\"params,omitempty\"`") {
		t.Fatalf("expected Params json.RawMessage fallback, got:\n%s", contentStr)
	}
	if !strings.Contains(contentStr, `"encoding/json"`) {
		t.Fatalf("expected encoding/json import for RawMessage, got:\n%s", contentStr)
	}
}

func TestGenerator_OpaqueUnknownDefinitionFallsBackToAny(t *testing.T) {
	testDir := "testdata_opaque_unknown"
	schemaFile := filepath.Join(testDir, "test.json")
	outputFile := filepath.Join(testDir, "output.go")

	defer os.RemoveAll(testDir)

	os.MkdirAll(testDir, 0755)
	os.WriteFile(schemaFile, []byte(`{
		"$defs": {
			"ExtRequest": {
				"description": "Opaque extension request"
			}
		}
	}`), 0644)

	config := &Config{
		InputFile:   schemaFile,
		OutputFile:  outputFile,
		PackageName: "test",
	}

	generator := NewGenerator(config)
	if err := generator.LoadSchema(); err != nil {
		t.Fatalf("Failed to load schema: %v", err)
	}
	if err := generator.Generate(); err != nil {
		t.Fatalf("Failed to generate: %v", err)
	}
	if err := generator.SaveToFile(); err != nil {
		t.Fatalf("Failed to save: %v", err)
	}

	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "type ExtRequest any") {
		t.Fatalf("expected opaque definition to fall back to any, got:\n%s", contentStr)
	}
}

func TestGenerator_JSONRPCResponseEnvelopeGeneratesStrongStruct(t *testing.T) {
	testDir := "testdata_response_envelope"
	schemaFile := filepath.Join(testDir, "test.json")
	outputFile := filepath.Join(testDir, "output.go")

	defer os.RemoveAll(testDir)

	os.MkdirAll(testDir, 0755)
	os.WriteFile(schemaFile, []byte(`{
		"$defs": {
			"RequestId": {
				"type": ["string", "integer"]
			},
			"Error": {
				"type": "object",
				"properties": {
					"message": {"type": "string"}
				}
			},
			"AgentResponse": {
				"description": "A response from the agent.",
				"anyOf": [
					{
						"type": "object",
						"properties": {
							"id": {"$ref": "#/$defs/RequestId"},
							"result": {}
						},
						"required": ["id", "result"]
					},
					{
						"type": "object",
						"properties": {
							"id": {"$ref": "#/$defs/RequestId"},
							"error": {"$ref": "#/$defs/Error"}
						},
						"required": ["id", "error"]
					}
				]
			}
		}
	}`), 0644)

	config := &Config{
		InputFile:   schemaFile,
		OutputFile:  outputFile,
		PackageName: "test",
	}

	generator := NewGenerator(config)
	if err := generator.LoadSchema(); err != nil {
		t.Fatalf("Failed to load schema: %v", err)
	}
	if err := generator.Generate(); err != nil {
		t.Fatalf("Failed to generate: %v", err)
	}
	if err := generator.SaveToFile(); err != nil {
		t.Fatalf("Failed to save: %v", err)
	}

	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "type AgentResponse struct") {
		t.Fatalf("expected AgentResponse struct to be generated, got:\n%s", contentStr)
	}
	if !strings.Contains(contentStr, "json:\"id\"`") || !strings.Contains(contentStr, "RequestID") {
		t.Fatalf("expected ID field in response envelope, got:\n%s", contentStr)
	}
	if !strings.Contains(contentStr, "Result json.RawMessage `json:\"result,omitempty\"`") {
		t.Fatalf("expected Result raw message field in response envelope, got:\n%s", contentStr)
	}
	if !strings.Contains(contentStr, "json:\"error,omitempty\"`") || !strings.Contains(contentStr, "*Error") {
		t.Fatalf("expected Error field in response envelope, got:\n%s", contentStr)
	}
	if !strings.Contains(contentStr, `"encoding/json"`) {
		t.Fatalf("expected encoding/json import for response envelope, got:\n%s", contentStr)
	}
}
