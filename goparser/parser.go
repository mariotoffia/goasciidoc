package goparser

import (
	"fmt"
	"go/ast"
	"go/build"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"strings"

	"github.com/mariotoffia/goasciidoc/utils"
)

// ParseSinglePackageWalkerFunc is used in conjunction with `PackageParserImpl.Process`.
//
// If the ParseSinglePackageWalker is returning an error, parsing will immediately stop
// and the error is returned.
type ParseSinglePackageWalkerFunc func(*GoPackage) error

// ParseInlineFile will parse the code provided.
//
// To simulate package names set the path to some level
// equal to or greater than GoModule.Base. Otherwise just
// set path "" to ignore.
func ParseInlineFile(mod *GoModule, path, code string) (*GoFile, error) {

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "", code, parser.ParseComments)

	if err != nil {
		return nil, err
	}

	conf := types.Config{
		Error: func(err error) {
			fmt.Printf("TODO: accumulate errors: %s\n", err.Error())
		},
		Importer: importer.ForCompiler(fset, "source", nil),
		Sizes:    types.SizesFor(build.Default.Compiler, build.Default.GOARCH),
	}

	info := &types.Info{
		Types: make(map[ast.Expr]types.TypeAndValue),
		Defs:  make(map[*ast.Ident]types.Object),
		Uses:  make(map[*ast.Ident]types.Object),
	}

	pkg, err := conf.Check(file.Name.Name, fset, []*ast.File{file}, info)

	if err != nil {
		panic(err)
	}

	return AstParseFileEx(mod, path, []byte(code), file, fset, []*ast.File{file}, pkg, info)

}

// PackageParserImpl implements enumeration of files using
// package per package instead of individual files.
//
// This is to increase performance since go resolution is per
// package / module basis.
type PackageParserImpl struct {
	files ModulePathAndFiles
	err   error
	// deepResolve determines if it should try to resolve third-party modules and packages.
	//
	// This is disabled by default since it is *very* time consuming.
	deepResolve bool
}

// NewPackageParserImpl creates a new `PackageParserImpl` that will
// use the _pf_ to enumerate paths and files.
func NewPackageParserImpl(pf PathAndFiles) *PackageParserImpl {

	return &PackageParserImpl{files: FromPathAndFiles(pf)}

}

// PackageParserConfig affects how the `NewPackageParserImplFromPaths` behaves.
type PackageParserConfig struct {
	// Test denotes if test files (ending with _test.go) should be included or not
	// (default not included)
	Test bool
	// Internal determines if internal folders are included or not (default not)
	Internal bool
	// UnderScore, when set to true it will include directories beginning with _
	UnderScore bool
}

// NewPackageParserImplFromPaths creates a new `PackageParserImpl` that will iterate directories
// (recursively) and add explicit files in the paths.
//
// It is possible to use relative paths or fully qualified paths along with '.'
// for current directory. The paths are stat:ed so it will check if it is a file
// or directory and do accordingly. If file it will ignore configuration and blindly
// accept the file.
//
// It uses the _config_ to determine which files / paths to include while enumerating
// _go_ files.
//
// If any errors occurs, it will set the error state.
func NewPackageParserImplFromPaths(config PackageParserConfig, paths ...string) *PackageParserImpl {

	pf, err := FromPaths(config, paths...)

	return &PackageParserImpl{
		files: FromPathAndFiles(pf),
		err:   err,
	}

}

func (pp *PackageParserImpl) Error() error {
	return pp.err
}

func (pp *PackageParserImpl) ClearError() *PackageParserImpl {
	pp.err = nil
	return pp
}

func (pp *PackageParserImpl) UseDeepResolve() *PackageParserImpl {
	pp.deepResolve = true
	return pp
}

func (pp *PackageParserImpl) Process(cb ParseSinglePackageWalkerFunc) *PackageParserImpl {

	for _, modulePath := range pp.files.ModulePaths() {

		if modulePath == NonModuleName {
			fmt.Printf("TODO: implement when 'NonModuleName' = '%s'\n", NonModuleName)
			continue
		}

		module := ToModule(modulePath)

		for _, packagePath := range pp.files.PackagePaths(modulePath) {

			files := pp.files[modulePath][packagePath]

			if idx, ok := utils.HasSuffixString(files, "go.mod"); ok {

				files = utils.RemoveString(files, idx)

			}

			parsedFiles, err := pp.parseFiles(module, packagePath, files)

			if err != nil {
				panic(err)
			}

			pp.callback(cb, packagePath, module, parsedFiles)

		}

	}

	return pp
}

func (pp *PackageParserImpl) callback(
	cb ParseSinglePackageWalkerFunc,
	fp string,
	module *GoModule,
	files []*GoFile,
) {

	pkg := &GoPackage{
		GoFile: GoFile{
			Module:    module,
			Package:   files[0].Package,
			FqPackage: files[0].FqPackage,
			FilePath:  fp,
			Decl:      files[0].Decl,
		},
		Files: files,
	}

	var b strings.Builder

	for _, gf := range files {

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

	if err := cb(pkg); err != nil {
		panic(err)
	}

}

func (pp *PackageParserImpl) parseFiles(
	mod *GoModule,
	packagePath string,
	files []string,
) ([]*GoFile, error) {

	if len(files) == 0 {
		panic("must specify at least one file to file to parse")
	}

	if len(packagePath) == 0 {
		panic("must have a package path")
	}

	astFiles := make([]*ast.File, len(files))
	fset := token.NewFileSet()

	for i, p := range files {

		file, err := parser.ParseFile(fset, p, nil, parser.ParseComments)

		if err != nil {
			return nil, err
		}

		astFiles[i] = file

	}

	goFiles := make([]*GoFile, len(files))

	// To import sources from vendor and import package names mismatching import path
	// , we use "source" compile.
	// SEE: https://github.com/golang/example/tree/master/gotypes#introduction
	conf := types.Config{
		Error: func(err error) {
			fmt.Printf("TODO: accumulate errors: %s\n", err.Error())
		},
		Importer: importer.ForCompiler(fset, "source", nil),
		Sizes:    types.SizesFor(build.Default.Compiler, build.Default.GOARCH),
	}

	info := &types.Info{
		Types: make(map[ast.Expr]types.TypeAndValue),
		Defs:  make(map[*ast.Ident]types.Object),
		Uses:  make(map[*ast.Ident]types.Object),
	}

	pkg, err := conf.Check(packagePath, fset, astFiles, info)

	if err != nil {
		panic(err)
	}

	for i := range files {

		goFile, err := AstParseFileEx(mod, files[i], nil, astFiles[i], fset, astFiles, pkg, info)

		if err != nil {
			return nil, err
		}

		goFiles[i] = goFile

	}

	return goFiles, nil

}
