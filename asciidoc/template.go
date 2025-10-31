package asciidoc

import (
	"bytes"
	"strings"
	"text/template"

	"github.com/mariotoffia/goasciidoc/goparser"
)

// TemplateType specifies the template type
type TemplateType string

const (
	// IndexTemplate is a template that binds all generated asciidoc files into one single index file
	// by referencing (or appending to this file).
	IndexTemplate TemplateType = "index"
	// PackageTemplate specifies that the template is a package
	PackageTemplate TemplateType = "package"
	// ImportTemplate specifies that the template renders a import
	ImportTemplate TemplateType = "import"
	// FunctionsTemplate is a template to render all functions for a given context (package, file)
	FunctionsTemplate TemplateType = "functions"
	// FunctionTemplate is a template to render a function
	FunctionTemplate TemplateType = "function"
	// InterfacesTemplate is a template to render a all interface defintions for a given context (package, file)
	InterfacesTemplate TemplateType = "interfaces"
	// InterfaceTemplate is a template to render a interface definition
	InterfaceTemplate TemplateType = "interface"
	// StructsTemplate specifies that the template renders all struct definitions for a given context (package, file)
	StructsTemplate TemplateType = "structs"
	// StructTemplate specifies that the template renders a struct definition
	StructTemplate TemplateType = "struct"
	// CustomVarTypeDefsTemplate is a template to render all variable type definitions for a given context (package, file)
	CustomVarTypeDefsTemplate TemplateType = "typedefvars"
	// CustomVarTypeDefTemplate is a template to render a type definition of a variable
	CustomVarTypeDefTemplate TemplateType = "typedefvar"
	// CustomFuncTypeDefsTemplate is a template to render all function type definitions for a given context (package, file)
	CustomFuncTypeDefsTemplate TemplateType = "typedeffuncs"
	// CustomFuncTypeDefTemplate is a template to render a function type definition
	CustomFuncTypeDefTemplate TemplateType = "typedeffunc"
	// VarDeclarationsTemplate is a template to render all variable definitions for a given context (package, file)
	VarDeclarationsTemplate TemplateType = "vars"
	// VarDeclarationTemplate is a template to render a variable definition
	VarDeclarationTemplate TemplateType = "var"
	// ConstDeclarationsTemplate is a template to render all const declaration entries for a given context (package, file)
	ConstDeclarationsTemplate TemplateType = "consts"
	// ConstDeclarationTemplate is a template to render a const declaration entry
	ConstDeclarationTemplate TemplateType = "const"
	// ReceiversTemplate is a template that renders receivers functions
	ReceiversTemplate TemplateType = "receivers"
)

func (tt TemplateType) String() string {
	return string(tt)
}

var defaultTemplateFuncs = template.FuncMap{
	"typeParams":         typeParamsSuffix,
	"nameWithTypeParams": nameWithTypeParams,
	"indent":             func(s string) string { return indent(s) },
	"typeSetItems":       typeSetItems,
}

// TemplateAndText is a wrapper of _template.Template_
// but also includes the original text representation
// of the template and not just the parsed tree.
type TemplateAndText struct {
	// Text is the actual template that got parsed by _template.Template_.
	Text string
	// Template is the instance of the parsed _Text_ including functions.
	Template *template.Template
}

// Template is handling all templates and actions
// to perform.
type Template struct {
	// Templates to use when rendering documentation
	Templates map[string]*TemplateAndText
}

