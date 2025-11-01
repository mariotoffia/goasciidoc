package goparser

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPackageAggregationSharedBetweenWalkerAndResolver(t *testing.T) {
	root := t.TempDir()

	writeFile := func(relPath, contents string) {
		path := filepath.Join(root, relPath)
		require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
		require.NoError(t, os.WriteFile(path, []byte(contents), 0o644))
	}

	writeFile("go.mod", "module example.com/test\n\ngo 1.24\n")
	writeFile("alpha/a.go", `// Package alpha doc.
package alpha

// Foo is declared so the package contains declarations.
type Foo struct{}
`)
	writeFile("alpha/b.go", `package alpha

type Bar struct{}
`)
	writeFile("beta/b.go", `// Package beta doc.
package beta

type Baz struct{}
`)

	mod, err := NewModule(filepath.Join(root, "go.mod"))
	require.NoError(t, err)

	cfg := ParseConfig{Module: mod}
	var walkerPkgs []*GoPackage
	err = ParseSinglePackageWalker(cfg, func(pkg *GoPackage) error {
		walkerPkgs = append(walkerPkgs, pkg)
		return nil
	}, root)
	require.NoError(t, err)

	resolver := NewResolver(ParseConfig{}, root)
	resolverPkgs, err := resolver.LoadAll()
	require.NoError(t, err)

	require.Len(t, walkerPkgs, 2)
	require.Len(t, resolverPkgs, 2)

	alpha := findPackageByName(t, walkerPkgs, "alpha")
	require.Len(t, alpha.Files, 2)
	assert.Equal(t, "Package alpha doc.", alpha.Doc)
	assert.Len(t, alpha.Structs, 2)

	beta := findPackageByName(t, resolverPkgs, "beta")
	require.Len(t, beta.Files, 1)
	assert.Equal(t, "Package beta doc.", beta.Doc)
	assert.Len(t, beta.Structs, 1)

	walkerNames := []string{walkerPkgs[0].Package, walkerPkgs[1].Package}
	resolverNames := []string{resolverPkgs[0].Package, resolverPkgs[1].Package}
	assert.ElementsMatch(t, walkerNames, resolverNames)
}

func findPackageByName(t *testing.T, pkgs []*GoPackage, name string) *GoPackage {
	t.Helper()
	for _, pkg := range pkgs {
		if pkg.Package == name {
			return pkg
		}
	}
	t.Fatalf("package %q not found", name)
	return nil
}
