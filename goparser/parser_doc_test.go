package goparser

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDocConcatinationModes(t *testing.T) {
	const source = `package sample

// Summary line
// providing high level context.


// .Example Usage
// demonstrate how to call the function.
func DoThing() {}
`

	fileNone, err := ParseInlineFileWithConfig(ParseConfig{}, "", source)
	require.NoError(t, err)
	require.Len(t, fileNone.StructMethods, 1)
	docNone := fileNone.StructMethods[0].Doc
	require.NotContains(t, docNone, "Summary line")

	fileFull, err := ParseInlineFileWithConfig(ParseConfig{
		DocConcatination: DocConcatinationFull,
	}, "", source)
	require.NoError(t, err)
	require.Len(t, fileFull.StructMethods, 1)
	docFull := fileFull.StructMethods[0].Doc

	require.Contains(t, docFull, ".Example Usage")
	require.True(t, strings.Contains(docFull, "Summary line"))
}
