package asciidoc

import (
	"bytes"
	"testing"

	"github.com/mariotoffia/goasciidoc/goparser"
	"github.com/stretchr/testify/assert"
)

func TestRenderPackage(t *testing.T) {
	src := `
	// The package foo is a sample package.
	package foo`

	f, err := goparser.ParseInlineFile(src)
	assert.Equal(t, nil, err)

	var buf bytes.Buffer

	x := NewTemplateWithOverrides(map[string]string{
		PackageTemplate.String(): `== {{ declaration .File }}
{{ .File.Doc }}`,
	}).NewContext(f)

	x.RenderPackage(&buf)

	assert.Equal(t, "== package foo\nThe package foo is a sample package.\n", buf.String())
}

func TestRenderImports(t *testing.T) {
	src := `	
	package foo
	
	import (
		// We import format here
		"fmt"
		// and time here :)
		"time"
	)
	func Bar() {
		fmt.Println(time.Now())
	}`

	f, err := goparser.ParseInlineFile(src)
	assert.Equal(t, nil, err)

	var buf bytes.Buffer

	x := NewTemplateWithOverrides(map[string]string{
		ImportTemplate.String(): `== Imports
[source, go]
----
{{ declaration .File }}
----

{{range .File.Imports}}{{if .Doc }}=== Import _{{ .Path }}_{{ cr }}{{ .Doc }}{{ cr }}{{ cr }}{{end}}{{end}}`,
	}).NewContext(f)

	x.RenderImports(&buf)

	assert.Equal(t,
		"== Imports\n[source, go]\n----\nimport (\n\t\"fmt\"\n\t\"time\"\n)\n----\n\n=== Import _fmt_\nWe import format here\n\n=== Import _time_\nand time here :)\n\n",
		buf.String())
}
