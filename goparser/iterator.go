package goparser

import (
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
	return ParseFiles(files...)
}

// ParseWalkerFunc is used in conjuction with ParseAnyWalk.
//
// If the ParseWalker is returning an error, parsing will immediately stop
// and the error is returned.
type ParseWalkerFunc func(*GoFile) error

// ParseAnyWalker is same as ParseAny, except that it will be fed one GoFile at the
// time and thus consume much less memory.
//
// It uses GetFilePaths and hence, the traversal is in sorted order, directory by directory.
func ParseAnyWalker(config ParseConfig, process ParseWalkerFunc, paths ...string) error {

	files, err := GetFilePaths(config, paths...)
	if err != nil {
		return err
	}

	for _, f := range files {

		goFile, err := ParseSingleFile(f)
		if err != nil {
			return err
		}

		if err := process(goFile); err != nil {
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
