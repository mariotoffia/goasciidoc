package asciidoc

import (
	"bytes"
	"strings"
	"testing"

	"github.com/mariotoffia/goasciidoc/goparser"
	"github.com/stretchr/testify/require"
)

func TestInterfaceMethodHeadingPlainWithSpacing(t *testing.T) {
	const code = `package sample

import "github.com/acme/coremodel"

type Call interface {
	// GetOperation retrieves the operation.
	GetOperation() coremodel.Operation
	// GetPath resolves the path.
	GetPath() string
}`

	goFile, err := goparser.ParseInlineFile(nil, "", code)
	require.NoError(t, err)

	overrides := loadTemplateOverrides(t, InterfaceTemplate)
	tmpl := NewTemplateWithOverrides(overrides)
	ctx := tmpl.NewContext(goFile)
	ctx.Config.SignatureStyle = "goasciidoc"
	ctx.Config.TypeLinks = TypeLinksInternalExternal

	require.Len(t, goFile.Interfaces, 1)

	var buf bytes.Buffer
	ctx.RenderInterface(&buf, goFile.Interfaces[0])
	doc := buf.String()

	// Headings should never include inline highlight spans.
	require.NotContains(t, doc, "==== <span")

	// Highlight block must exist for the signature.
	require.Contains(t, doc, `<div class="listingblock signature">`)

	// Ensure the heading uses the plain signature string with external type link.
	require.Contains(
		t,
		doc,
		"==== GetOperation() link:https://pkg.go.dev/github.com/acme/coremodel#Operation[coremodel.Operation]",
	)

	// Ensure the highlight block is separated from the next heading by at least one blank line.
	require.NotContains(t, doc, "++++====")
	require.Contains(t, doc, "++++\n\n")
	require.Contains(t, doc, "\n\n==== GetPath() string")

	// Validate subsequent headings only contain plain text.
	for _, line := range strings.Split(doc, "\n") {
		if strings.HasPrefix(line, "==== ") {
			require.NotContains(t, line, "<span", "heading contains highlight markup")
		}
	}
}
