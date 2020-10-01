package asciidoc

import (
	"encoding/json"
	"io"
	"os/user"

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
	// TypedefFun is current function type defintion.
	TypeDefFunc *goparser.GoMethod
	// VarAssignment is current variable assignment using var keyword
	VarAssignment *goparser.GoAssignment
	// ConstAssignment is current const definition and value assignment
	ConstAssignment *goparser.GoAssignment
	// Config contains the configuration of this context.
	Config *TemplateContextConfig
	// Index is configuration to render the index template
	Index *IndexConfig
}

// TemplateContextConfig contains configuration parameters how templates
// renders the content and the TemplateContexts behaves.
type TemplateContextConfig struct {
	// IncludeMethodCode determines if the code is included in the documentation or not.
	// Default not included.
	IncludeMethodCode bool
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
			creator: t.creator,
			File:    t.File,
			Module:  t.Module,
			Config:  t.Config,
		}

	}

	return &TemplateContext{
		creator:         t.creator,
		File:            t.File,
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

	if "" != overrides {

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

// RenderPackage will render the package defintion onto the provided writer.
func (t *TemplateContext) RenderPackage(wr io.Writer) *TemplateContext {

	if err := t.creator.Templates[PackageTemplate.String()].Execute(wr, t.Clone(true /*clean*/)); nil != err {
		panic(err)
	}

	return t
}

// RenderImports will render the imports section onto the provided writer.
func (t *TemplateContext) RenderImports(wr io.Writer) *TemplateContext {

	if err := t.creator.Templates[ImportTemplate.String()].Execute(wr, t.Clone(true /*clean*/)); nil != err {
		panic(err)
	}

	return t
}

// RenderFunctions will render all functions for GoFile/GoPackage onto the provided writer.
func (t *TemplateContext) RenderFunctions(wr io.Writer) *TemplateContext {

	if err := t.creator.Templates[FunctionsTemplate.String()].Execute(wr, t.Clone(true /*clean*/)); nil != err {
		panic(err)
	}

	return t
}

// RenderFunction will render a single function section onto the provided writer.
func (t *TemplateContext) RenderFunction(wr io.Writer, f *goparser.GoStructMethod) *TemplateContext {

	q := t.Clone(true /*clean*/)
	q.Function = f

	if err := t.creator.Templates[FunctionTemplate.String()].Execute(wr, q); nil != err {
		panic(err)
	}

	return t
}

// RenderInterfaces will render all interfaces for GoFile/GoPackage onto the provided writer.
func (t *TemplateContext) RenderInterfaces(wr io.Writer) *TemplateContext {

	if err := t.creator.Templates[InterfacesTemplate.String()].Execute(wr, t.Clone(true /*clean*/)); nil != err {
		panic(err)
	}

	return t
}

// RenderInterface will render a single interface section onto the provided writer.
func (t *TemplateContext) RenderInterface(wr io.Writer, i *goparser.GoInterface) *TemplateContext {

	q := t.Clone(true /*clean*/)
	q.Interface = i

	if err := t.creator.Templates[InterfaceTemplate.String()].Execute(wr, q); nil != err {
		panic(err)
	}

	return t
}

// RenderStructs will render all structs for GoFile/GoPackage onto the provided writer.
func (t *TemplateContext) RenderStructs(wr io.Writer) *TemplateContext {

	if err := t.creator.Templates[StructsTemplate.String()].Execute(wr, t.Clone(true /*clean*/)); nil != err {
		panic(err)
	}

	return t
}

// RenderStruct will render a single struct section onto the provided writer.
func (t *TemplateContext) RenderStruct(wr io.Writer, s *goparser.GoStruct) *TemplateContext {

	q := t.Clone(true /*clean*/)
	q.Struct = s

	if err := t.creator.Templates[StructTemplate.String()].Execute(wr, q); nil != err {
		panic(err)
	}

	return t
}

// RenderVarTypeDefs will render all variable type definitions for GoFile/GoPackage onto the provided writer.
func (t *TemplateContext) RenderVarTypeDefs(wr io.Writer) *TemplateContext {

	if err := t.creator.Templates[CustomVarTypeDefsTemplate.String()].Execute(wr, t.Clone(true /*clean*/)); nil != err {
		panic(err)
	}

	return t
}

// RenderVarTypeDef will render a single variable typedef section onto the provided writer.
func (t *TemplateContext) RenderVarTypeDef(wr io.Writer, td *goparser.GoCustomType) *TemplateContext {

	q := t.Clone(true /*clean*/)
	q.TypeDefVar = td

	if err := t.creator.Templates[CustomVarTypeDefTemplate.String()].Execute(wr, q); nil != err {
		panic(err)
	}

	return t
}

// RenderVarDeclarations will render all variable declarations for GoFile/GoPackage onto the provided writer.
func (t *TemplateContext) RenderVarDeclarations(wr io.Writer) *TemplateContext {

	if err := t.creator.Templates[VarDeclarationsTemplate.String()].Execute(wr, t.Clone(true /*clean*/)); nil != err {
		panic(err)
	}

	return t
}

// RenderVarDeclaration will render a single variable declaration section onto the provided writer.
func (t *TemplateContext) RenderVarDeclaration(wr io.Writer, a *goparser.GoAssignment) *TemplateContext {

	q := t.Clone(true /*clean*/)
	q.VarAssignment = a

	if err := t.creator.Templates[VarDeclarationTemplate.String()].Execute(wr, q); nil != err {
		panic(err)
	}

	return t
}

// RenderConstDeclarations will render all const declarations for GoFile/GoPackage onto the provided writer.
func (t *TemplateContext) RenderConstDeclarations(wr io.Writer) *TemplateContext {

	if err := t.creator.Templates[ConstDeclarationsTemplate.String()].Execute(wr, t.Clone(true /*clean*/)); nil != err {
		panic(err)
	}

	return t
}

// RenderConstDeclaration will render a single const declaration section onto the provided writer.
func (t *TemplateContext) RenderConstDeclaration(wr io.Writer, a *goparser.GoAssignment) *TemplateContext {

	q := t.Clone(true /*clean*/)
	q.ConstAssignment = a

	if err := t.creator.Templates[ConstDeclarationTemplate.String()].Execute(wr, q); nil != err {
		panic(err)
	}

	return t
}

// RenderTypeDefFuncs will render all type definitions for GoFile/GoPackage onto the provided writer.
func (t *TemplateContext) RenderTypeDefFuncs(wr io.Writer) *TemplateContext {

	if err := t.creator.Templates[CustomFuncTypeDefsTemplate.String()].Execute(wr, t.Clone(true /*clean*/)); nil != err {
		panic(err)
	}

	return t
}

// RenderTypeDefFunc will render a single typedef section onto the provided writer.
func (t *TemplateContext) RenderTypeDefFunc(wr io.Writer, td *goparser.GoMethod) *TemplateContext {

	q := t.Clone(true /*clean*/)
	q.TypeDefFunc = td

	if err := t.creator.Templates[CustomFuncTypeDefTemplate.String()].Execute(wr, q); nil != err {
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

	if err := t.creator.Templates[IndexTemplate.String()].Execute(wr, q); nil != err {
		panic(err)
	}

	return t
}
