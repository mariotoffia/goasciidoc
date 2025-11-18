package asciidoc

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"github.com/mariotoffia/goasciidoc/goparser"
)

// CreateTemplateWithOverrides creates a new instance of _Template_
// and add the possible _Provider.overrides_ into it.
func (p *Producer) CreateTemplateWithOverrides() *Template {
	return NewTemplateWithOverrides(p.overrides)
}

// Generate will execute the generation of the documentation
func (p *Producer) Generate() {

	p.debugf("Generate: starting with %d include path(s)", len(p.paths))

	// Package-level rendering takes precedence
	if p.packageMode != PackageModeNone {
		p.generateSeparatePackages()
		p.debugf("Generate: completed package-level rendering")
		return
	}

	// Dispatch based on sub-module mode
	switch p.subModuleMode {
	case SubModuleSingle:
		p.generateMergedModules()
	case SubModuleSeparate:
		p.generateSeparateModules()
	default:
		// SubModuleNone - original single module behavior
		p.generateSingleModule()
	}

	p.debugf("Generate: completed for %d include path(s)", len(p.paths))
}

// generateSingleModule handles the original single-module generation
func (p *Producer) generateSingleModule() {
	t := p.CreateTemplateWithOverrides()
	w := tabwriter.NewWriter(p.createWriter(), 4, 4, 4, ' ', 0)

	overviewpaths := p.overviewpaths
	if len(overviewpaths) == 0 {
		overviewpaths = []string{
			"overview.adoc",
			"_design/overview.adoc",
		}
	}

	indexdone := !p.index

	err := goparser.ParseSinglePackageWalker(
		p.parseconfig,
		p.getProcessFunc(t, w, indexdone, overviewpaths),
		p.paths...,
	)

	if nil != err {
		panic(err)
	}

	w.Flush()
}

// generateMergedModules generates documentation for all modules in a single file
func (p *Producer) generateMergedModules() {
	if p.parseconfig.Workspace == nil {
		// No workspace, fall back to single module
		p.generateSingleModule()
		return
	}

	t := p.CreateTemplateWithOverrides()
	w := tabwriter.NewWriter(p.createWriter(), 4, 4, 4, ' ', 0)

	overviewpaths := p.overviewpaths
	if len(overviewpaths) == 0 {
		overviewpaths = []string{
			"overview.adoc",
			"_design/overview.adoc",
		}
	}

	indexdone := !p.index

	// Process each module in the workspace
	for i, module := range p.parseconfig.Workspace.Modules {
		p.debugf("Generate: processing module %s (%d/%d)", module.Name, i+1, len(p.parseconfig.Workspace.Modules))

		// Create a temporary parseconfig for this module
		moduleConfig := p.parseconfig
		moduleConfig.Module = module

		// Determine paths for this module
		modulePaths := p.getModulePaths(module)
		if len(modulePaths) == 0 {
			p.debugf("Generate: no paths found for module %s, using module base", module.Name)
			modulePaths = []string{module.Base}
		}

		err := goparser.ParseSinglePackageWalker(
			moduleConfig,
			p.getProcessFunc(t, w, indexdone, overviewpaths),
			modulePaths...,
		)

		if err != nil {
			p.debugf("Generate: error processing module %s: %v", module.Name, err)
			// Continue with other modules
			continue
		}

		// Only render index once
		indexdone = true
	}

	w.Flush()
}

