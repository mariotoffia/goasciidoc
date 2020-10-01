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
