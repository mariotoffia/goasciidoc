package goparser

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
)

func TestIsExportedNormalization(t *testing.T) {
	cases := map[string]bool{
		"Foo":                     true,
		"*Foo":                    true,
		"[]*Foo":                  true,
		"map[string]*pkg.Foo":     true,
		"chan<- Foo":              true,
		"<-chan Foo":              true,
		"(...Foo)":                true,
		"pkg.Bar":                 true,
		"[]string":                false,
		"map[int]*pkg.unexported": false,
		"func()":                  false,
	}

	for input, expected := range cases {
		if got := isExported(input); got != expected {
			t.Fatalf("expected isExported(%q) to be %v, got %v", input, expected, got)
		}
	}
}

func TestBuildVarAssignmentUsesValueSpecSourceWhenNoDoc(t *testing.T) {
	src := `package foo

const (
	TypeKindUnknown TypeKind = iota
	TypeKindIdent
)`

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "enum.go", src, parser.ParseComments)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	var genDecl *ast.GenDecl
	for _, decl := range file.Decls {
		if d, ok := decl.(*ast.GenDecl); ok && d.Tok == token.CONST {
			genDecl = d
			break
		}
	}
	if genDecl == nil {
		t.Fatalf("const declaration not found")
	}

	var targetSpec *ast.ValueSpec
	for _, spec := range genDecl.Specs {
		valueSpec, ok := spec.(*ast.ValueSpec)
		if !ok || len(valueSpec.Names) == 0 {
			continue
		}
		if valueSpec.Names[0].Name == "TypeKindIdent" {
			targetSpec = valueSpec
			break
		}
	}
	if targetSpec == nil {
		t.Fatalf("value spec for TypeKindIdent not found")
	}

	assignments := buildVarAssignment(&GoFile{}, genDecl, targetSpec, fileSource{
		data: []byte(src),
		fset: fset,
	})

	if len(assignments) != 1 {
		t.Fatalf("expected one assignment, got %d", len(assignments))
	}

	got := assignments[0].Decl
	want := "TypeKindIdent"
	if got != want {
		t.Fatalf("expected Decl %q, got %q", want, got)
	}

	if assignments[0].Doc != "" {
		t.Fatalf("expected no doc for Decl %q, got %q", want, assignments[0].Doc)
	}
}