// generateSeparateModules generates separate documentation files for each module
func (p *Producer) generateSeparateModules() {
	if p.parseconfig.Workspace == nil {
		// No workspace, fall back to single module
		p.generateSingleModule()
		return
	}

	overviewpaths := p.overviewpaths
	if len(overviewpaths) == 0 {
		overviewpaths = []string{
			"overview.adoc",
			"_design/overview.adoc",
		}
	}

	// Track generated module files for master index
	var moduleFiles []string

	// Process each module separately
	for i, module := range p.parseconfig.Workspace.Modules {
		p.debugf("Generate: processing module %s (%d/%d) as separate file", module.Name, i+1, len(p.parseconfig.Workspace.Modules))

		// Create separate output file for this module
		moduleOutfile := p.getModuleOutputFile(module)
		p.debugf("Generate: writing module %s to %s", module.Name, moduleOutfile)
		moduleFiles = append(moduleFiles, moduleOutfile)

		// Save original settings
		origOutfile := p.outfile
		origWriter := p.writer

		// Set module-specific output
		p.outfile = moduleOutfile
		p.writer = nil // Force file creation

		// Create template and writer for this module
		t := p.CreateTemplateWithOverrides()
		w := tabwriter.NewWriter(p.createWriter(), 4, 4, 4, ' ', 0)

		// Create module-specific parseconfig
		moduleConfig := p.parseconfig
		moduleConfig.Module = module

		// Determine paths for this module
		modulePaths := p.getModulePaths(module)
		if len(modulePaths) == 0 {
			p.debugf("Generate: no paths found for module %s, using module base", module.Name)
			modulePaths = []string{module.Base}
		}

		indexdone := !p.index

		err := goparser.ParseSinglePackageWalker(
			moduleConfig,
			p.getProcessFunc(t, w, indexdone, overviewpaths),
			modulePaths...,
		)

		w.Flush()

		if err != nil {
			p.debugf("Generate: error processing module %s: %v", module.Name, err)
		}

		// Restore original settings
		p.outfile = origOutfile
		p.writer = origWriter
	}

	// Create master index file that includes all modules
	p.generateMasterIndex(moduleFiles)
}

// getModulePaths returns the paths to scan for a specific module
func (p *Producer) getModulePaths(module *goparser.GoModule) []string {
	// If paths are explicitly specified, filter for this module
	if len(p.paths) > 0 {
		var modulePaths []string
		for _, path := range p.paths {
			// Check if path belongs to this module
			absPath, err := filepath.Abs(path)
			if err != nil {
				continue
			}
			if strings.HasPrefix(absPath, module.Base) {
				modulePaths = append(modulePaths, path)
			}
		}
		return modulePaths
	}

	// Default: use module base directory
	return []string{module.Base}
}

// getModuleOutputFile returns the output filename for a module in separate mode
func (p *Producer) getModuleOutputFile(module *goparser.GoModule) string {
	shortName := goparser.ModuleShortName(module)

	if p.outfile == "" {
		// Use module base with module name
		return filepath.Join(module.Base, shortName+".adoc")
	}

	// If outfile specified, use it as a base
	dir := filepath.Dir(p.outfile)
	ext := filepath.Ext(p.outfile)
	base := filepath.Base(p.outfile)
	base = base[:len(base)-len(ext)]

	return filepath.Join(dir, base+"-"+shortName+ext)
}

// generateSeparatePackages generates separate documentation files for each package
func (p *Producer) generateSeparatePackages() {
	overviewpaths := p.overviewpaths
	if len(overviewpaths) == 0 {
		overviewpaths = []string{
			"overview.adoc",
			"_design/overview.adoc",
		}
	}

	// Collect all packages from the module(s)
	packages, err := p.collectAllPackages()
	if err != nil {
		p.debugf("Generate: error collecting packages: %v", err)
		return
	}

	if len(packages) == 0 {
		p.debugf("Generate: no packages found")
		return
	}

	p.debugf("Generate: found %d package(s) to document", len(packages))

	// Track package files for master index
	var packageFiles []string
	packageInfoMap := make(map[string]*PackageInfo)

	// Process each package separately
	for i, pkg := range packages {
		p.debugf("Generate: processing package %s (%d/%d)", pkg.FqPackage, i+1, len(packages))

		// Create separate output file for this package
		pkgOutfile := p.getPackageOutputFile(pkg, i)
		p.debugf("Generate: writing package %s to %s", pkg.FqPackage, pkgOutfile)
		packageFiles = append(packageFiles, pkgOutfile)

		// Store package info for cross-referencing
		packageInfoMap[pkg.FqPackage] = &PackageInfo{
			Package:  pkg,
			Outfile:  pkgOutfile,
			Anchor:   fmt.Sprintf("pkg-%d", i+1),
			Index:    i,
		}

		// Save original settings
		origOutfile := p.outfile
		origWriter := p.writer
		origIndex := p.index

		// Set package-specific output
		p.outfile = pkgOutfile
		p.writer = nil // Force file creation
		p.index = true // Each package should have its header

		// Create template and writer for this package
		t := p.CreateTemplateWithOverrides()
		w := tabwriter.NewWriter(p.createWriter(), 4, 4, 4, ' ', 0)

		// Create package-specific parseconfig
		pkgConfig := p.parseconfig
		if pkgConfig.Module != nil {
			// Temporarily set module for this package
			pkgConfig.Module = pkg.Module
		}

		// Process just this package
		err := p.renderPackage(t, w, pkg, packageInfoMap, overviewpaths)

		w.Flush()

		if err != nil {
			p.debugf("Generate: error processing package %s: %v", pkg.FqPackage, err)
		}

		// Restore original settings
		p.outfile = origOutfile
		p.writer = origWriter
		p.index = origIndex
	}

	// Create master index file that includes/links all packages
	p.generatePackageMasterIndex(packageFiles, packageInfoMap)
}

