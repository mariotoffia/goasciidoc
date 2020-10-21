package goparser

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func dummyModule() *GoModule {
	mod, _ := NewModuleFromBuff("/tmp/test-asciidoc/go.mod",
		[]byte(`module github.com/mariotoffia/goasciidoc/tests
	go 1.14`))
	mod.Version = "0.0.1"

	return mod
}

func TestParsePackageDoc(t *testing.T) {
	src := `
// The package foo is a sample package.
package foo`

	m := dummyModule()
	f, err := ParseInlineFile(m, m.Base+"/mypkg/file.go", src)
	assert.NoError(t, err)

	assert.Equal(t, "foo", f.Package)
	assert.Equal(t, "github.com/mariotoffia/goasciidoc/tests/mypkg", f.FqPackage)
	assert.Equal(t, "The package foo is a sample package.", f.Doc)
	assert.Equal(t, "package foo", f.Decl)
	assert.Equal(t, "/tmp/test-asciidoc/mypkg/file.go", f.FilePath)
}

func TestParseImportDoc(t *testing.T) {
	src := `package foo

import (
	// Importing fmt before time
	"fmt"
	// This is the time import
	"time"
)

func bar() {
	fmt.Println(time.Now())
}`

	m := dummyModule()
	f, err := ParseInlineFile(m, m.Base+"/mypkg/file.go", src)
	assert.NoError(t, err)

	assert.Equal(t, "Importing fmt before time", f.Imports[0].Doc)
	assert.Equal(t, "This is the time import", f.Imports[1].Doc)
	assert.Equal(t, "import (\n\t\"fmt\"\n\t\"time\"\n)", f.DeclImports())
	assert.Equal(t, "import (\n\t// Importing fmt before time\n\t\"fmt\"\n\t// This is the time import\n\t\"time\"\n)", f.ImportFullDecl)
}

func TestParsePrivateFunction(t *testing.T) {
	src := `package foo
import ( 
	"fmt" 
	"time" 
)

// bar is a private function that prints out current time
func bar() {
	fmt.Println(time.Now())
}`

	m := dummyModule()
	f, err := ParseInlineFile(m, m.Base+"/mypkg/file.go", src)
	assert.NoError(t, err)

	assert.Equal(t, "bar is a private function that prints out current time", f.StructMethods[0].Doc)
	assert.Equal(t, "func bar()", f.StructMethods[0].Decl)
	assert.Equal(t, "func bar() {\n\tfmt.Println(time.Now())\n}", f.StructMethods[0].FullDecl)
}

func TestParseExportedFunction(t *testing.T) {
	src := `package foo
import ( 
	"fmt" 
	"time" 
)

// Bar is a exported function that prints out current time
func Bar() {
	fmt.Println(time.Now())
}`

	m := dummyModule()
	f, err := ParseInlineFile(m, m.Base+"/mypkg/file.go", src)
	assert.NoError(t, err)

	assert.Equal(t, "Bar is a exported function that prints out current time", f.StructMethods[0].Doc)
}

func TestParseMultilineCppStyleComment(t *testing.T) {
	src := `package foo
import ( 
	"fmt" 
	"time" 
)

// Bar is a private function that prints out current time
//
// This function is exported!
func Bar() {
	fmt.Println(time.Now())
}`

	m := dummyModule()
	f, err := ParseInlineFile(m, m.Base+"/mypkg/file.go", src)
	assert.NoError(t, err)

	assert.Equal(t, "Bar is a private function that prints out current time\n\nThis function is exported!", f.StructMethods[0].Doc)
}

func TestParseMultilineCStyleComment(t *testing.T) {
	src := `package foo
import ( 
	"fmt" 
	"time" 
)

/* Bar is a private function that prints out current time
   This function is exported!
 */
func Bar() {
	fmt.Println(time.Now())
}`

	m := dummyModule()
	f, err := ParseInlineFile(m, m.Base+"/mypkg/file.go", src)
	assert.NoError(t, err)

	assert.Equal(t, " Bar is a private function that prints out current time\n   This function is exported!", f.StructMethods[0].Doc)
}

