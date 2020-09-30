package asciidoc

import (
	"io"
	"text/template"

	"github.com/mariotoffia/goasciidoc/goparser"
)

// Template is handling all templates and actions
// to perform.
type Template struct {
	// Templates to use when rendering documentation
	Templates map[string]*template.Template
}

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
			"package": createTemplate("package", templatePackage, overrides, template.FuncMap{
				"declaration": func(f *goparser.GoFile) string {
					return f.DeclPackage()
				},
			}),
			"import": createTemplate("import", templateImports, overrides, template.FuncMap{
				"declaration": func(f *goparser.GoFile) string {
					return f.DeclImports()
				},
			}),
			"function":  createTemplate("function", templateFunction, overrides, template.FuncMap{}),
			"interface": createTemplate("interface", templateInterface, overrides, template.FuncMap{}),
			"struct":    createTemplate("struct", templateStruct, overrides, template.FuncMap{}),
			"typedef":   createTemplate("typedef", templateCustomTypeDefintion, overrides, template.FuncMap{}),
			"var":       createTemplate("var", templateVarAssignment, overrides, template.FuncMap{}),
			"const":     createTemplate("const", templateConstAssignment, overrides, template.FuncMap{}),
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

// Creator returns the template created this context.
func (t *TemplateContext) Creator() *Template {
	return t.creator
}

// RenderPackage will render the package defintion onto the provided writer.
func (t *TemplateContext) RenderPackage(wr io.Writer) *TemplateContext {

	if err := t.creator.Templates["package"].Execute(wr, t); nil != err {
		panic(err)
	}

	return t
}

// createTemplate will create a template named name and parses the str
// as template. If fails it will panic with the parse error.
//
// If name is found in override map it will use that string to parse the template
// instead of the provided str.
func createTemplate(name, str string, overrides map[string]string, fm template.FuncMap) *template.Template {

	if s, ok := overrides[name]; ok {
		str = s
	}

	pt, err := template.New(name).Funcs(fm).Parse(str)
	if err != nil {
		panic(err)
	}
	return pt

}
