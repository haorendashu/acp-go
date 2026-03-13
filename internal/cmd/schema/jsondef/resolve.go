package jsondef

import (
	"fmt"
	"slices"
	"strings"
)

// goAcronyms maps lowercase words to their Go-conventional uppercase form.
var goAcronyms = map[string]string{
	"id": "ID", "url": "URL", "uri": "URI",
	"http": "HTTP", "https": "HTTPS", "json": "JSON",
	"api": "API", "sql": "SQL", "ssh": "SSH",
	"tcp": "TCP", "udp": "UDP", "ip": "IP",
	"html": "HTML", "css": "CSS", "xml": "XML",
	"rpc": "RPC", "tls": "TLS", "ssl": "SSL",
	"eof": "EOF", "sse": "SSE", "mcp": "MCP",
	"fs": "FS", "ui": "UI", "io": "IO",
}

// ApplyGoAcronyms converts known acronyms in a PascalCase name to uppercase.
// e.g., "SessionId" → "SessionID", "McpServer" → "MCPServer"
func ApplyGoAcronyms(name string) string {
	// Split on uppercase boundaries
	var words []string
	var word strings.Builder
	for i, r := range name {
		if i > 0 && r >= 'A' && r <= 'Z' {
			words = append(words, word.String())
			word.Reset()
		}
		word.WriteRune(r)
	}
	if word.Len() > 0 {
		words = append(words, word.String())
	}

	for i, w := range words {
		if acronym, ok := goAcronyms[strings.ToLower(w)]; ok {
			words[i] = acronym
		}
	}
	return strings.Join(words, "")
}

// ResolveRef extracts the definition name from a $ref string.
// e.g., "#/$defs/ContentBlock" → "ContentBlock"
func ResolveRef(ref string) string {
	return strings.TrimPrefix(ref, "#/$defs/")
}

// ResolveRefGo extracts the definition name from a $ref and applies Go naming conventions.
func ResolveRefGo(ref string) string {
	return ApplyGoAcronyms(ResolveRef(ref))
}

// FilterNonNull returns schemas that are not null type.
func FilterNonNull(schemas []*Schema) []*Schema {
	var result []*Schema
	for _, s := range schemas {
		if !IsNullType(s) {
			result = append(result, s)
		}
	}
	return result
}

// IsNullType checks if a schema represents a null type.
func IsNullType(s *Schema) bool {
	return len(s.Type) == 1 && s.Type[0] == "null"
}

// IsNullable checks if a schema is nullable (contains null in type or in anyOf/oneOf).
func IsNullable(s *Schema) bool {
	if s.Type.Contains("null") {
		return true
	}
	if slices.ContainsFunc(s.OneOf, func(sub *Schema) bool { return IsNullable(sub) }) {
		return true
	}
	return slices.ContainsFunc(s.AnyOf, func(sub *Schema) bool { return IsNullable(sub) })
}

// IsString checks if the schema represents a string type.
func IsString(s *Schema) bool {
	return len(s.Type) > 0 && s.Type[0] == "string"
}

// IsUnionOfRefs checks if all schemas in the list are $ref types or null types,
// with at least one $ref present. Also handles allOf wrapping pattern,
// but only when the variant has no additional properties (which would indicate
// a discriminated union rather than a simple ref union).
func IsUnionOfRefs(schemas []*Schema) bool {
	if len(schemas) <= 1 {
		return false
	}
	hasRef := false
	for _, s := range schemas {
		if s.Ref != "" {
			hasRef = true
		} else if len(s.AllOf) == 1 && s.AllOf[0].Ref != "" && s.Properties == nil {
			hasRef = true
		} else if IsNullType(s) {
			// null type is allowed
		} else {
			return false
		}
	}
	return hasRef
}

// DetectDiscriminator analyzes variants to find a field that:
// - exists in ALL variants as a const value
// - has a DIFFERENT value in each variant
func DetectDiscriminator(variants []*Schema) string {
	if len(variants) < 2 {
		return ""
	}

	// Collect all property names that have const values across all variants
	candidateFields := make(map[string]map[string]bool) // fieldName → set of const values

	for _, schema := range variants {
		if schema.Properties == nil {
			continue
		}
		for propName, propSchema := range schema.Properties {
			if propSchema.Const != nil {
				switch propSchema.Const.Value.(type) {
				case string, float64:
					if candidateFields[propName] == nil {
						candidateFields[propName] = make(map[string]bool)
					}
					valueStr := fmt.Sprintf("%v", propSchema.Const.Value)
					candidateFields[propName][valueStr] = true
				}
			}
		}
	}

	// Find fields that appear in ALL (or all-but-default) variants with DIFFERENT values
	for fieldName, values := range candidateFields {
		variantCount := 0
		for _, schema := range variants {
			if schema.Properties != nil {
				if prop, exists := schema.Properties[fieldName]; exists && prop.Const != nil {
					variantCount++
				}
			}
		}
		// All variants have discriminator
		if variantCount == len(variants) && len(values) == len(variants) {
			return fieldName
		}
		// Allow default variants: pure allOf/$ref without properties (no discriminator const)
		if variantCount >= 2 && len(values) == variantCount {
			allDefaultsValid := true
			for _, schema := range variants {
				hasDisc := false
				if schema.Properties != nil {
					if prop, exists := schema.Properties[fieldName]; exists && prop.Const != nil {
						hasDisc = true
					}
				}
				if !hasDisc {
					// Remaining variant must be a pure allOf ref (default variant)
					if len(schema.AllOf) == 0 || schema.AllOf[0].Ref == "" {
						allDefaultsValid = false
						break
					}
				}
			}
			if allDefaultsValid {
				return fieldName
			}
		}
	}

	return ""
}

// mapJSONTypeToGo maps a single JSON Schema type string to a Go type name.
func mapJSONTypeToGo(jsonType string) string {
	switch jsonType {
	case "string":
		return "string"
	case "boolean":
		return "bool"
	case "number":
		return "float64"
	case "integer":
		return "int64"
	default:
		return jsonType
	}
}

// GetGoTypeName converts a JSON Schema type to a Go type name.
func GetGoTypeName(s *Schema) string {
	if len(s.Type) == 0 {
		if s.Ref != "" {
			return ResolveRefGo(s.Ref)
		}
		return "any"
	}

	// Handle map types: type=object with additionalProperties and no properties
	baseType := s.Type[0]
	if baseType == "null" && len(s.Type) == 2 {
		for _, t := range s.Type {
			if t != "null" {
				baseType = t
				break
			}
		}
	}
	if baseType == "object" && s.Properties == nil && s.AdditionalProperties != nil && s.AdditionalProperties.Schema != nil {
		valueType := GetGoTypeName(s.AdditionalProperties.Schema)
		return "map[string]" + valueType
	}

	typeName := GetPrimitiveGoType(s)
	nullable := len(s.Type) == 2 && s.Type.Contains("null")

	if nullable {
		switch typeName {
		case "bool", "int64", "float64":
			typeName = "*" + typeName
		}
	}

	return typeName
}

// GetPrimitiveGoType returns the Go type for a primitive JSON Schema type (no pointer wrapping).
func GetPrimitiveGoType(s *Schema) string {
	if len(s.Type) == 0 {
		return "any"
	}
	for _, t := range s.Type {
		if t != "null" {
			return mapJSONTypeToGo(t)
		}
	}
	return "any"
}
