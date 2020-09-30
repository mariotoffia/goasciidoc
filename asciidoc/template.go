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

// RenderPackage will render the package defintion onto the provided writer.
func (t *Template) RenderPackage(wr io.Writer, file *goparser.GoFile) *Template {

	if err := t.Templates["package"].Execute(wr, file); nil != err {
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
