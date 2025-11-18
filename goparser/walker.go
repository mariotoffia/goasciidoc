package goparser

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/mariotoffia/goasciidoc/goparser/utils"
)

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

	matcher, err := utils.NewRegexMatcher(config.Excludes)
	if err != nil {
		return nil, err
	}

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

		root := filepath.Clean(p)
		err = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {

			if err != nil {
				return err
			}

			normalizedPath := filepath.ToSlash(path)
			relPath := normalizedPath
			if rel, relErr := filepath.Rel(root, path); relErr == nil {
				relPath = filepath.ToSlash(rel)
			}

			if match, pattern := matcher.Match(normalizedPath, relPath); match {
				if info.IsDir() {
					debugf(
						config.Debug,
						"GetFilePaths: skipped directory %s (excluded by %s)",
						path,
						pattern,
					)
					return filepath.SkipDir
				}
				debugf(
					config.Debug,
					"GetFilePaths: skipped file %s (excluded by %s)",
					path,
					pattern,
				)
				return nil
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
