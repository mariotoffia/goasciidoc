package goparser

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindWorkspace(t *testing.T) {
	// Get absolute path to test fixtures
	testRoot, err := filepath.Abs("../.temp-files/tests/workspace-test")
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}

	// Skip test if fixtures don't exist
	if _, err := os.Stat(testRoot); os.IsNotExist(err) {
		t.Skip("Test fixtures not found, skipping workspace tests")
	}

	tests := []struct {
		name      string
		startPath string
		wantError bool
		wantCount int // expected number of modules
	}{
		{
			name:      "find workspace from root",
			startPath: testRoot,
			wantError: false,
			wantCount: 3,
		},
		{
			name:      "find workspace from subdirectory",
			startPath: filepath.Join(testRoot, "module1"),
			wantError: false,
			wantCount: 3,
		},
		{
			name:      "find workspace from nested subdirectory",
			startPath: filepath.Join(testRoot, "subdir/module3"),
			wantError: false,
			wantCount: 3,
		},
		{
			name:      "no workspace found",
			startPath: "/tmp",
			wantError: true,
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			workspace, err := FindWorkspace(tt.startPath)

			if tt.wantError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				if err != ErrWorkspaceNotFound {
					t.Errorf("Expected ErrWorkspaceNotFound, got %v", err)
				}
				return
			}

			if err != nil {
				t.Fatalf("FindWorkspace() error = %v", err)
			}

			if workspace == nil {
				t.Fatal("Expected workspace, got nil")
			}

			if len(workspace.Modules) != tt.wantCount {
				t.Errorf("Expected %d modules, got %d", tt.wantCount, len(workspace.Modules))
			}

			// Verify workspace base is correct
			if workspace.Base != testRoot {
				t.Errorf("Expected base %s, got %s", testRoot, workspace.Base)
			}

			// Verify modules are loaded
			expectedModules := map[string]bool{
				"github.com/test/module1": false,
				"github.com/test/module2": false,
				"github.com/test/module3": false,
			}

			for _, mod := range workspace.Modules {
				if _, exists := expectedModules[mod.Name]; exists {
					expectedModules[mod.Name] = true
				}
			}

			for name, found := range expectedModules {
				if !found {
					t.Errorf("Module %s not found in workspace", name)
				}
			}
		})
	}
}

func TestLoadWorkspace(t *testing.T) {
	testRoot, err := filepath.Abs("../.temp-files/tests/workspace-test")
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}

	workPath := filepath.Join(testRoot, "go.work")

	// Skip if fixtures don't exist
	if _, err := os.Stat(workPath); os.IsNotExist(err) {
		t.Skip("Test fixtures not found, skipping")
	}

	workspace, err := LoadWorkspace(workPath)
	if err != nil {
		t.Fatalf("LoadWorkspace() error = %v", err)
	}

	if workspace == nil {
		t.Fatal("Expected workspace, got nil")
	}

	if len(workspace.Modules) != 3 {
		t.Errorf("Expected 3 modules, got %d", len(workspace.Modules))
	}

	if workspace.FilePath != workPath {
		t.Errorf("Expected FilePath %s, got %s", workPath, workspace.FilePath)
	}
}

func TestWorkspaceContainsModule(t *testing.T) {
	workspace := &GoWorkspace{
		Modules: []*GoModule{
			{Name: "github.com/test/module1"},
			{Name: "github.com/test/module2"},
			{Name: "github.com/other/package"},
		},
	}

	tests := []struct {
		name       string
		importPath string
		want       bool
	}{
		{
			name:       "exact match",
			importPath: "github.com/test/module1",
			want:       true,
		},
		{
			name:       "subpackage match",
			importPath: "github.com/test/module1/internal/service",
			want:       true,
		},
		{
			name:       "no match",
			importPath: "github.com/external/package",
			want:       false,
		},
		{
			name:       "partial match should not work",
			importPath: "github.com/test",
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := workspace.ContainsModule(tt.importPath)
			if got != tt.want {
				t.Errorf("ContainsModule(%q) = %v, want %v", tt.importPath, got, tt.want)
			}
		})
	}
}

