package asciidoc

import (
	"bytes"
	"testing"

	"github.com/mariotoffia/goasciidoc/goparser"
	"github.com/stretchr/testify/assert"
)

func dummyModule() *goparser.GoModule {
	mod, _ := goparser.NewModuleFromBuff("/tmp/test-asciidoc/go.mod",
		[]byte(`module github.com/mariotoffia/goasciidoc/tests
	go 1.14`))
	mod.Version = "0.0.1"

	return mod
}
func TestRenderPackageWithModule(t *testing.T) {
	src := `
	// The package mypkg is a sample package.
	package mypkg`

	m := dummyModule()
	f, err := goparser.ParseInlineFile(m, m.Base+"/mypkg/file.go", src)

	assert.NoError(t, err)

	var buf bytes.Buffer

	x := NewTemplateWithOverrides(map[string]string{
		PackageTemplate.String(): `== {{if .File.FqPackage}}package {{.File.FqPackage}}{{else}}{{.File.Decl}}{{end}}
{{ .File.Doc }}`,
	}).NewContext(f)

	x.RenderPackage(&buf)

	assert.Equal(t, "== package github.com/mariotoffia/goasciidoc/tests/mypkg\nThe package mypkg is a sample package.", buf.String())
}

func TestRenderPackageWithoutModule(t *testing.T) {
	src := `
	// The package mypkg is a sample package.
	package mypkg`

	f, err := goparser.ParseInlineFile(nil, "", src)

	assert.NoError(t, err)

	var buf bytes.Buffer

	x := NewTemplateWithOverrides(map[string]string{
		PackageTemplate.String(): `== {{if .File.FqPackage}}package {{.File.FqPackage}}{{else}}{{.File.Decl}}{{end}}
{{ .File.Doc }}`,
	}).NewContext(f)

	x.RenderPackage(&buf)

	assert.Equal(t, "== package mypkg\nThe package mypkg is a sample package.", buf.String())
}

func TestRenderImports(t *testing.T) {
	src := `	
	package mypkg
	
	import (
		// We import format here
		"fmt"
		// and time here :)
		"time"
	)
	func Bar() {
		fmt.Println(time.Now())
	}`

	m := dummyModule()
	f, err := goparser.ParseInlineFile(m, m.Base+"/mypkg/file.go", src)
	assert.NoError(t, err)

	var buf bytes.Buffer

	x := NewTemplateWithOverrides(map[string]string{
		ImportTemplate.String(): `=== Imports
[source, go]
----
{{ render . }}
----
{{range .File.Imports}}{{if .Doc }}{{ cr }}==== Import _{{ .Path }}_{{ cr }}{{ .Doc }}{{ cr }}{{end}}{{end}}`,
	}).NewContext(f)

	x.RenderImports(&buf)

	assert.Equal(t,
		"=== Imports\n[source, go]\n----\nimport (\n\t\"fmt\"\n\t\"time\"\n)\n----\n\n==== Import _fmt_\nWe import format here\n\n==== Import _time_\nand time here :)\n",
		buf.String())
}

func TestRenderSingleFunction(t *testing.T) {
	src := `	
	package mypkg
	
	import (
		"fmt"
		"time"
	)
	// Bar is a public function that outputs
	// current time and return zero.
	func Bar() int {
		fmt.Println(time.Now())
		return 0
	}`

	m := dummyModule()
	f, err := goparser.ParseInlineFile(m, m.Base+"/mypkg/file.go", src)
	assert.NoError(t, err)

	var buf bytes.Buffer

	x := NewTemplateWithOverrides(map[string]string{
		FunctionTemplate.String(): `=== {{ .Function.Name }}
[source, go]
----
{{ .Function.Decl }}
----
		
{{ .Function.Doc }}{{ if .Config.IncludeMethodCode }}{{cr}}[source, go]{{cr}}----{{cr}}{{ .Function.FullDecl }}{{cr}}----{{end}}`,
	}).NewContext(f)

	x.RenderFunction(&buf, f.StructMethods[0])

	assert.Equal(t, "=== Bar\n[source, go]\n----\nfunc Bar() int\n----\n\t\t\nBar is a public function that outputs\ncurrent time and return zero.", buf.String())
}

func TestRenderSingleFunctionWithCode(t *testing.T) {
	src := `	
	package mypkg
	
	import (
		"fmt"
		"time"
	)
	// Bar is a public function that outputs
	// current time and return zero.
	func Bar() int {
		fmt.Println(time.Now())
		return 0
	}`

	m := dummyModule()
	f, err := goparser.ParseInlineFile(m, m.Base+"/mypkg/file.go", src)
	assert.NoError(t, err)

	var buf bytes.Buffer

	x := NewTemplateWithOverrides(map[string]string{
		FunctionTemplate.String(): `=== {{ .Function.Name }}
[source, go]
----
{{ .Function.Decl }}
----

{{ .Function.Doc }}
{{ if .Config.IncludeMethodCode }}{{cr}}[source, go]{{cr}}----{{cr}}{{ .Function.FullDecl }}{{cr}}----{{end}}`,
	}).NewContextWithConfig(f, &TemplateContextConfig{IncludeMethodCode: true})

	x.RenderFunction(&buf, f.StructMethods[0])

	assert.Equal(t,
		"=== Bar\n[source, go]\n----\nfunc Bar() int\n----\n\nBar is a public function that outputs\ncurrent time and return zero.\n\n[source, go]"+
			"\n----\nfunc Bar() int {\n\t\tfmt.Println(time.Now())\n\t\treturn 0\n\t}\n----",
		buf.String())
}

func TestRenderFunctions(t *testing.T) {
	src := `	
package mypkg

import (
	"fmt"
	"time"
	"testing"
	"github.com/stretchr/testify/assert"
)
// Bar is a public function that outputs
// current time and return zero.
func Bar() int {
	fmt.Println(time.Now())
	return 0
}

// _Fubbo_ is a testing function that uses
// many tricks in the book. 
// [TIP]
// .Simplify Configuration
// ====
// Try to use a simple test config
// ====
func Fubbo(t *testing.T) {
	fmt.Println("hello world from test")
	assert.Equal(t, "hello", "world")
}`

	m := dummyModule()
	f, err := goparser.ParseInlineFile(m, m.Base+"/mypkg/file.go", src)
	assert.NoError(t, err)

	var buf bytes.Buffer

	x := NewTemplateWithOverrides(map[string]string{
		FunctionsTemplate.String(): `== Functions
{{range .File.StructMethods}}
{{- render $ .}}
{{end}}`,
	}).NewContext(f)

	x.RenderFunctions(&buf)

	assert.Equal(t,
		"== Functions\n=== Bar\n[source, go]\n----\nfunc Bar() int\n----\n\nBar is a public function that outputs\n"+
			"current time and return zero.\n\n=== Fubbo\n[source, go]\n----\nfunc Fubbo(t *testing.T)\n----\n\n_Fubbo_ is"+
			" a testing function that uses\nmany tricks in the book.\n[TIP]\n.Simplify Configuration\n====\nTry to use a simple test config\n====\n\n",
		buf.String())
}
