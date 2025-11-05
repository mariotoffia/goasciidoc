package asciidoc

import (
	"io"
	"os"
	"path/filepath"
	"text/tabwriter"

	"github.com/mariotoffia/goasciidoc/goparser"
)

// CreateTemplateWithOverrides creates a new instance of _Template_
// and add the possible _Provider.overrides_ into it.
func (p *Producer) CreateTemplateWithOverrides() *Template {
	return NewTemplateWithOverrides(p.overrides)
}

// Generate will execute the generation of the documentation
func (p *Producer) Generate() {

	t := p.CreateTemplateWithOverrides()
	w := tabwriter.NewWriter(p.createWriter(), 4, 4, 4, ' ', 0)

	overviewpaths := p.overviewpaths
	if len(overviewpaths) == 0 {
		overviewpaths = []string{
			"overview.adoc",
			"_design/overview.adoc",
		}
	}

	indexdone := !p.index

	err := goparser.ParseSinglePackageWalker(
		p.parseconfig,
		p.getProcessFunc(t, w, indexdone, overviewpaths),
		p.paths...,
	)

	if nil != err {
		panic(err)
	}

	w.Flush()
}

func (p *Producer) createWriter() io.Writer {

	if p.writer != nil {
		return p.writer
	}

	if p.outfile == "" {
		p.outfile = filepath.Join(p.parseconfig.Module.Base, "docs.adoc")
	}

	dir := filepath.Dir(p.outfile)
	if !dirExists(dir) {
		os.MkdirAll(dir, os.ModePerm)
	}

	wr, err := os.Create(p.outfile)
	if err != nil {
		panic(err)
	}

	return wr

}

// getProcessFunc will return the function to feed the template system to generate the documentation output.
//
// If `p.macro` is set to `true`, it will wrap the document template / writer function with a macro substitution
// function using function chaining.
func (p *Producer) getProcessFunc(
	t *Template,
	w io.Writer,
	indexdone bool,
	overviewpaths []string) goparser.ParseSinglePackageWalkerFunc {

	processor := func(pkg *goparser.GoPackage) error {

		tc := t.NewContextWithConfig(&pkg.GoFile, pkg, &TemplateContextConfig{
			IncludeMethodCode:    false,
			PackageOverviewPaths: overviewpaths,
			Private:              p.private,
			TypeLinks:            p.typeLinks,
			SignatureStyle:       p.signatureStyle,
		})

		if !indexdone {

			ic := tc.DefaultIndexConfig(p.indexconfig)
			if p.highlighter != "" {
				ic.Highlighter = p.highlighter
			}
			if !p.toc {
				ic.TocTitle = "" // disables toc generation
			}

			tc.RenderIndex(w, ic)
			indexdone = true
		}

		tc.RenderPackage(w)

		if len(pkg.Imports) > 0 {
			tc.RenderImports(w)
		}
		if len(pkg.Interfaces) > 0 {
			tc.RenderInterfaces(w)
		}
		if len(pkg.Structs) > 0 {
			tc.RenderStructs(w)
		}
		if len(pkg.CustomTypes) > 0 {
			tc.RenderVarTypeDefs(w)
		}
		if len(pkg.ConstAssignments) > 0 {
			tc.RenderConstDeclarations(w)
		}
		if len(pkg.CustomFuncs) > 0 {
			tc.RenderTypeDefFuncs(w)
		}
		if len(pkg.VarAssignments) > 0 {
			tc.RenderVarDeclarations(w)
		}
		if len(pkg.StructMethods) > 0 {
			tc.RenderFunctions(w)
		}

		return nil
	}

	if p.macro {

		return getProcessMacroFunc(processor)

	}

	return processor
}
