package astgen

import (
	"fmt"
	"go/ast"
	"go/token"
	"strings"
)

// StructField represents a field definition for struct generation.
type StructField struct {
	Name    string
	Type    ast.Expr
	Tag     string
	Comment string
	Embed   bool // anonymous embedding (no field name)
}

// ConstEntry represents a constant in a const block.
type ConstEntry struct {
	Name    string
	Type    string // optional
	Value   string
	Comment string
}

// StructVarEntry represents a field+value pair for anonymous struct variables.
type StructVarEntry struct {
	Name  string
	Type  string
	Value string
}

// TypeDef creates a type definition: type Name UnderlyingType
func TypeDef(name, underlying string, comment string) *ast.GenDecl {
	spec := &ast.TypeSpec{
		Name: Ident(name),
		Type: TypeExpr(underlying),
	}

	decl := &ast.GenDecl{
		Tok:   token.TYPE,
		Specs: []ast.Spec{spec},
	}

	if comment != "" {
		decl.Doc = makeCommentGroup(comment)
	}

	return decl
}

// TypeAlias creates a type alias: type Name = UnderlyingType
func TypeAlias(name, underlying string, comment string) *ast.GenDecl {
	spec := &ast.TypeSpec{
		Name:   Ident(name),
		Assign: 1, // non-zero means type alias
		Type:   TypeExpr(underlying),
	}

	decl := &ast.GenDecl{
		Tok:   token.TYPE,
		Specs: []ast.Spec{spec},
	}

	if comment != "" {
		decl.Doc = makeCommentGroup(comment)
	}

	return decl
}

// StructDef creates a struct type definition.
func StructDef(name string, fields []StructField, comment string) *ast.GenDecl {
	fieldList := make([]*ast.Field, 0, len(fields))
	for _, f := range fields {
		field := &ast.Field{
			Type: f.Type,
		}
		if !f.Embed {
			field.Names = []*ast.Ident{Ident(f.Name)}
		}
		if f.Tag != "" {
			field.Tag = FieldTag(f.Tag)
		}
		if f.Comment != "" {
			field.Doc = makeCommentGroup(f.Comment)
		}
		fieldList = append(fieldList, field)
	}

	spec := &ast.TypeSpec{
		Name: Ident(name),
		Type: &ast.StructType{
			Fields: &ast.FieldList{
				List: fieldList,
			},
		},
	}

	decl := &ast.GenDecl{
		Tok:   token.TYPE,
		Specs: []ast.Spec{spec},
	}

	if comment != "" {
		decl.Doc = makeCommentGroup(comment)
	}

	return decl
}

// EmptyStructDef creates an empty struct type definition.
func EmptyStructDef(name string, comment string) *ast.GenDecl {
	return StructDef(name, nil, comment)
}

// ConstBlockSource generates a const block as raw Go source text.
// This ensures correct comment placement for enum value descriptions.
func ConstBlockSource(consts []ConstEntry) string {
	var sb strings.Builder
	sb.WriteString("const (\n")
	for _, c := range consts {
		if c.Comment != "" {
			for _, line := range strings.Split(c.Comment, "\n") {
				if line != "" {
					sb.WriteString(fmt.Sprintf("\t// %s\n", line))
				} else {
					sb.WriteString("\t//\n")
				}
			}
		}
		if c.Type != "" {
			sb.WriteString(fmt.Sprintf("\t%s %s = %s\n", c.Name, c.Type, c.Value))
		} else {
			sb.WriteString(fmt.Sprintf("\t%s = %s\n", c.Name, c.Value))
		}
	}
	sb.WriteString(")")
	return sb.String()
}

// SingleConstSource creates a single const declaration as raw source.
func SingleConstSource(name, typ, value, comment string) string {
	return ConstBlockSource([]ConstEntry{{
		Name:    name,
		Type:    typ,
		Value:   value,
		Comment: comment,
	}})
}

// VarAnonymousStruct creates a var declaration with an anonymous struct literal.
func VarAnonymousStruct(name string, entries []StructVarEntry, comment string) *ast.GenDecl {
	// Build struct field list
	fields := make([]*ast.Field, 0, len(entries))
	kvs := make([]ast.Expr, 0, len(entries)*2)

	for _, e := range entries {
		fields = append(fields, &ast.Field{
			Names: []*ast.Ident{Ident(e.Name)},
			Type:  TypeExpr(e.Type),
		})
		kvs = append(kvs, &ast.KeyValueExpr{
			Key:   Ident(e.Name),
			Value: StringLit(e.Value),
		})
	}

	structType := &ast.StructType{
		Fields: &ast.FieldList{List: fields},
	}

	compositeLit := &ast.CompositeLit{
		Type: structType,
		Elts: kvs,
	}

	spec := &ast.ValueSpec{
		Names:  []*ast.Ident{Ident(name)},
		Values: []ast.Expr{compositeLit},
	}

	decl := &ast.GenDecl{
		Tok:   token.VAR,
		Specs: []ast.Spec{spec},
	}

	if comment != "" {
		decl.Doc = makeCommentGroup(comment)
	}

	return decl
}

func makeCommentGroup(text string) *ast.CommentGroup {
	lines := strings.Split(text, "\n")
	var list []*ast.Comment
	for _, line := range lines {
		if line != "" {
			list = append(list, &ast.Comment{Text: "// " + line})
		} else {
			list = append(list, &ast.Comment{Text: "//"})
		}
	}
	return &ast.CommentGroup{List: list}
}