func TestInterfaceDefinitionComment(t *testing.T) {
	src := `package foo

// IInterface is a interface comment
type IInterface interface {
	Name() string
}`

	m := dummyModule()
	f, err := ParseInlineFile(m, m.Base+"/mypkg/file.go", src)
	assert.NoError(t, err)

	assert.Equal(t, "IInterface is a interface comment", f.Interfaces[0].Doc)
	assert.Equal(t, "type IInterface interface", f.Interfaces[0].Decl)
	assert.Equal(t, "type IInterface interface {\n\tName() string\n}", f.Interfaces[0].FullDecl)
}

func TestInterfaceMethodComment(t *testing.T) {
	src := `package foo
type IInterface interface {
	// Name returns the name of the thing
	Name() string
}`

	m := dummyModule()
	f, err := ParseInlineFile(m, m.Base+"/mypkg/file.go", src)
	assert.NoError(t, err)

	assert.Equal(t, "Name returns the name of the thing", f.Interfaces[0].Methods[0].Doc)
	assert.Equal(t, "Name() string", f.Interfaces[0].Methods[0].Decl)
	assert.Equal(t, "Name() string", f.Interfaces[0].Methods[0].FullDecl)
}

func TestStructDefinitionComment(t *testing.T) {

	src := `package foo

// MyStruct is a structure of nonsense
type MyStruct struct {}`

	m := dummyModule()
	f, err := ParseInlineFile(m, m.Base+"/mypkg/file.go", src)
	assert.NoError(t, err)

	assert.Equal(t, "MyStruct is a structure of nonsense", f.Structs[0].Doc)
}

func TestStructFieldComment(t *testing.T) {

	src := `package foo

type MyStruct struct {
	// Name is a fine name inside MyStruct
	Name string
}`

	m := dummyModule()
	f, err := ParseInlineFile(m, m.Base+"/mypkg/file.go", src)
	assert.NoError(t, err)

	assert.Equal(t, "Name is a fine name inside MyStruct", f.Structs[0].Fields[0].Doc)
	assert.Equal(t, "string", f.Structs[0].Fields[0].Type)
	assert.Equal(t, "Name string", f.Structs[0].Fields[0].Decl)
}

func TestNestedAnonymousStructDefinitionComment(t *testing.T) {

	src := `package foo

// MyStruct is a structure of nonsense
type MyStruct struct {
	// Inline the struct
	Inline struct {
		// FooBar is a even more nonsense variable
		FooBar int
	}
}`

	m := dummyModule()
	f, err := ParseInlineFile(m, m.Base+"/mypkg/file.go", src)
	assert.NoError(t, err)

	assert.Equal(t, "MyStruct is a structure of nonsense", f.Structs[0].Doc)
	assert.Equal(t, "MyStruct", f.Structs[0].Name)
	assert.Equal(t, "type MyStruct struct", f.Structs[0].Decl)
	assert.Equal(t, "type MyStruct struct {\n\t// Inline the struct\n\tInline struct {\n\t\t// FooBar is a even more nonsense variable\n\t\tFooBar int\n\t}\n}", f.Structs[0].FullDecl)

	assert.Equal(t, "Inline the struct", f.Structs[0].Fields[0].Doc)
	assert.Equal(t, "Inline", f.Structs[0].Fields[0].Name)
	assert.Equal(t, "Inline struct {\n\t\t// FooBar is a even more nonsense variable\n\t\tFooBar int\n\t}", f.Structs[0].Fields[0].Decl)

	assert.Equal(t, "FooBar is a even more nonsense variable", f.Structs[0].Fields[0].Nested.Fields[0].Doc)
	assert.Equal(t, "FooBar", f.Structs[0].Fields[0].Nested.Fields[0].Name)
	assert.Equal(t, "FooBar int", f.Structs[0].Fields[0].Nested.Fields[0].Decl)
}