// PackageInfo holds metadata about a package for cross-referencing
type PackageInfo struct {
	Package *goparser.GoPackage
	Outfile string
	Anchor  string
	Index   int
}

// collectAllPackages collects all packages from the configured module(s)
func (p *Producer) collectAllPackages() ([]*goparser.GoPackage, error) {
	var packages []*goparser.GoPackage
	collector := &packageCollector{packages: &packages}

	if p.parseconfig.Workspace != nil {
		// Multi-module workspace: collect from all modules
		for _, module := range p.parseconfig.Workspace.Modules {
			moduleConfig := p.parseconfig
			moduleConfig.Module = module

			modulePaths := p.getModulePaths(module)
			if len(modulePaths) == 0 {
				modulePaths = []string{module.Base}
			}

			err := goparser.ParseSinglePackageWalker(
				moduleConfig,
				collector.collectFunc,
				modulePaths...,
			)

			if err != nil {
				p.debugf("collectAllPackages: error processing module %s: %v", module.Name, err)
			}
		}
	} else {
		// Single module
		err := goparser.ParseSinglePackageWalker(
			p.parseconfig,
			collector.collectFunc,
			p.paths...,
		)

		if err != nil {
			return nil, err
		}
	}

	return packages, nil
}

// packageCollector is a helper to collect packages during parsing
type packageCollector struct {
	packages *[]*goparser.GoPackage
}

func (pc *packageCollector) collectFunc(pkg *goparser.GoPackage) error {
	*pc.packages = append(*pc.packages, pkg)
	return nil
}

// renderPackage renders a single package to its output file
func (p *Producer) renderPackage(t *Template, w io.Writer, pkg *goparser.GoPackage, packageInfoMap map[string]*PackageInfo, overviewpaths []string) error {
	if pkg == nil || len(pkg.Files) == 0 {
		return nil
	}

	// Build package references (internal and external)
	pkgRefs := p.buildPackageReferences(pkg, packageInfoMap)

	// Render using package-standalone template
	firstFile := pkg.Files[0]

	// Create index config
	indexConfig := &IndexConfig{
		Title:       "Package " + pkg.FqPackage,
		Highlighter: p.highlighter,
		TocLevels:   3,
		DocType:     "article",
		TocTitle:    "Table of Contents",
	}

	ctx := &TemplateContext{
		creator:     t,
		File:        firstFile,
		Package:     pkg,
		Module:      pkg.Module,
		Index:       indexConfig,
		PackageRefs: pkgRefs,
		Config: &TemplateContextConfig{
			Private:          p.private,
			TypeLinks:        p.typeLinks,
			SignatureStyle:   p.signatureStyle,
			RenderOptions:    p.renderOptions,
			PackageMode:      p.packageMode,
		},
	}

	// Render package header (standalone mode detected via .Index presence)
	if err := t.Templates[PackageTemplate.String()].Template.Execute(w, ctx); err != nil {
		return fmt.Errorf("error rendering package header: %v", err)
	}

	// Render package contents (imports, interfaces, structs, functions, etc.)
	for _, file := range pkg.Files {
		fileCtx := ctx.Clone(false)
		fileCtx.File = file
		fileCtx.Package = pkg

		// Render imports
		if len(file.Imports) > 0 {
			fileCtx.RenderImports(w)
		}

		// Render interfaces
		if len(file.Interfaces) > 0 {
			fileCtx.RenderInterfaces(w)
		}

		// Render structs
		if len(file.Structs) > 0 {
			fileCtx.RenderStructs(w)
		}

		// Render functions
		if len(file.StructMethods) > 0 {
			fileCtx.RenderFunctions(w)
		}

		// Render variables
		if len(file.VarAssignments) > 0 {
			fileCtx.RenderVarDeclarations(w)
		}

		// Render constants
		if len(file.ConstAssignments) > 0 {
			fileCtx.RenderConstDeclarations(w)
		}
	}

	// Render package references section
	if pkgRefs != nil && (len(pkgRefs.Internal) > 0 || len(pkgRefs.External) > 0) {
		if err := t.Templates[PackageRefsTemplate.String()].Template.Execute(w, ctx); err != nil {
			p.debugf("renderPackage: error rendering package references: %v", err)
		}
	}

	return nil
}