func TestWorkspaceModuleForPath(t *testing.T) {
	module1 := &GoModule{Name: "github.com/test/module1"}
	module2 := &GoModule{Name: "github.com/test/module2"}
	longModule := &GoModule{Name: "github.com/test/module1/submodule"}

	workspace := &GoWorkspace{
		Modules: []*GoModule{module1, module2, longModule},
	}

	tests := []struct {
		name       string
		importPath string
		want       *GoModule
	}{
		{
			name:       "exact match module1",
			importPath: "github.com/test/module1",
			want:       module1,
		},
		{
			name:       "subpackage of module2",
			importPath: "github.com/test/module2/internal",
			want:       module2,
		},
		{
			name:       "longest match wins",
			importPath: "github.com/test/module1/submodule/pkg",
			want:       longModule,
		},
		{
			name:       "no match",
			importPath: "github.com/external/package",
			want:       nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := workspace.ModuleForPath(tt.importPath)
			if got != tt.want {
				var gotName, wantName string
				if got != nil {
					gotName = got.Name
				}
				if tt.want != nil {
					wantName = tt.want.Name
				}
				t.Errorf("ModuleForPath(%q) = %v, want %v", tt.importPath, gotName, wantName)
			}
		})
	}
}

func TestModuleShortName(t *testing.T) {
	tests := []struct {
		name       string
		moduleName string
		want       string
	}{
		{
			name:       "simple name",
			moduleName: "github.com/user/project",
			want:       "project",
		},
		{
			name:       "nested module",
			moduleName: "github.com/user/project/submodule",
			want:       "submodule",
		},
		{
			name:       "name with dots",
			moduleName: "github.com/user/my.project",
			want:       "my-project",
		},
		{
			name:       "name with underscores",
			moduleName: "github.com/user/my_project",
			want:       "my-project",
		},
		{
			name:       "nil module",
			moduleName: "",
			want:       "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var module *GoModule
			if tt.moduleName != "" {
				module = &GoModule{Name: tt.moduleName}
			}

			got := ModuleShortName(module)
			if got != tt.want {
				t.Errorf("ModuleShortName(%v) = %q, want %q", tt.moduleName, got, tt.want)
			}
		})
	}
}

func TestFindModuleOrWorkspace(t *testing.T) {
	workspaceRoot, err := filepath.Abs("../.temp-files/tests/workspace-test")
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}

	singleModRoot, err := filepath.Abs("../.temp-files/tests/single-module-test")
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}

	// Skip if fixtures don't exist
	if _, err := os.Stat(workspaceRoot); os.IsNotExist(err) {
		t.Skip("Test fixtures not found, skipping")
	}

	tests := []struct {
		name          string
		startPath     string
		wantWorkspace bool
		wantModule    bool
		wantError     bool
	}{
		{
			name:          "find workspace from root",
			startPath:     workspaceRoot,
			wantWorkspace: true,
			wantModule:    false,
			wantError:     false,
		},
		{
			name:          "find module when starting from module subdirectory",
			startPath:     filepath.Join(workspaceRoot, "module1"),
			wantWorkspace: false,
			wantModule:    true,
			wantError:     false,
		},
		{
			name:          "find single module",
			startPath:     singleModRoot,
			wantWorkspace: false,
			wantModule:    true,
			wantError:     false,
		},
		{
			name:          "prefer go.work over go.mod in same directory",
			startPath:     workspaceRoot,
			wantWorkspace: true, // Should find go.work, not go.mod in subdir
			wantModule:    false,
			wantError:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			workspace, module, err := FindModuleOrWorkspace(tt.startPath)

			if tt.wantError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("FindModuleOrWorkspace() error = %v", err)
			}

			if tt.wantWorkspace && workspace == nil {
				t.Error("Expected workspace, got nil")
			}

			if !tt.wantWorkspace && workspace != nil {
				t.Errorf("Expected no workspace, got %v", workspace)
			}

			if tt.wantModule && module == nil {
				t.Error("Expected module, got nil")
			}

			if !tt.wantModule && module != nil {
				t.Errorf("Expected no module, got %v", module)
			}

			// Ensure exactly one is non-nil
			if (workspace == nil) == (module == nil) {
				t.Error("Expected exactly one of workspace or module to be non-nil")
			}
		})
	}
}