func TestNestedStructDefinitionComment(t *testing.T) {

	src := `package foo

// Inline is a struct to be embedded.
type Inline struct {
	// FooBar is a even more nonsense variable
	FooBar int
}

// MyStruct is a structure of nonsense
type MyStruct struct {
	// Inline the struct
	Ins Inline
}`

	m := dummyModule()
	f, err := ParseInlineFile(m, m.Base+"/mypkg/file.go", src)
	assert.NoError(t, err)

	assert.Equal(t, "Inline is a struct to be embedded.", f.Structs[0].Doc)
	assert.Equal(t, "FooBar is a even more nonsense variable", f.Structs[0].Fields[0].Doc)
	assert.Equal(t, "FooBar", f.Structs[0].Fields[0].Name)
	assert.Equal(t, "FooBar int", f.Structs[0].Fields[0].Decl)

	assert.Equal(t, "MyStruct is a structure of nonsense", f.Structs[1].Doc)
	assert.Equal(t, "MyStruct", f.Structs[1].Name)
	assert.Equal(t, "type MyStruct struct", f.Structs[1].Decl)
	assert.Equal(t, "type MyStruct struct {\n\t// Inline the struct\n\tIns Inline\n}", f.Structs[1].FullDecl)

	assert.Equal(t, "Inline the struct", f.Structs[1].Fields[0].Doc)
	assert.Equal(t, "Ins", f.Structs[1].Fields[0].Name)
	assert.Equal(t, "Ins Inline", f.Structs[1].Fields[0].Decl)
}

func TestCustomTypePrimitive(t *testing.T) {

	src := `package foo

// This is a simple custom type
type MyType int`

	m := dummyModule()
	f, err := ParseInlineFile(m, m.Base+"/mypkg/file.go", src)
	assert.NoError(t, err)

	assert.Equal(t, "This is a simple custom type", f.CustomTypes[0].Doc)
	assert.Equal(t, "MyType", f.CustomTypes[0].Name)
	assert.Equal(t, "int", f.CustomTypes[0].Type)
	assert.Equal(t, "type MyType int", f.CustomTypes[0].Decl)
}

func TestCustomTypeStructType(t *testing.T) {

	src := `package foo

import "time"
// This is a struct custom type
type MyType time.Time`

	m := dummyModule()
	f, err := ParseInlineFile(m, m.Base+"/mypkg/file.go", src)
	assert.NoError(t, err)

	assert.Equal(t, "This is a struct custom type", f.CustomTypes[0].Doc)
	assert.Equal(t, "MyType", f.CustomTypes[0].Name)
	assert.Equal(t, "time.Time", f.CustomTypes[0].Type)
	assert.Equal(t, "type MyType time.Time", f.CustomTypes[0].Decl)
}

func TestCustomFunctionDefinition(t *testing.T) {

	src := `package foo

// This is a simple custom function to walk around with
type ParseWalkerFunc func(int) error`

	m := dummyModule()
	f, err := ParseInlineFile(m, m.Base+"/mypkg/file.go", src)
	assert.NoError(t, err)
	assert.Equal(t, "This is a simple custom function to walk around with", f.CustomFuncs[0].Doc)
	assert.Equal(t, "ParseWalkerFunc", f.CustomFuncs[0].Name)
	assert.Equal(t, "type ParseWalkerFunc func(int) error", f.CustomFuncs[0].Decl)
	assert.Equal(t, "type ParseWalkerFunc func(int) error", f.CustomFuncs[0].FullDecl)
}

func TestSingleLineMultiVarDeclaration(t *testing.T) {
	src := `package foo

// This is a simple variable declaration
var pelle, anna = 17, 19`

	m := dummyModule()
	f, err := ParseInlineFile(m, m.Base+"/mypkg/file.go", src)
	assert.NoError(t, err)

	assert.Equal(t, "This is a simple variable declaration", f.VarAssignments[0].Doc)
	assert.Equal(t, "pelle", f.VarAssignments[0].Name)
	assert.Equal(t, "var pelle, anna = 17, 19", f.VarAssignments[0].Decl)
	assert.Equal(t, "var pelle, anna = 17, 19", f.VarAssignments[0].FullDecl)
	assert.Equal(t, "This is a simple variable declaration", f.VarAssignments[1].Doc)
	assert.Equal(t, "anna", f.VarAssignments[1].Name)
	assert.Equal(t, "var pelle, anna = 17, 19", f.VarAssignments[1].Decl)
	assert.Equal(t, "var pelle, anna = 17, 19", f.VarAssignments[1].FullDecl)
}