// getPackageOutputFile returns the output filename for a package
func (p *Producer) getPackageOutputFile(pkg *goparser.GoPackage, index int) string {
	// Sanitize package name for filename
	pkgName := strings.ReplaceAll(pkg.FqPackage, "/", "_")
	pkgName = strings.ReplaceAll(pkgName, "\\", "_")

	if p.outfile == "" {
		// Use package base with package name
		if pkg.FilePath != "" {
			return filepath.Join(filepath.Dir(pkg.FilePath), pkgName+".adoc")
		}
		return pkgName + ".adoc"
	}

	// If outfile specified, create folder structure relative to it
	dir := filepath.Dir(p.outfile)

	// Create subdirectory for packages
	pkgDir := filepath.Join(dir, "packages")

	return filepath.Join(pkgDir, pkgName+".adoc")
}

// buildPackageReferences builds internal and external package references
func (p *Producer) buildPackageReferences(pkg *goparser.GoPackage, packageInfoMap map[string]*PackageInfo) *PackageReferences {
	refs := &PackageReferences{
		Internal: []PackageRef{},
		External: []PackageRef{},
	}

	if pkg == nil {
		return refs
	}

	// Track unique imports
	importMap := make(map[string]bool)

	// Collect all imports from all files in the package
	for _, file := range pkg.Files {
		for _, imp := range file.Imports {
			if imp.Path == "" {
				continue
			}
			importMap[imp.Path] = true
		}
	}

	// Categorize imports as internal or external
	modulePath := ""
	if pkg.Module != nil {
		modulePath = pkg.Module.Name
	}

	for impPath := range importMap {
		// Check if it's an internal package
		if modulePath != "" && strings.HasPrefix(impPath, modulePath) {
			// Internal package
			if pkgInfo, found := packageInfoMap[impPath]; found {
				masterDir := filepath.Dir(p.outfile)
				relPath, err := filepath.Rel(masterDir, pkgInfo.Outfile)
				if err != nil {
					relPath = filepath.Base(pkgInfo.Outfile)
				}

				refs.Internal = append(refs.Internal, PackageRef{
					Name:   impPath,
					Anchor: pkgInfo.Anchor,
					File:   relPath,
				})
			}
		} else {
			// External package
			refs.External = append(refs.External, PackageRef{
				Name: impPath,
			})
		}
	}

	return refs
}

