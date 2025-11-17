package goparser

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func getPwd() string {
	pwd, _ := os.Getwd()
	return pwd
}

func TestModuleBasePathIsTakenFromPathParam(t *testing.T) {
	data := `module github.com/mariotoffia/goasciidoc`
	path := getPwd() + "go.mod"

	m, err := NewModuleFromBuff(path, []byte(data))
	assert.NoError(t, err)
	assert.Equal(t, path, m.FilePath)
	assert.Equal(t, filepath.Dir(path), m.Base)
}

func TestParseWithOnlyModuleLine(t *testing.T) {
	data := `module github.com/mariotoffia/goasciidoc`
	path := getPwd() + "go.mod"

	m, err := NewModuleFromBuff(path, []byte(data))
	assert.NoError(t, err)
	assert.Equal(t, "github.com/mariotoffia/goasciidoc", m.Name)
}

func TestParseWithNoModuleLineMustFail(t *testing.T) {
	data := `go 1.14`
	path := getPwd() + "go.mod"

	_, err := NewModuleFromBuff(path, []byte(data))
	assert.Error(t, err)
}

func TestParseModuleNameGoVersionAndRequires(t *testing.T) {
	data := `module github.com/mariotoffia/goasciidoc

	require (
		golang.org/x/mod v0.3.0
		github.com/stretchr/testify v1.6.1
	)
	
	go 1.14`
	path := getPwd() + "go.mod"

	m, err := NewModuleFromBuff(path, []byte(data))
	assert.NoError(t, err)
	assert.Equal(t, "github.com/mariotoffia/goasciidoc", m.Name)
	assert.Equal(t, "1.14", m.GoVersion)
	// Require are skipped
}

func TestFindAllModules(t *testing.T) {
	workspaceRoot, err := filepath.Abs("../.temp-files/tests/workspace-test")
	assert.NoError(t, err)

	// Skip if fixtures don't exist
	if _, err := os.Stat(workspaceRoot); os.IsNotExist(err) {
		t.Skip("Test fixtures not found, skipping FindAllModules tests")
	}

	t.Run("find all modules in workspace", func(t *testing.T) {
		modules, err := FindAllModules(workspaceRoot)
		assert.NoError(t, err)
		assert.NotNil(t, modules)
		assert.Equal(t, 3, len(modules), "Should find exactly 3 modules")

		// Verify all expected modules are found
		moduleNames := make(map[string]bool)
		for _, mod := range modules {
			moduleNames[mod.Name] = true
		}

		assert.True(t, moduleNames["github.com/test/module1"], "Should find module1")
		assert.True(t, moduleNames["github.com/test/module2"], "Should find module2")
		assert.True(t, moduleNames["github.com/test/module3"], "Should find module3")
	})

	multiModRoot, err := filepath.Abs("../.temp-files/tests/multi-module-no-workspace")
	assert.NoError(t, err)

	if _, err := os.Stat(multiModRoot); err == nil {
		t.Run("find modules without workspace", func(t *testing.T) {
			modules, err := FindAllModules(multiModRoot)
			assert.NoError(t, err)
			assert.NotNil(t, modules)
			assert.Equal(t, 2, len(modules), "Should find exactly 2 modules")

			moduleNames := make(map[string]bool)
			for _, mod := range modules {
				moduleNames[mod.Name] = true
			}

			assert.True(t, moduleNames["github.com/test/moda"], "Should find moda")
			assert.True(t, moduleNames["github.com/test/modb"], "Should find modb")
		})
	}

	singleModRoot, err := filepath.Abs("../.temp-files/tests/single-module-test")
	assert.NoError(t, err)

	if _, err := os.Stat(singleModRoot); err == nil {
		t.Run("find single module", func(t *testing.T) {
			modules, err := FindAllModules(singleModRoot)
			assert.NoError(t, err)
			assert.NotNil(t, modules)
			assert.Equal(t, 1, len(modules), "Should find exactly 1 module")
			assert.Equal(t, "github.com/test/singlemodule", modules[0].Name)
		})
	}

	t.Run("empty directory returns empty slice", func(t *testing.T) {
		tempDir := t.TempDir()
		modules, err := FindAllModules(tempDir)
		assert.NoError(t, err)
		assert.Equal(t, 0, len(modules), "Should find no modules in empty directory")
	})
}

func TestLoadModule(t *testing.T) {
	singleModRoot, err := filepath.Abs("../.temp-files/tests/single-module-test")
	assert.NoError(t, err)

	modPath := filepath.Join(singleModRoot, "go.mod")

	// Skip if fixtures don't exist
	if _, err := os.Stat(modPath); os.IsNotExist(err) {
		t.Skip("Test fixtures not found, skipping LoadModule test")
	}

	module, err := LoadModule(modPath)
	assert.NoError(t, err)
	assert.NotNil(t, module)
	assert.Equal(t, "github.com/test/singlemodule", module.Name)
	assert.Equal(t, modPath, module.FilePath)
	assert.Equal(t, singleModRoot, module.Base)
}