func TestPrimitiveConst(t *testing.T) {
	src := `package foo

const (
	// Bubben is a int of one
	Bubben int = 1
)`

	m := dummyModule()
	f, err := ParseInlineFile(m, m.Base+"/mypkg/file.go", src)
	assert.NoError(t, err)
	assert.Equal(t, "Bubben is a int of one", f.ConstAssignments[0].Doc)
	assert.Equal(t, "Bubben", f.ConstAssignments[0].Name)
	assert.Equal(t, "Bubben int = 1", f.ConstAssignments[0].Decl)
	assert.Equal(t, "const (\n\t// Bubben is a int of one\n\tBubben int = 1\n)", f.ConstAssignments[0].FullDecl)
}

func TestMultiplePrimitiveConst(t *testing.T) {
	src := `package foo

const (
	// Bubben is a int of one
	Bubben int = 1
	// Apan is next to come
	Apan int = 4
)`

	m := dummyModule()
	f, err := ParseInlineFile(m, m.Base+"/mypkg/file.go", src)
	assert.NoError(t, err)
	assert.Equal(t, "Bubben is a int of one", f.ConstAssignments[0].Doc)
	assert.Equal(t, "Bubben", f.ConstAssignments[0].Name)
	assert.Equal(t, "Apan is next to come", f.ConstAssignments[1].Doc)
	assert.Equal(t, "Apan", f.ConstAssignments[1].Name)
}

func TestCustomTypeConst(t *testing.T) {
	src := `package foo

// Apan is a custom type
type Apan int

const (
	// Bubben is first to come
	Bubben Apan = iota
	// Next, crying out loud, is Olle
	GrinOlle
)`

	m := dummyModule()
	f, err := ParseInlineFile(m, m.Base+"/mypkg/file.go", src)
	assert.NoError(t, err)

	assert.Equal(t, "Apan is a custom type", f.CustomTypes[0].Doc)
	assert.Equal(t, "Apan", f.CustomTypes[0].Name)
	assert.Equal(t, "int", f.CustomTypes[0].Type)

	assert.Equal(t, "Bubben is first to come", f.ConstAssignments[0].Doc)
	assert.Equal(t, "Bubben", f.ConstAssignments[0].Name)
	assert.Equal(t, "Next, crying out loud, is Olle", f.ConstAssignments[1].Doc)
	assert.Equal(t, "GrinOlle", f.ConstAssignments[1].Name)
}

func TestVarInsideCodeIsDiscarded(t *testing.T) {
	src := `package foo

func boo() {
	var DiscardMe int
	DiscardMe = 9
	if DiscardMe != 9 {
		return
	}
}
`

	m := dummyModule()
	f, err := ParseInlineFile(m, m.Base+"/mypkg/file.go", src)
	assert.NoError(t, err)
	assert.Equal(t, "boo", f.StructMethods[0].Name)
	assert.Equal(t, 0, len(f.VarAssignments))
}

func TestParseStructFunction(t *testing.T) {
	src := `package foo
import ( 
	"fmt" 
	"time" 
)

type MyStruct struct {}

// Bar is a method bound to MyStruct
func (ms *MyStruct) Bar() string {
	fmt.Println(time.Now())
	return "now"
}`

	m := dummyModule()
	f, err := ParseInlineFile(m, m.Base+"/mypkg/file.go", src)
	assert.NoError(t, err)

	assert.Equal(t, "Bar is a method bound to MyStruct", f.StructMethods[0].Doc)
	assert.Equal(t, "func (ms *MyStruct) Bar() string", f.StructMethods[0].Decl)
	assert.Equal(t, "*MyStruct", f.StructMethods[0].Receivers[0])
}

func TestFunctionBoundToStruct(t *testing.T) {
	src := `package foo

// Apan is a custom type
type Apan int

func (a Apan) DoWork(msg string) string {
	return "hello: " + msg
}
`

	m := dummyModule()
	f, err := ParseInlineFile(m, m.Base+"/mypkg/file.go", src)
	assert.NoError(t, err)
	assert.Equal(t, "Apan", f.StructMethods[0].Receivers[0])
	assert.Equal(t, "DoWork", f.StructMethods[0].Name)

	fmt.Println(f.StructMethods[0])
}
