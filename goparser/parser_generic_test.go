package goparser

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStructWithSingleGenericArgument(t *testing.T) {

	src := `package foo

// MyType accepts int or string for the _Foo_ property
type MyType[T int | string] struct {
	Foo T
}`

	m := dummyModule()
	f, err := ParseInlineFile(m, m.Base+"/mypkg/file.go", src)
	assert.NoError(t, err)
	require.NotNil(t, f)
}

func TestStructWithSingleGenericArgumentAndSimpleReceiverMethod(t *testing.T) {

	src := `package foo

// MyType accepts int or string for the _Foo_ property
type MyType[T int | string] struct {
	Foo T
}

func (m *MyType[T]) Bar() *MyType[T] {
	return m
}`

	m := dummyModule()
	f, err := ParseInlineFile(m, m.Base+"/mypkg/file.go", src)
	assert.NoError(t, err)
	require.NotNil(t, f)
}
