package goparser

import (
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"golang.org/x/tools/go/packages"
)

type packageLoader struct {
	mu      sync.Mutex
	module  *GoModule
	cache   map[string][]*packages.Package
	loadDur map[string]time.Duration
}

var (
	defaultPackageLoaderMu sync.Mutex
	defaultPackageLoader   *packageLoader
)

func newPackageLoader(mod *GoModule) *packageLoader {
	return &packageLoader{
		module:  mod,
		cache:   make(map[string][]*packages.Package),
		loadDur: make(map[string]time.Duration),
	}
}

func (pl *packageLoader) load(dir string, includeTests bool, debug DebugFunc) ([]*packages.Package, error) {
	if dir == "" {
		return nil, fmt.Errorf("package loader: empty directory")
	}

	ensureLocalModulePreference()

	absDir, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}

	cacheKey := fmt.Sprintf("%s|tests=%t", absDir, includeTests)

	pl.mu.Lock()
	if pkgs, ok := pl.cache[cacheKey]; ok {
		debugf(debug, "packageLoader: reuse %s (%d package(s)) loaded in %s", cacheKey, len(pkgs), pl.loadDur[cacheKey])
		pl.mu.Unlock()
		return pkgs, nil
	}
	pl.mu.Unlock()

	mode := packages.NeedName |
		packages.NeedFiles |
		packages.NeedCompiledGoFiles |
		packages.NeedImports |
		packages.NeedDeps |
		packages.NeedSyntax |
		packages.NeedTypes |
		packages.NeedTypesInfo |
		packages.NeedModule

	cfg := &packages.Config{
		Mode:  mode,
		Dir:   absDir,
		Tests: includeTests,
	}

	start := time.Now()
	pkgs, err := packages.Load(cfg, ".")
	duration := time.Since(start)

	if err != nil {
		return nil, fmt.Errorf("package loader: load %s failed: %w", absDir, err)
	}

	if len(pkgs) == 0 {
		return nil, fmt.Errorf("package loader: load %s returned no packages", absDir)
	}

	pl.mu.Lock()
	pl.cache[cacheKey] = pkgs
	pl.loadDur[cacheKey] = duration
	pl.mu.Unlock()

	debugf(debug, "packageLoader: loaded %s (%d package(s)) tests=%t in %s", absDir, len(pkgs), includeTests, duration)

	return pkgs, nil
}

func (gm *GoModule) getPackageLoader() *packageLoader {
	gm.pkgLoaderMu.Lock()
	defer gm.pkgLoaderMu.Unlock()

	if gm.pkgLoader == nil {
		gm.pkgLoader = newPackageLoader(gm)
	}

	return gm.pkgLoader
}

func getSharedPackageLoader(mod *GoModule) *packageLoader {
	if mod != nil {
		return mod.getPackageLoader()
	}

	defaultPackageLoaderMu.Lock()
	defer defaultPackageLoaderMu.Unlock()

	if defaultPackageLoader == nil {
		defaultPackageLoader = newPackageLoader(nil)
	}

	return defaultPackageLoader
}