// generatePackageMasterIndex creates a master index file that includes/links all package files
func (p *Producer) generatePackageMasterIndex(packageFiles []string, packageInfoMap map[string]*PackageInfo) {
	if p.outfile == "" {
		// No master index file specified
		return
	}

	p.debugf("Generate: creating package master index file %s", p.outfile)

	// Create directory if needed
	dir := filepath.Dir(p.outfile)
	if !dirExists(dir) {
		os.MkdirAll(dir, os.ModePerm)
	}

	// Create master index file
	f, err := os.Create(p.outfile)
	if err != nil {
		p.debugf("Generate: error creating master index file: %v", err)
		return
	}
	defer f.Close()

	// Create template and context for rendering
	t := p.CreateTemplateWithOverrides()

	// Create index config
	indexConfig := &IndexConfig{
		Highlighter: p.highlighter,
		TocLevels:   3,
		DocType:     "book",
		TocTitle:    "Table of Contents",
	}

	// Determine project name
	if p.parseconfig.Module != nil {
		indexConfig.Title = p.parseconfig.Module.Name + " - Package Documentation"
	} else if p.parseconfig.Workspace != nil && len(p.parseconfig.Workspace.Modules) > 0 {
		indexConfig.Title = p.parseconfig.Workspace.Modules[0].Name + " - Package Documentation"
	} else {
		indexConfig.Title = "Package Documentation"
	}

	// Create template context for index header
	ctx := &TemplateContext{
		creator: t,
		Index:   indexConfig,
		Config: &TemplateContextConfig{
			PackageMode:        p.packageMode,
			PackageModeInclude: p.packageMode == PackageModeInclude,
		},
	}

	// Render index header
	if err := t.Templates[IndexTemplate.String()].Template.Execute(f, ctx); err != nil {
		p.debugf("Generate: error rendering index header: %v", err)
		return
	}

	// Create ordered list of packages
	packages := make([]*PackageInfo, len(packageFiles))
	for _, pkgInfo := range packageInfoMap {
		packages[pkgInfo.Index] = pkgInfo
	}

	// Include each package using package-ref template
	masterDir := filepath.Dir(p.outfile)
	for i, pkgInfo := range packages {
		if pkgInfo == nil {
			continue
		}

		// Calculate relative path from master file to package file
		relPath, err := filepath.Rel(masterDir, packageFiles[i])
		if err != nil {
			// Fallback to basename if relative path fails
			relPath = filepath.Base(packageFiles[i])
		}

		// Create context for package
		pkgCtx := &TemplateContext{
			creator:       t,
			Package:       pkgInfo.Package,
			PackageFile:   relPath,
			PackageAnchor: pkgInfo.Anchor,
			Config: &TemplateContextConfig{
				PackageMode:        p.packageMode,
				PackageModeInclude: p.packageMode == PackageModeInclude,
			},
		}

		// Execute package-ref template
		if err := t.Templates[PackageRefTemplate.String()].Template.ExecuteTemplate(f, "package-ref", pkgCtx); err != nil {
			p.debugf("Generate: error rendering package %s: %v", pkgInfo.Package.FqPackage, err)
			continue
		}

		fmt.Fprintf(f, "\n")
	}

	p.debugf("Generate: package master index file created with %d package includes/links", len(packageFiles))
}

// generateMasterIndex creates a master index file that includes all module files
func (p *Producer) generateMasterIndex(moduleFiles []string) {
	if p.outfile == "" {
		// No master index file specified
		return
	}

	p.debugf("Generate: creating master index file %s", p.outfile)

	// Create directory if needed
	dir := filepath.Dir(p.outfile)
	if !dirExists(dir) {
		os.MkdirAll(dir, os.ModePerm)
	}

	// Create master index file
	f, err := os.Create(p.outfile)
	if err != nil {
		p.debugf("Generate: error creating master index file: %v", err)
		return
	}
	defer f.Close()

	// Create template and context for rendering
	t := p.CreateTemplateWithOverrides()

	// Create index config
	indexConfig := &IndexConfig{
		Highlighter: p.highlighter,
		TocLevels:   3,
		DocType:     "book",
		TocTitle:    "Table of Contents",
	}

	// Determine workspace/project name
	if p.parseconfig.Workspace != nil && len(p.parseconfig.Workspace.Modules) > 0 {
		firstModule := p.parseconfig.Workspace.Modules[0]
		if firstModule != nil {
			indexConfig.Title = firstModule.Name
		}
	} else {
		indexConfig.Title = "Multi-Module Project"
	}

	// Create template context for index header
	ctx := &TemplateContext{
		creator:   t,
		Workspace: p.parseconfig.Workspace,
		Index:     indexConfig,
		Config: &TemplateContextConfig{
			ModuleModeInclude: false, // Don't use include mode for overview
		},
	}

	// Render index header with overview
	if err := t.Templates[IndexTemplate.String()].Template.Execute(f, ctx); err != nil {
		p.debugf("Generate: error rendering index header: %v", err)
		return
	}

	// Include each module using module template
	masterDir := filepath.Dir(p.outfile)
	for i, moduleFile := range moduleFiles {
		module := p.parseconfig.Workspace.Modules[i]

		// Calculate relative path from master file to module file
		relPath, err := filepath.Rel(masterDir, moduleFile)
		if err != nil {
			// Fallback to basename if relative path fails
			relPath = filepath.Base(moduleFile)
		}

		// Create context for module
		moduleCtx := &TemplateContext{
			creator:      t,
			Module:       module,
			ModuleFile:   relPath,
			ModuleAnchor: fmt.Sprintf("module-%d", i+1),
			Config: &TemplateContextConfig{
				ModuleModeInclude: true, // Use include mode
			},
		}

		// Execute module template
		if err := t.Templates[IndexTemplate.String()].Template.ExecuteTemplate(f, "module", moduleCtx); err != nil {
			p.debugf("Generate: error rendering module %s: %v", module.Name, err)
			continue
		}

		fmt.Fprintf(f, "\n")
	}

	p.debugf("Generate: master index file created with %d module includes", len(moduleFiles))
}

