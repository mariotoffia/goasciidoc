package asciidoc

import (
	"text/template"

	"github.com/mariotoffia/goasciidoc/goparser"
)

// TemplateType specifies the template type
type TemplateType string

const (
	// PackageTemplate specifies that the template is a package
	PackageTemplate TemplateType = "package"
	// ImportTemplate specifies that the template renders a import
	ImportTemplate TemplateType = "import"
	// FunctionTemplate is a template to render a function
	FunctionTemplate TemplateType = "function"
	// InterfaceTemplate is a template to render a interface defintion
	InterfaceTemplate TemplateType = "interface"
	// StructTemplate specifies that the template renders a struct defenition
	StructTemplate TemplateType = "struct"
	// CustomVarTypeDefTemplate is a template to render a type definition of a variable
	CustomVarTypeDefTemplate TemplateType = "typedefvar"
	// CustomFuncTYpeDefTemplate is a template to render a function type definition
	CustomFuncTYpeDefTemplate TemplateType = "typedeffunc"
	// VarDeclarationTemplate is a template to render a variable definition
	VarDeclarationTemplate TemplateType = "var"
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
			PackageTemplate.String(): createTemplate(PackageTemplate, templatePackage, overrides, template.FuncMap{}),
			ImportTemplate.String(): createTemplate(ImportTemplate, templateImports, overrides, template.FuncMap{
				"declaration": func(f *goparser.GoFile) string {
					return f.DeclImports()
				},
				"cr": func() string {
					return "\n"
				},
			}),
			FunctionTemplate.String():          createTemplate(FunctionTemplate, templateFunction, overrides, template.FuncMap{}),
			InterfaceTemplate.String():         createTemplate(InterfaceTemplate, templateInterface, overrides, template.FuncMap{}),
			StructTemplate.String():            createTemplate(StructTemplate, templateStruct, overrides, template.FuncMap{}),
			CustomVarTypeDefTemplate.String():  createTemplate(CustomVarTypeDefTemplate, templateCustomTypeDefintion, overrides, template.FuncMap{}),
			VarDeclarationTemplate.String():    createTemplate(VarDeclarationTemplate, templateVarAssignment, overrides, template.FuncMap{}),
			ConstDeclarationTemplate.String():  createTemplate(ConstDeclarationTemplate, templateConstAssignment, overrides, template.FuncMap{}),
			CustomFuncTYpeDefTemplate.String(): createTemplate(CustomFuncTYpeDefTemplate, templateCustomFuncDefintion, overrides, template.FuncMap{}),
		},
	}

}

// NewContext creates a new context to be used for rendering.
func (t *Template) NewContext(f *goparser.GoFile) *TemplateContext {

	return &TemplateContext{
		creator: t,
		File:    f,
	}

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
