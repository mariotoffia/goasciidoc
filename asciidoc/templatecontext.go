package asciidoc

import (
	"io"

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
}

// TemplateContextConfig contains configuration parameters how templates
// renders the content and the TemplateContexts behaves.
type TemplateContextConfig struct {
	// IncludeMethodCode determines if the code is included in the documentation or not.
	// Default not included.
	IncludeMethodCode bool
}

// Clone will clone the context.
func (t *TemplateContext) Clone(clean bool) *TemplateContext {

	if clean {

		return &TemplateContext{
			creator: t.creator,
			File:    t.File,
			Config:  t.Config,
		}

	}

	return &TemplateContext{
		creator:         t.creator,
		File:            t.File,
		Struct:          t.Struct,
		Function:        t.Function,
		Interface:       t.Interface,
		TypeDefVar:      t.TypeDefVar,
		TypeDefFunc:     t.TypeDefFunc,
		VarAssignment:   t.VarAssignment,
		ConstAssignment: t.ConstAssignment,
		Config:          t.Config,
	}
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
