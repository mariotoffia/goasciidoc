package goparser

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
)

// GoFile represents a complete file
type GoFile struct {
	Package          string
	Path             string
	Doc              string
	Decl             string
	ImportFullDecl   string
	Structs          []*GoStruct
	Interfaces       []*GoInterface
	Imports          []*GoImport
	StructMethods    []*GoStructMethod
	CustomTypes      []*GoCustomType
	CustomFuncs      []*GoMethod
	VarAssigments    []*GoAssignment
	ConstAssignments []*GoAssignment
}

// ImportPath is for TODO:
func (g *GoFile) ImportPath() (string, error) {
	importPath, err := filepath.Abs(g.Path)
	if err != nil {
		return "", err
	}

	importPath = strings.Replace(importPath, "\\", "/", -1)

	goPath := strings.Replace(os.Getenv("GOPATH"), "\\", "/", -1)
	importPath = strings.TrimPrefix(importPath, goPath)
	importPath = strings.TrimPrefix(importPath, "/src/")

	importPath = strings.TrimSuffix(importPath, filepath.Base(importPath))
	importPath = strings.TrimSuffix(importPath, "/")

	return importPath, nil
}

// DeclImports emits the imports
func (g *GoFile) DeclImports() string {
	if len(g.Imports) == 0 {
		return ""
	}

	if len(g.Imports) == 1 {
		return fmt.Sprintf(`import "%s"`, g.Imports[0].Path)
	}

	s := "import (\n"
	for _, i := range g.Imports {
		s += fmt.Sprintf("\t\"%s\"\n", i.Path)
	}

	return s + ")"
}

// GoAssignment represents a single var assignment e.g. var pelle = 10
type GoAssignment struct {
	File *GoFile
	Name string
	Doc  string
	// Decl will be the same if multi var assignment on same row e.g. var pelle, lisa = 10, 19
	// then both pelle and list will have 'var pelle, lisa = 10, 19' as Decl
	Decl     string
	FullDecl string
}

// GoCustomType is a custom type definition
type GoCustomType struct {
	File *GoFile
	Name string
	Doc  string
	Type string
	Decl string
}

// GoImport represents a import of a package
type GoImport struct {
	File *GoFile
	Doc  string
	Name string
	Path string
}

// GoInterface specifies a interface definition
type GoInterface struct {
	File     *GoFile
	Doc      string
	Decl     string
	FullDecl string
	Name     string
	Methods  []*GoMethod
}

// GoMethod is a method on a struct, interface or just plain function
type GoMethod struct {
	File     *GoFile
	Name     string
	Doc      string
	Decl     string
	FullDecl string
	Params   []*GoType
	Results  []*GoType
}

// GoStructMethod is a GoMethod but has receivers and is positioned on a struct.
type GoStructMethod struct {
	GoMethod
	Receivers []string
}

// GoType represents a go type such as a array, map, custom type etc.
type GoType struct {
	File       *GoFile
	Name       string
	Type       string
	Underlying string
	Inner      []*GoType
}

// GoStruct represents a struct
type GoStruct struct {
	File     *GoFile
	Doc      string
	Decl     string
	FullDecl string
	Name     string
	Fields   []*GoField
}

// GoField is a field in a file or struct
type GoField struct {
	File   *GoFile
	Struct *GoStruct
	Doc    string
	Decl   string
	Name   string
	Type   string
	Tag    *GoTag
}

// GoTag is a tag on a struct field
type GoTag struct {
	File  *GoFile
	Field *GoField
	Value string
}

// Get returns a struct tag with the specified name e.g. json
func (g *GoTag) Get(key string) string {
	tag := strings.Replace(g.Value, "`", "", -1)
	return reflect.StructTag(tag).Get(key)
}

// Prefix is for an import - guess what prefix will be used
// in type declarations.  For examples:
//    "strings" -> "strings"
//    "net/http/httptest" -> "httptest"
// Libraries where the package name does not match
// will be mis-identified.
func (g *GoImport) Prefix() string {
	if g.Name != "" {
		return g.Name
	}

	path := strings.Trim(g.Path, "\"")
	lastSlash := strings.LastIndex(path, "/")
	if lastSlash == -1 {
		return path
	}

	return path[lastSlash+1:]
}
