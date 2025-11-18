package asciidoc

import (
	"bytes"
	"strings"
	texttemplate "text/template"

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
	// PackageRefTemplate is used in master index to reference packages
	PackageRefTemplate TemplateType = "package-ref"
	// PackageRefsTemplate renders package dependencies/references
	PackageRefsTemplate TemplateType = "package-refs"
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

var defaultTemplateFuncs = texttemplate.FuncMap{
	"typeParams":         typeParamsSuffix,
	"nameWithTypeParams": nameWithTypeParams,
	"indent":             func(s string) string { return indent(s) },
	"typeSetItems":       typeSetItems,
	"trimnl":             func(s string) string { return strings.TrimRight(s, "\n") },
	"add":                func(a, b int) int { return a + b },
	"fieldSummary": func(t *TemplateContext, f *goparser.GoField) string {
		return t.fieldSummary(f)
	},
	"fieldHeading": func(t *TemplateContext, f *goparser.GoField) string {
		return t.fieldHeading(f)
	},
	"typeAnchor": func(t *TemplateContext, node interface{}) string {
		return t.typeAnchor(node)
	},
	"functionSignatureDoc": func(t *TemplateContext, m *goparser.GoStructMethod) *SignatureDoc {
		return t.functionSignatureDoc(m)
	},
	"methodSignatureDoc": func(t *TemplateContext, m *goparser.GoMethod, owner []*goparser.GoType) *SignatureDoc {
		return t.methodSignatureDoc(m, owner)
	},
	"funcTypeSignatureDoc": func(t *TemplateContext, m *goparser.GoMethod) *SignatureDoc {
		return t.funcTypeSignatureDoc(m)
	},
	"signatureHighlightBlocks": func(t *TemplateContext, doc *SignatureDoc) []SignatureHighlightBlock {
		return t.signatureHighlightBlocks(doc)
	},
	"signaturePlain": func(t *TemplateContext, doc *SignatureDoc) string {
		return t.signaturePlain(doc)
	},
	"linkedTypeSetDocs": func(t *TemplateContext, types []*goparser.GoType) []*SignatureDoc {
		return t.linkedTypeSetDocs(types)
	},
	"hasJSONTag": func(s *goparser.GoStruct) bool {
		return s.HasJSONTag()
	},
	"hasYAMLTag": func(s *goparser.GoStruct) bool {
		return s.HasYAMLTag()
	},
	"toJSON": func(s *goparser.GoStruct) string {
		return s.ToJSON()
	},
	"toYAML": func(s *goparser.GoStruct) string {
		return s.ToYAML()
	},
}

// TemplateAndText is a wrapper of _template.Template_
// but also includes the original text representation
// of the template and not just the parsed tree.
type TemplateAndText struct {
	// Text is the actual template that got parsed by _template.Template_.
	Text string
	// Template is the instance of the parsed _Text_ including functions.
	Template *texttemplate.Template
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
			IndexTemplate.String():       createTemplate(IndexTemplate, "", overrides, texttemplate.FuncMap{}),
			PackageTemplate.String():     createTemplate(PackageTemplate, "", overrides, texttemplate.FuncMap{}),
			PackageRefTemplate.String():  createTemplate(PackageRefTemplate, "", overrides, texttemplate.FuncMap{}),
			PackageRefsTemplate.String(): createTemplate(PackageRefsTemplate, "", overrides, texttemplate.FuncMap{}),
			ImportTemplate.String(): createTemplate(ImportTemplate, "", overrides, texttemplate.FuncMap{
				"render": func(t *TemplateContext) string { return t.File.DeclImports() },
			}),
			FunctionsTemplate.String(): createTemplate(FunctionsTemplate, "", overrides, texttemplate.FuncMap{
				"render": func(t *TemplateContext, f *goparser.GoStructMethod) string {
					var buf bytes.Buffer
					t.RenderFunction(&buf, f)
					return buf.String()
				},
				"notreceiver": func(t *TemplateContext, f *goparser.GoStructMethod) bool {
					return len(f.Receivers) == 0
				},
			}),
			FunctionTemplate.String(): createTemplate(FunctionTemplate, "", overrides, texttemplate.FuncMap{}),
			InterfacesTemplate.String(): createTemplate(InterfacesTemplate, "", overrides, texttemplate.FuncMap{
				"render": func(t *TemplateContext, i *goparser.GoInterface) string {
					var buf bytes.Buffer
					t.RenderInterface(&buf, i)
					return buf.String()
				},
			}),
			InterfaceTemplate.String(): createTemplate(InterfaceTemplate, "", overrides, texttemplate.FuncMap{
				"tabifylast": func(decl string) string {
					idx := strings.LastIndex(decl, " ")
					if -1 == idx {
						return decl
					}
					return decl[:idx] + "\t" + decl[idx+1:]
				},
			}),
			StructsTemplate.String(): createTemplate(StructsTemplate, "", overrides, texttemplate.FuncMap{
				"render": func(t *TemplateContext, s *goparser.GoStruct) string {
					var buf bytes.Buffer
					t.RenderStruct(&buf, s)
					return buf.String()
				},
			}),
			StructTemplate.String(): createTemplate(StructTemplate, "", overrides, texttemplate.FuncMap{
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
			ReceiversTemplate.String(): createTemplate(ReceiversTemplate, "", overrides, texttemplate.FuncMap{}),
			CustomVarTypeDefsTemplate.String(): createTemplate(CustomVarTypeDefsTemplate, "", overrides, texttemplate.FuncMap{
				"render": func(t *TemplateContext, td *goparser.GoCustomType) string {
					var buf bytes.Buffer
					t.RenderVarTypeDef(&buf, td)
					return buf.String()
				},
			}),
			CustomVarTypeDefTemplate.String(): createTemplate(CustomVarTypeDefTemplate, "", overrides, texttemplate.FuncMap{
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
			VarDeclarationsTemplate.String(): createTemplate(VarDeclarationsTemplate, "", overrides, texttemplate.FuncMap{
				"render": func(t *TemplateContext, a *goparser.GoAssignment) string {
					var buf bytes.Buffer
					t.RenderVarDeclaration(&buf, a)
					return buf.String()
				},
			}),
			VarDeclarationTemplate.String(): createTemplate(VarDeclarationTemplate, "", overrides, texttemplate.FuncMap{}),
			ConstDeclarationsTemplate.String(): createTemplate(ConstDeclarationsTemplate, "", overrides, texttemplate.FuncMap{
				"tabify": func(decl string) string { return strings.Replace(decl, " ", "\t", 1) },
				"render": func(t *TemplateContext, a *goparser.GoAssignment) string {
					var buf bytes.Buffer
					t.RenderConstDeclaration(&buf, a)
					return buf.String()
				},
			}),
			ConstDeclarationTemplate.String(): createTemplate(ConstDeclarationTemplate, "", overrides, texttemplate.FuncMap{}),
			CustomFuncTypeDefsTemplate.String(): createTemplate(CustomFuncTypeDefsTemplate, "", overrides, texttemplate.FuncMap{
				"render": func(t *TemplateContext, td *goparser.GoMethod) string {
					var buf bytes.Buffer
					t.RenderTypeDefFunc(&buf, td)
					return buf.String()
				},
			}),
			CustomFuncTypeDefTemplate.String(): createTemplate(CustomFuncTypeDefTemplate, "", overrides, texttemplate.FuncMap{}),
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
	if strings.TrimSpace(config.SignatureStyle) == "" {
		config.SignatureStyle = "source"
	} else {
		config.SignatureStyle = strings.ToLower(strings.TrimSpace(config.SignatureStyle))
	}

	tc := &TemplateContext{
		creator:     t,
		File:        f,
		Package:     p,
		Module:      f.Module,
		Config:      config,
		Docs:        map[string]string{},
		importCache: map[*goparser.GoFile]map[string]string{},
	}

	return tc
}

// createTemplate will create a template named name and parses the str
// as template. If fails it will panic with the parse error.
//
// If name is found in override map it will use that string to parse the template
// instead of the provided str.
func createTemplate(name TemplateType, str string, overrides map[string]string, fm texttemplate.FuncMap) *TemplateAndText {

	if s, ok := overrides[name.String()]; ok {
		str = s
	}

	pt, err := texttemplate.New(name.String()).Funcs(defaultTemplateFuncs).Funcs(fm).Parse(str)
	if err != nil {
		panic(err)
	}
	return &TemplateAndText{
		Text:     str,
		Template: pt,
	}

}
