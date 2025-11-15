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

	"golang.org/x/tools/go/packages"
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
	buildTagsSet := make(map[string]struct{})
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
		// Collect unique build tags from all files
		for _, tag := range gf.BuildTags {
			buildTagsSet[tag] = struct{}{}
		}
	}

	// Convert build tags set to slice
	if len(buildTagsSet) > 0 {
		pkg.BuildTags = make([]string, 0, len(buildTagsSet))
		for tag := range buildTagsSet {
			pkg.BuildTags = append(pkg.BuildTags, tag)
		}
		sort.Strings(pkg.BuildTags)
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

func collectPackages(
	config ParseConfig,
	groups map[string][]string,
) ([]*GoPackage, error) {
	module := config.Module
	debug := config.Debug
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

		debugf(debug, "collectPackages: parsing directory %s with %d file(s)", dir, len(files))

		sort.Strings(files)
		goFiles, err := parseFiles(config, files...)
		if err != nil {
			return nil, err
		}
		debugf(debug, "collectPackages: parsed directory %s", dir)

		pkg := aggregatePackage(module, dir, goFiles)
		if pkg != nil {
			packages = append(packages, pkg)
			debugf(
				debug,
				"collectPackages: aggregated package %s (%d file(s))",
				pkg.Package,
				len(pkg.Files),
			)
		}
	}

	return packages, nil
}

// ParseSingleFile parses a single file at the same time
//
// If a module is passed, it will calculate package relative to that
func ParseSingleFile(mod *GoModule, path string) (*GoFile, error) {

	return parseSingleFileWithConfig(ParseConfig{Module: mod}, path)
}

func parseSingleFileWithConfig(config ParseConfig, path string) (*GoFile, error) {

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, nil, parser.ParseComments)

	if err != nil {
		return nil, err
	}

	files := []*ast.File{file}
	info, typeErr := typeCheckPackage(config.Module, fset, files, nil)
	recordTypeCheckError(config.Module, path, typeErr)

	prev := activeDocConcatenation
	activeDocConcatenation = config.DocConcatenation
	defer func() { activeDocConcatenation = prev }()

	return parseFile(config.Module, path, nil, file, fset, info)

}

// ParseFiles parses one or more files
func ParseFiles(mod *GoModule, paths ...string) ([]*GoFile, error) {
	return parseFiles(ParseConfig{Module: mod}, paths...)
}

func parseFiles(config ParseConfig, paths ...string) ([]*GoFile, error) {
	if len(paths) == 0 {
		return nil, fmt.Errorf("must specify at least one path to file to parse")
	}

	prev := activeDocConcatenation
	activeDocConcatenation = config.DocConcatenation
	defer func() { activeDocConcatenation = prev }()

	if config.Module == nil {
		return parseFilesLegacy(nil, config.Debug, paths...)
	}

	goFiles, err := parseFilesWithPackages(config, paths...)
	if err != nil {
		if shouldFallbackToLegacy(err) {
			debugf(config.Debug, "ParseFiles: falling back to legacy parser due to: %v", err)
			return parseFilesLegacy(config.Module, config.Debug, paths...)
		}
		return nil, err
	}

	return goFiles, nil
}

func shouldFallbackToLegacy(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	if msg == "" {
		return false
	}
	return strings.Contains(msg, "go.mod file not found") || strings.Contains(msg, "no packages")
}

