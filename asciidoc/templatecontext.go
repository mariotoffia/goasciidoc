package asciidoc

import (
	"encoding/json"
	"io"
	"os/user"
	"path/filepath"

	"github.com/mariotoffia/goasciidoc/goparser"
)

// TemplateContext is a context that may be used to render
// a GoFile. Depending on the template, different fields are
// populated in this struct.
type TemplateContext struct {
	// creator is the template created this context.
	creator *Template
	// File is the complete file. This property is always present.
	//
	// For package and imports, this is the only one to access
	File *goparser.GoFile
	// Package where the `File` resides under. Most of the time
	// is `Package` and `File` the same since rendering is done
	// on package level.
	Package *goparser.GoPackage
	// Module for the context
	Module *goparser.GoModule
	// Struct is the current GoStruct
	Struct *goparser.GoStruct
	// Function is the current function
	Function *goparser.GoStructMethod
	// Interface is the current GoInterface
	Interface *goparser.GoInterface
	// TypeDefVar is current variable type definition
	TypeDefVar *goparser.GoCustomType
	// TypedefFun is current function type definition.
	TypeDefFunc *goparser.GoMethod
	// VarAssignment is current variable assignment using var keyword
	VarAssignment *goparser.GoAssignment
	// ConstAssignment is current const definition and value assignment
	ConstAssignment *goparser.GoAssignment
	// Config contains the configuration of this context.
	Config *TemplateContextConfig
	// Index is configuration to render the index template
	Index *IndexConfig
	// Receiver is the current receivers to be rendered.
	Receiver []*goparser.GoStructMethod
	// Docs is a map that contains filepaths to various asciidoc documents
	// that can be included.
	//
	// .Available Documents
	// |===
	// |Name |Comment
	//
	// |package-overview
	// |This is a absolute path to a overview document for the current package.
	//
	// |===
	Docs map[string]string
	// importCache caches import alias lookups per file for linking.
	importCache map[*goparser.GoFile]map[string]string
}

// TemplateContextConfig contains configuration parameters how templates
// renders the content and the TemplateContexts behaves.
type TemplateContextConfig struct {
	// IncludeMethodCode determines if the code is included in the documentation or not.
	// Default not included.
	IncludeMethodCode bool
	// PackageOverviewPaths paths to search for package overview relative the package path.
	//
	// It searches the order as they appear in this array until found, then terminates. It is
	// not possible to have two _*.adoc_ inclusions.
	//
	// .Example Paths
	// |===
	// |Example |Comment
	//
	// |overview.adoc
	// |This expects the overview.adoc to be in the same folders as the other go files in the package.
	//
	// |_design/package-summary.adoc
	// |This tells the renderer to look for _package-summary.adoc_ in _package path/_design_ folder.
	//
	// |===
	PackageOverviewPaths []string
	// Private indicates if it shall include private as well. By default only Exported is rendered.
	Private bool
	// TypeLinks determines how type references are rendered.
	TypeLinks TypeLinkMode
}

// IndexConfig is configuration to use when generating index template
type IndexConfig struct {
	// Title is the title of the index document, if omitted it uses the module name (if present)
	Title string `json:"title,omitempty"`
	// Version is the version stamped as version attribute, if omitted it uses module version (if any)
	Version string `json:"version,omitempty"`
	// AuthorName is the full name of the author e.g. Mario Toffia (if none is set, default to current user)
	AuthorName string `json:"author,omitempty"`
	// AuthorEmail is the email of the author e.g. mario.toffia@bullen.se
	AuthorEmail string `json:"email,omitempty"`
	// Highlighter is the source highlighter to use - default is 'highlightjs'
	Highlighter string `json:"highlight,omitempty"`
	// TocTitle is the title of the generated table of contents (if set a toc is generated)
	// Default is 'Table of Contents', hence by default a TOC is generated.
	TocTitle string `json:"toc,omitempty"`
	// TocLevels determines how many levels shall it include, default 3
	TocLevels int `json:"toclevel,omitempty"`
	// A fully qualified or relative output path to where to search for images
	ImageDir string `json:"images,omitempty"`
	// HomePage is the url to homepage
	HomePage string `json:"web,omitempty"`
	// DocType determines the document type, default is book
	DocType string `json:"doctype,omitempty"`
}

// Clone will clone the context.
func (t *TemplateContext) Clone(clean bool) *TemplateContext {

	if clean {

		return &TemplateContext{
			creator:     t.creator,
			Package:     t.Package,
			File:        t.File,
			Module:      t.Module,
			Config:      t.Config,
			Docs:        t.Docs,
			importCache: t.importCache,
		}

	}

	return &TemplateContext{
		creator:         t.creator,
		File:            t.File,
		Package:         t.Package,
		Module:          t.Module,
		Struct:          t.Struct,
		Function:        t.Function,
		Interface:       t.Interface,
		TypeDefVar:      t.TypeDefVar,
		TypeDefFunc:     t.TypeDefFunc,
		VarAssignment:   t.VarAssignment,
		ConstAssignment: t.ConstAssignment,
		Config:          t.Config,
		Index:           t.Index,
		Docs:            t.Docs,
		Receiver:        t.Receiver,
		importCache:     t.importCache,
	}
}