func (p *Producer) createWriter() io.Writer {

	if p.writer != nil {
		p.debugf("createWriter: using caller-supplied writer")
		return p.writer
	}

	if p.outfile == "" {
		p.outfile = filepath.Join(p.parseconfig.Module.Base, "docs.adoc")
	}

	dir := filepath.Dir(p.outfile)
	if !dirExists(dir) {
		os.MkdirAll(dir, os.ModePerm)
	}

	wr, err := os.Create(p.outfile)
	if err != nil {
		panic(err)
	}

	p.debugf("createWriter: writing output to %s", p.outfile)
	return wr

}

// getProcessFunc will return the function to feed the template system to generate the documentation output.
//
// If `p.macro` is set to `true`, it will wrap the document template / writer function with a macro substitution
// function using function chaining.
func (p *Producer) getProcessFunc(
	t *Template,
	w io.Writer,
	indexdone bool,
	overviewpaths []string) goparser.ParseSinglePackageWalkerFunc {

	processor := func(pkg *goparser.GoPackage) error {

		p.debugf("Render: package %s (%d file(s))", pkg.Package, len(pkg.Files))

		tc := t.NewContextWithConfig(&pkg.GoFile, pkg, &TemplateContextConfig{
			IncludeMethodCode:    false,
			PackageOverviewPaths: overviewpaths,
			Private:              p.private,
			TypeLinks:            p.typeLinks,
			SignatureStyle:       p.signatureStyle,
			RenderOptions:        p.renderOptions,
			SubModuleMode:        p.subModuleMode,
		})

		// Set workspace if available
		if p.parseconfig.Workspace != nil {
			tc.Workspace = p.parseconfig.Workspace
		}

		if !indexdone {

			p.debugf("Render: emitting index for package %s", pkg.Package)

			ic := tc.DefaultIndexConfig(p.indexconfig)
			if p.highlighter != "" {
				ic.Highlighter = p.highlighter
			}
			if !p.toc {
				ic.TocTitle = "" // disables toc generation
			}

			tc.RenderIndex(w, ic)
			indexdone = true
		}

		tc.RenderPackage(w)

		if len(pkg.Imports) > 0 {
			p.debugf("Render: package %s imports section", pkg.Package)
			tc.RenderImports(w)
		}
		if len(pkg.Interfaces) > 0 {
			p.debugf("Render: package %s interfaces section", pkg.Package)
			tc.RenderInterfaces(w)
		}
		if len(pkg.Structs) > 0 {
			p.debugf("Render: package %s structs section", pkg.Package)
			tc.RenderStructs(w)
		}
		if len(pkg.CustomTypes) > 0 {
			p.debugf("Render: package %s custom type definitions", pkg.Package)
			tc.RenderVarTypeDefs(w)
		}
		if len(pkg.ConstAssignments) > 0 {
			p.debugf("Render: package %s const declarations", pkg.Package)
			tc.RenderConstDeclarations(w)
		}
		if len(pkg.CustomFuncs) > 0 {
			p.debugf("Render: package %s custom function definitions", pkg.Package)
			tc.RenderTypeDefFuncs(w)
		}
		if len(pkg.VarAssignments) > 0 {
			p.debugf("Render: package %s variable declarations", pkg.Package)
			tc.RenderVarDeclarations(w)
		}
		if len(pkg.StructMethods) > 0 {
			p.debugf("Render: package %s struct methods", pkg.Package)
			tc.RenderFunctions(w)
		}

		return nil
	}

	if p.macro {

		return getProcessMacroFunc(p.parseconfig, processor)

	}

	return processor
}