// NewTemplateWithOverrides creates a new template with the ability to easily
// override defaults.
func NewTemplateWithOverrides(overrides map[string]string) *Template {

	return &Template{
		Templates: map[string]*TemplateAndText{
			IndexTemplate.String():   createTemplate(IndexTemplate, "", overrides, template.FuncMap{}),
			PackageTemplate.String(): createTemplate(PackageTemplate, "", overrides, template.FuncMap{}),
			ImportTemplate.String(): createTemplate(ImportTemplate, "", overrides, template.FuncMap{
				"render": func(t *TemplateContext) string { return t.File.DeclImports() },
			}),
			FunctionsTemplate.String(): createTemplate(FunctionsTemplate, "", overrides, template.FuncMap{
				"render": func(t *TemplateContext, f *goparser.GoStructMethod) string {
					var buf bytes.Buffer
					t.RenderFunction(&buf, f)
					return buf.String()
				},
				"notreceiver": func(t *TemplateContext, f *goparser.GoStructMethod) bool {
					return len(f.Receivers) == 0
				},
			}),
			FunctionTemplate.String(): createTemplate(FunctionTemplate, "", overrides, template.FuncMap{}),
			InterfacesTemplate.String(): createTemplate(InterfacesTemplate, "", overrides, template.FuncMap{
				"render": func(t *TemplateContext, i *goparser.GoInterface) string {
					var buf bytes.Buffer
					t.RenderInterface(&buf, i)
					return buf.String()
				},
			}),
			InterfaceTemplate.String(): createTemplate(InterfaceTemplate, "", overrides, template.FuncMap{
				"tabifylast": func(decl string) string {
					idx := strings.LastIndex(decl, " ")
					if -1 == idx {
						return decl
					}
					return decl[:idx] + "\t" + decl[idx+1:]
				},
			}),
			StructsTemplate.String(): createTemplate(StructsTemplate, "", overrides, template.FuncMap{
				"render": func(t *TemplateContext, s *goparser.GoStruct) string {
					var buf bytes.Buffer
					t.RenderStruct(&buf, s)
					return buf.String()
				},
			}),
			StructTemplate.String(): createTemplate(StructTemplate, "", overrides, template.FuncMap{
				"tabify": func(decl string) string { return strings.Replace(decl, " ", "\t", 1) },
				"render": func(t *TemplateContext, s *goparser.GoStruct) string {
					var buf bytes.Buffer
					t.RenderStruct(&buf, s)
					return buf.String()
				},
				"renderReceivers": func(t *TemplateContext, receiver string) string {
					var buf bytes.Buffer
					t.RenderReceiverFunctions(&buf, receiver)
					return buf.String()
				},
				"hasReceivers": func(t *TemplateContext, receiver string) bool {
					if nil != t.Package {
						return len(t.Package.FindMethodsByReceiver(receiver)) > 0
					}
					return len(t.File.FindMethodsByReceiver(receiver)) > 0
				},
			}),
			ReceiversTemplate.String(): createTemplate(ReceiversTemplate, "", overrides, template.FuncMap{}),
			CustomVarTypeDefsTemplate.String(): createTemplate(CustomVarTypeDefsTemplate, "", overrides, template.FuncMap{
				"render": func(t *TemplateContext, td *goparser.GoCustomType) string {
					var buf bytes.Buffer
					t.RenderVarTypeDef(&buf, td)
					return buf.String()
				},
			}),
			CustomVarTypeDefTemplate.String(): createTemplate(CustomVarTypeDefTemplate, "", overrides, template.FuncMap{
				"renderReceivers": func(t *TemplateContext, receiver string) string {
					var buf bytes.Buffer
					t.RenderReceiverFunctions(&buf, receiver)
					return buf.String()
				},
				"hasReceivers": func(t *TemplateContext, receiver string) bool {
					if nil != t.Package {
						return len(t.Package.FindMethodsByReceiver(receiver)) > 0
					}
					return len(t.File.FindMethodsByReceiver(receiver)) > 0
				},
			}),
			VarDeclarationsTemplate.String(): createTemplate(VarDeclarationsTemplate, "", overrides, template.FuncMap{
				"render": func(t *TemplateContext, a *goparser.GoAssignment) string {
					var buf bytes.Buffer
					t.RenderVarDeclaration(&buf, a)
					return buf.String()
				},
			}),
			VarDeclarationTemplate.String(): createTemplate(VarDeclarationTemplate, "", overrides, template.FuncMap{}),
			ConstDeclarationsTemplate.String(): createTemplate(ConstDeclarationsTemplate, "", overrides, template.FuncMap{
				"tabify": func(decl string) string { return strings.Replace(decl, " ", "\t", 1) },
				"render": func(t *TemplateContext, a *goparser.GoAssignment) string {
					var buf bytes.Buffer
					t.RenderConstDeclaration(&buf, a)
					return buf.String()
				},
			}),
			ConstDeclarationTemplate.String(): createTemplate(ConstDeclarationTemplate, "", overrides, template.FuncMap{}),
			CustomFuncTypeDefsTemplate.String(): createTemplate(CustomFuncTypeDefsTemplate, "", overrides, template.FuncMap{
				"render": func(t *TemplateContext, td *goparser.GoMethod) string {
					var buf bytes.Buffer
					t.RenderTypeDefFunc(&buf, td)
					return buf.String()
				},
			}),
			CustomFuncTypeDefTemplate.String(): createTemplate(CustomFuncTypeDefTemplate, "", overrides, template.FuncMap{}),
		},
	}

}

// NewContext creates a new context to be used for rendering.
func (t *Template) NewContext(f *goparser.GoFile) *TemplateContext {
	return t.NewContextWithConfig(f, nil, &TemplateContextConfig{})
}

// NewContextWithConfig creates a new context with configuration.
//
// If configuration is nil, it will use default configuration.
func (t *Template) NewContextWithConfig(
	f *goparser.GoFile,
	p *goparser.GoPackage,
	config *TemplateContextConfig) *TemplateContext {

	if nil == config {
		config = &TemplateContextConfig{}
	}

	tc := &TemplateContext{
		creator: t,
		File:    f,
		Package: p,
		Module:  f.Module,
		Config:  config,
		Docs:    map[string]string{},
	}

	return tc
}

// createTemplate will create a template named name and parses the str
// as template. If fails it will panic with the parse error.
//
// If name is found in override map it will use that string to parse the template
// instead of the provided str.
func createTemplate(name TemplateType, str string, overrides map[string]string, fm template.FuncMap) *TemplateAndText {

	if s, ok := overrides[name.String()]; ok {
		str = s
	}

	pt, err := template.New(name.String()).Funcs(defaultTemplateFuncs).Funcs(fm).Parse(str)
	if err != nil {
		panic(err)
	}
	return &TemplateAndText{
		Text:     str,
		Template: pt,
	}

}
