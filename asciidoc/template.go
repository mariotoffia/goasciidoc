package asciidoc

import (
	"bytes"
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
	// InterfaceTemplate is a template to render a interface defintion
	InterfaceTemplate TemplateType = "interface"
	// StructsTemplate specifies that the template renders all struct defenitions for a given context (package, file)
	StructsTemplate TemplateType = "structs"
	// StructTemplate specifies that the template renders a struct defenition
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
)

func (tt TemplateType) String() string {
	return string(tt)
}

// Template is handling all templates and actions
// to perform.
type Template struct {
	// Templates to use when rendering documentation
	Templates map[string]*template.Template
}

// NewTemplate creates a new set of templates to be used
func NewTemplate() *Template {
	return NewTemplateWithOverrides(map[string]string{})
}

// NewTemplateWithOverrides creates a new template with the ability to easily
// override defaults.
func NewTemplateWithOverrides(overrides map[string]string) *Template {

	return &Template{
		Templates: map[string]*template.Template{
			IndexTemplate.String(): createTemplate(IndexTemplate, templateIndex, overrides, template.FuncMap{
				"cr": func() string { return "\n" },
			}),
			PackageTemplate.String(): createTemplate(PackageTemplate, templatePackage, overrides, template.FuncMap{}),
			ImportTemplate.String(): createTemplate(ImportTemplate, templateImports, overrides, template.FuncMap{
				"render": func(t *TemplateContext) string { return t.File.DeclImports() },
				"cr":     func() string { return "\n" },
			}),
			FunctionsTemplate.String(): createTemplate(FunctionsTemplate, templateFunctions, overrides, template.FuncMap{
				"cr": func() string { return "\n" },
				"render": func(t *TemplateContext, f *goparser.GoStructMethod) string {
					var buf bytes.Buffer
					t.RenderFunction(&buf, f)
					return buf.String()
				},
			}),
			FunctionTemplate.String(): createTemplate(FunctionTemplate, templateFunction, overrides, template.FuncMap{
				"cr": func() string { return "\n" },
			}),
			InterfacesTemplate.String(): createTemplate(InterfacesTemplate, templateInterfaces, overrides, template.FuncMap{
				"render": func(t *TemplateContext, i *goparser.GoInterface) string {
					var buf bytes.Buffer
					t.RenderInterface(&buf, i)
					return buf.String()
				},
			}),
			InterfaceTemplate.String(): createTemplate(InterfaceTemplate, templateInterface, overrides, template.FuncMap{}),
			StructsTemplate.String(): createTemplate(StructsTemplate, templateStructs, overrides, template.FuncMap{
				"render": func(t *TemplateContext, s *goparser.GoStruct) string {
					var buf bytes.Buffer
					t.RenderStruct(&buf, s)
					return buf.String()
				},
			}),
			StructTemplate.String(): createTemplate(StructTemplate, templateStruct, overrides, template.FuncMap{}),
			CustomVarTypeDefsTemplate.String(): createTemplate(CustomVarTypeDefsTemplate, templateCustomTypeDefintions, overrides, template.FuncMap{
				"render": func(t *TemplateContext, td *goparser.GoCustomType) string {
					var buf bytes.Buffer
					t.RenderVarTypeDef(&buf, td)
					return buf.String()
				},
			}),
			CustomVarTypeDefTemplate.String(): createTemplate(CustomVarTypeDefTemplate, templateCustomTypeDefintion, overrides, template.FuncMap{}),
			VarDeclarationsTemplate.String(): createTemplate(VarDeclarationsTemplate, templateVarAssignments, overrides, template.FuncMap{
				"render": func(t *TemplateContext, a *goparser.GoAssignment) string {
					var buf bytes.Buffer
					t.RenderVarDeclaration(&buf, a)
					return buf.String()
				},
			}),
			VarDeclarationTemplate.String(): createTemplate(VarDeclarationTemplate, templateVarAssignment, overrides, template.FuncMap{}),
			ConstDeclarationsTemplate.String(): createTemplate(ConstDeclarationsTemplate, templateConstAssignments, overrides, template.FuncMap{
				"render": func(t *TemplateContext, a *goparser.GoAssignment) string {
					var buf bytes.Buffer
					t.RenderConstDeclaration(&buf, a)
					return buf.String()
				},
			}),
			ConstDeclarationTemplate.String(): createTemplate(ConstDeclarationTemplate, templateConstAssignment, overrides, template.FuncMap{}),
			CustomFuncTypeDefsTemplate.String(): createTemplate(CustomFuncTypeDefsTemplate, templateCustomFuncDefintions, overrides, template.FuncMap{
				"render": func(t *TemplateContext, td *goparser.GoMethod) string {
					var buf bytes.Buffer
					t.RenderTypeDefFunc(&buf, td)
					return buf.String()
				},
			}),
			CustomFuncTypeDefTemplate.String(): createTemplate(CustomFuncTypeDefTemplate, templateCustomFuncDefintion, overrides, template.FuncMap{}),
		},
	}

}

// NewContext creates a new context to be used for rendering.
func (t *Template) NewContext(f *goparser.GoFile) *TemplateContext {
	return t.NewContextWithConfig(f, &TemplateContextConfig{})
}

// NewContextWithConfig creates a new context with configuration.
//
// If configuration is nil, it will use default configuration.
func (t *Template) NewContextWithConfig(f *goparser.GoFile, config *TemplateContextConfig) *TemplateContext {

	if nil == config {
		config = &TemplateContextConfig{}
	}

	tc := &TemplateContext{
		creator: t,
		File:    f,
		Module:  f.Module,
		Config:  config,
	}

	return tc
}

// createTemplate will create a template named name and parses the str
// as template. If fails it will panic with the parse error.
//
// If name is found in override map it will use that string to parse the template
// instead of the provided str.
func createTemplate(name TemplateType, str string, overrides map[string]string, fm template.FuncMap) *template.Template {

	if s, ok := overrides[name.String()]; ok {
		str = s
	}

	pt, err := template.New(name.String()).Funcs(fm).Parse(str)
	if err != nil {
		panic(err)
	}
	return pt

}
