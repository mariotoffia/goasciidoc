package parserutils

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/mariotoffia/goasciidoc/utils"
)

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

				if !strings.HasSuffix(file, ".go") {
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

		if _, ok := utils.ContainsString(files, "go.mod"); ok {

			list = append(list, path)

		}
	}

	return list
}
