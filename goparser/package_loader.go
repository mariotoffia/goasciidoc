package goparser

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
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
	buildTags     []string
	allBuildTags  bool
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

func (pl *packageLoader) load(dir string, includeTests bool, buildTags []string, allBuildTags bool, debug DebugFunc) ([]*packages.Package, error) {
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

	// Check if we need to reload due to different build tags
	if pl.preloaded && (!equalStringSlices(pl.buildTags, buildTags) || pl.allBuildTags != allBuildTags) {
		debugf(debug, "packageLoader: build tags changed, forcing reload")
		pl.preloaded = false
		pl.packagesByDir = make(map[string][]*packages.Package)
		pl.allPackages = nil
	}

	pl.buildTags = buildTags
	pl.allBuildTags = allBuildTags

	if err := pl.ensureModuleLoadedLocked(absDir, debug); err != nil {
		return nil, err
	}

	pkgs := pl.packagesByDir[absDir]
	if len(pkgs) == 0 {
		return nil, fmt.Errorf("package loader: no packages found for %s", absDir)
	}

	debugf(debug, "packageLoader: reuse %s (%d package(s)) tests=%t tags=%v from module cache preloaded in %s", absDir, len(pkgs), includeTests, buildTags, pl.loadDuration)

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

	// Configure build tags
	if pl.allBuildTags {
		// Load all build tags by discovering them from the source
		discoveredTags, err := discoverBuildTags(rootDir, debug)
		if err != nil {
			debugf(debug, "packageLoader: failed to discover build tags: %v", err)
		} else if len(discoveredTags) > 0 {
			cfg.BuildFlags = []string{"-tags=" + strings.Join(discoveredTags, ",")}
			debugf(debug, "packageLoader: discovered build tags: %v", discoveredTags)
		}
	} else if len(pl.buildTags) > 0 {
		// Use explicitly specified build tags
		cfg.BuildFlags = []string{"-tags=" + strings.Join(pl.buildTags, ",")}
		debugf(debug, "packageLoader: using build tags: %v", pl.buildTags)
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

// equalStringSlices compares two string slices for equality
func equalStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// discoverBuildTags scans Go source files to find all unique build tags
func discoverBuildTags(rootDir string, debug DebugFunc) ([]string, error) {
	tagSet := make(map[string]bool)
	buildTagRegex := regexp.MustCompile(`^//\s*\+build\s+(.+)$`)
	goBuildRegex := regexp.MustCompile(`^//go:build\s+(.+)$`)

	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			// Skip vendor and .git directories
			name := info.Name()
			if name == "vendor" || name == ".git" || strings.HasPrefix(name, ".") {
				return filepath.SkipDir
			}
			return nil
		}

		if !strings.HasSuffix(path, ".go") {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return nil // Skip files we can't open
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		lineCount := 0
		for scanner.Scan() && lineCount < 20 { // Only check first 20 lines
			line := strings.TrimSpace(scanner.Text())
			lineCount++

			// Stop at package declaration
			if strings.HasPrefix(line, "package ") {
				break
			}

			// Check for //go:build directives (newer style)
			if matches := goBuildRegex.FindStringSubmatch(line); len(matches) > 1 {
				extractBuildTags(matches[1], tagSet)
			}

			// Check for // +build directives (older style)
			if matches := buildTagRegex.FindStringSubmatch(line); len(matches) > 1 {
				extractBuildTags(matches[1], tagSet)
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Convert set to sorted slice
	tags := make([]string, 0, len(tagSet))
	for tag := range tagSet {
		tags = append(tags, tag)
	}

	debugf(debug, "packageLoader: discovered %d unique build tags", len(tags))
	return tags, nil
}

// extractBuildTags parses build constraint expressions and extracts individual tags
func extractBuildTags(expr string, tagSet map[string]bool) {
	// Remove parentheses and split by logical operators
	expr = strings.ReplaceAll(expr, "(", " ")
	expr = strings.ReplaceAll(expr, ")", " ")
	expr = strings.ReplaceAll(expr, "&&", " ")
	expr = strings.ReplaceAll(expr, "||", " ")
	expr = strings.ReplaceAll(expr, ",", " ")

	fields := strings.Fields(expr)
	for _, field := range fields {
		field = strings.TrimSpace(field)
		// Skip negations and built-in tags
		if strings.HasPrefix(field, "!") {
			continue
		}
		// Skip GOOS and GOARCH values (common but not custom build tags we want)
		if isBuiltinConstraint(field) {
			continue
		}
		if field != "" {
			tagSet[field] = true
		}
	}
}

// isBuiltinConstraint checks if a tag is a built-in Go constraint
func isBuiltinConstraint(tag string) bool {
	builtins := []string{
		"linux", "darwin", "windows", "freebsd", "openbsd", "netbsd", "dragonfly", "solaris", "plan9", "aix", "js",
		"amd64", "386", "arm", "arm64", "ppc64", "ppc64le", "mips", "mipsle", "mips64", "mips64le", "s390x", "riscv64", "wasm",
		"cgo", "race", "msan", "asan",
	}
	for _, builtin := range builtins {
		if tag == builtin {
			return true
		}
	}
	return false
}

// multiModuleLoader manages package loading across multiple modules
type multiModuleLoader struct {
	mu      sync.Mutex
	loaders map[string]*packageLoader // module base path -> loader
	modules []*GoModule
}

// newMultiModuleLoader creates a new multi-module package loader
func newMultiModuleLoader(modules []*GoModule) *multiModuleLoader {
	loaders := make(map[string]*packageLoader)
	for _, mod := range modules {
		loaders[mod.Base] = mod.getPackageLoader()
	}

	return &multiModuleLoader{
		loaders: loaders,
		modules: modules,
	}
}

// load determines which module owns the path and delegates to appropriate loader
func (ml *multiModuleLoader) load(dir string, includeTests bool, buildTags []string, allBuildTags bool, debug DebugFunc) ([]*packages.Package, error) {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}

	ml.mu.Lock()
	defer ml.mu.Unlock()

	// Find which module owns this directory
	var owningModule *GoModule
	var owningLoader *packageLoader

	for _, mod := range ml.modules {
		if strings.HasPrefix(absDir, mod.Base) {
			// Found a match - check if it's more specific than current match
			if owningModule == nil || len(mod.Base) > len(owningModule.Base) {
				owningModule = mod
				owningLoader = ml.loaders[mod.Base]
			}
		}
	}

	if owningLoader == nil {
		return nil, fmt.Errorf("multiModuleLoader: no module found for directory %s", absDir)
	}

	debugf(debug, "multiModuleLoader: using module %s for directory %s", owningModule.Name, absDir)

	// Unlock before delegating to avoid deadlock
	ml.mu.Unlock()
	defer ml.mu.Lock()

	return owningLoader.load(dir, includeTests, buildTags, allBuildTags, debug)
}

// preload loads all packages for all modules (can be done in parallel)
func (ml *multiModuleLoader) preload(includeTests bool, buildTags []string, allBuildTags bool, debug DebugFunc) error {
	var wg sync.WaitGroup
	errChan := make(chan error, len(ml.loaders))

	ml.mu.Lock()
	loaders := make([]*packageLoader, 0, len(ml.loaders))
	modules := make([]*GoModule, 0, len(ml.modules))
	for _, mod := range ml.modules {
		loaders = append(loaders, ml.loaders[mod.Base])
		modules = append(modules, mod)
	}
	ml.mu.Unlock()

	for i, loader := range loaders {
		wg.Add(1)
		go func(l *packageLoader, m *GoModule) {
			defer wg.Done()
			debugf(debug, "multiModuleLoader: preloading module %s", m.Name)
			_, err := l.load(m.Base, includeTests, buildTags, allBuildTags, debug)
			if err != nil {
				errChan <- fmt.Errorf("failed to preload module %s: %w", m.Name, err)
			}
		}(loader, modules[i])
	}

	wg.Wait()
	close(errChan)

	// Return first error if any
	if err := <-errChan; err != nil {
		return err
	}

	return nil
}
