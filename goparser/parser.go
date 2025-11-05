package goparser

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type DebugFunc func(format string, args ...interface{})

func debugf(fn DebugFunc, format string, args ...interface{}) {
	if fn == nil {
		return
	}

	fn("goparser: "+format, args...)
}

func recordTypeCheckError(mod *GoModule, context string, err error) {
	if err == nil || mod == nil {
		return
	}

	message := fmt.Sprintf("type-check error for %s: %v", context, err)
	mod.AddUnresolvedDeclaration(UnresolvedDecl{
		Message: message,
	})
}

func aggregatePackage(module *GoModule, dir string, goFiles []*GoFile) *GoPackage {
	if len(goFiles) == 0 {
		return nil
	}

	pkgModule := module
	if pkgModule == nil {
		pkgModule = goFiles[0].Module
	}

	pkg := &GoPackage{
		GoFile: GoFile{
			Module:    pkgModule,
			Package:   goFiles[0].Package,
			FqPackage: goFiles[0].FqPackage,
			FilePath:  dir,
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

	pkg.Doc = strings.TrimSuffix(b.String(), "\n")
	return pkg
}

func groupFilesByDir(paths []string) map[string][]string {
	result := make(map[string][]string)
	for _, p := range paths {
		if !strings.HasSuffix(p, ".go") {
			continue
		}
		dir := filepath.Dir(p)
		result[dir] = append(result[dir], p)
	}
	return result
}

func collectPackages(module *GoModule, groups map[string][]string) ([]*GoPackage, error) {
	if len(groups) == 0 {
		return nil, nil
	}

	keys := make([]string, 0, len(groups))
	for dir := range groups {
		keys = append(keys, dir)
	}
	sort.Strings(keys)

	packages := make([]*GoPackage, 0, len(keys))
	for _, dir := range keys {
		files := groups[dir]
		if len(files) == 0 {
			continue
		}

		sort.Strings(files)
		goFiles, err := ParseFiles(module, files...)
		if err != nil {
			return nil, err
		}

		pkg := aggregatePackage(module, dir, goFiles)
		if pkg != nil {
			packages = append(packages, pkg)
		}
	}

	return packages, nil
}

// ParseSingleFile parses a single file at the same time
//
// If a module is passed, it will calculate package relative to that
func ParseSingleFile(mod *GoModule, path string) (*GoFile, error) {

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, nil, parser.ParseComments)

	if err != nil {
		return nil, err
	}

	files := []*ast.File{file}
	info, typeErr := typeCheckPackage(mod, fset, files)
	recordTypeCheckError(mod, path, typeErr)

	return parseFile(mod, path, nil, file, fset, files, info)

}

// ParseFiles parses one or more files
func ParseFiles(mod *GoModule, paths ...string) ([]*GoFile, error) {

	if len(paths) == 0 {
		return nil, fmt.Errorf("must specify at least one path to file to parse")
	}

	type fileContext struct {
		bucketKey string
		file      *ast.File
	}

	type packageBucket struct {
		fset    *token.FileSet
		files   []*ast.File
		paths   []string
		info    *types.Info
		err     error
		pkgName string
		pkgDir  string
	}

	buckets := make(map[string]*packageBucket)
	fileContexts := make([]fileContext, len(paths))

	for i, p := range paths {
		initialFset := token.NewFileSet()
		file, err := parser.ParseFile(initialFset, p, nil, parser.ParseComments)
		if err != nil {
			return nil, err
		}

		absPath, err := filepath.Abs(p)
		if err != nil {
			return nil, err
		}
		dir := filepath.Dir(absPath)
		key := fmt.Sprintf("%s:%s", dir, file.Name.Name)

		bucket, ok := buckets[key]
		if !ok {
			bucket = &packageBucket{
				fset:    initialFset,
				files:   []*ast.File{file},
				paths:   []string{p},
				pkgName: file.Name.Name,
				pkgDir:  dir,
			}
			buckets[key] = bucket
			fileContexts[i] = fileContext{bucketKey: key, file: file}
			continue
		}

		parsedFile, err := parser.ParseFile(bucket.fset, p, nil, parser.ParseComments)
		if err != nil {
			return nil, err
		}

		bucket.files = append(bucket.files, parsedFile)
		bucket.paths = append(bucket.paths, p)
		fileContexts[i] = fileContext{bucketKey: key, file: parsedFile}
	}

	for key, bucket := range buckets {
		info, typeErr := typeCheckPackage(mod, bucket.fset, bucket.files)
		bucket.info = info
		bucket.err = typeErr
		if typeErr != nil {
			recordTypeCheckError(mod, key, typeErr)
		}
	}

	goFiles := make([]*GoFile, len(paths))
	for i, p := range paths {
		ctx := fileContexts[i]
		bucket := buckets[ctx.bucketKey]
		if bucket == nil {
			return nil, fmt.Errorf("internal error: missing package bucket for %s", p)
		}

		goFile, err := parseFile(mod, p, nil, ctx.file, bucket.fset, bucket.files, bucket.info)
		if err != nil {
			return nil, err
		}
		goFiles[i] = goFile
	}
	return goFiles, nil
}

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
	files := []*ast.File{file}
	info, typeErr := typeCheckPackage(mod, fset, files)
	recordTypeCheckError(mod, path, typeErr)

	return parseFile(mod, path, []byte(code), file, fset, files, info)
}

