package goparser

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Contains all issues, reported on GitHub, for the goparser

// Issue15 is a bug that notes that ArrayType is not implemented
// `not-implemented typeSpec.Type.(type) = *ast.ArrayType`.
func TestIssue15(t *testing.T) {
	src := `package foo

	type Color string
	type Label struct {
		Name        string
		Description string
		color       Color
	}
	
	// LabelSet is a custom slice type
	type LabelSet []Label`

	m := dummyModule()
	f, err := ParseInlineFile(m, m.Base+"/mypkg/file.go", src)
	require.NoError(t, err)
	require.NotNil(t, f)

	assert.Equal(t, "[]Label", f.CustomTypes[1].Type)
	assert.Equal(t, "type LabelSet []Label", f.CustomTypes[1].Decl)
	assert.Equal(t, "LabelSet is a custom slice type", f.CustomTypes[1].Doc)

}