// DefaultIndexConfig creates a default index configuration that may be used in RenderIndex
// function.
//
// The overrides are specifies as a json document, only properties set in the JSON document will
// override default IndexConfig.
func (t *TemplateContext) DefaultIndexConfig(overrides string) *IndexConfig {

	ic := &IndexConfig{
		Highlighter: "highlightjs",
		TocLevels:   3,
		DocType:     "book",
		TocTitle:    "Table of Contents",
	}

	if t.Module != nil {
		ic.Title = t.Module.Name
		ic.Version = t.Module.Version
	}

	user, err := user.Current()
	if err != nil {
		panic(err)
	}

	ic.AuthorName = user.Username

	if overrides != "" {

		if err := json.Unmarshal([]byte(overrides), ic); err != nil {
			panic(err)
		}

	}

	return ic
}

// Creator returns the template created this context.
func (t *TemplateContext) Creator() *Template {
	return t.creator
}

// RenderPrivate will enable non exported to be rendered.
func (t *TemplateContext) RenderPrivate() *TemplateContext {

	if nil == t.Config {
		panic("Config is nil while trying to configure it!")
	}

	t.Config.Private = true
	return t
}

// RenderPackage will render the package defintion onto the provided writer.
//
// Depending on if a package overview asciidoc document is found it will prioritize that before
// the go package documentation. Hence it will use either _PackageTemplate_ or
// _PackageIncludeOverviewTemplate_ depending if found a ascii doc overview document.
func (t *TemplateContext) RenderPackage(wr io.Writer) *TemplateContext {

	fp := t.resolvePackageOverview()
	if fp != "" {
		t.Docs["package-overview"] = fp
	}

	if err := t.creator.Templates[PackageTemplate.String()].Template.Execute(wr, t.Clone(true /*clean*/)); nil != err {
		panic(err)
	}

	return t
}

// RenderImports will render the imports section onto the provided writer.
func (t *TemplateContext) RenderImports(wr io.Writer) *TemplateContext {

	if err := t.creator.Templates[ImportTemplate.String()].Template.Execute(wr, t.Clone(true /*clean*/)); nil != err {
		panic(err)
	}

	return t
}

// RenderFunctions will render all functions for GoFile/GoPackage onto the provided writer.
func (t *TemplateContext) RenderFunctions(wr io.Writer) *TemplateContext {

	if err := t.creator.Templates[FunctionsTemplate.String()].Template.Execute(wr, t.Clone(true /*clean*/)); nil != err {
		panic(err)
	}

	return t
}

// RenderReceiverFunctions will render all receiver functions for a given receiver, albeit a custom type or a struct.
func (t *TemplateContext) RenderReceiverFunctions(wr io.Writer, receiver string) *TemplateContext {

	q := t.Clone(true /*clean*/)

	if t.Package == nil {
		q.Receiver = t.File.FindMethodsByReceiver(receiver)
	} else {
		q.Receiver = t.Package.FindMethodsByReceiver(receiver)
	}

	if err := t.creator.Templates[ReceiversTemplate.String()].Template.Execute(wr, q); nil != err {
		panic(err)
	}

	return t
}

// RenderFunction will render a single function section onto the provided writer.
func (t *TemplateContext) RenderFunction(wr io.Writer, f *goparser.GoStructMethod) *TemplateContext {

	q := t.Clone(true /*clean*/)
	q.Function = f

	if err := t.creator.Templates[FunctionTemplate.String()].Template.Execute(wr, q); nil != err {
		panic(err)
	}

	return t
}

// RenderInterfaces will render all interfaces for GoFile/GoPackage onto the provided writer.
func (t *TemplateContext) RenderInterfaces(wr io.Writer) *TemplateContext {

	if err := t.creator.Templates[InterfacesTemplate.String()].Template.Execute(wr, t.Clone(true /*clean*/)); nil != err {
		panic(err)
	}

	return t
}

// RenderInterface will render a single interface section onto the provided writer.
func (t *TemplateContext) RenderInterface(wr io.Writer, i *goparser.GoInterface) *TemplateContext {

	q := t.Clone(true /*clean*/)
	q.Interface = i

	if err := t.creator.Templates[InterfaceTemplate.String()].Template.Execute(wr, q); nil != err {
		panic(err)
	}

	return t
}

// RenderStructs will render all structs for GoFile/GoPackage onto the provided writer.
func (t *TemplateContext) RenderStructs(wr io.Writer) *TemplateContext {

	if err := t.creator.Templates[StructsTemplate.String()].Template.Execute(wr, t.Clone(true /*clean*/)); nil != err {
		panic(err)
	}

	return t
}