func parseFilesWithPackages(config ParseConfig, paths ...string) ([]*GoFile, error) {
	if len(paths) == 0 {
		return nil, fmt.Errorf("must specify at least one path to file to parse")
	}

	mod := config.Module
	debug := config.Debug
	loader := getSharedPackageLoader(mod)

	type fileContext struct {
		pkg  *packages.Package
		file *ast.File
	}

	contexts := make([]fileContext, len(paths))
	absPaths := make([]string, len(paths))
	dirIndexes := make(map[string][]int)

	for i, p := range paths {
		debugf(debug, "ParseFiles: preparing %s", p)

		absPath, err := filepath.Abs(p)
		if err != nil {
			return nil, err
		}
		cleanAbs := filepath.Clean(absPath)
		absPaths[i] = cleanAbs

		dir := filepath.Dir(cleanAbs)
		dirIndexes[dir] = append(dirIndexes[dir], i)
	}

	for dir, indexes := range dirIndexes {
		includeTests := false
		for _, idx := range indexes {
			if strings.HasSuffix(paths[idx], "_test.go") {
				includeTests = true
				break
			}
		}

		debugf(debug, "ParseFiles: loading packages for %s (tests=%t)", dir, includeTests)

		packages, err := loader.load(dir, includeTests, config.BuildTags, config.AllBuildTags, debug)
		if err != nil {
			return nil, err
		}

		fileMap := make(map[string]fileContext)

		for _, pkg := range packages {
			if pkg == nil {
				continue
			}

			for _, pkgErr := range pkg.Errors {
				if debug != nil {
					debugf(debug, "packageLoader: %s error: %v", pkg.PkgPath, pkgErr)
				}
				recordTypeCheckError(mod, pkg.PkgPath, pkgErr)
			}

			for _, syntax := range pkg.Syntax {
				if syntax == nil {
					continue
				}
				pos := pkg.Fset.PositionFor(syntax.Pos(), false)
				filename := filepath.Clean(pos.Filename)
				if filename == "" {
					continue
				}

				if _, exists := fileMap[filename]; !exists {
					fileMap[filename] = fileContext{
						pkg:  pkg,
						file: syntax,
					}
				}
			}
		}

		for _, idx := range indexes {
			abs := absPaths[idx]
			ctx, ok := fileMap[abs]
			if !ok {
				// File not loaded - likely excluded by build constraints
				debugf(debug, "ParseFiles: skipping %s (excluded by build constraints)", abs)
				continue
			}
			contexts[idx] = ctx
		}
	}

	// Build result excluding skipped files
	goFiles := make([]*GoFile, 0, len(paths))
	for i, path := range paths {
		ctx := contexts[i]
		if ctx.pkg == nil || ctx.file == nil {
			// This file was skipped (build constraint mismatch)
			debugf(debug, "ParseFiles: file %s was excluded by build constraints", path)
			continue
		}

		info := ctx.pkg.TypesInfo
		if info == nil {
			info = &types.Info{
				Types: make(map[ast.Expr]types.TypeAndValue),
				Defs:  make(map[*ast.Ident]types.Object),
				Uses:  make(map[*ast.Ident]types.Object),
			}
		}

		goFile, err := parseFile(mod, path, nil, ctx.file, ctx.pkg.Fset, info)
		if err != nil {
			return nil, err
		}
		debugf(debug, "ParseFiles: built GoFile for %s", path)
		goFiles = append(goFiles, goFile)
	}

	debugf(debug, "ParseFiles: completed %d file(s) (%d skipped by build constraints)", len(goFiles), len(paths)-len(goFiles))
	return goFiles, nil
}

func parseFilesLegacy(mod *GoModule, debug DebugFunc, paths ...string) ([]*GoFile, error) {
	type fileContext struct {
		bucketKey string
		file      *ast.File
	}

	type packageBucket struct {
		fset    *token.FileSet
		files   []*ast.File
		info    *types.Info
		pkgName string
		pkgDir  string
	}

	buckets := make(map[string]*packageBucket)
	fileContexts := make([]fileContext, len(paths))

	for i, p := range paths {
		debugf(debug, "ParseFiles[legacy]: parsing %s", p)

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
				pkgName: file.Name.Name,
				pkgDir:  dir,
			}
			buckets[key] = bucket
			fileContexts[i] = fileContext{bucketKey: key, file: file}
			continue
		}

		debugf(debug, "ParseFiles[legacy]: reusing fileset for %s", key)

		parsedFile, err := parser.ParseFile(bucket.fset, p, nil, parser.ParseComments)
		if err != nil {
			return nil, err
		}

		bucket.files = append(bucket.files, parsedFile)
		fileContexts[i] = fileContext{bucketKey: key, file: parsedFile}
	}

	for key, bucket := range buckets {
		debugf(debug, "ParseFiles[legacy]: type-checking %s (%d file(s))", key, len(bucket.files))

		info, typeErr := typeCheckPackage(mod, bucket.fset, bucket.files, debug)
		bucket.info = info
		debugf(debug, "ParseFiles[legacy]: type-check completed for %s", key)

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

		debugf(debug, "ParseFiles[legacy]: building GoFile for %s", p)

		goFile, err := parseFile(mod, p, nil, ctx.file, bucket.fset, bucket.info)
		if err != nil {
			return nil, err
		}
		debugf(debug, "ParseFiles[legacy]: built GoFile for %s", p)

		goFiles[i] = goFile
	}

	debugf(debug, "ParseFiles[legacy]: completed %d file(s)", len(paths))
	return goFiles, nil
}

