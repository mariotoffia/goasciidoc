package asciidoc

import (
	"bytes"
	"fmt"
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
		"package": `==  {{ declaration .File }}
{{ .File.Doc }}`,
	}).NewContext(f)

	x.RenderPackage(&buf)

	fmt.Println(buf.String())
}