// RenderStruct will render a single struct section onto the provided writer.
func (t *TemplateContext) RenderStruct(wr io.Writer, s *goparser.GoStruct) *TemplateContext {

	q := t.Clone(true /*clean*/)
	q.Struct = s

	if err := t.creator.Templates[StructTemplate.String()].Template.Execute(wr, q); nil != err {
		panic(err)
	}

	return t
}

// RenderVarTypeDefs will render all variable type definitions for GoFile/GoPackage onto the provided writer.
func (t *TemplateContext) RenderVarTypeDefs(wr io.Writer) *TemplateContext {

	if err := t.creator.Templates[CustomVarTypeDefsTemplate.String()].Template.Execute(wr, t.Clone(true /*clean*/)); nil != err {
		panic(err)
	}

	return t
}

// RenderVarTypeDef will render a single variable typedef section onto the provided writer.
func (t *TemplateContext) RenderVarTypeDef(wr io.Writer, td *goparser.GoCustomType) *TemplateContext {

	q := t.Clone(true /*clean*/)
	q.TypeDefVar = td

	if err := t.creator.Templates[CustomVarTypeDefTemplate.String()].Template.Execute(wr, q); nil != err {
		panic(err)
	}

	return t
}

// RenderVarDeclarations will render all variable declarations for GoFile/GoPackage onto the provided writer.
func (t *TemplateContext) RenderVarDeclarations(wr io.Writer) *TemplateContext {

	if err := t.creator.Templates[VarDeclarationsTemplate.String()].Template.Execute(wr, t.Clone(true /*clean*/)); nil != err {
		panic(err)
	}

	return t
}

// RenderVarDeclaration will render a single variable declaration section onto the provided writer.
func (t *TemplateContext) RenderVarDeclaration(wr io.Writer, a *goparser.GoAssignment) *TemplateContext {

	q := t.Clone(true /*clean*/)
	q.VarAssignment = a

	if err := t.creator.Templates[VarDeclarationTemplate.String()].Template.Execute(wr, q); nil != err {
		panic(err)
	}

	return t
}

// RenderConstDeclarations will render all const declarations for GoFile/GoPackage onto the provided writer.
func (t *TemplateContext) RenderConstDeclarations(wr io.Writer) *TemplateContext {

	if err := t.creator.Templates[ConstDeclarationsTemplate.String()].Template.Execute(wr, t.Clone(true /*clean*/)); nil != err {
		panic(err)
	}

	return t
}

// RenderConstDeclaration will render a single const declaration section onto the provided writer.
func (t *TemplateContext) RenderConstDeclaration(wr io.Writer, a *goparser.GoAssignment) *TemplateContext {

	q := t.Clone(true /*clean*/)
	q.ConstAssignment = a

	if err := t.creator.Templates[ConstDeclarationTemplate.String()].Template.Execute(wr, q); nil != err {
		panic(err)
	}

	return t
}

// RenderTypeDefFuncs will render all type definitions for GoFile/GoPackage onto the provided writer.
func (t *TemplateContext) RenderTypeDefFuncs(wr io.Writer) *TemplateContext {

	if err := t.creator.Templates[CustomFuncTypeDefsTemplate.String()].Template.Execute(wr, t.Clone(true /*clean*/)); nil != err {
		panic(err)
	}

	return t
}

// RenderTypeDefFunc will render a single typedef section onto the provided writer.
func (t *TemplateContext) RenderTypeDefFunc(wr io.Writer, td *goparser.GoMethod) *TemplateContext {

	q := t.Clone(true /*clean*/)
	q.TypeDefFunc = td

	if err := t.creator.Templates[CustomFuncTypeDefTemplate.String()].Template.Execute(wr, q); nil != err {
		panic(err)
	}

	return t
}

// RenderIndex will render the complete index page for all GoFiles/GoPackages onto the provided writer.
//
// If nil is provided as IndexConfig it will use the default config.
func (t *TemplateContext) RenderIndex(wr io.Writer, ic *IndexConfig) *TemplateContext {

	if nil == ic {
		ic = t.DefaultIndexConfig("")
	}

	q := t.Clone(true /*clean*/)
	q.Index = ic

	if err := t.creator.Templates[IndexTemplate.String()].Template.Execute(wr, q); nil != err {
		panic(err)
	}

	return t
}

// resolvePackageOverview will search the list of inclusion try to resolve any file and return the filepath.
//
// If it fails, an empty string is returned. This uses the _TemplateConfig.PackageOverviewPaths_
// list to resolve the data. The first hit of the absolute filepath will be returned.
func (t *TemplateContext) resolvePackageOverview() string {

	if len(t.Config.PackageOverviewPaths) == 0 {
		return ""
	}

	base := t.File.FilePath

	for _, p := range t.Config.PackageOverviewPaths {

		var fp string
		if base == "" {
			fp = p
		} else {
			fp = filepath.Join(base, p)
		}

		if fileExists(fp) {
			afp, err := filepath.Abs(fp)
			if err != nil {
				panic(err)
			}

			return afp
		}
	}

	return ""
}