// ParseInlineFile will parse the code provided.
//
// To simulate package names set the path to some level
// equal to or greater than GoModule.Base. Otherwise just
// set path "" to ignore.
func ParseInlineFile(mod *GoModule, path, code string) (*GoFile, error) {
	return ParseInlineFileWithConfig(ParseConfig{Module: mod}, path, code)
}

func ParseInlineFileWithConfig(config ParseConfig, path, code string) (*GoFile, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "", code, parser.ParseComments)
	if err != nil {
		return nil, err
	}
	files := []*ast.File{file}
	info, typeErr := typeCheckPackage(config.Module, fset, files, nil)
	recordTypeCheckError(config.Module, path, typeErr)

	prev := activeDocConcatenation
	activeDocConcatenation = config.DocConcatenation
	defer func() { activeDocConcatenation = prev }()

	return parseFile(config.Module, path, []byte(code), file, fset, info)
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
type DocConcatenationMode int

const (
	DocConcatenationNone DocConcatenationMode = iota
	DocConcatenationFull
)

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
	// DocConcatenation controls how doc comments split by blank lines are handled.
	DocConcatenation DocConcatenationMode
	// BuildTags specifies build tags to use when loading packages.
	// Each string represents a set of comma-separated tags (e.g., "linux,amd64").
	// If empty, default build constraints apply.
	BuildTags []string
	// AllBuildTags when set to true, attempts to discover and load all build tags.
	AllBuildTags bool
	// IgnoreMarkdownHeadings when set to true, replaces markdown headings (#, ##, etc.) in comments with their text content
	IgnoreMarkdownHeadings bool
}

// end::parse-config[]

var activeDocConcatenation = DocConcatenationNone

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
	prev := activeDocConcatenation
	activeDocConcatenation = config.DocConcatenation
	defer func() { activeDocConcatenation = prev }()
	return parseFiles(config, files...)
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
func ParseSingleFileWalker(
	config ParseConfig,
	process ParseSingleFileWalkerFunc,
	paths ...string,
) error {

	debugf(config.Debug, "ParseSingleFileWalker: resolving files from %d path(s)", len(paths))

	files, err := GetFilePaths(config, paths...)
	if err != nil {
		return err
	}

	debugf(config.Debug, "ParseSingleFileWalker: walking %d file(s)", len(files))

	prev := activeDocConcatenation
	activeDocConcatenation = config.DocConcatenation
	defer func() { activeDocConcatenation = prev }()

	for _, f := range files {

		goFile, err := parseSingleFileWithConfig(config, f)
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
func ParseSinglePackageWalker(
	config ParseConfig,
	process ParseSinglePackageWalkerFunc,
	paths ...string,
) error {

	debugf(config.Debug, "ParseSinglePackageWalker: starting with %d path(s)", len(paths))

	files, err := GetFilePaths(config, paths...)
	if err != nil {
		return err
	}

	debugf(config.Debug, "ParseSinglePackageWalker: collected %d file(s)", len(files))

	groups := groupFilesByDir(files)
	debugf(config.Debug, "ParseSinglePackageWalker: grouped into %d director(ies)", len(groups))

	packages, err := collectPackages(config, groups)
	if err != nil {
		return err
	}

	debugf(config.Debug, "ParseSinglePackageWalker: built %d package(s)", len(packages))

	for _, pkg := range packages {
		debugf(
			config.Debug,
			"ParseSinglePackageWalker: processing package %s (%d file(s))",
			pkg.Package,
			len(pkg.Files),
		)
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
