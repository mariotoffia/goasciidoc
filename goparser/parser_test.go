package goparser

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParsePackageDoc(t *testing.T) {
	src := `
// The package foo is a sample package.
package foo`

	f, err := ParseInlineFile(src)
	assert.Equal(t, nil, err)

	assert.Equal(t, "foo", f.Package)
	assert.Equal(t, "The package foo is a sample package.", f.Doc)
	assert.Equal(t, "package foo", f.Decl)
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

	f, err := ParseInlineFile(src)
	assert.Equal(t, nil, err)

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

	f, err := ParseInlineFile(src)
	assert.Equal(t, nil, err)

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

// Bar is a private function that prints out current time
func Bar() {
	fmt.Println(time.Now())
}`

	f, err := ParseInlineFile(src)
	assert.Equal(t, nil, err)

	assert.Equal(t, "Bar is a private function that prints out current time", f.StructMethods[0].Doc)
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

	f, err := ParseInlineFile(src)
	assert.Equal(t, nil, err)

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

	f, err := ParseInlineFile(src)
	assert.Equal(t, nil, err)

	assert.Equal(t, " Bar is a private function that prints out current time\n   This function is exported!", f.StructMethods[0].Doc)
}

func TestInterfaceDefinitionComment(t *testing.T) {
	src := `package foo

// IInterface is a interface comment
type IInterface interface {
	Name() string
}`

	f, err := ParseInlineFile(src)
	assert.Equal(t, nil, err)

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

	f, err := ParseInlineFile(src)
	assert.Equal(t, nil, err)

	assert.Equal(t, "Name returns the name of the thing", f.Interfaces[0].Methods[0].Doc)
	assert.Equal(t, "Name() string", f.Interfaces[0].Methods[0].Decl)
	assert.Equal(t, "Name() string", f.Interfaces[0].Methods[0].FullDecl)
}

func TestStructDefinitionComment(t *testing.T) {

	src := `package foo

// MyStruct is a structure of nonsense
type MyStruct struct {}`

	f, err := ParseInlineFile(src)
	assert.Equal(t, nil, err)

	assert.Equal(t, "MyStruct is a structure of nonsense", f.Structs[0].Doc)
}

func TestStructFieldComment(t *testing.T) {

	src := `package foo

type MyStruct struct {
	// Name is a fine name inside MyStruct
	Name string
}`

	f, err := ParseInlineFile(src)
	assert.Equal(t, nil, err)

	assert.Equal(t, "Name is a fine name inside MyStruct", f.Structs[0].Fields[0].Doc)
	assert.Equal(t, "string", f.Structs[0].Fields[0].Type)
	assert.Equal(t, "Name string", f.Structs[0].Fields[0].Decl)
}

func TestCustomType(t *testing.T) {

	src := `package foo

// This is a simple custom type
type MyType int`

	f, err := ParseInlineFile(src)
	assert.Equal(t, nil, err)

	assert.Equal(t, "This is a simple custom type", f.CustomTypes[0].Doc)
	assert.Equal(t, "MyType", f.CustomTypes[0].Name)
	assert.Equal(t, "int", f.CustomTypes[0].Type)
	assert.Equal(t, "type MyType int", f.CustomTypes[0].Decl)
}

func TestCustomFunctionDefinition(t *testing.T) {

	src := `package foo

// This is a simple custom function to walk around with
type ParseWalkerFunc func(int) error`

	f, err := ParseInlineFile(src)
	assert.Equal(t, nil, err)
	assert.Equal(t, "This is a simple custom function to walk around with", f.CustomFuncs[0].Doc)
	assert.Equal(t, "ParseWalkerFunc", f.CustomFuncs[0].Name)
	assert.Equal(t, "type ParseWalkerFunc func(int) error", f.CustomFuncs[0].Decl)
	assert.Equal(t, "type ParseWalkerFunc func(int) error", f.CustomFuncs[0].FullDecl)
}

func TestSingleLineMultiVarDeclr(t *testing.T) {
	src := `package foo

// This is a simple variable declaration
var pelle, anna = 17, 19`

	f, err := ParseInlineFile(src)
	assert.Equal(t, nil, err)

	assert.Equal(t, "This is a simple variable declaration", f.VarAssigments[0].Doc)
	assert.Equal(t, "pelle", f.VarAssigments[0].Name)
	assert.Equal(t, "var pelle, anna = 17, 19", f.VarAssigments[0].Decl)
	assert.Equal(t, "var pelle, anna = 17, 19", f.VarAssigments[0].FullDecl)
	assert.Equal(t, "This is a simple variable declaration", f.VarAssigments[1].Doc)
	assert.Equal(t, "anna", f.VarAssigments[1].Name)
	assert.Equal(t, "var pelle, anna = 17, 19", f.VarAssigments[1].Decl)
	assert.Equal(t, "var pelle, anna = 17, 19", f.VarAssigments[1].FullDecl)
}

func TestPrimitiveConst(t *testing.T) {
	src := `package foo

const (
	// Bubben is a int of one
	Bubben int = 1
)`

	f, err := ParseInlineFile(src)
	assert.Equal(t, nil, err)
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

	f, err := ParseInlineFile(src)
	assert.Equal(t, nil, err)
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

	f, err := ParseInlineFile(src)
	assert.Equal(t, nil, err)

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

	f, err := ParseInlineFile(src)
	assert.Equal(t, nil, err)
	assert.Equal(t, "boo", f.StructMethods[0].Name)
	assert.Equal(t, 0, len(f.VarAssigments))
}
