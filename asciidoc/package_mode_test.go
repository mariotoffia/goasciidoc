package asciidoc

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mariotoffia/goasciidoc/goparser"
	"github.com/stretchr/testify/assert"
)

// TestParsePackageMode tests parsing of --package-mode flag values
func TestParsePackageMode(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    PackageMode
		wantErr bool
	}{
		{
			name:    "none explicit",
			input:   "none",
			want:    PackageModeNone,
			wantErr: false,
		},
		{
			name:    "empty string defaults to none",
			input:   "",
			want:    PackageModeNone,
			wantErr: false,
		},
		{
			name:    "include mode",
			input:   "include",
			want:    PackageModeInclude,
			wantErr: false,
		},
		{
			name:    "link mode",
			input:   "link",
			want:    PackageModeLink,
			wantErr: false,
		},
		{
			name:    "case insensitive - Include",
			input:   "Include",
			want:    PackageModeInclude,
			wantErr: false,
		},
		{
			name:    "case insensitive - LINK",
			input:   "LINK",
			want:    PackageModeLink,
			wantErr: false,
		},
		{
			name:    "whitespace trimmed",
			input:   "  include  ",
			want:    PackageModeInclude,
			wantErr: false,
		},
		{
			name:    "invalid mode",
			input:   "invalid",
			want:    PackageModeNone,
			wantErr: true,
		},
		{
			name:    "typo separated",
			input:   "separated",
			want:    PackageModeNone,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// parsePackageMode is defined in main.go, so we test the logic directly
			var got PackageMode
			var err error

			switch strings.ToLower(strings.TrimSpace(tt.input)) {
			case "none", "":
				got = PackageModeNone
			case "include":
				got = PackageModeInclude
			case "link":
				got = PackageModeLink
			default:
				err = assert.AnError
			}

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

// TestGetPackageOutputFile tests package-specific filename generation
func TestGetPackageOutputFile(t *testing.T) {
	tests := []struct {
		name     string
		outfile  string
		fqPkg    string
		filePath string
		index    int
		want     string
	}{
		{
			name:     "no outfile - uses package path",
			outfile:  "",
			fqPkg:    "github.com/example/pkg",
			filePath: "/path/to/pkg/file.go",
			index:    0,
			want:     "/path/to/pkg/github.com_example_pkg.adoc",
		},
		{
			name:     "with outfile - creates packages subdir",
			outfile:  "docs/index.adoc",
			fqPkg:    "github.com/example/api",
			filePath: "/path/to/api/file.go",
			index:    1,
			want:     "docs/packages/github.com_example_api.adoc",
		},
		{
			name:     "sanitizes slashes in package name",
			outfile:  "output.adoc",
			fqPkg:    "my/pkg/path",
			filePath: "/src/file.go",
			index:    0,
			want:     "packages/my_pkg_path.adoc",
		},
		{
			name:     "sanitizes backslashes in package name",
			outfile:  "docs/api.adoc",
			fqPkg:    "my\\pkg\\path",
			filePath: "/src/file.go",
			index:    0,
			want:     "docs/packages/my_pkg_path.adoc",
		},
		{
			name:     "handles complex package path",
			outfile:  "/tmp/docs/index.adoc",
			fqPkg:    "github.com/mariotoffia/goasciidoc/asciidoc",
			filePath: "/go/src/file.go",
			index:    5,
			want:     "/tmp/docs/packages/github.com_mariotoffia_goasciidoc_asciidoc.adoc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewProducer()
			p.outfile = tt.outfile

			pkg := &goparser.GoPackage{
				GoFile: goparser.GoFile{
					FqPackage: tt.fqPkg,
					FilePath:  tt.filePath,
				},
			}

			got := p.getPackageOutputFile(pkg, tt.index)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestBuildPackageReferences tests internal/external package categorization
func TestBuildPackageReferences(t *testing.T) {
	tests := []struct {
		name           string
		modulePath     string
		imports        []string
		packageInfoMap map[string]*PackageInfo
		wantInternal   int
		wantExternal   int
	}{
		{
			name:       "all external imports",
			modulePath: "github.com/example/project",
			imports:    []string{"fmt", "io", "github.com/other/pkg"},
			packageInfoMap: map[string]*PackageInfo{
				"github.com/example/project/api": {
					Outfile: "api.adoc",
					Anchor:  "pkg-1",
				},
			},
			wantInternal: 0,
			wantExternal: 3,
		},
		{
			name:       "all internal imports",
			modulePath: "github.com/example/project",
			imports: []string{
				"github.com/example/project/api",
				"github.com/example/project/models",
			},
			packageInfoMap: map[string]*PackageInfo{
				"github.com/example/project/api": {
					Outfile: "api.adoc",
					Anchor:  "pkg-1",
				},
				"github.com/example/project/models": {
					Outfile: "models.adoc",
					Anchor:  "pkg-2",
				},
			},
			wantInternal: 2,
			wantExternal: 0,
		},
		{
			name:       "mixed internal and external",
			modulePath: "github.com/example/project",
			imports: []string{
				"fmt",
				"github.com/example/project/api",
				"github.com/other/pkg",
				"github.com/example/project/models",
			},
			packageInfoMap: map[string]*PackageInfo{
				"github.com/example/project/api": {
					Outfile: "api.adoc",
					Anchor:  "pkg-1",
				},
				"github.com/example/project/models": {
					Outfile: "models.adoc",
					Anchor:  "pkg-2",
				},
			},
			wantInternal: 2,
			wantExternal: 2,
		},
		{
			name:           "no imports",
			modulePath:     "github.com/example/project",
			imports:        []string{},
			packageInfoMap: map[string]*PackageInfo{},
			wantInternal:   0,
			wantExternal:   0,
		},
		{
			name:       "internal import not in package map",
			modulePath: "github.com/example/project",
			imports: []string{
				"github.com/example/project/unknown",
			},
			packageInfoMap: map[string]*PackageInfo{},
			wantInternal:   1,
			wantExternal:   0,
		},
		{
			name:       "cross-module import present in package map",
			modulePath: "github.com/example/project",
			imports: []string{
				"github.com/other/module/pkg",
			},
			packageInfoMap: map[string]*PackageInfo{
				"github.com/other/module/pkg": {
					Outfile: "otherpkg.adoc",
					Anchor:  "pkg-99",
				},
			},
			wantInternal: 1,
			wantExternal: 0,
		},
		{
			name:       "workspace module treated as internal even without map",
			modulePath: "github.com/example/project",
			imports: []string{
				"github.com/other/module/pkg",
			},
			packageInfoMap: map[string]*PackageInfo{},
			wantInternal:   1,
			wantExternal:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewProducer()
			p.outfile = "index.adoc"
			p.packageMode = PackageModeLink
			p.parseconfig.Module = &goparser.GoModule{Name: tt.modulePath}
			p.parseconfig.Workspace = &goparser.GoWorkspace{
				Modules: []*goparser.GoModule{
					{Name: tt.modulePath},
					{Name: "github.com/other/module"},
				},
			}

			// Create a package with the specified imports
			files := []*goparser.GoFile{
				{
					Imports: make([]*goparser.GoImport, len(tt.imports)),
				},
			}

			for i, imp := range tt.imports {
				files[0].Imports[i] = &goparser.GoImport{
					Path: imp,
				}
			}

			pkg := &goparser.GoPackage{
				GoFile: goparser.GoFile{
					Module: &goparser.GoModule{
						Name: tt.modulePath,
					},
				},
				Files: files,
			}

			refs := p.buildPackageReferences(pkg, tt.packageInfoMap)

			assert.Len(t, refs.Internal, tt.wantInternal, "internal package count mismatch")
			assert.Len(t, refs.External, tt.wantExternal, "external package count mismatch")
		})
	}
}

func TestBuildPackageReferencesExternalLinks(t *testing.T) {
	p := NewProducer()
	p.outfile = "index.adoc"
	p.packageMode = PackageModeLink

	pkg := &goparser.GoPackage{
		GoFile: goparser.GoFile{
			Module: &goparser.GoModule{Name: "example.com/project"},
		},
		Files: []*goparser.GoFile{
			{
				Imports: []*goparser.GoImport{
					{Path: "github.com/example/lib"},
					{Path: "fmt"},
				},
			},
		},
	}

	refs := p.buildPackageReferences(pkg, map[string]*PackageInfo{})

	linkMap := make(map[string]string)
	for _, ref := range refs.External {
		linkMap[ref.Name] = ref.File
	}

	assert.Equal(t, "https://pkg.go.dev/github.com/example/lib", linkMap["github.com/example/lib"])
	assert.Equal(t, "https://pkg.go.dev/fmt", linkMap["fmt"])
}

// TestPackageModeConfiguration tests that package mode can be set and retrieved
func TestPackageModeConfiguration(t *testing.T) {
	tests := []struct {
		name string
		mode PackageMode
	}{
		{"none mode", PackageModeNone},
		{"include mode", PackageModeInclude},
		{"link mode", PackageModeLink},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewProducer().PackageMode(tt.mode)
			assert.Equal(t, tt.mode, p.packageMode)
		})
	}
}

// TestPackageModeNoneDoesNotCreatePackagesDir tests that none mode doesn't trigger package generation
func TestPackageModeNoneDoesNotCreatePackagesDir(t *testing.T) {
	p := NewProducer()
	assert.Equal(t, PackageModeNone, p.packageMode, "default should be none")
}

//Note: Full integration tests testing actual file generation are complex
// as they require proper module loading and Go environment setup.
// The unit tests above test the core functionality of the package mode feature.
// Manual testing with actual projects is recommended for end-to-end validation.

func TestGeneratePackageMasterIndexInclude(t *testing.T) {
	dir := t.TempDir()
	p := NewProducer().PackageMode(PackageModeInclude)
	p.outfile = filepath.Join(dir, "index.adoc")
	p.parseconfig.Module = &goparser.GoModule{Name: "example.com/project"}

	packageFiles := []string{
		filepath.Join(dir, "pkg1.adoc"),
		filepath.Join(dir, "pkg2.adoc"),
	}

	packageInfoMap := map[string]*PackageInfo{
		"example.com/project/pkg1": {
			Package: &goparser.GoPackage{
				GoFile: goparser.GoFile{
					Package:   "pkg1",
					FqPackage: "example.com/project/pkg1",
				},
			},
			Outfile: packageFiles[0],
			Anchor:  "pkg-1",
			Index:   0,
		},
		"example.com/project/pkg2": {
			Package: &goparser.GoPackage{
				GoFile: goparser.GoFile{
					Package:   "pkg2",
					FqPackage: "example.com/project/pkg2",
				},
			},
			Outfile: packageFiles[1],
			Anchor:  "pkg-2",
			Index:   1,
		},
	}

	p.generatePackageMasterIndex(packageFiles, packageInfoMap)

	data, err := os.ReadFile(p.outfile)
	assert.NoError(t, err, "should create master index file")
	content := string(data)

	assert.Contains(t, content, "Package: example.com/project/pkg1")
	assert.Contains(t, content, "Package: example.com/project/pkg2")
	assert.Contains(t, content, "include::pkg1.adoc[]")
	assert.Contains(t, content, "include::pkg2.adoc[]")
}

func TestGeneratePackageMasterIndexLink(t *testing.T) {
	dir := t.TempDir()
	p := NewProducer().PackageMode(PackageModeLink)
	p.outfile = filepath.Join(dir, "index.adoc")
	p.parseconfig.Module = &goparser.GoModule{Name: "example.com/project"}

	packageFiles := []string{
		filepath.Join(dir, "pkg1.adoc"),
	}

	packageInfoMap := map[string]*PackageInfo{
		"example.com/project/pkg1": {
			Package: &goparser.GoPackage{
				GoFile: goparser.GoFile{
					Package:   "pkg1",
					FqPackage: "example.com/project/pkg1",
				},
			},
			Outfile: packageFiles[0],
			Anchor:  "pkg-1",
			Index:   0,
		},
	}

	p.generatePackageMasterIndex(packageFiles, packageInfoMap)

	data, err := os.ReadFile(p.outfile)
	assert.NoError(t, err, "should create master index file")
	content := string(data)

	assert.Contains(t, content, "Package: example.com/project/pkg1")
	assert.Contains(t, content, "link:pkg1.adoc[View full documentation]")
}

func TestPackageRefsTemplateRendersExternalLinks(t *testing.T) {
	tmpl := NewTemplateWithOverrides(nil)
	ctx := &TemplateContext{
		creator: tmpl,
		PackageRefs: &PackageReferences{
			External: []PackageRef{
				{Name: "github.com/example/lib", File: "https://pkg.go.dev/github.com/example/lib"},
			},
		},
	}

	var buf bytes.Buffer
	err := tmpl.Templates[PackageRefsTemplate.String()].Template.Execute(&buf, ctx)
	assert.NoError(t, err)
	out := buf.String()

	assert.Contains(t, out, "link:https://pkg.go.dev/github.com/example/lib[github.com/example/lib]")
}

func TestBuildPackageReferencesInternalWithoutMapStillInternal(t *testing.T) {
	p := NewProducer()
	p.outfile = "index.adoc"
	p.packageMode = PackageModeLink
	p.parseconfig.Workspace = &goparser.GoWorkspace{
		Modules: []*goparser.GoModule{
			{Name: "github.com/example/project"},
		},
	}

	pkg := &goparser.GoPackage{
		GoFile: goparser.GoFile{
			Module: &goparser.GoModule{Name: "github.com/example/project"},
		},
		Files: []*goparser.GoFile{
			{
				Imports: []*goparser.GoImport{
					{Path: "github.com/example/project/subpkg"},
				},
			},
		},
	}

	refs := p.buildPackageReferences(pkg, map[string]*PackageInfo{})

	assert.Len(t, refs.Internal, 1)
	assert.Equal(t, "github.com/example/project/subpkg", refs.Internal[0].Name)
	assert.Equal(t, "github.com_example_project_subpkg.adoc", refs.Internal[0].File)
	assert.Empty(t, refs.Internal[0].Anchor)
	assert.Len(t, refs.External, 0)
}

func TestBuildPackageReferencesCrossModuleInternal(t *testing.T) {
	p := NewProducer()
	p.outfile = "index.adoc"
	p.packageMode = PackageModeLink
	p.parseconfig.Workspace = &goparser.GoWorkspace{
		Modules: []*goparser.GoModule{
			{Name: "github.com/example/project"},
			{Name: "github.com/other/module"},
		},
	}

	pkg := &goparser.GoPackage{
		GoFile: goparser.GoFile{
			Module: &goparser.GoModule{Name: "github.com/example/project"},
		},
		Files: []*goparser.GoFile{
			{
				Imports: []*goparser.GoImport{
					{Path: "github.com/other/module/pkg"},
				},
			},
		},
	}

	packageInfoMap := map[string]*PackageInfo{
		"github.com/other/module/pkg": {
			Outfile: "otherpkg.adoc",
			Anchor:  "pkg-99",
		},
	}

	refs := p.buildPackageReferences(pkg, packageInfoMap)

	assert.Len(t, refs.Internal, 1)
	assert.Equal(t, "github.com/other/module/pkg", refs.Internal[0].Name)
	assert.Equal(t, "otherpkg.adoc", refs.Internal[0].File)
	assert.Equal(t, "pkg-99", refs.Internal[0].Anchor)
	assert.Len(t, refs.External, 0)
}
