package goparser

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// GoFile represents a complete file
type GoFile struct {
	Module *GoModule
	// Package is the single package name where as FqPackage is the
	// fully qualified package (if Module) has been set.
	Package string
	// FqPackage is the fully qualified package name (if Module field)
	// is set to calculate the fq package name
	FqPackage        string
	FilePath         string
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
	importPath, err := filepath.Abs(g.FilePath)
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
