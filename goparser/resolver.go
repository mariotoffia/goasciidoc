package goparser

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

// Resolver pure purpose is to resolve `GoFile`, `GoStructMethod` to
// `GoTag` and all other types in between.
type Resolver interface {
}

// ResolverImpl is the implementation of a `Resolver` where it operarates on
// a `GoModule` level.
type ResolverImpl struct {
	// module is the resolvers workspace.
	//
	// It may have sub-modules through contained `Resolver` instances.
	module *GoModule

	// resolvers is a map containing `Resolver` for each `GoModule` that this _module_
	// references and makes use of.
	//resolvers map[string] /*module name*/ Resolver

	// config is the configuration that this `Resolver` adheres to.
	config ParseConfig
	// Fully qualified filepath to this module (where _go.mod_ resides).
	filepath string
}

// NewResolver creates a new `Resolver` from the filepath to the _go.mod_ file
// or directory where _go.mod_ resides.
func NewResolver(config ParseConfig, filepath string) Resolver {

	res := &ResolverImpl{
		config: config,
	}

	return res.resolveModule(filepath)
}

func (r *ResolverImpl) resolveModule(fp string) Resolver {

	if !strings.HasSuffix(fp, "go.mod") {

		var err error
		fp, err = filepath.Abs(filepath.Join(fp, "go.mod"))

		if err != nil {
			panic(err)
		}

	}

	m, err := NewModule(fp)

	if err != nil {
		panic(err)
	}

	r.module = m
	r.filepath = fp

	return r
}

func (r *ResolverImpl) LoadAll() error {

	files, err := GetFilePaths(r.config, r.filepath)
	if err != nil {
		return err
	}

	m := map[string][]string{}
	for _, f := range files {

		dir := filepath.Dir(f)
		if list, ok := m[dir]; ok {
			m[dir] = append(list, f)
		} else {
			m[dir] = []string{f}
		}
	}

	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	for _, k := range keys {

		v := m[k]

		goFiles, err := ParseFiles(r.module, v...)
		if err != nil {
			return err
		}

		pkg := &GoPackage{
			GoFile: GoFile{
				Module:    r.module,
				Package:   goFiles[0].Package,
				FqPackage: goFiles[0].FqPackage,
				FilePath:  k,
				Decl:      goFiles[0].Decl,
			},
			Files: goFiles,
		}

		var b strings.Builder
		for _, gf := range goFiles {

			if gf.Doc != "" {
				fmt.Fprintf(&b, "%s\n", gf.Doc)
			}
			if len(gf.Structs) > 0 {
				pkg.Structs = append(pkg.Structs, gf.Structs...)
			}
			if len(gf.Interfaces) > 0 {
				pkg.Interfaces = append(pkg.Interfaces, gf.Interfaces...)
			}
			if len(gf.Imports) > 0 {
				pkg.Imports = append(pkg.Imports, gf.Imports...)
			}
			if len(gf.StructMethods) > 0 {
				pkg.StructMethods = append(pkg.StructMethods, gf.StructMethods...)
			}
			if len(gf.CustomTypes) > 0 {
				pkg.CustomTypes = append(pkg.CustomTypes, gf.CustomTypes...)
			}
			if len(gf.CustomFuncs) > 0 {
				pkg.CustomFuncs = append(pkg.CustomFuncs, gf.CustomFuncs...)
			}
			if len(gf.VarAssignments) > 0 {
				pkg.VarAssignments = append(pkg.VarAssignments, gf.VarAssignments...)
			}
			if len(gf.ConstAssignments) > 0 {
				pkg.ConstAssignments = append(pkg.ConstAssignments, gf.ConstAssignments...)
			}

		}

		pkg.Doc = b.String()
	}

	return nil
}
