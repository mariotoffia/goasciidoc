package goparser

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/mariotoffia/goasciidoc/utils"
)

// ModulePathAndFiles is keyed with module path and `PathAndFiles` for
// that module.
type ModulePathAndFiles map[string]PathAndFiles

// NonModuleName is the key for non module `PathAndFiles`
const NonModuleName = "__no-module"

func FromPathAndFiles(pf PathAndFiles) ModulePathAndFiles {

	mpf := ModulePathAndFiles{}

	mods := pf.ModulePaths()

	for _, mod := range mods {

		mpf[mod] = PathAndFiles{}

	}

	mpf[NonModuleName] = PathAndFiles{}

	for fp, files := range pf {

		pos, ok := utils.GetMostNarrowPath(fp, mods)
		if !ok {

			mpf[NonModuleName][fp] = files
			continue

		}

		mpf[mods[pos]][fp] = files

	}

	if len(mpf[NonModuleName]) == 0 {
		delete(mpf, NonModuleName)
	}

	return mpf
}

// ModulePaths returns a sorted list of module paths.
func (mpf ModulePathAndFiles) ModulePaths() []string {

	var list []string

	for k := range mpf {
		list = append(list, k)
	}

	sort.Strings(list)

	return list
}

// PackagePaths returns a sorted set of package paths for the specified module.
func (mpf ModulePathAndFiles) PackagePaths(module string) []string {

	var list []string

	for k := range mpf[module] {
		list = append(list, k)
	}

	sort.Strings(list)

	return list

}

// PathAndFiles is keyed with path and fully qualified filenames
type PathAndFiles map[string][]string

// FromPaths will iterate directories (recursively) and add explicit files
// in the paths.
//
// It is possible to use relative paths or fully qualified paths along with '.'
// for current directory. The paths are stat:ed so it will check if it is a file
// or directory and do accordingly. If file it will ignore configuration and blindly
// accept the file.
func FromPaths(config PackageParserConfig, paths ...string) (PathAndFiles, error) {
	files := PathAndFiles{}

	appendFile := func(filePath string) {

		key := filepath.Dir(filePath)

		if f, ok := files[key]; ok {

			files[key] = append(f, filePath)

		} else {

			files[key] = []string{filePath}

		}

	}

	for _, p := range paths {

		fileInfo, err := os.Stat(p)
		if err != nil {
			return nil, err
		}

		if !fileInfo.IsDir() {
			appendFile(p)
			continue
		}

		err = filepath.Walk(p,
			func(path string, info os.FileInfo, err error) error {

				if err != nil {
					return err
				}

				if info.IsDir() {
					return nil
				}

				file := info.Name()

				if !strings.HasSuffix(file, ".go") &&
					!strings.HasSuffix(file, "go.mod") {

					return nil

				}

				if strings.HasSuffix(file, "_test.go") {

					if config.Test {
						appendFile(path)
					}

					return nil
				}

				dir := filepath.Dir(path)

				if strings.Contains(dir, "/internal/") {

					if config.Internal {
						appendFile(path)
					}

					return nil
				}

				if strings.Contains(dir, "/_") {

					if config.UnderScore {
						appendFile(path)
					}

					return nil
				}

				appendFile(path)
				return nil

			})

		if err != nil {
			return nil, err
		}

	}

	// Predictable file order
	for k, v := range files {

		sort.Slice(v, func(i, j int) bool {
			return v[i] < v[j]
		})

		files[k] = v

	}

	return files, nil

}

// ModulePaths iterates the _pf_ for all paths that do contain _go.mod_.
func (pf PathAndFiles) ModulePaths() []string {

	var list []string

	for path, files := range pf {

		if _, ok := utils.HasSuffixString(files, "go.mod"); ok {

			list = append(list, path)

		}
	}

	return list
}

// ToModule uses a path to where _go.mod_ resides. It may
// also be a complete filepath to _go.mod_.
func ToModule(fp string) *GoModule {

	if !strings.HasSuffix(fp, "go.mod") {
		fp = filepath.Join(fp, "go.mod")
	}

	m, err := NewModule(fp)
	if err != nil {
		panic(err)
	}

	return m
}
