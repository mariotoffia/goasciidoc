package asciidoc

import (
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

	// Process each module separately
	for i, module := range p.parseconfig.Workspace.Modules {
		p.debugf("Generate: processing module %s (%d/%d) as separate file", module.Name, i+1, len(p.parseconfig.Workspace.Modules))

		// Create separate output file for this module
		moduleOutfile := p.getModuleOutputFile(module)
		p.debugf("Generate: writing module %s to %s", module.Name, moduleOutfile)

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
