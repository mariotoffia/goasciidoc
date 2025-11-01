package asciidoc

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mariotoffia/goasciidoc/goparser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRenderingIncludesTypeParametersAndTypeSets(t *testing.T) {
	const code = `package sample

import "io"

type Pair[K comparable, V any] struct {
	Key   K
	Value V
}

type List[T any] []T

type Constraint[T any] interface {
	~[]T | *List[T]
	io.Reader
	Do(T) error
}

type Set[T any] struct{}

func Transform[T any](in T) T {
	return in
}

type Mapper[K comparable, V any] func(K) V
`

	goFile, err := goparser.ParseInlineFile(nil, "", code)
	require.NoError(t, err)

	overrides := loadTemplateOverrides(t, StructTemplate, InterfaceTemplate, FunctionTemplate, CustomFuncTypeDefTemplate)
	tmpl := NewTemplateWithOverrides(overrides)
	ctx := tmpl.NewContext(goFile)

	require.NotEmpty(t, goFile.Structs)
	require.Len(t, goFile.Structs[0].TypeParams, 2)
	var structBuf bytes.Buffer
	ctx.RenderStruct(&structBuf, goFile.Structs[0])
	structDoc := structBuf.String()
	require.Contains(t, structDoc, "=== Pair[K comparable, V any]")
	require.Contains(t, structDoc, "type Pair[K comparable, V any] struct")

	require.NotEmpty(t, goFile.Interfaces)
	require.Len(t, goFile.Interfaces[0].TypeSet, 3)
	var ifaceBuf bytes.Buffer
	ctx.RenderInterface(&ifaceBuf, goFile.Interfaces[0])
	ifaceDoc := ifaceBuf.String()
	require.Contains(t, ifaceDoc, "=== Constraint[T any]")
	require.Contains(t, ifaceDoc, "~[]T | *List[T]")
	require.Contains(t, ifaceDoc, "io.Reader")
	require.Contains(t, ifaceDoc, "* `~[]T`")
	require.Contains(t, ifaceDoc, "* `*List[T]`")
	require.Contains(t, ifaceDoc, "* `io.Reader`")

	require.NotEmpty(t, goFile.StructMethods)
	transform := findMethodByName(goFile.StructMethods, "Transform")
	require.NotNil(t, transform)
	var funcBuf bytes.Buffer
	ctx.RenderFunction(&funcBuf, transform)
	funcDoc := funcBuf.String()
	t.Log(funcDoc)
	require.Contains(t, funcDoc, "=== Transform[T any]")
	require.Contains(t, funcDoc, "<span class=\"hljs-keyword\">func</span> Transform[T any]")
	require.Contains(t, funcDoc, "<pre class=\"highlightjs highlight\"><code class=\"language-go hljs\">")

	require.NotEmpty(t, goFile.CustomFuncs)
	mapper := findCustomFuncByName(goFile.CustomFuncs, "Mapper")
	require.NotNil(t, mapper)
	var typeBuf bytes.Buffer
	ctx.RenderTypeDefFunc(&typeBuf, mapper)
	typeDoc := typeBuf.String()
	require.Contains(t, typeDoc, "=== Mapper[K comparable, V any]")
	require.Contains(t, typeDoc, "<span class=\"hljs-keyword\">type</span> Mapper[K comparable, V any] <span class=\"hljs-keyword\">func</span>")
	require.Contains(t, typeDoc, "<pre class=\"highlightjs highlight\"><code class=\"language-go hljs\">")
}

func TestInterfaceWithoutTypeSetOmitsSection(t *testing.T) {
	const code = `package sample

type NoSet interface {
	Do()
}
`

	goFile, err := goparser.ParseInlineFile(nil, "", code)
	require.NoError(t, err)

	overrides := loadTemplateOverrides(t, InterfaceTemplate)
	tmpl := NewTemplateWithOverrides(overrides)
	ctx := tmpl.NewContext(goFile)

	require.Len(t, goFile.Interfaces, 1)
	var buf bytes.Buffer
	ctx.RenderInterface(&buf, goFile.Interfaces[0])
	doc := buf.String()
	require.Contains(t, doc, "=== NoSet")
	assert.NotContains(t, doc, "==== Type Set")
}

func loadTemplateOverrides(t *testing.T, types ...TemplateType) map[string]string {
	include := append([]TemplateType{SignatureTemplate}, types...)
	overrides := make(map[string]string, len(include))
	seen := map[TemplateType]struct{}{}
	for _, tt := range include {
		if _, exists := seen[tt]; exists {
			continue
		}
		seen[tt] = struct{}{}
		path := filepath.Join("..", "defaults", tt.String()+".gtpl")
		data, err := os.ReadFile(path)
		require.NoErrorf(t, err, "failed to read template %s", path)
		overrides[tt.String()] = string(data)
	}
	return overrides
}

func findMethodByName(methods []*goparser.GoStructMethod, name string) *goparser.GoStructMethod {
	for _, m := range methods {
		if strings.EqualFold(m.Name, name) {
			return m
		}
	}
	return nil
}

func findCustomFuncByName(methods []*goparser.GoMethod, name string) *goparser.GoMethod {
	for _, m := range methods {
		if strings.EqualFold(m.Name, name) {
			return m
		}
	}
	return nil
}
