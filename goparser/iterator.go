package goparser

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// ParseConfig to use when invoking ParseAny()
type ParseConfig struct {
	// Test denotes if test files (ending with _test.go) should be included or not
	// (default not included)
	Test bool
	// Internal determines if internal folders are included or not (default not)
	Internal bool
	// UnderScore, when set to true it will include directories beginning with _
	UnderScore bool
	// Optional module to resolve fully qualified package paths
	Module *GoModule
}

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

	files, err := GetFilePaths(config, paths...)
	if err != nil {
		return nil, err
	}
	return ParseFiles(config.Module, files...)
}

// ParseSingleFileWalkerFunc is used in conjuction with ParseSingleFileWalker.
//
// If the ParseSingleFileWalker is returning an error, parsing will immediately stop
// and the error is returned.
type ParseSingleFileWalkerFunc func(*GoFile) error

// ParseSingleFileWalker is same as ParseAny, except that it will be fed one GoFile at the
// time and thus consume much less memory.
//
// It uses GetFilePaths and hence, the traversal is in sorted order, directory by directory.
func ParseSingleFileWalker(config ParseConfig, process ParseSingleFileWalkerFunc, paths ...string) error {

	files, err := GetFilePaths(config, paths...)
	if err != nil {
		return err
	}

	for _, f := range files {

		goFile, err := ParseSingleFile(config.Module, f)
		if err != nil {
			return err
		}

		if err := process(goFile); err != nil {
			return err
		}

	}

	return nil
}

// ParseSinglePackageWalkerFunc is used in conjuction with ParseSinglePackageWalker.
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

	files, err := GetFilePaths(config, paths...)
	if err != nil {
		return err
	}

	m := map[string][]string{}
	for _, f := range files {

		dir := filepath.Dir(f)
		if list, ok := m[dir]; ok {
			list = append(list, f)
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

		goFiles, err := ParseFiles(config.Module, v...)
		if err != nil {
			return err
		}

		pkg := &GoPackage{
			Files:   goFiles,
			Package: goFiles[0].Package,
			Path:    k,
			Decl:    goFiles[0].Decl,
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
			if len(gf.VarAssigments) > 0 {
				pkg.VarAssigments = append(pkg.VarAssigments, gf.VarAssigments...)
			}
			if len(gf.ConstAssignments) > 0 {
				pkg.ConstAssignments = append(pkg.ConstAssignments, gf.ConstAssignments...)
			}

		}

		pkg.Doc = b.String()

		if err := process(pkg); err != nil {
			return err
		}
	}

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

	for _, p := range paths {

		fileInfo, err := os.Stat(p)
		if err != nil {
			return nil, err
		}

		if !fileInfo.IsDir() {
			files = append(files, p)
			continue
		}

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
					files = append(files, file)
				}

				return nil
			}

			dir := filepath.Dir(path)

			if strings.Contains(dir, "/Internal/") {

				if config.Internal {
					files = append(files, file)
				}

				return nil
			}

			if strings.Contains(dir, "/_") {

				if config.UnderScore {
					files = append(files, file)
				}

				return nil
			}

			files = append(files, file)
			return nil
		})

		if err != nil {
			return nil, err
		}
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i] < files[j]
	})

	return files, nil
}
