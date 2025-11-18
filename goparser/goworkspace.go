package goparser

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/mod/modfile"
)

var (
	// ErrWorkspaceNotFound is returned when no go.work file is found
	ErrWorkspaceNotFound = errors.New("go.work file not found")
	// ErrNoModulesFound is returned when workspace has no modules
	ErrNoModulesFound = errors.New("no modules found in workspace")
)

// GoWorkspace represents a Go workspace (go.work file)
type GoWorkspace struct {
	File      *modfile.WorkFile        // Parsed go.work content
	FilePath  string                   // Absolute path to go.work
	Base      string                   // Directory containing go.work
	Modules   []*GoModule              // All modules in workspace
	ModuleMap map[string]*GoModule     // Module name -> GoModule (for fast lookup)
}

// FindWorkspace searches for go.work file starting from path
// It walks up the directory tree until a go.work file is found
// Returns ErrWorkspaceNotFound if no workspace found
func FindWorkspace(path string) (*GoWorkspace, error) {
	// Make path absolute
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	// If path is a file, start from its directory
	info, err := os.Stat(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat path: %w", err)
	}

	searchPath := absPath
	if !info.IsDir() {
		searchPath = filepath.Dir(absPath)
	}

	// Walk up directory tree
	for {
		workPath := filepath.Join(searchPath, "go.work")
		if _, err := os.Stat(workPath); err == nil {
			// Found go.work
			return LoadWorkspace(workPath)
		}

		// Move to parent directory
		parent := filepath.Dir(searchPath)
		if parent == searchPath {
			// Reached root
			break
		}
		searchPath = parent
	}

	return nil, ErrWorkspaceNotFound
}

// LoadWorkspace loads a workspace from the specified go.work file
func LoadWorkspace(workPath string) (*GoWorkspace, error) {
	// Read go.work file
	data, err := os.ReadFile(workPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read go.work: %w", err)
	}

	// Parse go.work file
	workFile, err := modfile.ParseWork(workPath, data, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to parse go.work: %w", err)
	}

	absWorkPath, err := filepath.Abs(workPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	workspace := &GoWorkspace{
		File:      workFile,
		FilePath:  absWorkPath,
		Base:      filepath.Dir(absWorkPath),
		Modules:   make([]*GoModule, 0),
		ModuleMap: make(map[string]*GoModule),
	}

	// Load each module from "use" directives
	for _, use := range workFile.Use {
		modulePath := use.Path
		// Module paths in go.work are relative to workspace root
		absModulePath := filepath.Join(workspace.Base, modulePath)

		// Find go.mod in this path
		modPath := filepath.Join(absModulePath, "go.mod")
		if _, err := os.Stat(modPath); err != nil {
			// Module doesn't have go.mod, skip
			continue
		}

		// Load the module
		module, err := LoadModule(modPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load module at %s: %w", modulePath, err)
		}

		workspace.Modules = append(workspace.Modules, module)
		workspace.ModuleMap[module.Name] = module
	}

	if len(workspace.Modules) == 0 {
		return nil, ErrNoModulesFound
	}

	return workspace, nil
}

// ContainsModule checks if an import path belongs to any workspace module
func (w *GoWorkspace) ContainsModule(importPath string) bool {
	for _, module := range w.Modules {
		if strings.HasPrefix(importPath, module.Name) {
			return true
		}
	}
	return false
}

// ModuleForPath returns the module that owns the given import path
// Returns nil if no module matches
func (w *GoWorkspace) ModuleForPath(importPath string) *GoModule {
	// Find the longest matching module name
	var bestMatch *GoModule
	bestMatchLen := 0

	for _, module := range w.Modules {
		if strings.HasPrefix(importPath, module.Name) {
			if len(module.Name) > bestMatchLen {
				bestMatch = module
				bestMatchLen = len(module.Name)
			}
		}
	}

	return bestMatch
}

// ModuleShortName returns a short name for a module suitable for anchors
// Uses the last component of the module path
func ModuleShortName(module *GoModule) string {
	if module == nil {
		return ""
	}

	// Get the last component of the module name
	parts := strings.Split(module.Name, "/")
	if len(parts) == 0 {
		return module.Name
	}

	shortName := parts[len(parts)-1]
	// Sanitize for use in anchors (remove special chars)
	shortName = strings.ReplaceAll(shortName, ".", "-")
	shortName = strings.ReplaceAll(shortName, "_", "-")
	return shortName
}

// FindModuleOrWorkspace walks up from path checking for go.work or go.mod
// Prefers go.work over go.mod at each directory level
// Returns workspace, module, error
// Exactly one of workspace or module will be non-nil on success
func FindModuleOrWorkspace(path string) (*GoWorkspace, *GoModule, error) {
	// Make path absolute
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	// If path is a file, start from its directory
	info, err := os.Stat(absPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to stat path: %w", err)
	}

	searchPath := absPath
	if !info.IsDir() {
		searchPath = filepath.Dir(absPath)
	}

	// Walk up directory tree
	for {
		// Check for go.work first (preferred)
		workPath := filepath.Join(searchPath, "go.work")
		if _, err := os.Stat(workPath); err == nil {
			workspace, err := LoadWorkspace(workPath)
			if err != nil {
				// go.work exists but couldn't load - return error
				return nil, nil, fmt.Errorf("found go.work but failed to load: %w", err)
			}
			return workspace, nil, nil
		}

		// Check for go.mod (fallback)
		modPath := filepath.Join(searchPath, "go.mod")
		if _, err := os.Stat(modPath); err == nil {
			module, err := LoadModule(modPath)
			if err != nil {
				// go.mod exists but couldn't load - return error
				return nil, nil, fmt.Errorf("found go.mod but failed to load: %w", err)
			}
			return nil, module, nil
		}

		// Move to parent directory
		parent := filepath.Dir(searchPath)
		if parent == searchPath {
			// Reached root without finding either
			break
		}
		searchPath = parent
	}

	return nil, nil, fmt.Errorf("no go.work or go.mod found for %s", absPath)
}
