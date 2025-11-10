package goparser

import (
	"errors"
	"fmt"
	"sort"
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
	BuildTags        []string // Build tags extracted from //go:build or // +build directives
	Structs          []*GoStruct
	Interfaces       []*GoInterface
	Imports          []*GoImport
	StructMethods    []*GoStructMethod
	CustomTypes      []*GoCustomType
	CustomFuncs      []*GoMethod
	VarAssignments   []*GoAssignment
	ConstAssignments []*GoAssignment
}

// FindMethodsByReceiver searches the file / package after struct and custom type receiver
// methods that matches the _receiver_ name.
func (g *GoFile) FindMethodsByReceiver(receiver string) []*GoStructMethod {

	list := []*GoStructMethod{}
	for i := range g.StructMethods {

		if contains(receiver, g.StructMethods[i].Receivers) {
			list = append(list, g.StructMethods[i])
		}

	}

	return list
}

// contains checks if any in the _arr_ matches the _name_. If found
// `true` is returned, otherwise `false` is returned.
func contains(name string, arr []string) bool {

	if len(arr) == 0 {
		return false
	}

	target := normalizeReceiverName(name)

	for i := range arr {
		if normalizeReceiverName(arr[i]) == target {
			return true
		}
	}

	return false
}

func normalizeReceiverName(name string) string {
	name = strings.TrimSpace(name)

	for len(name) > 0 {
		switch name[0] {
		case '*', '&':
			name = name[1:]
			continue
		}
		break
	}

	if idx := strings.Index(name, "["); idx != -1 {
		name = name[:idx]
	}

	if idx := strings.LastIndex(name, "."); idx != -1 {
		name = name[idx+1:]
	}

	return name
}

// ImportPath resolves the import path.
func (g *GoFile) ImportPath() (string, error) {
	if g.Module != nil {
		if g.FqPackage != "" {
			return g.FqPackage, nil
		}
		if resolved, err := g.Module.ResolvePackage(g.FilePath); err == nil {
			g.FqPackage = resolved
			return resolved, nil
		} else if !errors.Is(err, ErrModuleNotConfigured) {
			return "", err
		}
	}

	module, err := FindModule(g.FilePath)
	if err != nil {
		return "", fmt.Errorf("unable to determine import path for %s: %w. Configure ParseConfig.Module or run NewModule to resolve package names", g.FilePath, err)
	}

	resolved, err := module.ResolvePackage(g.FilePath)
	if err != nil {
		return "", err
	}

	g.Module = module
	g.FqPackage = resolved

	return resolved, nil
}

// DeclImports emits the imports
func (g *GoFile) DeclImports() string {
	if len(g.Imports) == 0 {
		return ""
	}

	if len(g.Imports) == 1 {

		if g.Imports[0].Name == "" {

			return fmt.Sprintf(`import "%s"`, g.Imports[0].Path)
		} else {

			return fmt.Sprintf(`import %s "%s"`, g.Imports[0].Name, g.Imports[0].Path)

		}
	}

	// Filter out any duplicate
	set := make(map[string]*GoImport)
	for i, imp := range g.Imports {

		if set[imp.Path] == nil {

			set[imp.Path] = g.Imports[i]

		}

	}

	// Sort imports
	keys := make([]string, 0, len(set))
	for k := range set {
		keys = append(keys, k)
	}

	sort.Slice(keys, func(i, j int) bool {
		iBasePkg := !strings.Contains(keys[i], "/")
		jBasePkg := !strings.Contains(keys[j], "/")
		if iBasePkg && !jBasePkg {
			return true
		}
		if !iBasePkg && jBasePkg {
			return false
		}
		return keys[i] < keys[j]
	})

	s := "import (\n"
	for _, k := range keys {

		imp := set[k]
		if imp.Name != "" {

			s += fmt.Sprintf("\t%s \"%s\"\n", imp.Name, k)

		} else {

			s += fmt.Sprintf("\t\"%s\"\n", k)

		}

	}

	return s + ")"
}
