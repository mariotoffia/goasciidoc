package goparser

import (
	"go/importer"
	"go/token"
	"go/types"
	"os"
	"strings"
	"sync"
	"time"
)

type moduleImporter struct {
	mu           sync.Mutex
	module       *GoModule
	importer     types.Importer
	importerFrom types.ImporterFrom
	seen         map[string]int
	debug        DebugFunc
}

var (
	defaultModuleImporterMu sync.Mutex
	defaultModuleImporter   *moduleImporter
	ensureLocalOnce         sync.Once
)

func ensureLocalModulePreference() {
	ensureLocalOnce.Do(func() {
		if _, ok := os.LookupEnv("GOPROXY"); !ok {
			// Favor the local module cache first while still permitting network fallback.
			// This mirrors Go's default proxy chain but makes the behavior explicit for callers
			// running in stripped or custom environments.
			os.Setenv("GOPROXY", "https://proxy.golang.org,direct")
		}
	})
}

func newModuleImporter(mod *GoModule, debug DebugFunc) *moduleImporter {
	ensureLocalModulePreference()

	base := importer.ForCompiler(token.NewFileSet(), "source", nil)

	mi := &moduleImporter{
		module:   mod,
		importer: base,
		seen:     make(map[string]int),
		debug:    debug,
	}

	if importerFrom, ok := base.(types.ImporterFrom); ok {
		mi.importerFrom = importerFrom
	}

	return mi
}

func (mi *moduleImporter) setDebug(debug DebugFunc) {
	mi.mu.Lock()
	defer mi.mu.Unlock()
	mi.debug = debug
}

func (mi *moduleImporter) Import(path string) (*types.Package, error) {
	return mi.ImportFrom(path, ".", 0)
}

func (mi *moduleImporter) ImportFrom(path, dir string, mode types.ImportMode) (*types.Package, error) {
	start := time.Now()

	mi.mu.Lock()
	count := mi.seen[path]
	debugFn := mi.debug
	mi.mu.Unlock()

	var (
		pkg *types.Package
		err error
	)

	if mi.importerFrom != nil {
		pkg, err = mi.importerFrom.ImportFrom(path, dir, mode)
	} else {
		pkg, err = mi.importer.Import(path)
	}

	if err == nil {
		mi.mu.Lock()
		mi.seen[path] = count + 1
		mi.mu.Unlock()
	}

	if debugFn != nil {
		status := "loaded"
		if count > 0 && err == nil {
			status = "cached"
		}

		duration := time.Since(start)

		if isLocalImport(path, mi) {
			status = "local-" + status
		}

		if err != nil {
			debugf(debugFn, "typeCheck: import %s from %s failed after %s: %v", path, dir, duration, err)
		} else {
			debugf(debugFn, "typeCheck: import %s from %s %s in %s", path, dir, status, duration)
		}
	}

	return pkg, err
}

func getSharedModuleImporter(mod *GoModule, debug DebugFunc) types.Importer {
	if mod != nil {
		return mod.getModuleImporter(debug)
	}

	defaultModuleImporterMu.Lock()
	defer defaultModuleImporterMu.Unlock()

	if defaultModuleImporter == nil {
		defaultModuleImporter = newModuleImporter(nil, debug)
	} else {
		defaultModuleImporter.setDebug(debug)
	}

	return defaultModuleImporter
}

func isLocalImport(path string, mi *moduleImporter) bool {
	if path == "" || mi == nil || mi.module == nil {
		return false
	}

	modPath := strings.TrimSuffix(mi.module.Name, "/")
	if modPath == "" {
		return false
	}

	return path == modPath || strings.HasPrefix(path, modPath+"/")
}
