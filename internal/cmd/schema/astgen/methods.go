package astgen

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
)

// CreateMethodFromSource parses a Go method source and returns the FuncDecl AST node.
// The source should be a complete method definition including func keyword.
func CreateMethodFromSource(source string) (*ast.FuncDecl, error) {
	// Wrap in a package to make it parseable
	src := fmt.Sprintf("package p\n\ntype __placeholder struct{}\n\n%s", source)

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "", src, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("failed to parse method source: %w", err)
	}

	// Find the method declaration (skip any type declarations)
	for _, decl := range file.Decls {
		if fn, ok := decl.(*ast.FuncDecl); ok {
			return fn, nil
		}
	}

	return nil, fmt.Errorf("no function declaration found in source")
}

// CreateFuncFromSource parses a Go function source (no receiver) and returns the FuncDecl.
func CreateFuncFromSource(source string) (*ast.FuncDecl, error) {
	src := fmt.Sprintf("package p\n\n%s", source)

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "", src, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("failed to parse function source: %w", err)
	}

	for _, decl := range file.Decls {
		if fn, ok := decl.(*ast.FuncDecl); ok {
			return fn, nil
		}
	}

	return nil, fmt.Errorf("no function declaration found in source")
}

// CreateDeclsFromSource parses Go source code and returns all declarations.
// Useful for generating multiple related declarations at once.
func CreateDeclsFromSource(source string) ([]ast.Decl, error) {
	src := fmt.Sprintf("package p\n\n%s", source)

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "", src, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("failed to parse source: %w", err)
	}

	return file.Decls, nil
}
