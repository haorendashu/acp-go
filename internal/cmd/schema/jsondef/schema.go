package jsondef

import (
	"encoding/json"
	"fmt"
	"os"
	"slices"
)

// Schema represents a JSON Schema definition.
// Only fields needed for Go type generation are included.
type Schema struct {
	Ref         string             `json:"$ref,omitempty"`
	Defs        map[string]*Schema `json:"$defs,omitempty"`
	Type        SchemaType         `json:"type,omitempty"`
	Description *string            `json:"description,omitempty"`
	Properties  map[string]*Schema `json:"properties,omitempty"`
	Required    []string           `json:"required,omitempty"`
	Items       *Schema            `json:"items,omitempty"`
	Enum        []json.RawMessage  `json:"enum,omitempty"`
	Const       *ConstValue        `json:"const,omitempty"`
	OneOf       []*Schema          `json:"oneOf,omitempty"`
	AnyOf       []*Schema          `json:"anyOf,omitempty"`
	AllOf       []*Schema          `json:"allOf,omitempty"`
	AdditionalProperties *AdditionalProps  `json:"additionalProperties,omitempty"`
	Discriminator       *Discriminator   `json:"discriminator,omitempty"`
	Default             json.RawMessage  `json:"default,omitempty"`
	Title               string           `json:"title,omitempty"`
}

// SchemaType handles JSON Schema "type" which can be a string or array of strings.
type SchemaType []string

func (st *SchemaType) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		*st = SchemaType{s}
		return nil
	}
	var arr []string
	if err := json.Unmarshal(data, &arr); err != nil {
		return fmt.Errorf("type must be a string or array of strings: %w", err)
	}
	*st = SchemaType(arr)
	return nil
}

func (st SchemaType) MarshalJSON() ([]byte, error) {
	if len(st) == 1 {
		return json.Marshal(st[0])
	}
	return json.Marshal([]string(st))
}

// Contains checks if the schema type list contains a specific type.
func (st SchemaType) Contains(t string) bool {
	return slices.Contains(st, t)
}

// ConstValue wraps a JSON const value.
type ConstValue struct {
	Value any
}

func (cv *ConstValue) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &cv.Value)
}

func (cv ConstValue) MarshalJSON() ([]byte, error) {
	return json.Marshal(cv.Value)
}

// StringValue returns the const value as a string, if it is one.
func (cv *ConstValue) StringValue() (string, bool) {
	s, ok := cv.Value.(string)
	return s, ok
}

// AdditionalProps represents JSON Schema additionalProperties, which can be a bool or a Schema.
type AdditionalProps struct {
	Bool   *bool
	Schema *Schema
}

func (ap *AdditionalProps) UnmarshalJSON(data []byte) error {
	var b bool
	if err := json.Unmarshal(data, &b); err == nil {
		ap.Bool = &b
		return nil
	}
	var s Schema
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("additionalProperties must be bool or schema: %w", err)
	}
	ap.Schema = &s
	return nil
}

// Discriminator represents JSON Schema discriminator.
type Discriminator struct {
	PropertyName string `json:"propertyName"`
}

// LoadSchema loads and parses a JSON Schema file.
func LoadSchema(path string) (*Schema, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read schema file %s: %w", path, err)
	}

	var schema Schema
	if err := json.Unmarshal(data, &schema); err != nil {
		return nil, fmt.Errorf("failed to parse schema file %s: %w", path, err)
	}

	return &schema, nil
}