// ParseConfig to use when invoking ParseAny, ParseSingleFileWalker, and
// ParseSinglePackageWalker.
//
// .ParserConfig
// [source,go]
// ----
// include::${gad:current:fq}[tag=parse-config,indent=0]
// ----
// <1> These are usually excluded since many testcases is not documented anyhow
// <2> As of _go 1.16_ it is recommended to *only* use module based parsing
// tag::parse-config[]
type ParseConfig struct {
	// Test denotes if test files (ending with _test.go) should be included or not
	// (default not included)
	Test bool // <1>
	// Internal determines if internal folders are included or not (default not)
	Internal bool
	// UnderScore, when set to true it will include directories beginning with _
	UnderScore bool
	// Optional module to resolve fully qualified package paths
	Module *GoModule // <2>
	// Debug collects debug statements during traversal.
	Debug DebugFunc
}

// end::parse-config[]

// ParseAny parses one or more directories (recursively) for go files. It is also possible
// to add files along with directories (or just files).
//
// It is possible to use relative paths or fully qualified paths along with '.'
// for current directory. The paths are stat:ed so it will check if it is a file
// or directory and do accordingly. If file it will ignore configuration and blindly
// accept the file.
//
// The example below parses from current directory down recursively and skips
// test, internal and underscore directories.
// Example: ParseAny(ParseConfig{}, ".")
//
// Next example will recursively add go files from src and one single test.go under
// directory dummy (both relative current directory).
// Example: ParseAny(ParseConfig{}, "./src", "./dummy/test.go")
func ParseAny(config ParseConfig, paths ...string) ([]*GoFile, error) {

	debugf(config.Debug, "ParseAny: resolving files from %d input path(s)", len(paths))

	files, err := GetFilePaths(config, paths...)
	if err != nil {
		return nil, err
	}
	debugf(config.Debug, "ParseAny: parsing %d collected file(s)", len(files))
	return ParseFiles(config.Module, files...)
}

// ParseSingleFileWalkerFunc is used in conjunction with ParseSingleFileWalker.
//
// If the ParseSingleFileWalker is returning an error, parsing will immediately stop
// and the error is returned.
type ParseSingleFileWalkerFunc func(*GoFile) error

// ParseSingleFileWalker is same as ParseAny, except that it will be fed one GoFile at the
// time and thus consume much less memory.
//
// It uses GetFilePaths and hence, the traversal is in sorted order, directory by directory.
func ParseSingleFileWalker(config ParseConfig, process ParseSingleFileWalkerFunc, paths ...string) error {

	debugf(config.Debug, "ParseSingleFileWalker: resolving files from %d path(s)", len(paths))

	files, err := GetFilePaths(config, paths...)
	if err != nil {
		return err
	}

	debugf(config.Debug, "ParseSingleFileWalker: walking %d file(s)", len(files))

	for _, f := range files {

		goFile, err := ParseSingleFile(config.Module, f)
		if err != nil {
			return err
		}

		debugf(config.Debug, "ParseSingleFileWalker: processing %s", f)

		if err := process(goFile); err != nil {
			return err
		}

	}

	return nil
}

