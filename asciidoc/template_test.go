package asciidoc

import (
	"bytes"
	"fmt"
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
{{range .File.Imports}}{{if .Doc }}{{"\n"}}==== Import _{{ .Path }}_{{"\n"}}{{ .Doc }}{{"\n"}}{{end}}{{end}}`,
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
		
{{ .Function.Doc }}{{ if .Config.IncludeMethodCode }}{{"\n"}}[source, go]{{"\n"}}----{{"\n"}}{{ .Function.FullDecl }}{{"\n"}}----{{end}}`,
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
{{ if .Config.IncludeMethodCode }}{{"\n"}}[source, go]{{"\n"}}----{{"\n"}}{{ .Function.FullDecl }}{{"\n"}}----{{end}}`,
	}).NewContextWithConfig(f, nil, &TemplateContextConfig{IncludeMethodCode: true})

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
		"== Functions\n=== Bar\n[source, go]\n----\nfunc Bar() int\n----\n\nBar is a public function that outputs\ncurrent time and return zero.\n\n\n=== Fubbo\n[source, go]\n----\nfunc Fubbo(t *testing.T)\n----\n\n_Fubbo_ is a testing function that uses\nmany tricks in the book.\n[TIP]\n.Simplify Configuration\n====\nTry to use a simple test config\n====\n\n\n",
		buf.String())
}

func TestRenderSingleInterface(t *testing.T) {
	src := `	
package mypkg

import "time"

// IInterface is a public interface.
type IInterface interface {
	// Bar is a public function that outputs
	// current time and return zero.
	Bar() int
	// baz is a private function that returns current time.
	baz() time.Time
}`

	m := dummyModule()
	f, err := goparser.ParseInlineFile(m, m.Base+"/mypkg/file.go", src)
	assert.NoError(t, err)

	var buf bytes.Buffer

	x := NewTemplateWithOverrides(map[string]string{
		InterfaceTemplate.String(): `=== {{ .Interface.Name }}
[source, go]
----
{{.Interface.Decl}} {
{{- range .Interface.Methods}}
	{{.Decl}}
{{- end}}
}
----
		
{{ .Interface.Doc }}
{{range .Interface.Methods}}
==== {{.Decl}}
{{.Doc}}
{{end}}`,
	}).NewContext(f)

	x.RenderInterface(&buf, f.Interfaces[0])

	assert.Equal(t,
		"=== IInterface\n[source, go]\n----\ntype IInterface interface {\n\tBar() "+
			"int\n\tbaz() time.Time\n}\n----\n\t\t\nIInterface is a public interface.\n\n"+
			"==== Bar() int\nBar is a public function that outputs\ncurrent time and "+
			"return zero.\n\n==== baz() time.Time\nbaz is a private function that returns current time.\n",
		buf.String())
}

func TestRenderMultipleInterfaces(t *testing.T) {
	src := `	
package mypkg

import "time"

// IInterface is a public interface.
type IInterface interface {
	// Bar is a public function that outputs
	// current time and return zero.
	Bar() int
	// baz is a private function that returns current time.
	baz() time.Time
}

// MyInterface is a plain interface to do misc stuff.
type MyInterface interface {
	// FooBot is a public method to do just that! ;)
	FooBot(i IInterface) string
}`

	m := dummyModule()
	f, err := goparser.ParseInlineFile(m, m.Base+"/mypkg/file.go", src)
	assert.NoError(t, err)

	var buf bytes.Buffer

	x := NewTemplateWithOverrides(map[string]string{
		InterfacesTemplate.String(): `== Interfaces
{{range .File.Interfaces}}
{{- render $ .}}
{{end}}`,
	}).NewContext(f)

	x.RenderInterfaces(&buf)

	assert.Equal(t, "== Interfaces\n=== IInterface\n[source, go]\n----\ntype IInterface interface {\n\tBar()\tint\n\tbaz()\ttime.Time\n}\n----\n\t\t\nIInterface is a public interface.\n\n==== Bar() int\nBar is a public function that outputs\ncurrent time and return zero.\n\n==== baz() time.Time\nbaz is a private function that returns current time.\n\n\n=== MyInterface\n[source, go]\n----\ntype MyInterface interface {\n\tFooBot(i IInterface)\tstring\n}\n----\n\t\t\nMyInterface is a plain interface to do misc stuff.\n\n==== FooBot(i IInterface) string\nFooBot is a public method to do just that! ;)\n\n\n",
		buf.String())
}

func TestRenderSingleStruct(t *testing.T) {
	src := `	
package mypkg

import "time"

// Person is a public struct describing
// a persons name, age and when he or
// she was born.
type Person struct {
	// Name is full name
	Name string
	// Born is when the person was born
	Born  time.Time
	// Age is how old this person is now
	Age uint8
}`

	m := dummyModule()
	f, err := goparser.ParseInlineFile(m, m.Base+"/mypkg/file.go", src)
	assert.NoError(t, err)

	var buf bytes.Buffer

	x := NewTemplateWithOverrides(map[string]string{
		StructTemplate.String(): `=== {{.Struct.Name}}
[source, go]
----
{{.Struct.Decl}} {
{{- range .Struct.Fields}}
	{{.Decl}}
{{- end}}
}
----
		
{{ .Struct.Doc }}
{{range .Struct.Fields}}
==== {{.Decl}}
{{.Doc}}
{{end}}`,
	}).NewContext(f)

	x.RenderStruct(&buf, f.Structs[0])

	assert.Equal(t,
		"=== Person\n[source, go]\n----\ntype Person struct {\n\tName string\n\tBorn time.Time\n\tAge uint8\n}\n"+
			"----\n\t\t\nPerson is a public struct describing\na persons name, age and when he or\nshe was born.\n\n"+
			"==== Name string\nName is full name\n\n==== Born time.Time\nBorn is when the person was born\n\n==== "+
			"Age uint8\nAge is how old this person is now\n",
		buf.String())
}

func TestRenderMultipleStructs(t *testing.T) {
	src := `	
package mypkg

import "time"

// Person is a public struct describing
// a persons name, age and when he or
// she was born.
type Person struct {
	// Name is full name
	Name string
	// Born is when the person was born
	Born  time.Time
	// Age is how old this person is now
	Age uint8
}

// Anka is a duck
type Anka struct {
	// Anka is a person like Kalle Anka
	Person
	// Loudness is the amplitude of the kvack!
	Loudness int32
}`

	m := dummyModule()
	f, err := goparser.ParseInlineFile(m, m.Base+"/mypkg/file.go", src)
	assert.NoError(t, err)

	var buf bytes.Buffer

	x := NewTemplateWithOverrides(map[string]string{
		StructsTemplate.String(): `== Structs
{{range .File.Structs}}
{{- render $ .}}
{{end}}`,
	}).NewContext(f)

	x.RenderStructs(&buf)

	assert.Equal(t, `== Structs
=== Person
[source, go]
----
type Person struct {
	Name	string
	Born	time.Time
	Age	uint8
}
----

Person is a public struct describing
a persons name, age and when he or
she was born.

==== Name string
Name is full name

==== Born time.Time
Born is when the person was born

==== Age uint8
Age is how old this person is now



=== Anka
[source, go]
----
type Anka struct {
	Person
	Loudness	int32
}
----

Anka is a duck

==== Person
Anka is a person like Kalle Anka

==== Loudness int32
Loudness is the amplitude of the kvack!



`,
		buf.String())
}

func TestRenderNestedAnonymousStruct(t *testing.T) {
	src := `	
package mypkg

import "time"

// MyStruct is a structure of nonsense
type MyStruct struct {
	// Inline the struct
	Inline struct {
		// FooBar is a even more nonsense variable
		FooBar int
	}
	// MyInt is happy to be after Inline
	MyInt int
}`

	m := dummyModule()
	f, err := goparser.ParseInlineFile(m, m.Base+"/mypkg/file.go", src)
	assert.NoError(t, err)

	var buf bytes.Buffer

	x := NewTemplateWithOverrides(map[string]string{
		StructTemplate.String(): `=== {{.Struct.Name}}
[source, go]
----
{{.Struct.Decl}} {
{{- range .Struct.Fields}}
	{{if .Nested}}{{.Nested.Name}}{{"\t"}}struct{{else}}{{tabify .Decl}}{{end}}
{{- end}}
}
----

{{ .Struct.Doc }}
{{- range .Struct.Fields}}{{if not .Nested}}
==== {{.Decl}}
{{.Doc}}
{{- end}}
{{end}}
{{range .Struct.Fields}}{{if .Nested}}{{render $ .Nested}}{{end}}{{end}}`,
	}).NewContext(f)

	x.RenderStruct(&buf, f.Structs[0])

	assert.Equal(t, `=== MyStruct
[source, go]
----
type MyStruct struct {
	Inline	struct
	MyInt	int
}
----

MyStruct is a structure of nonsense

==== MyInt int
MyInt is happy to be after Inline

=== Inline
[source, go]
----
struct {
	FooBar	int
}
----

Inline the struct
==== FooBar int
FooBar is a even more nonsense variable

`,
		buf.String())
}

func TestRenderNestedKnownStruct(t *testing.T) {
	src := `	
package mypkg

import "time"

// This is inline struct
type Inline struct {
	// FooBar in the inline struct
	FooBar int
}
// MyStruct is a structure of nonsense
type MyStruct struct {
	// Inline the struct
	Ins Inline
	// MyInt is happy to be after Inline
	MyInt int
}`

	m := dummyModule()
	f, err := goparser.ParseInlineFile(m, m.Base+"/mypkg/file.go", src)
	assert.NoError(t, err)

	var buf bytes.Buffer

	x := NewTemplateWithOverrides(map[string]string{
		StructTemplate.String(): `=== {{.Struct.Name}}
[source, go]
----
{{.Struct.Decl}} {
{{- range .Struct.Fields}}
	{{if .Nested}}{{.Nested.Name}} struct{{else}}{{tabify .Decl}}{{end}}
{{- end}}
}
----

{{ .Struct.Doc }}
{{- range .Struct.Fields}}{{if not .Nested}}
==== {{.Decl}}
{{.Doc}}
{{- end}}
{{end}}
{{range .Struct.Fields}}{{if .Nested}}{{render $ .Nested}}{{end}}{{end}}`,
	}).NewContext(f)

	x.RenderStruct(&buf, f.Structs[1])

	assert.Equal(t, `=== MyStruct
[source, go]
----
type MyStruct struct {
	Ins	Inline
	MyInt	int
}
----

MyStruct is a structure of nonsense
==== Ins Inline
Inline the struct

==== MyInt int
MyInt is happy to be after Inline

`,
		buf.String())
}

func TestRenderSingleVarTypeDef(t *testing.T) {
	src := `	
package mypkg

// MyVarTypeDef is a type that wraps a int to a custom type
type MyVarTypeDef int`

	m := dummyModule()
	f, err := goparser.ParseInlineFile(m, m.Base+"/mypkg/file.go", src)
	assert.NoError(t, err)

	var buf bytes.Buffer

	x := NewTemplateWithOverrides(map[string]string{
		CustomVarTypeDefTemplate.String(): `=== {{.TypeDefVar.Name}}
[source, go]
----
{{.TypeDefVar.Decl}}
----
{{.TypeDefVar.Doc}}`,
	}).NewContext(f)

	x.RenderVarTypeDef(&buf, f.CustomTypes[0])

	assert.Equal(t,
		"=== MyVarTypeDef\n[source, go]\n----\ntype MyVarTypeDef int\n----\nMyVarTypeDef is a type that wraps a int to a custom type",
		buf.String())
}

func TestRenderMultipleVarTypeDefs(t *testing.T) {
	src := `	
package mypkg

// MyVarTypeDef is a type that wraps a int to a custom type
type MyVarTypeDef int

// NextVar is a another custom typedef for a variable.
type NextVar string`

	m := dummyModule()
	f, err := goparser.ParseInlineFile(m, m.Base+"/mypkg/file.go", src)
	assert.NoError(t, err)

	var buf bytes.Buffer

	x := NewTemplateWithOverrides(map[string]string{
		CustomVarTypeDefsTemplate.String(): `== Variable Type Definitions
{{range .File.CustomTypes}}
{{render $ .}}
{{end}}`,
	}).NewContext(f)

	x.RenderVarTypeDefs(&buf)

	assert.Equal(t,
		"== Variable Type Definitions\n\n=== MyVarTypeDef\n[source, go]\n----\ntype MyVarTypeDef int\n----\nMyVarTypeDef is a type that wraps a int to a custom type\n\n\n=== NextVar\n[source, go]\n----\ntype NextVar string\n----\nNextVar is a another custom typedef for a variable.\n\n",
		buf.String())
}

func TestRenderSingleVarDeclaration(t *testing.T) {
	src := `	
package mypkg

// MyVar is a var declaration that is exported
var MyVar int = 99`

	m := dummyModule()
	f, err := goparser.ParseInlineFile(m, m.Base+"/mypkg/file.go", src)
	assert.NoError(t, err)

	var buf bytes.Buffer

	x := NewTemplateWithOverrides(map[string]string{
		VarDeclarationTemplate.String(): `=== {{.VarAssignment.Name}}
[source, go]
----
{{.VarAssignment.FullDecl}}
----
{{.VarAssignment.Doc}}`,
	}).NewContext(f)

	x.RenderVarDeclaration(&buf, f.VarAssignments[0])

	assert.Equal(t,
		"=== MyVar\n[source, go]\n----\nvar MyVar int = 99\n----\nMyVar is a var declaration that is exported",
		buf.String())
}

func TestRenderMultipleVarDeclarations(t *testing.T) {
	src := `	
package mypkg

// MyVar is a var declaration that is exported
var MyVar int = 99

// dryrun determines if the lambda will affect resources or just log
var dryrun = false`

	m := dummyModule()
	f, err := goparser.ParseInlineFile(m, m.Base+"/mypkg/file.go", src)
	assert.NoError(t, err)

	var buf bytes.Buffer

	x := NewTemplateWithOverrides(map[string]string{
		VarDeclarationsTemplate.String(): `== Variables
{{range .File.VarAssignments}}
{{render $ .}}
{{end}}`,
	}).NewContext(f)

	x.RenderVarDeclarations(&buf)

	assert.Equal(t,
		"== Variables\n\n=== MyVar\n[source, go]\n----\nvar MyVar int = 99\n----\nMyVar is a var declaration that is exported\n\n\n=== dryrun\n[source, go]\n----\nvar dryrun = false\n----\ndryrun determines if the lambda will affect resources or just log\n\n",
		buf.String())
}

func TestRenderSingleConstDeclaration(t *testing.T) {
	src := `	
package mypkg

const (
	// MyConstVar is just to demonstrate a single const declaration
	MyConstVar string = "apa"
)`

	m := dummyModule()
	f, err := goparser.ParseInlineFile(m, m.Base+"/mypkg/file.go", src)
	assert.NoError(t, err)

	var buf bytes.Buffer

	x := NewTemplateWithOverrides(map[string]string{
		ConstDeclarationTemplate.String(): `=== {{.ConstAssignment.Name}}
[source, go]
----
{{.ConstAssignment.Decl}}
----
{{.ConstAssignment.Doc}}`,
	}).NewContext(f)

	x.RenderConstDeclaration(&buf, f.ConstAssignments[0])

	assert.Equal(t,
		"=== MyConstVar\n[source, go]\n----\nMyConstVar string = \"apa\"\n----\nMyConstVar is just to demonstrate a single const declaration",
		buf.String())
}

func TestRenderMultipleConstDeclarations(t *testing.T) {
	src := `	
package mypkg

const (
	// MyConstVar is just to demonstrate a single const declaration
	MyConstVar string = "apa"
	// NextVar is more trixy...
	NextVar string = "next"
)`

	m := dummyModule()
	f, err := goparser.ParseInlineFile(m, m.Base+"/mypkg/file.go", src)
	assert.NoError(t, err)

	var buf bytes.Buffer

	x := NewTemplateWithOverrides(map[string]string{
		ConstDeclarationsTemplate.String(): `=== Constants
[source, go]
----
const (
	{{- range .File.ConstAssignments}}
	{{.Decl}}
	{{- end}}
)
----
{{range .File.ConstAssignments}}
{{render $ .}}
{{end}}`,
	}).NewContext(f)

	x.RenderConstDeclarations(&buf)

	assert.Equal(t,
		"=== Constants\n[source, go]\n----\nconst (\n\tMyConstVar string = \"apa\"\n\tNextVar string = \"next\"\n)\n----\n\n=== MyConstVar\n[source, go]\n----\nMyConstVar string = \"apa\"\n----\nMyConstVar is just to demonstrate a single const declaration\n\n\n=== NextVar\n[source, go]\n----\nNextVar string = \"next\"\n----\nNextVar is more trixy...\n\n",
		buf.String())
}

func TestRenderSingleTypeDefFunc(t *testing.T) {
	src := `	
package mypkg

// Parse is a function that gets an id and a message and 
// is expected to return an array of tokenized data
// or _error_ if fails.
type Parse func(id, msg string) ([]string, error)`

	m := dummyModule()
	f, err := goparser.ParseInlineFile(m, m.Base+"/mypkg/file.go", src)
	assert.NoError(t, err)

	var buf bytes.Buffer

	x := NewTemplateWithOverrides(map[string]string{
		CustomFuncTypeDefTemplate.String(): `=== {{.TypeDefFunc.Name}}
[source, go]
----
{{.TypeDefFunc.Decl}}
----
{{.TypeDefFunc.Doc}}`,
	}).NewContext(f)

	x.RenderTypeDefFunc(&buf, f.CustomFuncs[0])

	assert.Equal(t,
		"=== Parse\n[source, go]\n----\ntype Parse func(id, msg string) ([]string, error)\n----"+
			"\nParse is a function that gets an id and a message and\nis expected to return an array "+
			"of tokenized data\nor _error_ if fails.",
		buf.String())
}

func TestRenderMultipleTypeDefFuncs(t *testing.T) {
	src := `	
package mypkg

// Parse is a function that gets an id and a message and 
// is expected to return an array of tokenized data
// or _error_ if fails.
type Parse func(id, msg string) ([]string, error)

// Visit is a visitor function that gets one chunk from the
// return value from Parse function.
type Visit func(chunk string) error`

	m := dummyModule()
	f, err := goparser.ParseInlineFile(m, m.Base+"/mypkg/file.go", src)
	assert.NoError(t, err)

	var buf bytes.Buffer

	x := NewTemplateWithOverrides(map[string]string{
		CustomFuncTypeDefsTemplate.String(): `=== Function Definitions
{{range .File.CustomFuncs}}
{{render $ .}}
{{end}}`,
	}).NewContext(f)

	x.RenderTypeDefFuncs(&buf)

	assert.Equal(t,
		"=== Function Definitions\n\n=== Parse\n[source, go]\n----\ntype Parse func(id, msg string) ([]string, error)\n----\nParse is a function that gets an id and a message and\nis expected to return an array of tokenized data\nor _error_ if fails.\n\n\n=== Visit\n[source, go]\n----\ntype Visit func(chunk string) error\n----\nVisit is a visitor function that gets one chunk from the\nreturn value from Parse function.\n\n",
		buf.String())
}

func TestRenderIndexWithDefaults(t *testing.T) {
	index := `= {{ .Index.Title }}
{{- if .Index.AuthorName}}{{"\n"}}:author_name: {{.Index.AuthorName}}{{"\n"}}:author: {author_name}{{end}}
{{- if .Index.AuthorEmail}}{{"\n"}}:author_email: {{.Index.AuthorEmail}}{{"\n"}}:email: {author_email}{{end}}
:source-highlighter: {{ .Index.Highlighter }}
{{- if .Index.TocTitle}}{{"\n"}}:toc:{{"\n"}}:toc-title: {{ .Index.TocTitle }}{{"\n"}}:toclevels: {{ .Index.TocLevels }}{{end}}
:icons: font
{{- if .Index.ImageDir}}{{"\n"}}:imagesdir: {{.Index.ImageDir}}{{end}}
{{- if .Index.HomePage}}{{"\n"}}:homepage: {{.Index.HomePage}}{{end}}
:kroki-default-format: svg
:doctype: {{.Index.DocType}}`

	m := dummyModule()
	f, err := goparser.ParseInlineFile(m, m.Base+"/mypkg/file.go", `package mypkg`)
	assert.NoError(t, err)

	var buf bytes.Buffer

	x := NewTemplateWithOverrides(map[string]string{IndexTemplate.String(): index}).
		NewContext(f)

	x.RenderIndex(&buf, nil)

	assert.Equal(t,
		"= github.com/mariotoffia/goasciidoc/tests\n:author_name: martoffi\n:author: {author_name}\n"+
			":source-highlighter: highlightjs\n:toc:\n:toc-title: Table of Contents\n:toclevels: 3\n:icons: font\n:kroki-default-format: svg\n:doctype: book",
		buf.String())
}

func TestRenderIndexWithAllSet(t *testing.T) {
	index := `= {{ .Index.Title }}
{{- if .Index.AuthorName}}{{"\n"}}:author_name: {{.Index.AuthorName}}{{"\n"}}:author: {author_name}{{end}}
{{- if .Index.AuthorEmail}}{{"\n"}}:author_email: {{.Index.AuthorEmail}}{{"\n"}}:email: {author_email}{{end}}
:source-highlighter: {{ .Index.Highlighter }}
{{- if .Index.TocTitle}}{{"\n"}}:toc:{{"\n"}}:toc-title: {{ .Index.TocTitle }}{{"\n"}}:toclevels: {{ .Index.TocLevels }}{{end}}
:icons: font
{{- if .Index.ImageDir}}{{"\n"}}:imagesdir: {{.Index.ImageDir}}{{end}}
{{- if .Index.HomePage}}{{"\n"}}:homepage: {{.Index.HomePage}}{{end}}
:kroki-default-format: svg
:doctype: {{.Index.DocType}}`

	m := dummyModule()
	f, err := goparser.ParseInlineFile(m, m.Base+"/mypkg/file.go", `package mypkg`)
	assert.NoError(t, err)

	var buf bytes.Buffer

	x := NewTemplateWithOverrides(map[string]string{IndexTemplate.String(): index}).
		NewContext(f)

	ic := x.DefaultIndexConfig(`{
		"author":"Mario Toffia", 
		"email": "mario.toffia@bullen.com",
		"web": "www.bullen.se",
		"images": "../meta/assets",
		"title": "Bullen Bakar Kaka",
		"toc": "Table of Contents",
		"toclevel": 2
		}`)

	x.RenderIndex(&buf, ic)

	assert.Equal(t,
		"= Bullen Bakar Kaka\n:author_name: Mario Toffia\n:author: {author_name}\n:author_email: mario.toffia@bullen.com\n"+
			":email: {author_email}\n:source-highlighter: highlightjs\n:toc:\n:toc-title: Table of Contents\n:toclevels: 2\n"+
			":icons: font\n:imagesdir: ../meta/assets\n:homepage: www.bullen.se\n:kroki-default-format: svg\n:doctype: book",
		buf.String())
}

func TestStructReceiverFunction(t *testing.T) {
	src := `	
	package mypkg

	// MyStruct is my and only struct
	type MyStruct struct {} 
	// Bar is a public function receives a string and returns a string.
	func (ms *MyStruct) Bar(msg string) string {
		return "hello: " + msg
	}`

	m := dummyModule()
	f, err := goparser.ParseInlineFile(m, m.Base+"/mypkg/file.go", src)
	assert.NoError(t, err)

	var buf bytes.Buffer

	x := NewTemplateWithOverrides(map[string]string{
		StructTemplate.String(): `=== {{.Struct.Name}}
[source, go]
----
{{.Struct.Decl}} {
{{- range .Struct.Fields}}
	{{if .Nested}}{{.Nested.Name}}{{"\t"}}struct{{else}}{{tabify .Decl}}{{end}}
{{- end}}
}
----

{{.Struct.Doc}}
{{range .Struct.Fields}}{{if not .Nested}}
==== {{.Decl}}
{{.Doc}}
{{- end}}
{{end}}
{{range .Struct.Fields}}{{if .Nested}}{{render $ .Nested}}{{end}}{{end}}
{{- if hasReceivers . .Struct.Name}}{{renderReceivers . .Struct.Name}}{{end}}
`,
		ReceiversTemplate.String(): `==== Receivers
{{range .Receiver}}
===== {{.Name}}
[source, go]
----
{{ .Decl }}
----

{{.Doc}}
{{- end}}
`,
	}).NewContextWithConfig(f, nil, &TemplateContextConfig{IncludeMethodCode: false})

	x.RenderStruct(&buf, f.Structs[0])

	fmt.Println(buf.String())
}
