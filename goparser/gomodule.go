package goparser

import (
	"errors"
	"fmt"
	"go/ast"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"golang.org/x/mod/modfile"
)

type UnresolvedDecl struct {
	Expr    ast.Expr
	Message string
}

var (
	ErrModuleNotConfigured  = errors.New("module not configured")
	ErrPackageOutsideModule = errors.New("path does not belong to module")
	ErrModuleNotFound       = errors.New("no go.mod found in parent directories")
)

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

	importerMu sync.Mutex
	importer   *moduleImporter
}

func (gm *GoModule) AddUnresolvedDeclaration(u UnresolvedDecl) *GoModule {

	gm.Unresolved = append(gm.Unresolved, u)
	return gm

}

func (gm *GoModule) getModuleImporter(debug DebugFunc) *moduleImporter {
	gm.importerMu.Lock()
	defer gm.importerMu.Unlock()

	if gm.importer == nil {
		gm.importer = newModuleImporter(gm, debug)
	} else {
		gm.importer.setDebug(debug)
	}

	return gm.importer
}

// ResolvePackage tries to resolve the full package import path for the provided file path.
// When the file resides outside of the module or the module lacks sufficient information,
// an error is returned describing the problem.
func (gm *GoModule) ResolvePackage(path string) (string, error) {

	if gm == nil {
		return "", ErrModuleNotConfigured
	}

	if gm.Base == "" || gm.Name == "" {
		return "", fmt.Errorf("module %q is missing base directory or module name", gm.FilePath)
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("resolve package for %q: %w", path, err)
	}

	dir := filepath.Dir(absPath)
	rel, err := filepath.Rel(gm.Base, dir)
	if err != nil {
		return "", fmt.Errorf("resolve package for %q: %w", path, err)
	}

	rel = filepath.ToSlash(rel)
	if rel == "." || rel == "" {
		return gm.Name, nil
	}

	if strings.HasPrefix(rel, "..") {
		return "", fmt.Errorf("%w: %s", ErrPackageOutsideModule, absPath)
	}

	return fmt.Sprintf("%s/%s", strings.TrimSuffix(gm.Name, "/"), strings.Trim(rel, "/")), nil

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

// FindModule walks parent directories of the provided path until it locates a go.mod file.
// It returns ErrModuleNotFound when no module file is present.
func FindModule(path string) (*GoModule, error) {

	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	start := absPath
	info, err := os.Stat(absPath)
	if err == nil && !info.IsDir() {
		start = filepath.Dir(absPath)
	}

	dir := start
	for {
		candidate := filepath.Join(dir, "go.mod")
		if _, err := os.Stat(candidate); err == nil {
			return NewModule(candidate)
		} else if !os.IsNotExist(err) {
			return nil, err
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return nil, fmt.Errorf("%w for %s", ErrModuleNotFound, absPath)
}