// ParseSinglePackageWalkerFunc is used in conjunction with ParseSinglePackageWalker.
//
// If the ParseSinglePackageWalker is returning an error, parsing will immediately stop
// and the error is returned.
type ParseSinglePackageWalkerFunc func(*GoPackage) error

// ParseSinglePackageWalker is same as ParseAny, except that it will be fed one GoPackage at the
// time and thus consume much less memory.
//
// It uses GetFilePaths and hence, the traversal is in sorted order, directory by directory. It will
// bundle all files in same directory and assign those to a GoPackage before invoking ParseSinglePackageWalkerFunc
func ParseSinglePackageWalker(config ParseConfig, process ParseSinglePackageWalkerFunc, paths ...string) error {

	debugf(config.Debug, "ParseSinglePackageWalker: starting with %d path(s)", len(paths))

	files, err := GetFilePaths(config, paths...)
	if err != nil {
		return err
	}

	debugf(config.Debug, "ParseSinglePackageWalker: collected %d file(s)", len(files))

	groups := groupFilesByDir(files)
	debugf(config.Debug, "ParseSinglePackageWalker: grouped into %d director(ies)", len(groups))

	packages, err := collectPackages(config.Module, groups)
	if err != nil {
		return err
	}

	debugf(config.Debug, "ParseSinglePackageWalker: built %d package(s)", len(packages))

	for _, pkg := range packages {
		debugf(config.Debug, "ParseSinglePackageWalker: processing package %s (%d file(s))", pkg.Package, len(pkg.Files))
		if err := process(pkg); err != nil {
			return err
		}
	}

	debugf(config.Debug, "ParseSinglePackageWalker: completed processing")

	return nil
}

// GetFilePaths will iterate directories (recursively) and add explicit files
// in the paths.
//
// It is possible to use relative paths or fully qualified paths along with '.'
// for current directory. The paths are stat:ed so it will check if it is a file
// or directory and do accordingly. If file it will ignore configuration and blindly
// accept the file.
func GetFilePaths(config ParseConfig, paths ...string) ([]string, error) {
	files := []string{}

	debugf(config.Debug, "GetFilePaths: walking %d root path(s)", len(paths))

	for _, p := range paths {

		debugf(config.Debug, "GetFilePaths: scanning %s", p)

		fileInfo, err := os.Stat(p)
		if err != nil {
			return nil, err
		}

		if !fileInfo.IsDir() {
			files = append(files, p)
			debugf(config.Debug, "GetFilePaths: added file %s", p)
			continue
		}

		debugf(config.Debug, "GetFilePaths: walking directory %s", p)
		before := len(files)

		err = filepath.Walk(p, func(path string, info os.FileInfo, err error) error {

			if err != nil {
				return err
			}

			if info.IsDir() {
				return nil
			}

			file := info.Name()

			if !strings.HasSuffix(file, ".go") {
				return nil
			}

			if strings.HasSuffix(file, "_test.go") {

				if config.Test {
					files = append(files, path)
				} else {
					debugf(config.Debug, "GetFilePaths: skipped test file %s", path)
				}

				return nil
			}

			dir := filepath.Dir(path)
			relDir := dir
			if rel, err := filepath.Rel(p, dir); err == nil {
				relDir = rel
			}
			dirSegments := strings.Split(filepath.ToSlash(relDir), "/")

			hasInternal := false
			hasUnderscore := false

			for _, segment := range dirSegments {
				if segment == "" || segment == "." {
					continue
				}
				if segment == ".." {
					continue
				}
				if segment == "internal" {
					hasInternal = true
				}
				if strings.HasPrefix(segment, "_") {
					hasUnderscore = true
				}
			}

			if hasInternal && !config.Internal {
				debugf(config.Debug, "GetFilePaths: skipped %s (internal directory)", path)
				return nil
			}

			if hasUnderscore && !config.UnderScore {
				debugf(config.Debug, "GetFilePaths: skipped %s (underscored directory)", path)
				return nil
			}

			files = append(files, path)
			return nil
		})

		if err != nil {
			return nil, err
		}

		debugf(config.Debug, "GetFilePaths: directory %s yielded %d file(s)", p, len(files)-before)
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i] < files[j]
	})

	debugf(config.Debug, "GetFilePaths: collected %d file(s) in total", len(files))

	return files, nil
}
