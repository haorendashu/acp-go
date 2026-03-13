package astgen

import (
	"go/ast"
	"go/token"
	"strconv"
	"strings"
)

// Ident creates a simple identifier.
func Ident(name string) *ast.Ident {
	return ast.NewIdent(name)
}

// TypeExpr converts a Go type name string to an ast.Expr.
// Handles basic types, pointers, slices, maps, and qualified names.
func TypeExpr(typeName string) ast.Expr {
	if typeName == "" || typeName == "any" || typeName == "interface{}" {
		return Ident("any")
	}

	// Pointer type
	if strings.HasPrefix(typeName, "*") {
		return PointerExpr(TypeExpr(typeName[1:]))
	}

	// Slice type
	if strings.HasPrefix(typeName, "[]") {
		return SliceExpr(TypeExpr(typeName[2:]))
	}

	// Map type
	if strings.HasPrefix(typeName, "map[") {
		return parseMapType(typeName)
	}

	// Qualified name (e.g., json.RawMessage)
	if dot := strings.LastIndex(typeName, "."); dot > 0 {
		pkg := typeName[:dot]
		name := typeName[dot+1:]
		return SelectorExpr(pkg, name)
	}

	return Ident(typeName)
}

// PointerExpr creates a pointer type expression: *T
func PointerExpr(elem ast.Expr) ast.Expr {
	return &ast.StarExpr{X: elem}
}

// SliceExpr creates a slice type expression: []T
func SliceExpr(elem ast.Expr) ast.Expr {
	return &ast.ArrayType{Elt: elem}
}

// ArrayExpr creates an array type expression: [N]T
func ArrayExpr(length ast.Expr, elem ast.Expr) ast.Expr {
	return &ast.ArrayType{Len: length, Elt: elem}
}

// MapExpr creates a map type expression: map[K]V
func MapExpr(key, value ast.Expr) ast.Expr {
	return &ast.MapType{Key: key, Value: value}
}

// SelectorExpr creates a qualified identifier: pkg.Name
func SelectorExpr(pkg, name string) ast.Expr {
	return &ast.SelectorExpr{
		X:   Ident(pkg),
		Sel: Ident(name),
	}
}

// InterfaceExpr creates an empty interface expression.
func InterfaceExpr() ast.Expr {
	return Ident("any")
}

// FieldTag creates a struct tag literal.
func FieldTag(tag string) *ast.BasicLit {
	if tag == "" {
		return nil
	}
	return &ast.BasicLit{
		Kind:  token.STRING,
		Value: "`" + tag + "`",
	}
}

// StringLit creates a string literal expression.
func StringLit(s string) *ast.BasicLit {
	return &ast.BasicLit{
		Kind:  token.STRING,
		Value: `"` + s + `"`,
	}
}

// IntLit creates an integer literal expression.
func IntLit(n int) *ast.BasicLit {
	return &ast.BasicLit{
		Kind:  token.INT,
		Value: strconv.Itoa(n),
	}
}

func parseMapType(typeName string) ast.Expr {
	// Find matching ] for the key type
	keyStart := 4 // after "map["
	bracketCount := 0
	keyEnd := -1

	for i := keyStart; i < len(typeName); i++ {
		switch typeName[i] {
		case '[':
			bracketCount++
		case ']':
			if bracketCount == 0 {
				keyEnd = i
			} else {
				bracketCount--
			}
		}
		if keyEnd != -1 {
			break
		}
	}

	if keyEnd == -1 {
		return Ident(typeName)
	}

	keyType := typeName[keyStart:keyEnd]
	valueType := typeName[keyEnd+1:]
	return MapExpr(TypeExpr(keyType), TypeExpr(valueType))
}
