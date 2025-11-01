package goparser

import (
	"fmt"
	"path/filepath"
	"strings"
)

type Resolver interface {
	LoadAll() ([]*GoPackage, error)
}

// ResolverImpl is the implementation of a `Resolver` where it operarates on
// a `GoModule` level.
type ResolverImpl struct {
	// module is the resolvers workspace.
	//
	// It may have sub-modules through contained `Resolver` instances.
	module *GoModule

	// resolvers is a map containing `Resolver` for each `GoModule` that this _module_
	// references and makes use of.
	//resolvers map[string] /*module name*/ Resolver

	// config is the configuration that this `Resolver` adheres to.
	config ParseConfig
	// Fully qualified filepath to this module (where _go.mod_ resides).
	filepath string
}

// NewResolver creates a new `Resolver` from the filepath to the _go.mod_ file
// or directory where _go.mod_ resides.
func NewResolver(config ParseConfig, filepath string) Resolver {

	res := &ResolverImpl{
		config: config,
	}

	return res.resolveModule(filepath)
}

func (r *ResolverImpl) resolveModule(fp string) Resolver {

	if !strings.HasSuffix(fp, "go.mod") {

		var err error
		fp, err = filepath.Abs(filepath.Join(fp, "go.mod"))

		if err != nil {
			panic(err)
		}

	}

	m, err := NewModule(fp)

	if err != nil {
		panic(err)
	}

	r.module = m
	r.filepath = filepath.Dir(fp)

	return r
}

func (r *ResolverImpl) LoadAll() ([]*GoPackage, error) {
	if r.module == nil {
		return nil, fmt.Errorf("resolver has no module configured")
	}

	root := r.filepath
	if root == "" {
		root = r.module.Base
	}

	files, err := GetFilePaths(r.config, root)
	if err != nil {
		return nil, err
	}

	groups := groupFilesByDir(files)
	return collectPackages(r.module, groups)
}
