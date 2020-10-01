package goparser

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"golang.org/x/mod/modfile"
)

// https://github.com/golang/mod

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
}

// ResolvePackage wil try to resolve the full package path
// bases on this module and the provided path.
//
// If it fails, it returns an empty string.
func (gm *GoModule) ResolvePackage(path string) string {
	if len(path) < len(gm.Base) {
		return ""
	}

	return path[:len(gm.Base)]
}

// NewModule creates a new module from go.mod pointed out in the
// inparam path parameter.
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
			"Must specify a module that atleast have a 'module' statement, path = %s buff = %s",
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
