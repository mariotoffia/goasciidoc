package goparser

import (
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParsePackageDoc(t *testing.T) {
	src := `
// The package foo is a sample package.
package foo`

	f, err := ParseInlineFile(src)
	assert.Equal(t, nil, err)

	if err != nil {
		log.Fatal(err)
	}

	assert.Equal(t, "foo", f.Package)
	assert.Equal(t, "The package foo is a sample package.\n", f.Doc)
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

	assert.Equal(t, "Importing fmt before time\n", f.Imports[0].Doc)
	assert.Equal(t, "This is the time import\n", f.Imports[1].Doc)
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

	assert.Equal(t, "bar is a private function that prints out current time\n", f.StructMethods[0].Doc)
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

	assert.Equal(t, "Bar is a private function that prints out current time\n", f.StructMethods[0].Doc)
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

	assert.Equal(t, "Bar is a private function that prints out current time\n\nThis function is exported!\n", f.StructMethods[0].Doc)
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

	assert.Equal(t, " Bar is a private function that prints out current time\n   This function is exported!\n", f.StructMethods[0].Doc)
}

func TestInterfaceDefinitionComment(t *testing.T) {
	src := `package foo

// IInterface is a interface comment
type IInterface interface {
	Name() string
}`

	f, err := ParseInlineFile(src)
	assert.Equal(t, nil, err)

	assert.Equal(t, "IInterface is a interface comment\n", f.Interfaces[0].Doc)
}

func TestInterfaceMethodComment(t *testing.T) {
	src := `package foo
type IInterface interface {
	// Name returns the name of the thing
	Name() string
}`

	f, err := ParseInlineFile(src)
	assert.Equal(t, nil, err)

	assert.Equal(t, "Name returns the name of the thing\n", f.Interfaces[0].Methods[0].Doc)
}

func TestStructDefinitionComment(t *testing.T) {

	src := `package foo

// MyStruct is a structure of nonsense
type MyStruct struct {}`

	f, err := ParseInlineFile(src)
	assert.Equal(t, nil, err)

	assert.Equal(t, "MyStruct is a structure of nonsense\n", f.Structs[0].Doc)
}

func TestStructFieldComment(t *testing.T) {

	src := `package foo

type MyStruct struct {
	// Name is a fine name inside MyStruct
	Name string
}`

	f, err := ParseInlineFile(src)
	assert.Equal(t, nil, err)

	assert.Equal(t, "Name is a fine name inside MyStruct\n", f.Structs[0].Fields[0].Doc)
}

func TestCustomType(t *testing.T) {

	src := `package foo

// This is a simple custom type
type MyType int`

	f, err := ParseInlineFile(src)
	assert.Equal(t, nil, err)

	assert.Equal(t, "This is a simple custom type\n", f.CustomTypes[0].Doc)
	assert.Equal(t, "MyType", f.CustomTypes[0].Name)
	assert.Equal(t, "int", f.CustomTypes[0].Type)
}

func TestVarDeclr(t *testing.T) {
	src := `package foo

// This is a simple variable declaration
var pelle, anna = 17, 19`

	f, err := ParseInlineFile(src)
	assert.Equal(t, nil, err)

	assert.Equal(t, "This is a simple variable declaration\n", f.VarAssigments[0].Doc)
	assert.Equal(t, "pelle", f.VarAssigments[0].Name)
}