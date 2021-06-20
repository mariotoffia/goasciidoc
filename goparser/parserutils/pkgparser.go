package parserutils

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"

	"github.com/mariotoffia/goasciidoc/goparser"
)

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

func (pp *PackageParserImpl) parseFiles(
	mod *goparser.GoModule,
	files ...string,
) ([]*goparser.GoFile, error) {

	if len(files) == 0 {
		return nil, fmt.Errorf("must specify at least one file to file to parse")
	}

	astFiles := make([]*ast.File, len(files))
	fsets := make([]*token.FileSet, len(files))

	for i, p := range files {

		// File: A File node represents a Go source file: https://golang.org/pkg/go/ast/#File
		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, p, nil, parser.ParseComments)

		if err != nil {
			return nil, err
		}

		astFiles[i] = file
		fsets[i] = fset

	}

	goFiles := make([]*goparser.GoFile, len(files))

	for i, p := range files {

		goFile, err := goparser.AstParseFile(mod, p, nil, astFiles[i], fsets[i], astFiles)

		if err != nil {
			return nil, err
		}

		goFiles[i] = goFile

	}

	return goFiles, nil

}
