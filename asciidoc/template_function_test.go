package asciidoc

import (
	"bytes"
	"strings"
	"testing"

	"github.com/mariotoffia/goasciidoc/goparser"
	"github.com/stretchr/testify/require"
)

func TestFunctionDocExampleSpacing(t *testing.T) {
	const code = `package sample

type HandleCOMTOptions struct{}

type DataPoint struct{}

type CbContext interface{}

// HandleCOMT will make sure that proper _COMT_ (or COMT_Cn - where n is a circuit number) is set in the datapoints
// if it is not present and a _COMT_ is found in the _pipeCtx_. It will also make sure to strip out _COMT_ when
// no control signal is found for either global or the specific circuit(s).
//
// .Example Usage
// [source,go]
// ----
// var (
// 	mid      string
// 	modifiedMid bool
// )
//
// dps, mid, modifiedMid = HandleCOMT( // <1>
// 	ctx, lid, pipeCtx, dps,
// 	HandleCOMTOptions{
// 		MID: "", // <2>
// 		Mapping: nil, // <3>
// 	},
// )
//
//	if modifiedMid {
//		_ = mid // <4>
//	}
//
// ----
// <1> Handle the _COMT_ datapoints and adjust the mid string accordingly.
// <2> Get the current mid string from the provider config.
// <3> This is typically provided by the provider sdk in the target function, otherwise use it
// from the provider config mapping table.
// <3> If there are any datapoints added/removed, update the mid string in the global params.
func HandleCOMT(ctx CbContext, lid string, pipeCtx map[string]any, dp []DataPoint, opts ...HandleCOMTOptions) ([]DataPoint, string, bool) {
	return nil, "", false
}
`

	goFile, err := goparser.ParseInlineFileWithConfig(
		goparser.ParseConfig{DocConcatenation: goparser.DocConcatenationFull},
		"",
		code,
	)
	require.NoError(t, err)

	overrides := loadTemplateOverrides(t, FunctionTemplate)
	tmpl := NewTemplateWithOverrides(overrides)
	ctx := tmpl.NewContext(goFile)
	ctx.Config.SignatureStyle = "goasciidoc"

	fn := findMethodByName(goFile.StructMethods, "HandleCOMT")
	require.NotNil(t, fn)

	var buf bytes.Buffer
	ctx.RenderFunction(&buf, fn)
	doc := buf.String()

	require.Contains(t, doc, "=== HandleCOMT")
	require.Contains(t, doc, ".Example Usage")
	require.Contains(t, doc, "+++\n\n")
	require.NotContains(t, doc, "+++[source,go]")
	require.Contains(t, doc, "\n[source,go]\n")
	require.Contains(t, doc, "\n----\n")
	require.Contains(t, doc, "<1> Handle the _COMT_ datapoints")

	for _, line := range strings.Split(doc, "\n") {
		if strings.HasPrefix(line, "==== ") {
			require.NotContains(t, line, "<span")
		}
	}
}
