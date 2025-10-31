package goparser

import (
	"fmt"
	"go/ast"
	"io/ioutil"
	"path/filepath"
	"strings"

	"golang.org/x/mod/modfile"
)

type UnresolvedDecl struct {
	Expr    ast.Expr
	Message string
}

// GoModule is a simple representation of a go.mod
type GoModule struct {
	// File is the actual parsed go.mod file
	File *modfile.File
	// FilePath is the filepath to the go module
	FilePath string
	// Base is where all other packages are relative to.
	//
	// This is usually the directory to the File field since
	// go.mod is usually in root project folder.
	Base string
	// Name of the module e.g. github.com/mariotoffia/goasciidoc
	Name string
	// Version of this module
	Version string
	// GoVersion specifies the required go version
	GoVersion string
	// UnresolvedDecl contains all unresolved declarations.
	Unresolved []UnresolvedDecl
}

func (gm *GoModule) AddUnresolvedDeclaration(u UnresolvedDecl) *GoModule {

	gm.Unresolved = append(gm.Unresolved, u)
	return gm

}

// ResolvePackage wil try to resolve the full package path
// bases on this module and the provided path.
//
// If it fails, it returns an empty string.
func (gm *GoModule) ResolvePackage(path string) string {

	if gm == nil || gm.Base == "" {
		return ""
	}

	dir := filepath.Dir(path)
	rel, err := filepath.Rel(gm.Base, dir)
	if err != nil {
		return ""
	}
	if rel == "." || rel == "" {
		return gm.Name // root package
	}

	rel = filepath.ToSlash(rel)
	if strings.HasPrefix(rel, "..") {
		return ""
	}

	return fmt.Sprintf("%s/%s", gm.Name, rel)

}

// NewModule creates a new module from go.mod pointed out in the
// in param path parameter.
func NewModule(path string) (*GoModule, error) {

	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return NewModuleFromBuff(path, data)
}

// NewModuleFromBuff creates a new module from the buff specified in
// the buff parameter and states that the buff is read from path.
func NewModuleFromBuff(path string, buff []byte) (*GoModule, error) {

	file, err := modfile.Parse(path, buff, nil)
	if err != nil {
		return nil, err
	}

	if file.Module == nil {

		return nil, fmt.Errorf(
			"must specify a module that at least have a 'module' statement, path = %s buff = %s",
			path, string(buff),
		)

	}

	goModule := &GoModule{
		File:     file,
		FilePath: path,
		Base:     filepath.Dir(path),
		Name:     file.Module.Mod.Path,
		Version:  file.Module.Mod.Version,
	}

	if file.Go != nil {
		goModule.GoVersion = file.Go.Version
	}

	return goModule, nil
}
