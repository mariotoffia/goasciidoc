package goparser

import (
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"golang.org/x/tools/go/packages"
)

type packageLoader struct {
	mu            sync.Mutex
	module        *GoModule
	preloaded     bool
	moduleDir     string
	allPackages   []*packages.Package
	packagesByDir map[string][]*packages.Package
	loadDuration  time.Duration
}

var (
	defaultPackageLoaderMu sync.Mutex
	defaultPackageLoader   *packageLoader
)

func newPackageLoader(mod *GoModule) *packageLoader {
	return &packageLoader{
		module:        mod,
		packagesByDir: make(map[string][]*packages.Package),
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

	pl.mu.Lock()
	defer pl.mu.Unlock()

	if err := pl.ensureModuleLoadedLocked(absDir, debug); err != nil {
		return nil, err
	}

	pkgs := pl.packagesByDir[absDir]
	if len(pkgs) == 0 {
		return nil, fmt.Errorf("package loader: no packages found for %s", absDir)
	}

	debugf(debug, "packageLoader: reuse %s (%d package(s)) tests=%t from module cache preloaded in %s", absDir, len(pkgs), includeTests, pl.loadDuration)

	return pkgs, nil
}

func (pl *packageLoader) ensureModuleLoadedLocked(hintDir string, debug DebugFunc) error {
	if pl.preloaded {
		return nil
	}

	rootDir := hintDir
	if pl.module != nil && pl.module.Base != "" {
		rootDir = pl.module.Base
	}

	rootDir = filepath.Clean(rootDir)

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
		Dir:   rootDir,
		Tests: true,
	}

	start := time.Now()
	pkgs, err := packages.Load(cfg, "./...")
	duration := time.Since(start)

	if err != nil {
		return fmt.Errorf("package loader: preload %s failed: %w", rootDir, err)
	}

	if len(pkgs) == 0 {
		return fmt.Errorf("package loader: preload %s returned no packages", rootDir)
	}

	pl.preloaded = true
	pl.moduleDir = rootDir
	pl.allPackages = pkgs
	pl.loadDuration = duration

	for _, pkg := range pkgs {
		dir := packageDirectory(pkg)
		if dir == "" {
			continue
		}
		dir = filepath.Clean(dir)
		pl.packagesByDir[dir] = append(pl.packagesByDir[dir], pkg)
	}

	debugf(debug, "packageLoader: preloaded module %s (%d package(s)) tests=true in %s", rootDir, len(pkgs), duration)

	return nil
}

func packageDirectory(pkg *packages.Package) string {
	candidates := make([]string, 0, len(pkg.GoFiles)+len(pkg.CompiledGoFiles)+len(pkg.OtherFiles)+len(pkg.IgnoredFiles))
	candidates = append(candidates, pkg.GoFiles...)
	candidates = append(candidates, pkg.CompiledGoFiles...)
	candidates = append(candidates, pkg.OtherFiles...)
	candidates = append(candidates, pkg.IgnoredFiles...)

	for _, file := range candidates {
		if file == "" {
			continue
		}
		return filepath.Dir(file)
	}

	return ""
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
