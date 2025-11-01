package asciidoc

import (
	"strings"
	"testing"

	"github.com/mariotoffia/goasciidoc/goparser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testContextWithMode(mode TypeLinkMode) *TemplateContext {
	tmpl := NewTemplateWithOverrides(nil)
	mod := &goparser.GoModule{Name: "example.com/mod"}
	file := &goparser.GoFile{
		Module:    mod,
		FqPackage: "example.com/mod/pkg",
		Package:   "pkg",
	}
	pkg := &goparser.GoPackage{}
	pkg.Module = mod
	pkg.FqPackage = file.FqPackage
	ctx := tmpl.NewContextWithConfig(file, pkg, &TemplateContextConfig{TypeLinks: mode})
	return ctx
}

func TestFieldSummaryLinksInternalType(t *testing.T) {
	ctx := testContextWithMode(TypeLinksInternal)

	widget := &goparser.GoStruct{Name: "Widget", File: ctx.File}
	container := &goparser.GoStruct{Name: "Container", File: ctx.File}
	ctx.Struct = container
	ctx.Package.Structs = []*goparser.GoStruct{container, widget}

	inner := &goparser.GoType{File: ctx.File, Type: "Widget", Kind: goparser.TypeKindIdent}
	pointer := &goparser.GoType{File: ctx.File, Type: "*Widget", Kind: goparser.TypeKindPointer, Inner: []*goparser.GoType{inner}}

	field := &goparser.GoField{
		Struct:   container,
		File:     ctx.File,
		Name:     "Child",
		Type:     "*Widget",
		Decl:     "Child *Widget",
		TypeInfo: pointer,
	}

	got := ctx.fieldSummary(field)
	expected := "Child\t*<<" + anchorID(ctx.File.FqPackage, "Widget") + ",Widget>>"
	assert.Equal(t, expected, got)
}

func TestFieldSummaryExternalTypeLink(t *testing.T) {
	ctx := testContextWithMode(TypeLinksInternalExternal)
	ctx.File.Imports = []*goparser.GoImport{{Name: "foo", Path: "github.com/acme/lib"}}

	inner := &goparser.GoType{File: ctx.File, Type: "foo.Bar", Kind: goparser.TypeKindSelector}
	field := &goparser.GoField{
		Struct:   &goparser.GoStruct{},
		File:     ctx.File,
		Name:     "Ext",
		Type:     "foo.Bar",
		Decl:     "Ext foo.Bar",
		TypeInfo: inner,
	}

	got := ctx.fieldSummary(field)
	expected := "Ext\tlink:https://pkg.go.dev/github.com/acme/lib#Bar[foo.Bar]"
	assert.Equal(t, expected, got)
}

func TestFieldSummaryDisabled(t *testing.T) {
	ctx := testContextWithMode(TypeLinksDisabled)
	inner := &goparser.GoType{File: ctx.File, Type: "Widget", Kind: goparser.TypeKindIdent}
	pointer := &goparser.GoType{File: ctx.File, Type: "*Widget", Kind: goparser.TypeKindPointer, Inner: []*goparser.GoType{inner}}
	field := &goparser.GoField{
		Struct:   &goparser.GoStruct{},
		File:     ctx.File,
		Name:     "Child",
		Type:     "*Widget",
		Decl:     "Child *Widget",
		TypeInfo: pointer,
	}

	got := ctx.fieldSummary(field)
	assert.Equal(t, "Child\t*Widget", got)
}

func TestFunctionSignatureLinksReceiversAndParams(t *testing.T) {
	ctx := testContextWithMode(TypeLinksInternalExternal)
	ctx.File.Imports = []*goparser.GoImport{{Path: "context"}}

	container := &goparser.GoStruct{Name: "Container", File: ctx.File, TypeParams: []*goparser.GoType{{Name: "T"}}}
	widget := &goparser.GoStruct{Name: "Widget", File: ctx.File}
	ctx.Package.Structs = []*goparser.GoStruct{container, widget}

	receiverInner := &goparser.GoType{File: ctx.File, Type: "Container", Kind: goparser.TypeKindIdent}
	receiver := &goparser.GoType{File: ctx.File, Name: "c", Type: "*Container", Kind: goparser.TypeKindPointer, Inner: []*goparser.GoType{receiverInner}}

	paramInner := &goparser.GoType{File: ctx.File, Type: "context.Context", Kind: goparser.TypeKindSelector}
	param := &goparser.GoType{File: ctx.File, Name: "ctx", Type: "context.Context", Kind: goparser.TypeKindSelector, Inner: []*goparser.GoType{paramInner}}

	resultInner := &goparser.GoType{File: ctx.File, Type: "Widget", Kind: goparser.TypeKindIdent}
	result := &goparser.GoType{File: ctx.File, Type: "*Widget", Kind: goparser.TypeKindPointer, Inner: []*goparser.GoType{resultInner}}

	fn := &goparser.GoStructMethod{
		GoMethod: goparser.GoMethod{
			File:    ctx.File,
			Name:    "Get",
			Params:  []*goparser.GoType{param},
			Results: []*goparser.GoType{result},
		},
		Receivers:     []string{"*Container"},
		ReceiverTypes: []*goparser.GoType{receiver},
	}

	sig := ctx.functionSignature(fn)
	expectedReceiver := "(c *<<" + anchorID(ctx.File.FqPackage, "Container") + ",Container>>)"
	expectedResult := "*<<" + anchorID(ctx.File.FqPackage, "Widget") + ",Widget>>"
	assert.Contains(t, sig, expectedReceiver)
	assert.Contains(t, sig, "ctx link:https://pkg.go.dev/context#Context[context.Context]")
	assert.Contains(t, sig, expectedResult)
}

func TestFunctionSignatureLeavesTypeParameters(t *testing.T) {
	ctx := testContextWithMode(TypeLinksInternal)

	container := &goparser.GoStruct{Name: "Container", File: ctx.File, TypeParams: []*goparser.GoType{{Name: "T"}}}
	ctx.Package.Structs = []*goparser.GoStruct{container}

	receiverInner := &goparser.GoType{File: ctx.File, Type: "Container[T]", Kind: goparser.TypeKindIndexList, Inner: []*goparser.GoType{
		{File: ctx.File, Type: "Container", Kind: goparser.TypeKindIdent},
		{File: ctx.File, Type: "T", Kind: goparser.TypeKindIdent},
	}}
	receiver := &goparser.GoType{File: ctx.File, Type: "*Container[T]", Kind: goparser.TypeKindPointer, Inner: []*goparser.GoType{receiverInner}}

	result := &goparser.GoType{File: ctx.File, Type: "T", Kind: goparser.TypeKindIdent}

	fn := &goparser.GoStructMethod{
		GoMethod: goparser.GoMethod{
			File:    ctx.File,
			Name:    "Value",
			Results: []*goparser.GoType{result},
		},
		Receivers:     []string{"*Container[T]"},
		ReceiverTypes: []*goparser.GoType{receiver},
	}

	sig := ctx.functionSignature(fn)
	assert.Contains(t, sig, "Value() T")
}

func TestFunctionSignatureHTMLLinks(t *testing.T) {
	ctx := testContextWithMode(TypeLinksInternalExternal)
	ctx.File.Imports = []*goparser.GoImport{{Path: "context"}}

	container := &goparser.GoStruct{Name: "Container", File: ctx.File}
	widget := &goparser.GoStruct{Name: "Widget", File: ctx.File}
	ctx.Package.Structs = []*goparser.GoStruct{container, widget}

	receiver := &goparser.GoType{File: ctx.File, Type: "*Container", Kind: goparser.TypeKindPointer, Inner: []*goparser.GoType{{File: ctx.File, Type: "Container", Kind: goparser.TypeKindIdent}}}
	param := &goparser.GoType{File: ctx.File, Name: "ctx", Type: "context.Context", Kind: goparser.TypeKindSelector}
	result := &goparser.GoType{File: ctx.File, Type: "*Widget", Kind: goparser.TypeKindPointer, Inner: []*goparser.GoType{{File: ctx.File, Type: "Widget", Kind: goparser.TypeKindIdent}}}

	fn := &goparser.GoStructMethod{
		GoMethod: goparser.GoMethod{
			File:    ctx.File,
			Name:    "Get",
			Params:  []*goparser.GoType{param},
			Results: []*goparser.GoType{result},
		},
		Receivers:     []string{"*Container"},
		ReceiverTypes: []*goparser.GoType{receiver},
	}

	doc := ctx.functionSignatureDoc(fn)
	require.NotNil(t, doc)
	htmlSig := signatureHTML(doc)
	assert.Contains(t, htmlSig, "<a href=\"#example-com-mod-pkg-Container\">Container</a>")
	assert.Contains(t, htmlSig, "<a href=\"https://pkg.go.dev/context#Context\">Context</a>")
	assert.Contains(t, htmlSig, "<a href=\"#example-com-mod-pkg-Widget\">Widget</a>")
}

func TestMethodSignatureHTML(t *testing.T) {
	ctx := testContextWithMode(TypeLinksInternal)
	method := &goparser.GoMethod{
		File:    ctx.File,
		Name:    "Format",
		Params:  []*goparser.GoType{{File: ctx.File, Type: "string", Kind: goparser.TypeKindIdent}},
		Results: []*goparser.GoType{{File: ctx.File, Type: "string", Kind: goparser.TypeKindIdent}},
	}

	doc := ctx.methodSignatureDoc(method, nil)
	require.NotNil(t, doc)
	assert.Equal(t, "Format(string) string", doc.Raw)
}

func TestFuncTypeSignatureHTML(t *testing.T) {
	ctx := testContextWithMode(TypeLinksInternal)
	ctx.Package.Structs = []*goparser.GoStruct{{Name: "Service", File: ctx.File}}
	inner := &goparser.GoType{File: ctx.File, Type: "*Service", Kind: goparser.TypeKindPointer, Inner: []*goparser.GoType{{File: ctx.File, Type: "Service", Kind: goparser.TypeKindIdent}}}
	fnType := &goparser.GoMethod{
		File:    ctx.File,
		Params:  []*goparser.GoType{inner},
		Results: []*goparser.GoType{{File: ctx.File, Type: "error", Kind: goparser.TypeKindIdent}},
	}

	doc := ctx.funcTypeSignatureDoc(fnType)
	require.NotNil(t, doc)
	htmlSig := signatureHTML(doc)
	assert.Contains(t, htmlSig, "<span class=\"hljs-keyword\">func</span>(")
	assert.Contains(t, htmlSig, "*<a href=\"#example-com-mod-pkg-Service\">Service</a>")
}

func signatureHTML(doc *SignatureDoc) string {
	var b strings.Builder
	for _, seg := range doc.Segments {
		if seg.Class != "" {
			b.WriteString(`<span class="`)
			b.WriteString(seg.Class)
			b.WriteString(`">`)
			b.WriteString(string(seg.Content))
			b.WriteString(`</span>`)
			continue
		}
		b.WriteString(string(seg.Content))
	}
	return b.String()
}
